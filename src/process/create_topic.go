package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
)

type createTopicProcess struct {
	database  Database
	confluent Confluent
	vault     Vault
}

func NewCreateTopicProcess(database Database, confluent Confluent, vault Vault) CreateTopicProcess {
	return &createTopicProcess{database, confluent, vault}
}

type CreateTopicProcessInput struct {
	CapabilityRootId models.CapabilityRootId
	ClusterId        models.ClusterId
	Topic            models.Topic
}

func (ctp *createTopicProcess) Process(ctx context.Context, input CreateTopicProcessInput) error {
	process, err := ctp.prepareProcess(ctx, input)
	if err != nil {
		return err
	}

	if process.State.IsCompleted() {
		// already completed => skip
		return nil
	}

	return PrepareSteps().
		Step(ensureServiceAccount).
		Step(ensureServiceAccountAcl).Until(func() bool { return process.State.HasClusterAccess }).
		Step(ensureServiceAccountApiKey).
		Step(ensureServiceAccountApiKeyAreStoredInVault).
		Step(ensureTopicIsCreated).
		Run(process)
}

func (ctp *createTopicProcess) prepareProcess(ctx context.Context, input CreateTopicProcessInput) (*Process, error) {
	session := ctp.database.NewSession(ctx)

	state, err := getOrCreateProcessState(session, input.CapabilityRootId, input.ClusterId, input.Topic)
	if err != nil {
		return nil, err
	}

	return NewProcess(ctx, session, ctp.confluent, ctp.vault, state), nil
}

func getOrCreateProcessState(repository stateRepository, capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, topic models.Topic) (*models.ProcessState, error) {
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

// region Steps

func ensureServiceAccount(process *Process) error {
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId
	service := process.service()

	fmt.Println("### EnsureServiceAccount")

	if process.State.HasServiceAccount {
		return nil
	}

	err := service.CreateServiceAccount(capabilityRootId, clusterId)
	if err != nil {
		return err
	}

	process.State.HasServiceAccount = true

	return err
}

func ensureServiceAccountAcl(process *Process) error {
	service := process.service()
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
	service := process.service()
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
	service := process.service()
	aws := process.Vault
	capabilityRootId := process.State.CapabilityRootId
	clusterId := process.State.ClusterId

	fmt.Println("### EnsureServiceAccountApiKeyAreStoredInVault")
	if process.State.HasApiKeyInVault {
		return nil
	}

	clusterAccess, err := service.GetClusterAccess(capabilityRootId, clusterId)
	if err != nil {
		return err
	}

	if err := aws.StoreApiKey(context.TODO(), capabilityRootId, clusterAccess.ClusterId, clusterAccess.ApiKey); err != nil {
		return err
	}

	process.State.HasApiKeyInVault = true

	return nil
}

func ensureTopicIsCreated(process *Process) error {
	fmt.Println("### EnsureTopicIsCreated")
	if process.State.IsCompleted() {
		return nil
	}

	err := process.createTopic()
	if err != nil {
		return err
	}

	process.State.MarkAsCompleted()

	return nil
}

// endregion
