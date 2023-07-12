package helpers

import (
	"fmt"
	"github.com/dfds/confluent-gateway/internal/models"
)

type TestVariablesConfig struct {
	TopicId          string
	TopicName        string
	ServiceAccountId models.ServiceAccountId
	CapabilityId     models.CapabilityId
}

func NewTestVariables(prefix string) *TestVariablesConfig {
	return &TestVariablesConfig{
		TopicId:          fmt.Sprintf("%s-topic-id", prefix),
		TopicName:        fmt.Sprintf("%s-topic-name", prefix),
		ServiceAccountId: models.ServiceAccountId(fmt.Sprintf("%s-service-account-id", prefix)),
		CapabilityId:     models.CapabilityId(fmt.Sprintf("%s-capability-id", prefix)),
	}
}
