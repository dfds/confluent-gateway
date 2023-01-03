package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
	"strings"
)

type createTopicProcess struct {
	logger    logging.Logger
	database  Database
	confluent Confluent
	vault     Vault
	registry  messaging.OutgoingMessageRegistry
}

func NewCreateTopicProcess(logger logging.Logger, database Database, confluent Confluent, vault Vault, registry messaging.OutgoingMessageRegistry) CreateTopicProcess {
	return &createTopicProcess{
		logger:    logger,
		database:  database,
		confluent: confluent,
		vault:     vault,
		registry:  registry,
	}
}

type CreateTopicProcessInput struct {
	CapabilityRootId models.CapabilityRootId
	ClusterId        models.ClusterId
	Topic            models.Topic
}

func (ctp *createTopicProcess) Process(ctx context.Context, input CreateTopicProcessInput) error {
	database := ctp.database.WithContext(ctx)

	state, err := ctp.getOrCreateProcessState(database, input)
	if err != nil {
		return err
	}

	if state.IsCompleted() {
		// already completed => skip
		return nil
	}

	return PrepareSteps().
		Step(ensureServiceAccount).
		Step(ensureServiceAccountAcl).Until(func(p *StepContext) bool { return p.HasClusterAccess() }).
		Step(ensureServiceAccountApiKey).
		Step(ensureServiceAccountApiKeyAreStoredInVault).
		Step(ensureTopicIsCreated).
		Run(func(step Step) error {
			return database.Transaction(func(tx Transaction) error {
				stepContext := ctp.getStepContext(ctx, tx, state)

				err := step(stepContext)
				if err != nil {
					return err
				}

				return tx.UpdateProcessState(state)
			})
		})
}

func (ctp *createTopicProcess) getOrCreateProcessState(database Database, input CreateTopicProcessInput) (*models.ProcessState, error) {
	var s *models.ProcessState

	err := database.Transaction(func(tx Transaction) error {
		outbox := ctp.getOutbox(tx)

		state, err := getOrCreateProcessState(tx, outbox, input)
		if err != nil {
			return err
		}

		s = state

		return nil
	})

	return s, err
}

func getOrCreateProcessState(repo stateRepository, outbox Outbox, input CreateTopicProcessInput) (*models.ProcessState, error) {
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

func (ctp *createTopicProcess) getStepContext(ctx context.Context, tx Transaction, state *models.ProcessState) *StepContext {
	newAccountService := NewAccountService(ctx, ctp.confluent, tx)
	vault := NewVaultService(ctx, ctp.vault)
	topic := NewTopicService(ctx, ctp.confluent)
	outbox := ctp.getOutbox(tx)

	return NewStepContext(state, newAccountService, vault, topic, outbox)
}

func (ctp *createTopicProcess) getOutbox(tx Transaction) *messaging.Outbox {
	return messaging.NewOutbox(ctp.logger, ctp.registry, tx)
}

// region Steps

func ensureServiceAccount(stepContext *StepContext) error {
	fmt.Println("### EnsureServiceAccount")
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
	fmt.Println("### EnsureServiceAccountAcl")
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
	fmt.Println("### EnsureServiceAccountApiKey")
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
	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
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
	fmt.Println("### EnsureTopicIsCreated")
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
