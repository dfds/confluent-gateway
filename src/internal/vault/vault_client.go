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

func (v *Vault) StoreApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId, apiKey models.ApiKey) error {

	parameterName := fmt.Sprintf("/capabilities/%s/kafka/%s/credentials", capabilityId, clusterId)
	v.logger.Trace("Storing api key {ApiKeyUserName} for capability {CapabilityId} in cluster {ClusterId} at location {ParameterName}", apiKey.Username, string(capabilityId), string(clusterId), parameterName)

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
		v.logger.Error(err, "Error when storing api key {ApiKeyUserName} for capability {CapabilityId} i cluster {ClusterId}", apiKey.Username, string(capabilityId), string(clusterId))
		return err
	}

	v.logger.Information("Successfully stored api key {ApiKeyUserName} for capability {CapabilityId} at location {ParameterName}", apiKey.Username, string(capabilityId), parameterName)

	return nil
}

func (v *Vault) QueryApiKey(ctx context.Context, capabilityId models.CapabilityId, clusterId models.ClusterId) (bool, error) {
	parameterName := fmt.Sprintf("/capabilities/%s/kafka/%s/credentials", capabilityId, clusterId)
	v.logger.Trace("Querying existence of API key for capability {CapabilityId} in cluster {ClusterId} at location {ParameterName}", string(capabilityId), string(clusterId), parameterName)

	/*
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
	*/
	return false, nil
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
