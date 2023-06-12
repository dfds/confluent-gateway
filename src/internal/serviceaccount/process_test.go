package serviceaccount

/*
import (
	"errors"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const someTopicName = "some-topic-name"

var serviceError = errors.New("service error")

func Test_createProcessState(t *testing.T) {
	tests := []struct {
		name                  string
		mock                  *mock
		input                 ProcessInput
		wantHasServiceAccount bool
		wantHasClusterAccess  bool
		wantHasApiKey         bool
		wantHasApiKeyInVault  bool
		wantIsCompleted       bool
		wantErr               assert.ErrorAssertionFunc
		wantEvent             *TopicProvisioningBegun
	}{
		{
			name: "ok",
			mock: &mock{},
			input: ProcessInput{
				CapabilityId: someCapabilityId,
				ClusterId:    someClusterId,
				Topic:        models.TopicDescription{Name: someTopicName},
			},
			wantHasServiceAccount: false,
			wantHasClusterAccess:  false,
			wantHasApiKey:         false,
			wantHasApiKeyInVault:  false,
			wantIsCompleted:       false,
			wantErr:               assert.NoError,
			wantEvent: &TopicProvisioningBegun{
				CapabilityId: string(someCapabilityId),
				ClusterId:    string(someClusterId),
				TopicName:    someTopicName,
			},
		},
		{
			name: "ok got service account",
			mock: &mock{ReturnServiceAccount: &models.ServiceAccount{}},
			input: ProcessInput{
				CapabilityId: someCapabilityId,
				ClusterId:    someClusterId,
				Topic:        models.TopicDescription{Name: someTopicName},
			},
			wantHasServiceAccount: true,
			wantHasClusterAccess:  false,
			wantHasApiKey:         false,
			wantHasApiKeyInVault:  false,
			wantIsCompleted:       false,
			wantErr:               assert.NoError,
			wantEvent: &TopicProvisioningBegun{
				CapabilityId: string(someCapabilityId),
				ClusterId:    string(someClusterId),
				TopicName:    someTopicName,
			},
		},
		{
			name: "ok got cluster access",
			mock: &mock{ReturnServiceAccount: &models.ServiceAccount{ClusterAccesses: []models.ClusterAccess{{ClusterId: someClusterId}}}},
			input: ProcessInput{
				CapabilityId: someCapabilityId,
				ClusterId:    someClusterId,
				Topic:        models.TopicDescription{Name: someTopicName},
			},
			wantHasServiceAccount: true,
			wantHasClusterAccess:  true,
			wantHasApiKey:         true,
			wantHasApiKeyInVault:  true,
			wantIsCompleted:       false,
			wantErr:               assert.NoError,
			wantEvent: &TopicProvisioningBegun{
				CapabilityId: string(someCapabilityId),
				ClusterId:    string(someClusterId),
				TopicName:    someTopicName,
			},
		},
		{
			name: "process already exists",
			mock: &mock{ReturnProcessState: &models.CreateProcess{
				CapabilityId:      someCapabilityId,
				ClusterId:         someClusterId,
				TopicName:         someTopicName,
				HasServiceAccount: true,
				HasClusterAccess:  true,
				HasApiKey:         true,
				HasApiKeyInVault:  true,
				CompletedAt:       nil,
			}},
			input: ProcessInput{
				CapabilityId: someCapabilityId,
				ClusterId:    someClusterId,
				Topic:        models.TopicDescription{Name: someTopicName},
			},
			wantHasServiceAccount: true,
			wantHasClusterAccess:  true,
			wantHasApiKey:         true,
			wantHasApiKeyInVault:  true,
			wantIsCompleted:       false,
			wantErr:               assert.NoError,
			wantEvent:             nil,
		},
		{
			name: "process already finished",
			mock: &mock{ReturnProcessState: &models.CreateProcess{
				CompletedAt: &time.Time{},
			}},
			input: ProcessInput{
				CapabilityId: someCapabilityId,
				ClusterId:    someClusterId,
				Topic:        models.TopicDescription{Name: someTopicName},
			},
			wantHasServiceAccount: false,
			wantHasClusterAccess:  false,
			wantHasApiKey:         false,
			wantHasApiKeyInVault:  false,
			wantIsCompleted:       false,
			wantErr:               assert.NoError,
			wantEvent: &TopicProvisioningBegun{
				CapabilityId: string(someCapabilityId),
				ClusterId:    string(someClusterId),
				TopicName:    someTopicName,
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getOrCreateProcessState(tt.mock, tt.mock, tt.input)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, someCapabilityId, got.CapabilityId)
			assert.Equal(t, someClusterId, got.ClusterId)
			assert.Equal(t, someTopicName, got.TopicName)
			assert.Equal(t, 0, got.TopicPartitions)
			assert.Equal(t, int64(0), got.TopicRetention)
			assert.Equal(t, tt.wantHasServiceAccount, got.HasServiceAccount)
			assert.Equal(t, tt.wantHasClusterAccess, got.HasClusterAccess)
			assert.Equal(t, tt.wantHasApiKey, got.HasApiKey)
			assert.Equal(t, tt.wantHasApiKeyInVault, got.HasApiKeyInVault)
			assert.Equal(t, tt.wantIsCompleted, got.IsCompleted())
			assert.Equal(t, tt.wantEvent, tt.mock.EventProduced)
		})
	}
}

type mock struct {
	ReturnProcessState   *models.CreateProcess
	ReturnServiceAccount *models.ServiceAccount
	EventProduced        *TopicProvisioningBegun
}

func (m *mock) GetCreateProcessState(models.CapabilityId, models.ClusterId, string) (*models.CreateProcess, error) {
	return m.ReturnProcessState, nil
}

func (m *mock) Produce(msg messaging.OutgoingMessage) error {
	m.EventProduced, _ = msg.(*TopicProvisioningBegun)

	return nil
}

func (m *mock) GetServiceAccount(models.CapabilityId) (*models.ServiceAccount, error) {
	return m.ReturnServiceAccount, nil
}

func (m *mock) SaveCreateProcessState(*models.CreateProcess) error {
	return nil
}

func Test_ensureServiceAccount(t *testing.T) {
	tests := []struct {
		name    string
		context *mocks.StepContextMock
		wantErr assert.ErrorAssertionFunc
		marked  bool
	}{
		{
			name:    "ok",
			context: &mocks.StepContextMock{},
			wantErr: assert.NoError,
			marked:  true,
		},
		{
			name:    "error",
			context: &mocks.StepContextMock{OnCreateServiceAccountError: serviceError},
			wantErr: assert.Error,
			marked:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureServiceAccountStep(tt.context))

			assert.Equal(t, tt.marked, tt.context.MarkServiceAccountAsReadyWasCalled)
		})
	}
}

func Test_ensureServiceAccountAcl(t *testing.T) {
	tests := []struct {
		name    string
		context *mocks.StepContextMock
		wantErr assert.ErrorAssertionFunc
		marked  bool
	}{
		{
			name:    "ok",
			context: &mocks.StepContextMock{ReturnClusterAccess: &models.ClusterAccess{}},
			wantErr: assert.NoError,
			marked:  true,
		},
		{
			name:    "create entry",
			context: &mocks.StepContextMock{ReturnClusterAccess: &models.ClusterAccess{Acl: []models.AclEntry{{}}}},
			wantErr: assert.NoError,
			marked:  false,
		},
		{
			name: "create entry error",
			context: &mocks.StepContextMock{
				ReturnClusterAccess:   &models.ClusterAccess{Acl: []models.AclEntry{{}}},
				OnCreateAclEntryError: serviceError,
			},
			wantErr: assert.Error,
			marked:  false,
		},
		{
			name:    "get or create error",
			context: &mocks.StepContextMock{OnGetOrCreateClusterAccessError: serviceError},
			wantErr: assert.Error,
			marked:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureServiceAccountAclStep(tt.context))

			assert.Equal(t, tt.marked, tt.context.MarkClusterAccessAsReadyWasCalled)
		})
	}
}

func Test_ensureServiceAccountApiKey(t *testing.T) {
	tests := []struct {
		name    string
		context *mocks.StepContextMock
		wantErr assert.ErrorAssertionFunc
		marked  bool
	}{
		{
			name:    "ok",
			context: &mocks.StepContextMock{ReturnClusterAccess: &models.ClusterAccess{}},
			wantErr: assert.NoError,
			marked:  true,
		},
		{
			name:    "get cluster access error",
			context: &mocks.StepContextMock{OnGetClusterAccessError: serviceError},
			wantErr: assert.Error,
			marked:  false,
		},
		{
			name:    "create api key error",
			context: &mocks.StepContextMock{OnCreateApiKeyError: serviceError},
			wantErr: assert.Error,
			marked:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureServiceAccountApiKeyStep(tt.context))

			assert.Equal(t, tt.marked, tt.context.MarkApiKeyAsReadyWasCalled)
		})
	}
}

func Test_ensureServiceAccountApiKeyAreStoredInVault(t *testing.T) {
	tests := []struct {
		name    string
		context *mocks.StepContextMock
		wantErr assert.ErrorAssertionFunc
		marked  bool
	}{
		{
			name:    "ok",
			context: &mocks.StepContextMock{ReturnClusterAccess: &models.ClusterAccess{}},
			wantErr: assert.NoError,
			marked:  true,
		},
		{
			name:    "get cluster access error",
			context: &mocks.StepContextMock{OnGetClusterAccessError: serviceError},
			wantErr: assert.Error,
			marked:  false,
		},
		{
			name: "store api key error",
			context: &mocks.StepContextMock{
				ReturnClusterAccess: &models.ClusterAccess{},
				OnStoreApiKeyError:  serviceError,
			},
			wantErr: assert.Error,
			marked:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureServiceAccountApiKeyAreStoredInVaultStep(tt.context))

			assert.Equal(t, tt.marked, tt.context.MarkApiKeyInVaultAsReadyWasCalled)
		})
	}
}

func Test_ensureTopicIsCreated(t *testing.T) {
	tests := []struct {
		name        string
		context     *mocks.StepContextMock
		wantErr     assert.ErrorAssertionFunc
		marked      bool
		eventRaised bool
	}{
		{
			name:        "ok",
			context:     &mocks.StepContextMock{},
			wantErr:     assert.NoError,
			marked:      true,
			eventRaised: true,
		},
		{
			name:        "create topic error",
			context:     &mocks.StepContextMock{OnCreateTopicError: serviceError},
			wantErr:     assert.Error,
			marked:      false,
			eventRaised: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureTopicIsCreatedStep(tt.context))

			assert.Equal(t, tt.marked, tt.context.MarkAsCompletedWasCalled)
			assert.Equal(t, tt.eventRaised, tt.context.TopicProvisionedEventWasRaised)
		})
	}
}
*/
