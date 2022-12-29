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
		Step(ensureServiceAccountAcl).Until(func(p *Process) bool { return p.HasClusterAccess() }).
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

type Process struct {
	State   *models.ProcessState
	Account AccountService
	Vault   VaultService
	Topic   TopicService
}

func (ctp *createTopicProcess) NewProcess(ctx context.Context, tx Transaction, state *models.ProcessState) *Process {
	return &Process{
		State:   state,
		Account: NewAccountService(ctx, ctp.confluent, tx),
		Vault:   NewVaultService(ctx, ctp.vault),
		Topic:   NewTopicService(ctx, ctp.confluent),
	}
}

func (p *Process) HasClusterAccess() bool {
	return p.State.HasClusterAccess
}

// region Steps

func ensureServiceAccount(process *Process) error {
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId
	fmt.Println("### EnsureServiceAccount")

	if process.State.HasServiceAccount {
		return nil
	}

	err := process.Account.CreateServiceAccount(capabilityRootId, clusterId)
	if err != nil {
		return err
	}

	process.State.HasServiceAccount = true

	return err
}

func ensureServiceAccountAcl(process *Process) error {
	service := process.Account
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId

	fmt.Println("### EnsureServiceAccountAcl")
	if process.State.HasClusterAccess {
		return nil
	}

	clusterAccess, err := service.GetOrCreateClusterAccess(capabilityRootId, clusterId)
	if err != nil {
		return err
	}

	entries := clusterAccess.GetAclPendingCreation()
	if len(entries) == 0 {
		// no acl entries left => mark as done
		process.State.HasClusterAccess = true
		return nil

	} else {
		nextEntry := entries[0]

		return service.CreateAclEntry(clusterId, clusterAccess.ServiceAccountId, &nextEntry)
	}
}

func ensureServiceAccountApiKey(process *Process) error {
	service := process.Account
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId

	fmt.Println("### EnsureServiceAccountApiKey")
	if process.State.HasApiKey {
		return nil
	}

	clusterAccess, err := service.GetClusterAccess(capabilityRootId, clusterId)
	if err != nil {
		return err
	}

	err2 := service.CreateApiKey(clusterAccess)
	if err2 != nil {
		return err2
	}

	process.State.HasApiKey = true
	return nil
}

func ensureServiceAccountApiKeyAreStoredInVault(process *Process) error {
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId

	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	if process.State.HasApiKeyInVault {
		return nil
	}

	clusterAccess, err := process.Account.GetClusterAccess(capabilityRootId, clusterId)
	if err != nil {
		return err
	}

	if err := process.Vault.StoreApiKey(capabilityRootId, clusterAccess); err != nil {
		return err
	}

	process.State.HasApiKeyInVault = true

	return nil
}

func ensureTopicIsCreated(process *Process) error {
	clusterId := process.State.ClusterId
	topic := process.State.Topic()

	fmt.Println("### EnsureTopicIsCreated")
	if process.State.IsCompleted() {
		return nil
	}

	err := process.Topic.CreateTopic(clusterId, topic)
	if err != nil {
		return err
	}

	process.State.MarkAsCompleted()

	return nil
}

// endregion
