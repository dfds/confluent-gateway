package create

import (
	"context"
	"testing"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
)

const someServiceAccountId = models.ServiceAccountId("sa-1234")
const someCapabilityId = models.CapabilityId("some-capability-id")
const someClusterId = models.ClusterId("some-cluster-id")
const someTopicId = "e72d7a14-b240-4ace-a8e0-27ee0b0ccb25"
const someUserAccountId = 1234

func TestAccountHelper_CreateServiceAccount_Ok(t *testing.T) {
	spy := &mocks.AccountRepository{}
	sut := NewAccountService(context.TODO(), spy)

	_, err := sut.GetServiceAccount(someCapabilityId)

	assert.NoError(t, err)
	assert.Equal(t, someServiceAccountId, spy.GotServiceAccount.Id)
	assert.Equal(t, models.MakeUserAccountId(someUserAccountId), spy.GotServiceAccount.UserAccountId)
	assert.Equal(t, someCapabilityId, spy.GotServiceAccount.CapabilityId)
	assert.Equal(t, someServiceAccountId, spy.GotServiceAccount.ClusterAccesses[0].ServiceAccountId)
	assert.Equal(t, someClusterId, spy.GotServiceAccount.ClusterAccesses[0].ClusterId)
}
