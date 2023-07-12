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
	"github.com/dfds/confluent-gateway/logging"
)

type vault struct {
	logger logging.Logger
	config aws.Config
}

func getApiParameter(input Input) string {
	switch input.OperationDestination {
	case OperationDestinationCluster:
		return fmt.Sprintf("/capabilities/%s/kafka/%s/credentials", input.CapabilityId, input.ClusterId)
	case OperationDestinationSchemaRegistry:
		return fmt.Sprintf("/capabilities/%s/kafka/%s/schemaregistry-credentials", input.CapabilityId, input.ClusterId)
	default:
		return ""
	}
}

func validateInput(input Input, isStoring bool) error {

	switch input.OperationDestination {
	case OperationDestinationCluster:
	case OperationDestinationSchemaRegistry:
	default:
		return errors.New(fmt.Sprintf("invalid operationDestination: %s", input.OperationDestination))
	}
	if input.CapabilityId == "" {
		return errors.New("capabilityId is required")
	}
	if input.ClusterId == "" {
		return errors.New("clusterId is required")
	}

	if !isStoring {
		return nil
	}

	if input.StoringInput == nil {
		return errors.New("storingInput is required")
	}
	if input.StoringInput.ApiKey.Username == "" {
		return errors.New("username is required")
	}
	if input.StoringInput.ApiKey.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (v *vault) StoreApiKey(ctx context.Context, input Input) error {
	err := validateInput(input, false)
	if err != nil {
		return err
	}

	apiKey := input.StoringInput.ApiKey
	parameterName := getApiParameter(input)
	v.logger.Information("Storing api key {ApiKeyUserName} for capability {CapabilityId} at location {ParameterName}", apiKey.Username, string(input.CapabilityId), parameterName)

	client := ssm.NewFromConfig(v.config)

	v.logger.Trace("Sending request to AWS Parameter Store")
	_, err = client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(parameterName),
		Value: aws.String(`{ "key": "` + apiKey.Username + `", "secret": "` + apiKey.Password + `" }`),
		Tags: []types.Tag{
			{
				Key:   aws.String("capabilityId"),
				Value: aws.String(string(input.CapabilityId)),
			},
			{
				Key:   aws.String("createdBy"),
				Value: aws.String("Kafka-Janitor"),
			},
		},
		Tier:      types.ParameterTierStandard,
		Type:      types.ParameterTypeSecureString,
		Overwrite: &input.StoringInput.Overwrite,
	})

	if err != nil {
		return fmt.Errorf("error when storing api key %s for capability %s at location %s", apiKey.Username, string(input.CapabilityId), parameterName)
	}

	v.logger.Information("Successfully stored api key {ApiKeyUserName} for capability {CapabilityId} at location {ParameterName}", apiKey.Username, string(input.CapabilityId), parameterName)

	return nil
}

func (v *vault) QueryApiKey(ctx context.Context, input Input) (bool, error) {
	err := validateInput(input, false)
	if err != nil {
		return false, err
	}

	parameterName := getApiParameter(input)
	v.logger.Trace("Querying existence of API key for capability {CapabilityId} in cluster {ClusterId} at location {ParameterName}", string(input.CapabilityId), string(input.ClusterId), parameterName)

	client := ssm.NewFromConfig(v.config)

	v.logger.Trace("Sending request to AWS Parameter Store")

	_, err = client.GetParameter(ctx, &ssm.GetParameterInput{
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

func (v *vault) DeleteApiKey(ctx context.Context, input Input) error {
	err := validateInput(input, false)
	if err != nil {
		return err
	}

	parameterName := getApiParameter(input)
	v.logger.Trace("Deleting API key for capability {CapabilityId} in cluster {ClusterId} at location {ParameterName}", string(input.CapabilityId), string(input.ClusterId), parameterName)

	client := ssm.NewFromConfig(v.config)

	v.logger.Trace("Sending request to AWS Parameter Store")

	_, err = client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(parameterName),
	})
	if err != nil {
		var pnf *types.ParameterNotFound
		if errors.As(err, &pnf) {
			return nil
		}
		return err
	}

	return nil
}

func NewDefaultConfig() (*aws.Config, error) {
	awsConfig, err := config.LoadDefaultConfig(context.Background())
	return &awsConfig, err
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

func NewVaultClient(logger logging.Logger, cfg *aws.Config) (Vault, error) {
	if cfg == nil {
		return nil, errors.New("cannot create a valid vault client with a nil config")
	}

	return &vault{
		logger: logger,
		config: *cfg,
	}, nil
}
