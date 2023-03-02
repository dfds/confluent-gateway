package create

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/internal/models"
	. "github.com/dfds/confluent-gateway/internal/process"
	"github.com/dfds/confluent-gateway/logging"
)

type process struct {
	logger    logging.Logger
	database  models.Database
	confluent Confluent
	vault     Vault
	factory   OutboxFactory
}

func NewProcess(logger logging.Logger, database models.Database, confluent Confluent, vault Vault, factory OutboxFactory) Process {
	return &process{
		logger:    logger,
		database:  database,
		confluent: confluent,
		vault:     vault,
		factory:   factory,
	}
}

type ProcessInput struct {
	CapabilityId models.CapabilityId
	ClusterId    models.ClusterId
	Topic        models.TopicDescription
}

func (p *process) Process(ctx context.Context, input ProcessInput) error {
	session := p.database.NewSession(ctx)

	state, err := p.prepareProcessState(session, input)
	if err != nil {
		if errors.Is(err, ErrTopicAlreadyExists) {
			// topic already exists => skip
			return nil
		}

		return err
	}

	if state.IsCompleted() {
		// already completed => skip
		return nil
	}

	return PrepareSteps[*StepContext]().
		Step(ensureServiceAccount).
		Step(ensureServiceAccountAcl).Until(func(c *StepContext) bool { return c.HasClusterAccess() }).
		Step(ensureServiceAccountApiKey).
		Step(ensureServiceAccountApiKeyAreStoredInVault).
		Step(ensureTopicIsCreated).
		Run(func(step func(*StepContext) error) error {
			return session.Transaction(func(tx models.Transaction) error {
				stepContext := p.getStepContext(ctx, tx, state)

				err := step(stepContext)
				if err != nil {
					return err
				}

				return tx.UpdateCreateProcessState(state)
			})
		})
}

func (p *process) prepareProcessState(session models.Session, input ProcessInput) (*models.CreateProcess, error) {
	var s *models.CreateProcess

	err := session.Transaction(func(tx models.Transaction) error {
		outbox := p.factory(tx)

		if err := ensureNewTopic(tx, input); err != nil {
			p.logger.Warning("{Topic} on {Cluster} for {Capability} already exists", input.Topic.Name, string(input.CapabilityId), string(input.ClusterId))
			return err
		}

		state, err := getOrCreateProcessState(tx, outbox, input)
		if err != nil {
			return err
		}

		s = state

		return nil
	})

	return s, err
}

var ErrTopicAlreadyExists = errors.New("topic already exists")

func ensureNewTopic(tx models.Transaction, input ProcessInput) error {
	capabilityId, clusterId, topicName := input.CapabilityId, input.ClusterId, input.Topic.Name

	topic, err := tx.GetTopic(capabilityId, clusterId, topicName)
	if err != nil {
		return err
	}

	if topic != nil {
		return ErrTopicAlreadyExists
	}

	return nil
}

type stateRepository interface {
	GetCreateProcessState(capabilityId models.CapabilityId, clusterId models.ClusterId, topicName string) (*models.CreateProcess, error)
	GetServiceAccount(capabilityId models.CapabilityId) (*models.ServiceAccount, error)
	SaveCreateProcessState(state *models.CreateProcess) error
}

func getOrCreateProcessState(repo stateRepository, outbox Outbox, input ProcessInput) (*models.CreateProcess, error) {
	capabilityId, clusterId, topic := input.CapabilityId, input.ClusterId, input.Topic

	state, err := repo.GetCreateProcessState(capabilityId, clusterId, topic.Name)
	if err != nil {
		return nil, err
	}

	if state != nil && !state.IsCompleted() {
		// is process is unfinished => continue
		return state, nil
	}

	serviceAccount, err := repo.GetServiceAccount(capabilityId)
	if err != nil {
		return nil, err
	}

	HasServiceAccount := false
	HasClusterAccess := false

	if serviceAccount != nil {
		HasServiceAccount = true
		_, HasClusterAccess = serviceAccount.TryGetClusterAccess(clusterId)
	}

	state = models.NewCreateProcess(capabilityId, clusterId, topic, HasServiceAccount, HasClusterAccess)

	if err := repo.SaveCreateProcessState(state); err != nil {
		return nil, err
	}

	err = outbox.Produce(&TopicProvisioningBegun{
		partitionKey: state.Id.String(),
		CapabilityId: string(capabilityId),
		ClusterId:    string(clusterId),
		TopicName:    topic.Name,
	})

	if err != nil {
		return nil, err
	}

	return state, nil
}

