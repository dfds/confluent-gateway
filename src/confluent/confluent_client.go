package confluent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dfds/confluent-gateway/models"
	"net/http"
	"strconv"
)

type createTopicRequest struct {
	TopicName         string   `json:"topic_name"`
	PartitionCount    int      `json:"partition_count"`
	ReplicationFactor int      `json:"replication_factor"`
	Configs           []config `json:"configs"`
}

type config struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type client struct {
	cloudApiAccess    models.CloudApiAccess
	clusterRepository models.ClusterRepository
}

func (c *client) CreateServiceAccount(ctx context.Context, name string, description string) models.ServiceAccountId {
	//TODO implement me
	panic("implement me")
}

func (c *client) CreateACLEntry(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId, entry models.AclDefinition) {
	//TODO implement me
	panic("implement me")
}

func (c *client) CreateApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) models.ApiKey {
	//TODO implement me
	panic("implement me")
}

func (c *client) CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int) {
	cluster, _ := c.clusterRepository.Get(ctx, clusterId)
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics", cluster.AdminApiEndpoint, clusterId)

	payload, _ := json.Marshal(createTopicRequest{
		TopicName:         name,
		PartitionCount:    partitions,
		ReplicationFactor: 3,
		Configs: []config{{
			Name:  "retention.ms",
			Value: strconv.Itoa(retention),
		}},
	})

	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(cluster.AdminApiKey.Username, cluster.AdminApiKey.Password)

	response, err := http.DefaultClient.Do(request)
	defer response.Body.Close()

	if err != nil {
		panic(response.Status)
	}
}

func NewConfluentClient(cloudApiAccess models.CloudApiAccess, clusterRepository models.ClusterRepository) models.ConfluentClient {
	return &client{cloudApiAccess: cloudApiAccess, clusterRepository: clusterRepository}
}
