package serviceaccount

/*
import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

const someServiceAccountId = models.ServiceAccountId("sa-1234")
const someCapabilityId = models.CapabilityId("some-capability-id")
const someClusterId = models.ClusterId("some-cluster-id")
const someTopicId = "e72d7a14-b240-4ace-a8e0-27ee0b0ccb25"
const someUserAccountId = 1234

func TestAccountHelper_CreateServiceAccount_Ok(t *testing.T) {
	stub := &mocks.MockClient{ReturnServiceAccountId: someServiceAccountId, ReturnUsers: []models.User{{Id: someUserAccountId, ResourceID: string(someServiceAccountId)}}}
	spy := &mocks.AccountRepository{}
	sut := NewAccountService(context.TODO(), stub, spy)

	err := sut.CreateServiceAccount(someCapabilityId, someClusterId)

	assert.NoError(t, err)
	assert.Equal(t, someServiceAccountId, spy.GotServiceAccount.Id)
	assert.Equal(t, models.MakeUserAccountId(someUserAccountId), spy.GotServiceAccount.UserAccountId)
	assert.Equal(t, someCapabilityId, spy.GotServiceAccount.CapabilityId)
	assert.Equal(t, someServiceAccountId, spy.GotServiceAccount.ClusterAccesses[0].ServiceAccountId)
	assert.Equal(t, someClusterId, spy.GotServiceAccount.ClusterAccesses[0].ClusterId)
}

func TestAccountHelper_CreateServiceAccount_ErrorCreatingServiceAccount(t *testing.T) {
	const errorText = "confluent failed"
	stub := &mocks.MockClient{OnCreateServiceAccountError: errors.New(errorText)}
	sut := NewAccountService(context.TODO(), stub, &mocks.AccountRepository{})

	err := sut.CreateServiceAccount(someCapabilityId, someClusterId)

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateServiceAccount_ErrorGettingUsers(t *testing.T) {
	const errorText = "confluent failed"
	stub := &mocks.MockClient{OnGetUsersError: errors.New(errorText)}
	sut := NewAccountService(context.TODO(), stub, &mocks.AccountRepository{})

	err := sut.CreateServiceAccount(someCapabilityId, someClusterId)

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateServiceAccount_MissingUserAccount(t *testing.T) {
	const errorText = "confluent failed"
	stub := &mocks.MockClient{ReturnServiceAccountId: someServiceAccountId}
	sut := NewAccountService(context.TODO(), stub, &mocks.AccountRepository{})

	err := sut.CreateServiceAccount(someCapabilityId, someClusterId)

	assert.EqualError(t, err, fmt.Sprintf("unable to find matching user account for %q", someServiceAccountId))
}

func TestAccountHelper_CreateServiceAccount_ErrorPersistingServiceAccount(t *testing.T) {
	const errorText = "db failed"
	repoStub := &mocks.AccountRepository{OnCreateServiceAccountError: errors.New(errorText)}
	clientStub := &mocks.MockClient{ReturnServiceAccountId: someServiceAccountId, ReturnUsers: []models.User{{Id: someUserAccountId, ResourceID: string(someServiceAccountId)}}}
	sut := NewAccountService(context.TODO(), clientStub, repoStub)

	err := sut.CreateServiceAccount(someCapabilityId, someClusterId)

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_GetOrCreateClusterAccess_GetExistingClusterAccess(t *testing.T) {
	clusterAccess := models.ClusterAccess{
		ClusterId:        someClusterId,
		ServiceAccountId: someServiceAccountId,
		ApiKey:           models.ApiKey{},
	}
	stub := &mocks.AccountRepository{
		ReturnServiceAccount: &models.ServiceAccount{ClusterAccesses: []models.ClusterAccess{clusterAccess}},
	}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, stub)

	result, err := sut.GetOrCreateClusterAccess(someCapabilityId, someClusterId)

	assert.NoError(t, err)
	assert.Equal(t, &clusterAccess, result)

}
func TestAccountHelper_GetOrCreateClusterAccess_GetExistingClusterAccessError(t *testing.T) {
	const errorText = "database error"
	stub := &mocks.AccountRepository{
		OnGetServiceAccountError: errors.New(errorText),
	}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, stub)

	_, err := sut.GetOrCreateClusterAccess(someCapabilityId, someClusterId)

	assert.EqualError(t, err, errorText)

}

func TestAccountHelper_GetOrCreateClusterAccess_CreateNewClusterAccess(t *testing.T) {
	spy := &mocks.AccountRepository{
		ReturnServiceAccount: &models.ServiceAccount{Id: someServiceAccountId, UserAccountId: models.MakeUserAccountId(someUserAccountId)},
	}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, spy)

	_, err := sut.GetOrCreateClusterAccess(someCapabilityId, someClusterId)

	assert.NoError(t, err)
	assert.Equal(t, spy.GotClusterAccess.ClusterId, someClusterId)
	assert.Equal(t, spy.GotClusterAccess.ServiceAccountId, someServiceAccountId)
	assert.Equal(t, spy.GotClusterAccess.UserAccountId, models.MakeUserAccountId(someUserAccountId))
	assert.Equal(t, spy.GotClusterAccess.ApiKey, models.ApiKey{})
}

func TestAccountHelper_GetOrCreateClusterAccess_DatabaseError(t *testing.T) {
	const errorText = "database error"
	stub := &mocks.AccountRepository{
		ReturnServiceAccount:       &models.ServiceAccount{},
		OnCreateClusterAccessError: errors.New(errorText),
	}

	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, stub)

	_, err := sut.GetOrCreateClusterAccess(someCapabilityId, someClusterId)

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_GetOrCreateClusterAccess_NoServiceAccount(t *testing.T) {
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, &mocks.AccountRepository{})

	_, err := sut.GetOrCreateClusterAccess(someCapabilityId, someClusterId)

	assert.EqualError(t, err, fmt.Sprintf("no service account for capability '%s' found", someCapabilityId))
}

func TestAccountHelper_GetClusterAccess_GetExistingClusterAccess(t *testing.T) {
	clusterAccess := models.ClusterAccess{
		ClusterId:        someClusterId,
		ServiceAccountId: someServiceAccountId,
		ApiKey:           models.ApiKey{},
	}
	stub := &mocks.AccountRepository{
		ReturnServiceAccount: &models.ServiceAccount{ClusterAccesses: []models.ClusterAccess{clusterAccess}},
	}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, stub)

	result, err := sut.GetClusterAccess(someCapabilityId, someClusterId)

	assert.NoError(t, err)
	assert.Equal(t, &clusterAccess, result)

}

func TestAccountHelper_GetClusterAccess_DatabaseError(t *testing.T) {
	const errorText = "database error"
	stub := &mocks.AccountRepository{
		OnGetServiceAccountError: errors.New(errorText),
	}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, stub)

	_, err := sut.GetClusterAccess(someCapabilityId, someClusterId)

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_GetClusterAccess_NoServiceAccount(t *testing.T) {
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, &mocks.AccountRepository{})

	_, err := sut.GetClusterAccess(someCapabilityId, someClusterId)

	assert.EqualError(t, err, fmt.Sprintf("no service account for capability '%s' found", someCapabilityId))
}

func TestAccountHelper_GetClusterAccess_NoClusterAccess(t *testing.T) {
	stub := &mocks.AccountRepository{ReturnServiceAccount: &models.ServiceAccount{Id: someServiceAccountId}}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, stub)

	_, err := sut.GetClusterAccess(someCapabilityId, someClusterId)

	assert.EqualError(t, err, fmt.Sprintf("no cluster access for service account '%s' found", someServiceAccountId))
}

func TestAccountHelper_CreateAclEntry_Ok(t *testing.T) {
	entry := &models.AclEntry{}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, &mocks.AccountRepository{})

	err := sut.CreateAclEntry(someClusterId, models.MakeUserAccountId(someUserAccountId), entry)

	assert.NoError(t, err)
	assert.NotNil(t, entry.CreatedAt)
}

func TestAccountHelper_CreateAclEntry_ConfluentError(t *testing.T) {
	const errorText = "confluent error"
	stub := &mocks.MockClient{OnCreateAclEntryError: errors.New(errorText)}
	sut := NewAccountService(context.TODO(), stub, &mocks.AccountRepository{})

	err := sut.CreateAclEntry(someClusterId, models.MakeUserAccountId(someUserAccountId), &models.AclEntry{})

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateAclEntry_DatabaseError(t *testing.T) {
	const errorText = "database error"
	stub := &mocks.AccountRepository{OnUpdateAclEntryError: errors.New(errorText)}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, stub)

	err := sut.CreateAclEntry(someClusterId, models.MakeUserAccountId(someUserAccountId), &models.AclEntry{})

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateApiKey_Ok(t *testing.T) {
	clusterAccess := &models.ClusterAccess{}
	apiKey := models.ApiKey{Username: "some-user", Password: "some-pass"}
	stub := &mocks.MockClient{ReturnApiKey: apiKey}
	sut := NewAccountService(context.TODO(), stub, &mocks.AccountRepository{})

	err := sut.CreateApiKey(clusterAccess)

	assert.NoError(t, err)
	assert.Equal(t, apiKey, clusterAccess.ApiKey)
}

func TestAccountHelper_CreateApiKey_ConfluentError(t *testing.T) {
	const errorText = "confluent error"
	stub := &mocks.MockClient{OnCreateApiKeyError: errors.New(errorText)}
	sut := NewAccountService(context.TODO(), stub, &mocks.AccountRepository{})

	err := sut.CreateApiKey(&models.ClusterAccess{})

	assert.EqualError(t, err, errorText)
}

func TestAccountHelper_CreateApiKey_DatabaseError(t *testing.T) {
	const errorText = "database error"
	stub := &mocks.AccountRepository{OnUpdateClusterAccessError: errors.New(errorText)}
	sut := NewAccountService(context.TODO(), &mocks.MockClient{}, stub)

	err := sut.CreateApiKey(&models.ClusterAccess{})

	assert.EqualError(t, err, errorText)
}
*/
