package process

import (
	"errors"
	"github.com/dfds/confluent-gateway/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

var serviceError = errors.New("service error")

func Test_ensureServiceAccount(t *testing.T) {
	tests := []struct {
		name    string
		mock    *ProcessMock
		wantErr assert.ErrorAssertionFunc
		wantOk  bool
	}{
		{
			name:    "ok",
			mock:    &ProcessMock{},
			wantErr: assert.NoError,
			wantOk:  true,
		},
		{
			name: "error",

			mock:    &ProcessMock{OnCreateServiceAccountError: serviceError},
			wantErr: assert.Error,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureServiceAccountStep(tt.mock))

			assert.Equal(t, tt.wantOk, tt.mock.HasServiceAccount)
		})
	}
}

func Test_ensureServiceAccountAcl(t *testing.T) {
	tests := []struct {
		name    string
		mock    *ProcessMock
		wantErr assert.ErrorAssertionFunc
		wantOk  bool
	}{
		{
			name:    "ok",
			mock:    &ProcessMock{ReturnClusterAccess: &models.ClusterAccess{}},
			wantErr: assert.NoError,
			wantOk:  true,
		},
		{
			name:    "create entry",
			mock:    &ProcessMock{ReturnClusterAccess: &models.ClusterAccess{Acl: []models.AclEntry{{}}}},
			wantErr: assert.NoError,
			wantOk:  false,
		},
		{
			name: "create entry error",
			mock: &ProcessMock{
				ReturnClusterAccess:   &models.ClusterAccess{Acl: []models.AclEntry{{}}},
				OnCreateAclEntryError: serviceError,
			},
			wantErr: assert.Error,
			wantOk:  false,
		},
		{
			name:    "get or create error",
			mock:    &ProcessMock{OnGetOrCreateClusterAccessError: serviceError},
			wantErr: assert.Error,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureServiceAccountAclStep(tt.mock))

			assert.Equal(t, tt.wantOk, tt.mock.HasClusterAccess)
		})
	}
}

func Test_ensureServiceAccountApiKey(t *testing.T) {
	tests := []struct {
		name    string
		mock    *ProcessMock
		wantErr assert.ErrorAssertionFunc
		wantOk  bool
	}{
		{
			name:    "ok",
			mock:    &ProcessMock{ReturnClusterAccess: &models.ClusterAccess{}},
			wantErr: assert.NoError,
			wantOk:  true,
		},
		{
			name:    "get cluster access error",
			mock:    &ProcessMock{OnGetClusterAccessError: serviceError},
			wantErr: assert.Error,
			wantOk:  false,
		},
		{
			name:    "create api key error",
			mock:    &ProcessMock{OnCreateApiKeyError: serviceError},
			wantErr: assert.Error,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureServiceAccountApiKeyStep(tt.mock))

			assert.Equal(t, tt.wantOk, tt.mock.HasApiKey)
		})
	}
}

func Test_ensureServiceAccountApiKeyAreStoredInVault(t *testing.T) {
	tests := []struct {
		name    string
		mock    *ProcessMock
		wantErr assert.ErrorAssertionFunc
		wantOk  bool
	}{
		{
			name:    "ok",
			mock:    &ProcessMock{ReturnClusterAccess: &models.ClusterAccess{}},
			wantErr: assert.NoError,
			wantOk:  true,
		},
		{
			name:    "get cluster access error",
			mock:    &ProcessMock{OnGetClusterAccessError: serviceError},
			wantErr: assert.Error,
			wantOk:  false,
		},
		{
			name: "store api key error",
			mock: &ProcessMock{
				ReturnClusterAccess: &models.ClusterAccess{},
				OnStoreApiKeyError:  serviceError,
			},
			wantErr: assert.Error,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureServiceAccountApiKeyAreStoredInVaultStep(tt.mock))

			assert.Equal(t, tt.wantOk, tt.mock.HasApiKeyInVault)
		})
	}
}

func Test_ensureTopicIsCreated(t *testing.T) {
	tests := []struct {
		name         string
		process      *Process
		mock         *ProcessMock
		wantErr      assert.ErrorAssertionFunc
		wantOk       bool
		wantProduced bool
	}{
		{
			name:         "ok",
			mock:         &ProcessMock{},
			wantErr:      assert.NoError,
			wantOk:       true,
			wantProduced: true,
		},
		{
			name:         "create topic error",
			mock:         &ProcessMock{OnCreateTopicError: serviceError},
			wantErr:      assert.Error,
			wantOk:       false,
			wantProduced: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureTopicIsCreatedStep(tt.mock))

			assert.Equal(t, tt.wantOk, tt.mock.IsCompleted)
			assert.Equal(t, tt.wantProduced, tt.mock.TopicProvisioned)
		})
	}
}

// region Test Doubles

type ProcessMock struct {
	ReturnClusterAccess *models.ClusterAccess

	HasServiceAccount bool
	HasClusterAccess  bool
	HasApiKey         bool
	HasApiKeyInVault  bool
	IsCompleted       bool
	TopicProvisioned  bool

	OnCreateServiceAccountError     error
	OnGetOrCreateClusterAccessError error
	OnCreateAclEntryError           error
	OnGetClusterAccessError         error
	OnCreateApiKeyError             error
	OnStoreApiKeyError              error
	OnCreateTopicError              error
}

func (p *ProcessMock) hasServiceAccount() bool {
	return false
}

func (p *ProcessMock) createServiceAccount() error {
	return p.OnCreateServiceAccountError
}

func (p *ProcessMock) markServiceAccountReady() {
	p.HasServiceAccount = true
}

func (p *ProcessMock) hasClusterAccess() bool {
	return false
}

func (p *ProcessMock) getOrCreateClusterAccess() (*models.ClusterAccess, error) {
	return p.ReturnClusterAccess, p.OnGetOrCreateClusterAccessError
}

func (p *ProcessMock) createAclEntry(*models.ClusterAccess, models.AclEntry) error {
	return p.OnCreateAclEntryError
}

func (p *ProcessMock) markClusterAccessReady() {
	p.HasClusterAccess = true
}

func (p *ProcessMock) hasApiKey() bool {
	return false
}

func (p *ProcessMock) getClusterAccess() (*models.ClusterAccess, error) {
	return p.ReturnClusterAccess, p.OnGetClusterAccessError
}

func (p *ProcessMock) createApiKey(*models.ClusterAccess) error {
	return p.OnCreateApiKeyError
}

func (p *ProcessMock) markApiKeyReady() {
	p.HasApiKey = true
}

func (p *ProcessMock) hasApiKeyInVault() bool {
	return false
}

func (p *ProcessMock) storeApiKey(*models.ClusterAccess) error {
	return p.OnStoreApiKeyError
}

func (p *ProcessMock) markApiKeyInVaultReady() {
	p.HasApiKeyInVault = true
}

func (p *ProcessMock) isCompleted() bool {
	return false
}

func (p *ProcessMock) createTopic() error {
	return p.OnCreateTopicError
}

func (p *ProcessMock) markAsCompleted() {
	p.IsCompleted = true
}

func (p *ProcessMock) topicProvisioned() error {
	p.TopicProvisioned = true
	return nil
}

// endregion
