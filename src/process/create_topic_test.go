package process

import (
	"errors"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/dfds/confluent-gateway/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

var serviceError = errors.New("service error")

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