func (p *process) getStepContext(ctx context.Context, tx models.Transaction, state *models.CreateProcess) *StepContext {
	logger := p.logger
	newAccountService := NewAccountService(ctx, p.confluent, tx)
	vault := NewVaultService(ctx, p.vault)
	topic := NewTopicService(ctx, p.confluent, tx)
	outbox := p.factory(tx)

	return NewStepContext(logger, state, newAccountService, vault, topic, outbox)
}

// region Steps

func ensureServiceAccount(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureServiceAccount")
	return ensureServiceAccountStep(stepContext)
}

type EnsureServiceAccountStep interface {
	HasServiceAccount() bool
	CreateServiceAccount() error
	MarkServiceAccountAsReady()
}

func ensureServiceAccountStep(step EnsureServiceAccountStep) error {
	if step.HasServiceAccount() {
		return nil
	}

	err := step.CreateServiceAccount()
	if err != nil {
		return err
	}

	step.MarkServiceAccountAsReady()

	return nil
}

func ensureServiceAccountAcl(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureServiceAccountAcl")
	return ensureServiceAccountAclStep(stepContext)
}

type EnsureServiceAccountAclStep interface {
	HasClusterAccess() bool
	GetOrCreateClusterAccess() (*models.ClusterAccess, error)
	CreateAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error
	MarkClusterAccessAsReady()
}

func ensureServiceAccountAclStep(step EnsureServiceAccountAclStep) error {
	if step.HasClusterAccess() {
		return nil
	}

	clusterAccess, err := step.GetOrCreateClusterAccess()
	if err != nil {
		return err
	}

	entries := clusterAccess.GetAclPendingCreation()
	if len(entries) == 0 {
		// no acl entries left => mark as done
		step.MarkClusterAccessAsReady()
		return nil

	} else {
		nextEntry := entries[0]

		return step.CreateAclEntry(clusterAccess, nextEntry)
	}
}

func ensureServiceAccountApiKey(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureServiceAccountApiKey")
	return ensureServiceAccountApiKeyStep(stepContext)
}

type EnsureServiceAccountApiKeyStep interface {
	HasApiKey() bool
	GetClusterAccess() (*models.ClusterAccess, error)
	CreateApiKey(clusterAccess *models.ClusterAccess) error
	MarkApiKeyAsReady()
}

func ensureServiceAccountApiKeyStep(step EnsureServiceAccountApiKeyStep) error {
	if step.HasApiKey() {
		return nil
	}

	clusterAccess, err := step.GetClusterAccess()
	if err != nil {
		return err
	}

	err = step.CreateApiKey(clusterAccess)
	if err != nil {
		return err
	}

	step.MarkApiKeyAsReady()
	return nil
}

func ensureServiceAccountApiKeyAreStoredInVault(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureServiceAccountApiKeyAreStoredInVault")
	return ensureServiceAccountApiKeyAreStoredInVaultStep(stepContext)
}

type EnsureServiceAccountApiKeyAreStoredInVaultStep interface {
	HasApiKeyInVault() bool
	GetClusterAccess() (*models.ClusterAccess, error)
	StoreApiKey(clusterAccess *models.ClusterAccess) error
	MarkApiKeyInVaultAsReady()
}

func ensureServiceAccountApiKeyAreStoredInVaultStep(step EnsureServiceAccountApiKeyAreStoredInVaultStep) error {
	if step.HasApiKeyInVault() {
		return nil
	}

	clusterAccess, err := step.GetClusterAccess()
	if err != nil {
		return err
	}

	if err = step.StoreApiKey(clusterAccess); err != nil {
		return err
	}

	step.MarkApiKeyInVaultAsReady()

	return nil
}

func ensureTopicIsCreated(stepContext *StepContext) error {
	stepContext.logger.Trace("Running {Step}", "EnsureTopicIsCreated")
	return ensureTopicIsCreatedStep(stepContext)
}

type EnsureTopicIsCreatedStep interface {
	IsCompleted() bool
	CreateTopic() error
	MarkAsCompleted()
	RaiseTopicProvisionedEvent() error
}

func ensureTopicIsCreatedStep(step EnsureTopicIsCreatedStep) error {
	if step.IsCompleted() {
		return nil
	}

	err := step.CreateTopic()
	if err != nil {
		return err
	}

	step.MarkAsCompleted()

	return step.RaiseTopicProvisionedEvent()
}

// endregion
