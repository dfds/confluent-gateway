package process

import (
	"context"
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
