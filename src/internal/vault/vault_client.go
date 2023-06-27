package vault

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
)

type Vault struct {
	logger logging.Logger
	config aws.Config
}

func getClusterApiParameter(capabilityId models.CapabilityId, clusterId models.ClusterId) string {
	return fmt.Sprintf("/capabilities/%s/kafka/%s/credentials", capabilityId, clusterId)
}

func getSchemaRegistryApiParameter(capabilityId models.CapabilityId, clusterId models.ClusterId) string {
	return fmt.Sprintf("/capabilities/%s/kafka/%s/schemaregistry-credentials", capabilityId, clusterId)
}

func (v *Vault) storeApiKey(ctx context.Context, capabilityId models.CapabilityId, parameterName string, apiKey models.ApiKey) error {

	if apiKey.Username == "" && apiKey.Password == "" {
		return fmt.Errorf("attempted to store api key with empty username and password for parameter at location %s", parameterName)
	}
	v.logger.Information("Storing api key {ApiKeyUserName} for capability {CapabilityId} at location {ParameterName}", apiKey.Username, string(capabilityId), parameterName)

	client := ssm.NewFromConfig(v.config)

	v.logger.Trace("Sending request to AWS Parameter Store")
	_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(parameterName),
		Value: aws.String(`{ "key": "` + apiKey.Username + `", "secret": "` + apiKey.Password + `" }`),
		Tags: []types.Tag{
			{
				Key:   aws.String("capabilityId"),
				Value: aws.String(string(capabilityId)),
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
		return fmt.Errorf("error when storing api key %s for capability %s at location %s", apiKey.Username, string(capabilityId), parameterName)
	}

	v.logger.Information("Successfully stored api key {ApiKeyUserName} for capability {CapabilityId} at location {ParameterName}", apiKey.Username, string(capabilityId), parameterName)

	return nil
}

func (v *Vault) StoreClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error {
	return v.storeApiKey(ctx, capabilityId, getClusterApiParameter(capabilityId, clusterId), apiKey)
}
func (v *Vault) StoreSchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error {
	return v.storeApiKey(ctx, capabilityId, getSchemaRegistryApiParameter(capabilityId, clusterId), apiKey)
}

func (v *Vault) queryApiKey(ctx context.Context, parameterName string, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	v.logger.Trace("Querying existence of API key for capability {CapabilityId} in cluster {ClusterId} at location {ParameterName}", string(capabilityId), string(clusterId), parameterName)

	client := ssm.NewFromConfig(v.config)

	v.logger.Trace("Sending request to AWS Parameter Store")

	_, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(parameterName),
	})
	if err != nil {
		var pnf *types.ParameterNotFound
		if errors.As(err, &pnf) {
			return false, nil
		}
		return false, err
	}

	return true, nil

}

func (v *Vault) QuerySchemaRegistryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	return v.queryApiKey(ctx, getSchemaRegistryApiParameter(capabilityId, clusterId), capabilityId, clusterId)
}

func (v *Vault) QueryClusterApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	return v.queryApiKey(ctx, getClusterApiParameter(capabilityId, clusterId), capabilityId, clusterId)
}

func NewDefaultConfig() (*aws.Config, error) {
	config, err := config.LoadDefaultConfig(context.Background())
	return &config, err
}

func NewTestConfig(url string) (*aws.Config, error) {
	cfg := aws.NewConfig()
	cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: url,
		}, nil
	})
	cfg.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			CanExpire: false,
			Expires:   time.Time{},
		}, nil
	})

	return cfg, nil
}

func NewVaultClient(logger logging.Logger, cfg *aws.Config) (*Vault, error) {
	if cfg == nil {
		return nil, errors.New("cannot create a valid vault client with a nil config")
	}

	return &Vault{
		logger: logger,
		config: *cfg,
	}, nil
}
