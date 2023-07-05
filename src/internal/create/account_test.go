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

func TestAccountHelper_CreateServiceAccount_Ok(t *testing.T) {
	spy := &mocks.AccountRepository{ReturnServiceAccount: &models.ServiceAccount{
		Id: someServiceAccountId,
		ClusterAccesses: []models.ClusterAccess{
			{
				ClusterId: someClusterId,
			},
		},
	}}
	sut := NewAccountService(context.TODO(), spy)

	access, err := sut.HasClusterAccess(someCapabilityId, someClusterId)

	assert.NoError(t, err)
	assert.Equal(t, true, access)
}
