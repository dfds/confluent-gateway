package vault

import "github.com/dfds/confluent-gateway/internal/models"

type OperationDestination string

const (
	OperationDestinationCluster        OperationDestination = "cluster"
	OperationDestinationSchemaRegistry OperationDestination = "schema-registry"
)

type Input struct {
	OperationDestination OperationDestination
	CapabilityId         models.CapabilityId
	ClusterId            models.ClusterId

	// storing fields
	StoringInput *StoringInput
}

type StoringInput struct {
	ApiKey    models.ApiKey
	Overwrite bool
}
