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

type client struct {
	cloudApiAccess    models.CloudApiAccess
	clusterRepository models.ClusterRepository
}

func (c *client) CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error) {
	url := c.cloudApiAccess.ApiEndpoint
	payload := `{
		"display_name": "` + name + `",
		"description": "` + description + `"
	}`

	request, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(c.cloudApiAccess.Username, c.cloudApiAccess.Password)

	response, err := http.DefaultClient.Do(request)
	defer response.Body.Close()

	if err != nil {
		//log -> response.Status
		return "", err
	}

	serviceAccountResponse := &createServiceAccountResponse{}
	derr := json.NewDecoder(response.Body).Decode(serviceAccountResponse)
	if derr != nil {
		return "", derr
	}

	return models.ServiceAccountId(serviceAccountResponse.Id), nil
}

type createServiceAccountResponse struct {
	Id string ` json:"id"`
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

	payload := `{
		"topic_name": "` + name + `",
		"partition_count": ` + strconv.Itoa(partitions) + `,
		"replication_factor": 3,
		"configs": [{
			"name": "retention.ms",
			"value": "` + strconv.Itoa(retention) + `"
		}]
	}`

	request, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
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
