package create

import (
	"context"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
	. "github.com/dfds/confluent-gateway/process"
	"github.com/satori/go.uuid"
	"strings"
)

type process struct {
	logger    logging.Logger
	database  models.Database
	confluent Confluent
	vault     Vault
	registry  messaging.OutgoingMessageRegistry
}

func NewProcess(logger logging.Logger, database models.Database, confluent Confluent, vault Vault, registry messaging.OutgoingMessageRegistry) Process {
	return &process{
		logger:    logger,
		database:  database,
		confluent: confluent,
		vault:     vault,
		registry:  registry,
	}
}

type ProcessInput struct {
	CapabilityRootId models.CapabilityRootId
	ClusterId        models.ClusterId
	Topic            models.TopicDescription
}

func (p *process) Process(ctx context.Context, input ProcessInput) error {
	session := p.database.NewSession(ctx)

	state, err := p.prepareProcessState(session, input)
	if err != nil {
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

				return tx.UpdateProcessState(state)
			})
		})
}

func (p *process) prepareProcessState(session models.Session, input ProcessInput) (*models.ProcessState, error) {
	var s *models.ProcessState

	err := session.Transaction(func(tx models.Transaction) error {
		outbox := p.getOutbox(tx)

		state, err := getOrCreateProcessState(tx, outbox, input)
		if err != nil {
			return err
		}

		s = state

		return nil
	})

	return s, err
}

type stateRepository interface {
	GetProcessState(capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topicName string) (*models.ProcessState, error)
	GetServiceAccount(capabilityRootId models.CapabilityRootId) (*models.ServiceAccount, error)
	CreateProcessState(state *models.ProcessState) error
}

func getOrCreateProcessState(repo stateRepository, outbox Outbox, input ProcessInput) (*models.ProcessState, error) {
	capabilityRootId, clusterId, topic := input.CapabilityRootId, input.ClusterId, input.Topic

	state, err := repo.GetProcessState(capabilityRootId, clusterId, topic.Name)
	if err != nil {
		return nil, err
	}

	if state != nil {
		return state, nil
	}

	if strings.HasSuffix(topic.Name, "-cg") {
		serviceAccount, err := repo.GetServiceAccount(capabilityRootId)
		if err != nil {
			return nil, err
		}

		HasServiceAccount := false
		HasClusterAccess := false

		if serviceAccount != nil {
			HasServiceAccount = true
			_, HasClusterAccess = serviceAccount.TryGetClusterAccess(clusterId)
		}

		state = models.NewProcessState(capabilityRootId, clusterId, topic, HasServiceAccount, HasClusterAccess)

		if err := repo.CreateProcessState(state); err != nil {
			return nil, err
		}

		err = outbox.Produce(&TopicProvisioningBegun{
			partitionKey:     state.Id.String(),
			CapabilityRootId: string(capabilityRootId),
			ClusterId:        string(clusterId),
			TopicName:        topic.Name,
		})

		if err != nil {
			return nil, err
		}
	} else {
		// TODO -- stop faking
		state = models.NewProcessState(capabilityRootId, clusterId, topic, true, true)
		state.HasApiKey = true
		state.HasApiKeyInVault = true
		state.MarkAsCompleted()

		if err := repo.CreateProcessState(state); err != nil {
			return nil, err
		}
	}

	return state, nil
}

func (p *process) getStepContext(ctx context.Context, tx models.Transaction, state *models.ProcessState) *StepContext {
	logger := p.logger
	newAccountService := NewAccountService(ctx, p.confluent, tx)
	vault := NewVaultService(ctx, p.vault)
	topic := NewTopicService(ctx, p.confluent, tx)
	outbox := p.getOutbox(tx)

	return NewStepContext(logger, state, newAccountService, vault, topic, outbox)
}

func (p *process) getOutbox(tx models.Transaction) *messaging.Outbox {
	return messaging.NewOutbox(p.logger, p.registry, tx, func() string { return uuid.NewV4().String() })
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
