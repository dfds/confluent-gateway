package confluent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"io"
	"net/http"
	"strconv"
	"time"
)

type CloudApiAccess struct {
	ApiEndpoint     string
	Username        string
	Password        string
	UserApiEndpoint string
}

func (a *CloudApiAccess) ApiKey() models.ApiKey {
	return models.ApiKey{Username: a.Username, Password: a.Password}
}

type Clusters interface {
	Get(clusterId models.ClusterId) (*models.Cluster, error)
}

type Client struct {
	logger         logging.Logger
	cloudApiAccess CloudApiAccess
	clusters       Clusters
}

func NewClient(logger logging.Logger, cloudApiAccess CloudApiAccess, repo Clusters) *Client {
	return &Client{logger: logger, cloudApiAccess: cloudApiAccess, clusters: repo}
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

type usersResponse struct {
	Users    []models.User `json:"users"`
	PageInfo struct {
		PageSize  int    `json:"page_size"`
		PageToken string `json:"page_token"`
	} `json:"page_info"`
}

func (c *Client) CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error) {
	url := c.cloudApiAccess.ApiEndpoint + "/iam/v2/service-accounts"
	payload := `{
		"display_name": "` + name + `",
		"description": "` + description + `"
	}`

	response, err := c.post(ctx, url, payload, c.cloudApiAccess.ApiKey())
	defer response.Body.Close()

	if err != nil {
		return "", err
	}

	serviceAccountResponse := &createServiceAccountResponse{}
	derr := json.NewDecoder(response.Body).Decode(serviceAccountResponse)
	if derr != nil {
		return "", derr
	}

	return models.ServiceAccountId(serviceAccountResponse.Id), nil
}

func (c *Client) post(ctx context.Context, url string, payload string, apiKey models.ApiKey) (*http.Response, error) {
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(payload)))
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(apiKey.Username, apiKey.Password)

	return c.getResponseReader(request, payload)
}

func (c *Client) getResponseReader(request *http.Request, payload string) (*http.Response, error) {
	url := request.URL.String()
	start := time.Now()
	response, err := http.DefaultClient.Do(request)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Error(err, "Closing response body failed")
		}
	}(response.Body)

	if err != nil {
		c.logger.Error(err, "{Method} {Url} failed", request.Method, url)
		return nil, err
	}

	elapsed := time.Since(start)

	c.logger.Trace("{Method} {Url}, Body: {Body}, StatusCode: {StatusCode}, Took: {Elapsed}",
		request.Method, url, payload, response.Status, elapsed.String())

	var buf bytes.Buffer
	tee := io.TeeReader(response.Body, &buf)
	content, _ := io.ReadAll(tee)
	response.Body = io.NopCloser(&buf)

	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		return response, nil
	}

	return response, fmt.Errorf("confluent client (%s) failed with status code %d: %s", url, response.StatusCode, content)
}

func (c *Client) CreateACLEntry(ctx context.Context, clusterId models.ClusterId, userAccountId models.UserAccountId, entry models.AclDefinition) error {
	cluster, err := c.clusters.Get(clusterId)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/acls", cluster.AdminApiEndpoint, clusterId)

	payload := `{
		"resource_type": "` + string(entry.ResourceType) + `",
		"resource_name": "` + string(entry.ResourceName) + `",
		"pattern_type": "` + string(entry.PatternType) + `",
		"principal": "` + string(userAccountId) + `",
		"host": "*",
		"operation": "` + string(entry.OperationType) + `",
		"permission": "` + string(entry.PermissionType) + `"
	}`

	response, err := c.post(ctx, url, payload, cluster.AdminApiKey)
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

	response, err := c.post(ctx, url, payload, c.cloudApiAccess.ApiKey())
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
	cluster, _ := c.clusters.Get(clusterId)
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

	response, err := c.post(ctx, url, payload, cluster.AdminApiKey)
	defer response.Body.Close()

	if err != nil {
		// log
	}

	return err
}

func (c *Client) GetUsers(ctx context.Context) ([]models.User, error) {
	url := c.cloudApiAccess.UserApiEndpoint

	response, err := c.get(ctx, url, c.cloudApiAccess.ApiKey())
	defer response.Body.Close()

	if err != nil {
		return nil, err
	}

	var users usersResponse
	if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users.Users, nil
}

func (c *Client) get(ctx context.Context, url string, apiKey models.ApiKey) (*http.Response, error) {
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	request.Header.Set("Accept", "application/json")
	request.SetBasicAuth(apiKey.Username, apiKey.Password)

	return c.getResponseReader(request, "")
}

func (c *Client) DeleteTopic(ctx context.Context, clusterId models.ClusterId, topicName string) error {
	cluster, _ := c.clusters.Get(clusterId)
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics/%s", cluster.AdminApiEndpoint, clusterId, topicName)

	response, err := c.delete(ctx, url, cluster.AdminApiKey)
	defer response.Body.Close()

	return err
}

func (c *Client) delete(ctx context.Context, url string, apiKey models.ApiKey) (*http.Response, error) {
	request, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	request.Header.Set("Accept", "application/json")
	request.SetBasicAuth(apiKey.Username, apiKey.Password)

	return c.getResponseReader(request, "")
}

type schemaPayload struct {
	SchemaType string `json:"schemaType"`
	Schema     string `json:"schema"`
}

var ErrNoSchemaRegistry = errors.New("no schema registry")

func (c *Client) RegisterSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string) error {
	cluster, _ := c.clusters.Get(clusterId)

	if len(cluster.SchemaRegistryApiEndpoint) == 0 {
		return ErrNoSchemaRegistry
	}

	url := fmt.Sprintf("%s/subjects/%s/versions", cluster.SchemaRegistryApiEndpoint, subject)

	payload, err := json.Marshal(schemaPayload{
		SchemaType: "JSON",
		Schema:     schema,
	})
	if err != nil {
		return err
	}

	// "Content-Type: application/vnd.schemaregistry.v1+json"

	response, err := c.post(ctx, url, string(payload), cluster.SchemaRegistryApiKey)
	defer response.Body.Close()

	if err != nil {
		// log
	}

	return err
}
