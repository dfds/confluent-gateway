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

type CloudApiAccess struct {
	ApiEndpoint string
	Username    string
	Password    string
}

type client struct {
	cloudApiAccess CloudApiAccess
	repo           ClusterRepository
}

type createServiceAccountResponse struct {
	Id string `json:"id"`
}

type createApiKeyResponse struct {
	Id   string `json:"id"`
	Spec struct {
		Secret string `json:"secret"`
	} `json:"spec"`
}

func (c *client) CreateServiceAccount(ctx context.Context, name string, description string) (*models.ServiceAccountId, error) {
	url := c.cloudApiAccess.ApiEndpoint + "/iam/v2/service-accounts"
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
		return nil, err
	}

	serviceAccountResponse := &createServiceAccountResponse{}
	derr := json.NewDecoder(response.Body).Decode(serviceAccountResponse)
	if derr != nil {
		return nil, derr
	}

	serviceAccountId := models.ServiceAccountId(serviceAccountResponse.Id)
	return &serviceAccountId, nil
}

func (c *client) CreateACLEntry(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId, entry models.AclDefinition) error {
	cluster, err := c.repo.GetClusterById(ctx, clusterId)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/acls", cluster.AdminApiEndpoint, clusterId)

	payload := `{
		"resource_type": "` + string(entry.ResourceType) + `",
		"resource_name": "` + string(entry.ResourceName) + `",
		"pattern_type": "` + string(entry.PatternType) + `",
		"principal": "User:` + string(serviceAccountId) + `",
		"host": "*",
		"operation": "` + string(entry.OperationType) + `",
		"permission": "` + string(entry.PermissionType) + `"
	}`

	request, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(cluster.AdminApiKey.Username, cluster.AdminApiKey.Password)

	response, err := http.DefaultClient.Do(request)
	defer response.Body.Close()

	if err != nil {
		// log
	}

	return err
}

func (c *client) CreateApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (*models.ApiKey, error) {
	url := c.cloudApiAccess.ApiEndpoint + "/iam/v2/api-keys"
	payload := `{
		"spec": {
			"display_name": "` + fmt.Sprintf("%s-%s", clusterId, serviceAccountId) + `",
			"description": "Created with Confluent Gateway",
			"owner": {
				"id": "` + string(serviceAccountId) + `"
			},
			"resource": {
				"id": "` + string(clusterId) + `"
			}
		}
	}`

	request, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(c.cloudApiAccess.Username, c.cloudApiAccess.Password)

	response, err := http.DefaultClient.Do(request)
	defer response.Body.Close()

	if err != nil {
		//log -> response.Status
		return nil, err
	}

	apiKeyResponse := &createApiKeyResponse{}
	derr := json.NewDecoder(response.Body).Decode(apiKeyResponse)
	if derr != nil {
		return nil, derr
	}

	return &models.ApiKey{
		Username: apiKeyResponse.Id,
		Password: apiKeyResponse.Spec.Secret,
	}, nil
}

func (c *client) CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int) error {
	cluster, _ := c.repo.GetClusterById(ctx, clusterId)
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
		// log
	}

	return err
}

type ClusterRepository interface {
	GetClusterById(ctx context.Context, id models.ClusterId) (models.Cluster, error)
}

func NewConfluentClient(cloudApiAccess CloudApiAccess, repo ClusterRepository) models.ConfluentClient {
	return &client{cloudApiAccess: cloudApiAccess, repo: repo}
}
