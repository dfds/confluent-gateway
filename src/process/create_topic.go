package process

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
	"github.com/satori/go.uuid"
	"time"
)

type CreateTopicProcess struct {
	database  Database
	confluent Confluent
	vault     Vault
}

func NewCreateTopicProcess(database Database, confluent Confluent, vault Vault) *CreateTopicProcess {
	return &CreateTopicProcess{database, confluent, vault}
}

type CreateTopicProcessInput struct {
	CapabilityRootId models.CapabilityRootId
	ClusterId        models.ClusterId
	Topic            models.Topic
}

func (ctp *CreateTopicProcess) Process(ctx context.Context, input CreateTopicProcessInput) error {
	p, err := ctp.prepareProcess(ctx, input)
	if err != nil {
		return err
	}

	if p.State.IsCompleted() {
		// already completed => skip
		return nil
	}

	return PrepareSteps().
		Step(ensureServiceAccount).
		Step(ensureServiceAccountAcl).Until(func() bool { return p.State.HasClusterAccess }).
		Step(ensureServiceAccountApiKey).
		Step(ensureServiceAccountApiKeyAreStoredInVault).
		Step(ensureTopicIsCreated).
		Run(p)
}

func (ctp *CreateTopicProcess) prepareProcess(ctx context.Context, input CreateTopicProcessInput) (*Process, error) {
	session := ctp.database.NewSession(ctx)

	state, err := session.Processes().GetProcessState(input.CapabilityRootId, input.ClusterId, input.Topic.Name)
	if err != nil {
		return nil, err
	}

	if state == nil {
		serviceAccount, err := session.ServiceAccounts().GetServiceAccount(input.CapabilityRootId)
		if err != nil {
			return nil, err
		}

		hasServiceAccount := false
		hasClusterAccess := false

		if serviceAccount != nil {
			hasServiceAccount = true
			_, hasClusterAccess = serviceAccount.TryGetClusterAccess(input.ClusterId)
		}

		state = &models.ProcessState{
			Id:                uuid.NewV4(),
			CapabilityRootId:  input.CapabilityRootId,
			ClusterId:         input.ClusterId,
			Topic:             input.Topic,
			HasServiceAccount: hasServiceAccount,
			HasClusterAccess:  hasClusterAccess,
			HasApiKey:         hasClusterAccess,
			HasApiKeyInVault:  hasClusterAccess,
			CreatedAt:         time.Now(),
			CompletedAt:       nil,
		}

		if err := session.Processes().CreateProcessState(state); err != nil {
			return nil, err
		}
	}

	return NewProcess(ctx, session, ctp.confluent, ctp.vault, state), nil
}
