package vault

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/models"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVault_StoreApiKey_SendsExpectedPayload(t *testing.T) {
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)

	sentRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := io.ReadAll(r.Body)
		sentRequest = string(body)
	}))

	defer server.Close()

	cfg := aws.NewConfig()
	cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: server.URL,
		}, nil
	})
	cfg.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			CanExpire: false,
			Expires:   time.Time{},
		}, nil
	})

	sut := vault{
		logger: logging.NilLogger(),
		config: *cfg,
	}

	stubCapabilityRootId := models.CapabilityRootId("foo")
	stubClusterId := models.ClusterId("bar")
	stubApiKey := models.ApiKey{
		Username: "baz",
		Password: "qux",
	}

	// act
	err := sut.StoreApiKey(ctx, stubCapabilityRootId, stubClusterId, stubApiKey)

	// assert
	assert.Nil(t, err)
	assert.JSONEq(
		t,
		`{
			"Name": "/capabilities/`+string(stubCapabilityRootId)+`/kafka/`+string(stubClusterId)+`/credentials",
			"Tier": "Standard",
			"Type": "SecureString",
			"Value": "{ \"key\": \"`+stubApiKey.Username+`\", \"secret\": \"`+stubApiKey.Password+`\" }",
			"Tags": [
				{
					"Key": "capabilityRootId",
					"Value": "`+string(stubCapabilityRootId)+`"
				},
				{
					"Key": "createdBy",
					"Value": "Kafka-Janitor"
				}
			]
		}`,
		sentRequest,
	)
}

func TestVault_StoreApiKey_ReturnsErrorWhenServerDoes(t *testing.T) {
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	cfg := aws.NewConfig()
	cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: server.URL,
		}, nil
	})
	cfg.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			CanExpire: false,
			Expires:   time.Time{},
		}, nil
	})

	sut := vault{
		logger: logging.NilLogger(),
		config: *cfg,
	}

	// act
	err := sut.StoreApiKey(
		ctx,
		models.CapabilityRootId("foo"),
		models.ClusterId("bar"),
		models.ApiKey{
			Username: "baz",
			Password: "qux",
		},
	)

	// assert
	assert.NotNil(t, err)
}
