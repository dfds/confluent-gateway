package process

import (
	"errors"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/dfds/confluent-gateway/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccountHelper_CreateServiceAccount_Ok(t *testing.T) {
	const serviceAccountId = models.ServiceAccountId("sa-1234")
	const capabilityRootId = models.CapabilityRootId("some-capability-root-id")
	const clusterId = models.ClusterId("some-cluster-id")

	spy := &mocks.AccountService{
		ReturnCapabilityRootId: capabilityRootId,
		ReturnClusterId:        clusterId,
		ReturnServiceAccountId: serviceAccountId,
	}
	err := NewAccountHelper(spy).CreateServiceAccount()

	assert.NoError(t, err)
	assert.Equal(t, serviceAccountId, spy.SavedServiceAccount.Id)
	assert.Equal(t, capabilityRootId, spy.SavedServiceAccount.CapabilityRootId)
	assert.Equal(t, serviceAccountId, spy.SavedServiceAccount.ClusterAccesses[0].ServiceAccountId)
	assert.Equal(t, clusterId, spy.SavedServiceAccount.ClusterAccesses[0].ClusterId)
}

func TestAccountHelper_CreateServiceAccount_ErrorCreatingServiceAccount(t *testing.T) {
	const errorText = "confluent failed"
	stub := &mocks.AccountService{OnCreateServiceAccountError: errors.New(errorText)}

	err := NewAccountHelper(stub).CreateServiceAccount()

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateServiceAccount_ErrorPersistingServiceAccount(t *testing.T) {
	const errorText = "db failed"
	stub := &mocks.AccountService{OnSaveServiceAccountError: errors.New(errorText)}

	err := NewAccountHelper(stub).CreateServiceAccount()

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_GetOrCreateClusterAccess_GetExistingClusterAccess(t *testing.T) {
	const serviceAccountId = models.ServiceAccountId("sa-1234")
	const capabilityRootId = models.CapabilityRootId("some-capability-root-id")
	const clusterId = models.ClusterId("some-cluster-id")
	clusterAccess := models.ClusterAccess{
		ClusterId:        clusterId,
		ServiceAccountId: serviceAccountId,
		ApiKey:           models.ApiKey{},
	}
	s := &mocks.AccountService{
		ReturnCapabilityRootId: capabilityRootId,
		ReturnClusterId:        clusterId,
		ReturnServiceAccount:   &models.ServiceAccount{ClusterAccesses: []models.ClusterAccess{clusterAccess}},
	}

	result, err := NewAccountHelper(s).GetOrCreateClusterAccess()

	assert.NoError(t, err)
	assert.Equal(t, &clusterAccess, result)

}

func TestAccountHelper_GetOrCreateClusterAccess_CreateNewClusterAccess(t *testing.T) {
	const serviceAccountId = models.ServiceAccountId("sa-1234")
	const capabilityRootId = models.CapabilityRootId("some-capability-root-id")
	const clusterId = models.ClusterId("some-cluster-id")
	spy := &mocks.AccountService{
		ReturnCapabilityRootId: capabilityRootId,
		ReturnClusterId:        clusterId,
		ReturnServiceAccount:   &models.ServiceAccount{Id: serviceAccountId},
	}

	_, err := NewAccountHelper(spy).GetOrCreateClusterAccess()

	assert.NoError(t, err)
	assert.Equal(t, spy.SavedClusterAccess.ClusterId, clusterId)
	assert.Equal(t, spy.SavedClusterAccess.ServiceAccountId, serviceAccountId)
	assert.Equal(t, spy.SavedClusterAccess.ApiKey, models.ApiKey{})
}

// TODO TestAccountHelper_GetOrCreateClusterAccess_CreateNewClusterAccessError

func TestAccountHelper_GetOrCreateClusterAccess_DatabaseError(t *testing.T) {
	const errorText = "database error"
	s := &mocks.AccountService{
		OnGetServiceAccountError: errors.New(errorText),
	}

	_, err := NewAccountHelper(s).GetOrCreateClusterAccess()

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_GetOrCreateClusterAccess_NoServiceAccount(t *testing.T) {
	_, err := NewAccountHelper(&mocks.AccountService{ReturnCapabilityRootId: "some-capability-root-id"}).GetOrCreateClusterAccess()

	assert.EqualError(t, err, "no service account for capability 'some-capability-root-id' found")
}

func TestAccountHelper_CreateAclEntry_Ok(t *testing.T) {
	entry := &models.AclEntry{}

	err := NewAccountHelper(&mocks.AccountService{}).CreateAclEntry("sa-1234", entry)

	assert.NoError(t, err)
	assert.NotNil(t, entry.CreatedAt)
}

func TestAccountHelper_CreateAclEntry_ConfluentError(t *testing.T) {
	const errorText = "confluent error"
	s := &mocks.AccountService{OnCreateAclEntryError: errors.New(errorText)}

	err := NewAccountHelper(s).CreateAclEntry("sa-1234", &models.AclEntry{})

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateAclEntry_DatabaseError(t *testing.T) {
	const errorText = "database error"
	s := &mocks.AccountService{OnUpdateAclEntryError: errors.New(errorText)}

	err := NewAccountHelper(s).CreateAclEntry("sa-1234", &models.AclEntry{})

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateApiKey_Ok(t *testing.T) {
	apiKey := models.ApiKey{
		Username: "some-user",
		Password: "some-pass",
	}
	s := &mocks.AccountService{ReturnApiKey: apiKey}
	clusterAccess := &models.ClusterAccess{}

	err := NewAccountHelper(s).CreateApiKey(clusterAccess)

	assert.NoError(t, err)
	assert.Equal(t, apiKey, clusterAccess.ApiKey)
}

func TestAccountHelper_CreateApiKey_ConfluentError(t *testing.T) {
	const errorText = "confluent error"
	s := &mocks.AccountService{OnCreateApiKeyError: errors.New(errorText)}

	err := NewAccountHelper(s).CreateApiKey(&models.ClusterAccess{})

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateApiKey_DatabaseError(t *testing.T) {
	const errorText = "database error"
	s := &mocks.AccountService{OnUpdateClusterAccessError: errors.New(errorText)}

	err := NewAccountHelper(s).CreateApiKey(&models.ClusterAccess{})

	assert.EqualError(t, err, errorText)
}
