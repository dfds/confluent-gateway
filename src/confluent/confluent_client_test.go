package confluent

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubClusterRepository struct {
	cluster models.Cluster
}

func (stub stubClusterRepository) Get(ctx context.Context, id models.ClusterId) (models.Cluster, error) {
	return stub.cluster, nil
}

func (stub stubClusterRepository) GetAll(ctx context.Context) ([]models.Cluster, error) {
	return []models.Cluster{stub.cluster}, nil
}

func TestCreateTopicCallsExpectedClusterAdminEndpoint(t *testing.T) {
	tests := []string{"foo", "bar", "baz", "qux"}

	for _, stubClusterId := range tests {
		t.Run(stubClusterId, func(t *testing.T) {
			usedEndpointUrl := ""
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				usedEndpointUrl = r.RequestURI
			}))
			defer server.Close()

			stubCluster := models.Cluster{
				ClusterId:        models.ClusterId(stubClusterId),
				Name:             "dummy",
				AdminApiEndpoint: server.URL,
				AdminApiKey: models.ApiKey{
					Username: "dummy",
					Password: "dummy",
				},
				BootstrapEndpoint: "dummy",
			}

			stubClient := client{
				cloudApiAccess:    models.CloudApiAccess{},
				clusterRepository: stubClusterRepository{cluster: stubCluster},
			}

			// act
			stubClient.CreateTopic(context.TODO(), stubCluster.ClusterId, "dummy", 1, 1)

			// assert
			expectedRelativeUrl := fmt.Sprintf("/kafka/v3/clusters/%s/topics", stubClusterId)

			assert.Equal(t, expectedRelativeUrl, usedEndpointUrl)
		})
	}
}

func TestCreateTopicSendsExpectedPayload(t *testing.T) {
	sentRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		body, _ := io.ReadAll(r.Body)
		sentRequest = string(body)
	}))

	defer server.Close()

	stubCluster := models.Cluster{
		ClusterId:        models.ClusterId("dummy"),
		Name:             "dummy",
		AdminApiEndpoint: server.URL,
		AdminApiKey: models.ApiKey{
			Username: "dummy",
			Password: "dummy",
		},
		BootstrapEndpoint: "dummy",
	}

	stubClient := client{
		cloudApiAccess:    models.CloudApiAccess{},
		clusterRepository: stubClusterRepository{cluster: stubCluster},
	}

	// act
	stubClient.CreateTopic(context.TODO(), stubCluster.ClusterId, "foo-topic-name", 1, 2)

	// assert
	assert.JSONEq(
		t,
		`{
			"topic_name": "foo-topic-name",
			"partition_count": 1,
			"replication_factor": 3,
			"configs": [{
				"name": "retention.ms",
				"value": "2"
			}]
		}`,
		sentRequest,
	)
}

func TestCreateTopicUsesExpectedApiKey(t *testing.T) {
	expected := "Basic " + b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "foo", "bar")))

	usedApiKey := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		usedApiKey = r.Header.Get("Authorization")
	}))

	defer server.Close()

	stubCluster := models.Cluster{
		ClusterId:        models.ClusterId("dummy"),
		Name:             "dummy",
		AdminApiEndpoint: server.URL,
		AdminApiKey: models.ApiKey{
			Username: "foo",
			Password: "bar",
		},
		BootstrapEndpoint: "dummy",
	}

	stubClient := client{
		cloudApiAccess:    models.CloudApiAccess{},
		clusterRepository: stubClusterRepository{cluster: stubCluster},
	}

	// act
	stubClient.CreateTopic(context.TODO(), stubCluster.ClusterId, "dummy", 1, 1)

	// assert
	assert.Equal(t, expected, usedApiKey)
}

// ---------------------------------------------------------------------------------------------------------

func TestCreateServiceAccountCallsExpectedEndpoint(t *testing.T) {
	tests := []string{"foo", "bar", "baz", "qux"}

	for _, stubEndpoint := range tests {
		t.Run(stubEndpoint, func(t *testing.T) {
			usedEndpointUrl := ""
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				usedEndpointUrl = r.RequestURI
			}))
			defer server.Close()

			stubClient := client{
				cloudApiAccess: models.CloudApiAccess{
					ApiEndpoint: server.URL + "/" + stubEndpoint,
					Username:    "dummy",
					Password:    "dummy",
				},
				clusterRepository: stubClusterRepository{},
			}

			// act
			stubClient.CreateServiceAccount(context.TODO(), "dummy", "dummy")

			// assert
			expectedRelativeUrl := "/" + stubEndpoint
			assert.Equal(t, expectedRelativeUrl, usedEndpointUrl)
		})
	}
}

func TestCreateServiceAccountSendsExpectedPayload(t *testing.T) {
	sentPayload := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		body, _ := io.ReadAll(r.Body)
		sentPayload = string(body)
	}))

	defer server.Close()

	stubClient := client{
		cloudApiAccess: models.CloudApiAccess{
			ApiEndpoint: server.URL,
			Username:    "dummy",
			Password:    "dummy",
		},
		clusterRepository: stubClusterRepository{},
	}

	// act
	stubClient.CreateServiceAccount(context.TODO(), "foo", "bar")

	// assert
	assert.JSONEq(
		t,
		`{
			"display_name": "foo",
			"description": "bar"
		}`,
		sentPayload,
	)
}

func TestCreateServiceAccountUsesExpectedApiKey(t *testing.T) {
	expected := "Basic " + b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "foo", "bar")))

	usedApiKey := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		usedApiKey = r.Header.Get("Authorization")
	}))

	defer server.Close()

	stubClient := client{
		cloudApiAccess: models.CloudApiAccess{
			ApiEndpoint: server.URL,
			Username:    "foo",
			Password:    "bar",
		},
		clusterRepository: stubClusterRepository{},
	}

	// act
	stubClient.CreateServiceAccount(context.TODO(), "dummy", "dummy")

	// assert
	assert.Equal(t, expected, usedApiKey)
}

func TestCreateServiceAccountReturnsExpectedServiceAccountId(t *testing.T) {
	expected := models.ServiceAccountId("sa-123456")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{ "id": "` + expected + `" }`))
	}))

	defer server.Close()

	stubClient := client{
		cloudApiAccess: models.CloudApiAccess{
			ApiEndpoint: server.URL,
			Username:    "dummy",
			Password:    "dummy",
		},
		clusterRepository: stubClusterRepository{},
	}

	// act
	serviceAccountId, _ := stubClient.CreateServiceAccount(context.TODO(), "dummy", "dummy")

	// assert
	assert.Equal(t, expected, serviceAccountId)
}