package helpers

import (
	"fmt"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/gofrs/uuid"
)

type TestVariablesConfig struct {
	TopicId          string
	TopicName        string
	ServiceAccountId models.ServiceAccountId
	CapabilityId     models.CapabilityId
}

func NewTestVariables(prefix string) *TestVariablesConfig {
	topicId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return &TestVariablesConfig{
		TopicId:          topicId.String(),
		TopicName:        fmt.Sprintf("%s-topic-name", prefix),
		ServiceAccountId: models.ServiceAccountId(fmt.Sprintf("%s-service-account-id", prefix)),
		CapabilityId:     models.CapabilityId(fmt.Sprintf("%s-capability-id", prefix)),
	}
}
