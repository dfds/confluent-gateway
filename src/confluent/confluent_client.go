package confluent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/models"
	"net/http"
	"strconv"
	"time"
)

type CloudApiAccess struct {
	ApiEndpoint string
	Username    string
	Password    string
}

func (a *CloudApiAccess) ApiKey() models.ApiKey {
	return models.ApiKey{Username: a.Username, Password: a.Password}
}

type ClusterRepository interface {
	GetClusterById(ctx context.Context, id models.ClusterId) (*models.Cluster, error)
}

type Client struct {
	logger         logging.Logger
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

func (c *Client) CreateServiceAccount(ctx context.Context, name string, description string) (*models.ServiceAccountId, error) {
	url := c.cloudApiAccess.ApiEndpoint + "/iam/v2/service-accounts"
	payload := `{
		"display_name": "` + name + `",
		"description": "` + description + `"
	}`

	response, err := c.post(url, payload, c.cloudApiAccess.ApiKey())
	defer response.Body.Close()

	if err != nil {
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

func (c *Client) post(url string, payload string, apiKey models.ApiKey) (*http.Response, error) {
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(apiKey.Username, apiKey.Password)

	start := time.Now()

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		c.logger.Error(&err, "POST {Url} failed", url)
		return nil, err
	}

	elapsed := time.Since(start)

	c.logger.Trace("POST {Url}, Body: {Body}, StatusCode: {StatusCode}, Took: {Elapsed}", url, payload, response.Status, elapsed.String())

	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		return response, nil
	}

	return response, fmt.Errorf("confluent client (%s) failed with status code %d", url, response.StatusCode)
}

func (c *Client) CreateACLEntry(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId, entry models.AclDefinition) error {
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

	response, err := c.post(url, payload, cluster.AdminApiKey)
	defer response.Body.Close()

	if err != nil {
		// log
	}

	return err
}

func (c *Client) CreateApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (*models.ApiKey, error) {
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

	response, err := c.post(url, payload, c.cloudApiAccess.ApiKey())
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

func (c *Client) CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int64) error {
	cluster, _ := c.repo.GetClusterById(ctx, clusterId)
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics", cluster.AdminApiEndpoint, clusterId)

	payload := `{
		"topic_name": "` + name + `",
		"partitions_count": ` + strconv.Itoa(partitions) + `,
		"replication_factor": 3,
		"configs": [{
			"name": "retention.ms",
			"value": "` + strconv.FormatInt(retention, 10) + `"
		}]
	}`

	response, err := c.post(url, payload, cluster.AdminApiKey)
	defer response.Body.Close()

	if err != nil {
		// log
	}

	return err
}

func NewConfluentClient(logger logging.Logger, cloudApiAccess CloudApiAccess, repo ClusterRepository) *Client {
	return &Client{logger: logger, cloudApiAccess: cloudApiAccess, repo: repo}
}
