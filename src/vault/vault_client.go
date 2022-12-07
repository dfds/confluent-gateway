package vault

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/models"
)

type vault struct {
	logger logging.Logger
	config aws.Config
}

func (v *vault) StoreApiKey(ctx context.Context, capabilityRootId models.CapabilityRootId, clusterId models.ClusterId, apiKey models.ApiKey) error {

	parameterName := fmt.Sprintf("/capabilities/%s/kafka/%s/credentials", capabilityRootId, clusterId)
	v.logger.Trace("Storing api key {ApiKeyUserName} for capability {CapabilityRootId} in cluster {ClusterId} at location {ParameterName}", apiKey.Username, string(capabilityRootId), string(clusterId), parameterName)

	client := ssm.NewFromConfig(v.config)

	v.logger.Trace("Sending request to AWS Parameter Store")
	_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(parameterName),
		Value: aws.String(`{ "key": "` + apiKey.Username + `", "secret": "` + apiKey.Password + `" }`),
		Tags: []types.Tag{
			{
				Key:   aws.String("capabilityRootId"),
				Value: aws.String(string(capabilityRootId)),
			},
			{
				Key:   aws.String("createdBy"),
				Value: aws.String("Kafka-Janitor"),
			},
		},
		Tier: types.ParameterTierStandard,
		Type: types.ParameterTypeSecureString,
	})

	if err != nil {
		v.logger.Error(&err, "Error when storing api key {ApiKeyUserName} for capability {CapabilityRootId} i cluster {ClusterId}", apiKey.Username, string(capabilityRootId), string(clusterId))
		return err
	}

	v.logger.Information("Successfully stored api key {ApiKeyUserName} for capability {CapabilityRootId} at location {ParameterName}", apiKey.Username, string(capabilityRootId), parameterName)

	return nil
}

func NewVaultClient(logger logging.Logger) (models.Vault, error) {
	config, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		logger.Error(&err, "Error when loading AWS config!")
		return nil, err
	}

	return &vault{
		logger: logger,
		config: config,
	}, nil
}
