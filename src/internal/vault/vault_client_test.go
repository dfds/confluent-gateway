package vault

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVault_StoreApiKey_SendsExpectedPayload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	sentRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := io.ReadAll(r.Body)
		sentRequest = string(body)
	}))

	defer server.Close()

	config, _ := NewTestConfig(server.URL)
	sut := vault{
		logger: logging.NilLogger(),
		config: *config,
	}

	stubCapabilityId := models.CapabilityId("foo")
	stubClusterId := models.ClusterId("bar")
	stubApiKey := models.ApiKey{
		Username: "baz",
		Password: "qux",
	}
	input := Input{
		OperationDestination: OperationDestinationCluster,
		CapabilityId:         stubCapabilityId,
		ClusterId:            stubClusterId,
		StoringInput: &StoringInput{
			ApiKey:    stubApiKey,
			Overwrite: false,
		},
	}

	// act
	err := sut.StoreApiKey(ctx, input)

	// assert
	assert.Nil(t, err)
	assert.JSONEq(
		t,
		`{
			"Name": "/capabilities/`+string(stubCapabilityId)+`/kafka/`+string(stubClusterId)+`/credentials",
			"Tier": "Standard",
			"Type": "SecureString",
			"Value": "{ \"key\": \"`+stubApiKey.Username+`\", \"secret\": \"`+stubApiKey.Password+`\" }",
			"Tags": [
				{
					"Key": "capabilityId",
					"Value": "`+string(stubCapabilityId)+`"
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
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	config, _ := NewTestConfig(server.URL)
	sut := vault{
		logger: logging.NilLogger(),
		config: *config,
	}
	input := Input{
		OperationDestination: OperationDestinationCluster,
		CapabilityId:         models.CapabilityId("foo"),
		ClusterId:            models.ClusterId("bar"),
		StoringInput: &StoringInput{
			ApiKey: models.ApiKey{
				Username: "baz",
				Password: "qux",
			},
			Overwrite: false,
		},
	}

	// act
	err := sut.StoreApiKey(ctx, input)

	// assert
	assert.NotNil(t, err)
}
