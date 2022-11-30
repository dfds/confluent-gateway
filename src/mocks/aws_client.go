package mocks

import (
	"fmt"
	"github.com/dfds/confluent-gateway/models"
)

type MockAwsClient struct {
}

func (m MockAwsClient) PutApiKey(apikey models.ApiKey) error {
	fmt.Printf("Putting API key\n")
	return nil
}
