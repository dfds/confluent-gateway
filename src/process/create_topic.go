package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
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

	state, err := getOrCreateProcessState(database, input)
	if err != nil {
		return err
	}

	if state.IsCompleted() {
		// already completed => skip
		return nil
	}

	return PrepareSteps().
		Step(ensureServiceAccount).
		Step(ensureServiceAccountAcl).Until(func(p *Process) bool { return p.hasClusterAccess() }).
		Step(ensureServiceAccountApiKey).
		Step(ensureServiceAccountApiKeyAreStoredInVault).
		Step(ensureTopicIsCreated).
		Run(func(step Step) error {
			return database.Transaction(func(tx Transaction) error {

				process := ctp.NewProcess(ctx, tx, state)

				err := step(process)
				if err != nil {
					return err
				}

				return tx.UpdateProcessState(state)
			})
		})
}

func getOrCreateProcessState(repository stateRepository, input CreateTopicProcessInput) (*models.ProcessState, error) {
	capabilityRootId, clusterId, topic := input.CapabilityRootId, input.ClusterId, input.Topic

	state, err := repository.GetProcessState(capabilityRootId, clusterId, topic.Name)
	if err != nil {
		return nil, err
	}

	if state == nil {
		//serviceAccount, err := repository.GetServiceAccount(capabilityRootId)
		//if err != nil {
		//	return nil, err
		//}
		//
		//hasServiceAccount := false
		//hasClusterAccess := false
		//
		//if serviceAccount != nil {
		//	hasServiceAccount = true
		//	_, hasClusterAccess = serviceAccount.TryGetClusterAccess(clusterId)
		//}
		//
		//state = models.NewProcessState(capabilityRootId, clusterId, topic, hasServiceAccount, hasClusterAccess)

		// TODO -- stop faking

		state = models.NewProcessState(capabilityRootId, clusterId, topic, true, true)
		state.HasApiKey = true
		state.HasApiKeyInVault = true
		state.MarkAsCompleted()

		if err := repository.CreateProcessState(state); err != nil {
			return nil, err
		}
	}
	return state, nil
}

func (ctp *createTopicProcess) NewProcess(ctx context.Context, tx Transaction, state *models.ProcessState) *Process {
	newAccountService := NewAccountService(ctx, ctp.confluent, tx)
	vault := NewVaultService(ctx, ctp.vault)
	topic := NewTopicService(ctx, ctp.confluent)
	outbox := messaging.NewOutbox(ctp.logger, ctp.registry, tx)

	return NewProcess(state, newAccountService, vault, topic, outbox)
}

// region Steps

func ensureServiceAccount(process *Process) error {
	fmt.Println("### EnsureServiceAccount")
	return ensureServiceAccountStep(process)
}

type EnsureServiceAccountStep interface {
	hasServiceAccount() bool
	createServiceAccount() error
	markServiceAccountReady()
}

func ensureServiceAccountStep(process EnsureServiceAccountStep) error {
	if process.hasServiceAccount() {
		return nil
	}

	err := process.createServiceAccount()
	if err != nil {
		return err
	}

	process.markServiceAccountReady()

	return nil
}

func ensureServiceAccountAcl(process *Process) error {
	fmt.Println("### EnsureServiceAccountAcl")
	return ensureServiceAccountAclStep(process)
}

type EnsureServiceAccountAclStep interface {
	hasClusterAccess() bool
	getOrCreateClusterAccess() (*models.ClusterAccess, error)
	createAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error
	markClusterAccessReady()
}

func ensureServiceAccountAclStep(process EnsureServiceAccountAclStep) error {
	if process.hasClusterAccess() {
		return nil
	}

	clusterAccess, err := process.getOrCreateClusterAccess()
	if err != nil {
		return err
	}

	entries := clusterAccess.GetAclPendingCreation()
	if len(entries) == 0 {
		// no acl entries left => mark as done
		process.markClusterAccessReady()
		return nil

	} else {
		nextEntry := entries[0]

		return process.createAclEntry(clusterAccess, nextEntry)
	}
}

func ensureServiceAccountApiKey(process *Process) error {
	fmt.Println("### EnsureServiceAccountApiKey")
	return ensureServiceAccountApiKeyStep(process)
}

type EnsureServiceAccountApiKeyStep interface {
	hasApiKey() bool
	getClusterAccess() (*models.ClusterAccess, error)
	createApiKey(clusterAccess *models.ClusterAccess) error
	markApiKeyReady()
}

func ensureServiceAccountApiKeyStep(process EnsureServiceAccountApiKeyStep) error {
	if process.hasApiKey() {
		return nil
	}

	clusterAccess, err := process.getClusterAccess()
	if err != nil {
		return err
	}

	err = process.createApiKey(clusterAccess)
	if err != nil {
		return err
	}

	process.markApiKeyReady()
	return nil
}

func ensureServiceAccountApiKeyAreStoredInVault(process *Process) error {
	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	return ensureServiceAccountApiKeyAreStoredInVaultStep(process)
}

type EnsureServiceAccountApiKeyAreStoredInVaultStep interface {
	hasApiKeyInVault() bool
	getClusterAccess() (*models.ClusterAccess, error)
	storeApiKey(clusterAccess *models.ClusterAccess) error
	markApiKeyInVaultReady()
}

func ensureServiceAccountApiKeyAreStoredInVaultStep(process EnsureServiceAccountApiKeyAreStoredInVaultStep) error {
	if process.hasApiKeyInVault() {
		return nil
	}

	clusterAccess, err := process.getClusterAccess()
	if err != nil {
		return err
	}

	if err = process.storeApiKey(clusterAccess); err != nil {
		return err
	}

	process.markApiKeyInVaultReady()

	return nil
}

func ensureTopicIsCreated(process *Process) error {
	fmt.Println("### EnsureTopicIsCreated")
	return ensureTopicIsCreatedStep(process)

}

type EnsureTopicIsCreatedStep interface {
	isCompleted() bool
	createTopic() error
	markAsCompleted()
	topicProvisioned() error
}

func ensureTopicIsCreatedStep(process EnsureTopicIsCreatedStep) error {
	if process.isCompleted() {
		return nil
	}

	err := process.createTopic()
	if err != nil {
		return err
	}

	process.markAsCompleted()

	return process.topicProvisioned()
}

// endregion
