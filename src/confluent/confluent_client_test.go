package confluent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
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
					Username: "dummy-username",
					Password: "dummy-password",
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
			if expectedRelativeUrl != usedEndpointUrl {
				t.Errorf("Unexpected cluster admin api endpoint of %s used", usedEndpointUrl)
			}
		})
	}
}

func TestCreateTopicCreatesTopicWithExpectedName(t *testing.T) {
	tests := []string{"foo", "bar", "baz", "qux"}

	for _, expectedTopicName := range tests {
		t.Run(expectedTopicName, func(t *testing.T) {
			sentRequest := &createTopicRequest{}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, sentRequest)
			}))
			defer server.Close()

			stubCluster := models.Cluster{
				ClusterId:        models.ClusterId("dummy"),
				Name:             "dummy",
				AdminApiEndpoint: server.URL,
				AdminApiKey: models.ApiKey{
					Username: "dummy-username",
					Password: "dummy-password",
				},
				BootstrapEndpoint: "dummy",
			}

			stubClient := client{
				cloudApiAccess:    models.CloudApiAccess{},
				clusterRepository: stubClusterRepository{cluster: stubCluster},
			}

			// act
			stubClient.CreateTopic(context.TODO(), stubCluster.ClusterId, expectedTopicName, 1, 1)

			// assert
			if expectedTopicName != sentRequest.TopicName {
				t.Errorf("Unexpected topic name of %s sent", sentRequest.TopicName)
			}
		})
	}
}

func TestCreateTopicCreatesTopicWithExpectedPartitionCount(t *testing.T) {
	tests := []int{1, 2, 3, 4}

	for _, expectedPartitionCount := range tests {
		t.Run(strconv.Itoa(expectedPartitionCount), func(t *testing.T) {
			sentRequest := &createTopicRequest{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, sentRequest)
			}))

			defer server.Close()

			stubCluster := models.Cluster{
				ClusterId:        models.ClusterId("dummy"),
				Name:             "dummy",
				AdminApiEndpoint: server.URL,
				AdminApiKey: models.ApiKey{
					Username: "dummy-username",
					Password: "dummy-password",
				},
				BootstrapEndpoint: "dummy",
			}

			stubClient := client{
				cloudApiAccess:    models.CloudApiAccess{},
				clusterRepository: stubClusterRepository{cluster: stubCluster},
			}

			// act
			stubClient.CreateTopic(context.TODO(), stubCluster.ClusterId, "dummy", expectedPartitionCount, 1)

			// assert
			if expectedPartitionCount != sentRequest.PartitionCount {
				t.Errorf("Unexpected partition count of %v sent", sentRequest.PartitionCount)
			}
		})
	}
}

func TestCreateTopicCreatesTopicWithExpectedReplicationFactor(t *testing.T) {
	sentRequest := &createTopicRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, sentRequest)
	}))

	defer server.Close()

	stubCluster := models.Cluster{
		ClusterId:        models.ClusterId("dummy"),
		Name:             "dummy",
		AdminApiEndpoint: server.URL,
		AdminApiKey: models.ApiKey{
			Username: "dummy-username",
			Password: "dummy-password",
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
	if sentRequest.ReplicationFactor != 3 {
		t.Errorf("Unexpected replication factor of %v sent", sentRequest.ReplicationFactor)
	}
}

func TestCreateTopicCreatesTopicWithExpectedRetention(t *testing.T) {
	tests := []int{1, 2, 3, 4}

	for _, expectedRetention := range tests {
		t.Run(strconv.Itoa(expectedRetention), func(t *testing.T) {
			sentRequest := &createTopicRequest{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, sentRequest)
			}))

			defer server.Close()

			stubCluster := models.Cluster{
				ClusterId:        models.ClusterId("dummy"),
				Name:             "dummy",
				AdminApiEndpoint: server.URL,
				AdminApiKey: models.ApiKey{
					Username: "dummy-username",
					Password: "dummy-password",
				},
				BootstrapEndpoint: "dummy",
			}

			stubClient := client{
				cloudApiAccess:    models.CloudApiAccess{},
				clusterRepository: stubClusterRepository{cluster: stubCluster},
			}

			// act
			stubClient.CreateTopic(context.TODO(), stubCluster.ClusterId, "dummy", 1, expectedRetention)

			// assert
			expectedConfigs := []config{{
				Name:  "retention.ms",
				Value: strconv.Itoa(expectedRetention),
			}}

			if !reflect.DeepEqual(expectedConfigs, sentRequest.Configs) {
				t.Errorf("Unexpected retention of %v sent", sentRequest.Configs)
			}
		})
	}
}
