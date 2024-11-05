package confluent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
)

type CloudApiAccess struct {
	ApiEndpoint     string
	Username        string
	Password        string
	UserApiEndpoint string
}

var ErrSchemaRegistryIdIsEmpty = errors.New("schema registry id is not found, manually add id to cluster table")
var ErrMissingSchemaRegistryIds = errors.New("unable to create schema registry role binding: cluster table has any or all missing ids: organization_id, environment_id, schema_registry_id")
var ErrApiKeyNotFoundForDeletion = errors.New("unable to delete api key: key not found in confluent")
var ErrFoundExistingServiceAccount = errors.New("unable to create service account, service name already in use")
var ErrNoServiceAccountFound = errors.New("unable to find requested service account")

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

type ConfluentClient interface {
	ListSchemas(ctx context.Context, subjectPrefix string, clusterId models.ClusterId) ([]models.Schema, error)
	CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error)
	GetServiceAccount(ctx context.Context, displayName string) (models.ServiceAccountId, error)
	CreateACLEntry(ctx context.Context, clusterId models.ClusterId, userAccountId models.UserAccountId, entry models.AclDefinition) error
	CreateClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error)
	CreateSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error)
	DeleteClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error
	DeleteSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error
	CreateServiceAccountRoleBinding(ctx context.Context, serviceAccount models.ServiceAccountId, clusterId models.ClusterId) error
	CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int64) error
	DeleteTopic(ctx context.Context, clusterId models.ClusterId, topicName string) error
	GetConfluentInternalUsers(ctx context.Context) ([]models.ConfluentInternalUser, error)
	RegisterSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version int32) error
	DeleteSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version string) error
	CountClusterApiKeys(ctx context.Context, serviceAccountId models.ServiceAccountId, clusterId models.ClusterId) (int, error)
	CountSchemaRegistryApiKeys(ctx context.Context, serviceAccountId models.ServiceAccountId, clusterId models.ClusterId) (int, error)
}

func NewClient(logger logging.Logger, cloudApiAccess CloudApiAccess, repo Clusters) ConfluentClient {
	return &Client{logger: logger, cloudApiAccess: cloudApiAccess, clusters: repo}
}

type createServiceAccountResponse struct {
	Id string `json:"id"`
}

type listServiceAccountsResponse struct {
	Data []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
	} `json:"data"`
}

type createApiKeyResponse struct {
	Id   string `json:"id"`
	Spec struct {
		Secret string `json:"secret"`
	} `json:"spec"`
}

type usersResponse struct {
	Users    []models.ConfluentInternalUser `json:"users"`
	PageInfo struct {
		PageSize  int    `json:"page_size"`
		PageToken string `json:"page_token"`
	} `json:"page_info"`
}

type listApiKeysResponse struct {
	APIVersion string `json:"api_version"`
	Kind       string `json:"kind"`
	Metadata   struct {
		First     string `json:"first"`
		Last      string `json:"last"`
		Prev      string `json:"prev"`
		Next      string `json:"next"`
		TotalSize int    `json:"total_size"`
	} `json:"metadata"`
	Data []struct {
		APIVersion string `json:"api_version"`
		Kind       string `json:"kind"`
		ID         string `json:"id"`
		Metadata   struct {
			Self         string `json:"self"`
			ResourceName string `json:"resource_name"`
			CreatedAt    string `json:"created_at"`
			UpdatedAt    string `json:"updated_at"`
			DeletedAt    string `json:"deleted_at"`
		} `json:"metadata"`
		Spec struct {
			Secret      string `json:"secret"`
			DisplayName string `json:"display_name"`
			Description string `json:"description"`
			Owner       struct {
				ID           string `json:"id"`
				Related      string `json:"related"`
				ResourceName string `json:"resource_name"`
				APIVersion   string `json:"api_version"`
				Kind         string `json:"kind"`
			} `json:"owner"`
			Resource struct {
				ID           string `json:"id"`
				Environment  string `json:"environment"`
				Related      string `json:"related"`
				ResourceName string `json:"resource_name"`
				APIVersion   string `json:"api_version"`
				Kind         string `json:"kind"`
			} `json:"resource"`
		} `json:"spec"`
	} `json:"data"`
}

type createRoleBindingResponse struct {
	Id string `json:"id"`
}

func (c *Client) ListSchemas(ctx context.Context, subjectPrefix string, clusterId models.ClusterId) ([]models.Schema, error) {
	cluster, err := c.clusters.Get(clusterId)

	if err != nil {
		return nil, err
	}

	if len(cluster.SchemaRegistryApiEndpoint) == 0 {
		return nil, ErrNoSchemaRegistry
	}

	url := fmt.Sprintf("%s/clusters/%s/schemas", cluster.SchemaRegistryApiEndpoint, clusterId)

	/*
		ignore subjectPrefix for now
		if subjectPrefix != "" {
		  url += "?subjectPrefix=" + subjectPrefix
		}
	*/

	response, err := c.get(ctx, url, cluster.SchemaRegistryApiKey)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch schemas: " + response.Status)
	}

	var schemas []models.Schema
	if err := json.NewDecoder(response.Body).Decode(&schemas); err != nil {
		return nil, err
	}

	return schemas, nil
}

func (c *Client) CreateServiceAccount(ctx context.Context, name string, description string) (models.ServiceAccountId, error) {
	url := c.cloudApiAccess.ApiEndpoint + "/iam/v2/service-accounts"
	payload := `{
		"display_name": "` + name + `",
		"description": "` + description + `"
	}`

	response, err := c.post(ctx, url, payload, c.cloudApiAccess.ApiKey())
	if response != nil && response.StatusCode == 409 {
		return "", ErrFoundExistingServiceAccount
	}
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	serviceAccountResponse := &createServiceAccountResponse{}
	derr := json.NewDecoder(response.Body).Decode(serviceAccountResponse)
	if derr != nil {
		return "", derr
	}

	return models.ServiceAccountId(serviceAccountResponse.Id), nil
}

func (c *Client) GetServiceAccount(ctx context.Context, displayName string) (models.ServiceAccountId, error) {
	url := c.cloudApiAccess.ApiEndpoint + "/iam/v2/service-accounts"

	response, err := c.get(ctx, url, c.cloudApiAccess.ApiKey())
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if err != nil {
		return "", err
	}

	serviceAccountResponse := &listServiceAccountsResponse{}
	decodeErr := json.NewDecoder(response.Body).Decode(serviceAccountResponse)
	if decodeErr != nil {
		return "", decodeErr
	}

	for _, accountData := range serviceAccountResponse.Data {
		if accountData.DisplayName == displayName {
			return models.ServiceAccountId(accountData.ID), nil
		}
	}

	return "", ErrNoServiceAccountFound
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
	if err != nil {
		c.logger.Error(err, "{Method} {Url} failed", request.Method, url)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Error(err, "Closing response body failed")
		}
	}(response.Body)

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

	return response, NewClientError(url, response.StatusCode, string(content))
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
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

func (c *Client) getSchemaRegistryId(clusterId models.ClusterId) (models.SchemaRegistryId, error) {
	cluster, err := c.clusters.Get(clusterId)
	if err != nil {
		return "", err
	}

	if cluster.SchemaRegistryId == "" {
		return "", ErrSchemaRegistryIdIsEmpty
	}
	return cluster.SchemaRegistryId, nil
}

func (c *Client) listApiKeys(ctx context.Context, serviceAccountId models.ServiceAccountId, resourceId string) (listApiKeysResponse, error) {
	url := fmt.Sprintf("%s/iam/v2/api-keys?spec.owner=%s&spec.resource=%s", c.cloudApiAccess.ApiEndpoint, string(serviceAccountId), resourceId)
	response, err := c.get(ctx, url, c.cloudApiAccess.ApiKey())
	if err != nil {
		return listApiKeysResponse{}, err
	}
	defer response.Body.Close()

	var resp listApiKeysResponse
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return listApiKeysResponse{}, err
	}

	return resp, nil

}
func (c *Client) deleteApiKey(ctx context.Context, apiKeyId string) error {
	url := fmt.Sprintf("%s/iam/v2/api-keys/%s", c.cloudApiAccess.ApiEndpoint, apiKeyId)
	_, err := c.delete(ctx, url, c.cloudApiAccess.ApiKey())
	if err != nil {
		return err
	}

	return nil

}

func (c *Client) CountClusterApiKeys(ctx context.Context, serviceAccountId models.ServiceAccountId, clusterId models.ClusterId) (int, error) {
	keysResponse, err := c.listApiKeys(ctx, serviceAccountId, string(clusterId))
	if err != nil {
		return 0, err
	}
	return keysResponse.Metadata.TotalSize, nil
}

func (c *Client) CountSchemaRegistryApiKeys(ctx context.Context, serviceAccountId models.ServiceAccountId, clusterId models.ClusterId) (int, error) {
	schemaRegistryId, err := c.getSchemaRegistryId(clusterId)
	if err != nil {
		return 0, err
	}
	keysResponse, err := c.listApiKeys(ctx, serviceAccountId, string(schemaRegistryId))
	if err != nil {
		return 0, err
	}
	return keysResponse.Metadata.TotalSize, nil
}

func (c *Client) createApiKey(ctx context.Context, resourceId string, serviceAccountId models.ServiceAccountId) (models.ApiKey, error) {
	url := c.cloudApiAccess.ApiEndpoint + "/iam/v2/api-keys"
	payload := `{
		"spec": {
			"display_name": "` + fmt.Sprintf("%s-%s", resourceId, serviceAccountId) + `",
			"description": "Created with Confluent Gateway",
			"owner": {
				"id": "` + string(serviceAccountId) + `"
			},
			"resource": {
				"id": "` + resourceId + `"
			}
		}
	}`

	response, err := c.post(ctx, url, payload, c.cloudApiAccess.ApiKey())
	defer response.Body.Close()

	if err != nil {
		//log -> response.Status
		return models.ApiKey{}, err
	}

	apiKeyResponse := &createApiKeyResponse{}
	derr := json.NewDecoder(response.Body).Decode(apiKeyResponse)
	if derr != nil {
		return models.ApiKey{}, derr
	}

	return models.ApiKey{
		Username: apiKeyResponse.Id,
		Password: apiKeyResponse.Spec.Secret,
	}, nil
}

func (c *Client) CreateClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error) {
	return c.createApiKey(ctx, string(clusterId), serviceAccountId)
}

func (c *Client) CreateSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) (models.ApiKey, error) {

	schemaRegistryId, err := c.getSchemaRegistryId(clusterId)
	if err != nil {
		return models.ApiKey{}, err
	}
	return c.createApiKey(ctx, string(schemaRegistryId), serviceAccountId)
}

func (c *Client) findResourceAndDeleteApiKey(ctx context.Context, serviceAccountId models.ServiceAccountId, resourceId string) error {
	listApiKeys, err := c.listApiKeys(ctx, serviceAccountId, resourceId)
	if err != nil {
		return err
	}
	for _, datum := range listApiKeys.Data {
		if datum.Spec.Resource.ID != resourceId {
			continue
		}

		err := c.deleteApiKey(ctx, datum.ID)
		if err != nil {
			return err
		}
		return nil
	}
	return ErrApiKeyNotFoundForDeletion

}

func (c *Client) DeleteClusterApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error {
	return c.findResourceAndDeleteApiKey(ctx, serviceAccountId, string(clusterId))
}

func (c *Client) DeleteSchemaRegistryApiKey(ctx context.Context, clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error {

	schemaRegistryId, err := c.getSchemaRegistryId(clusterId)
	if err != nil {
		return err
	}
	resourceId := string(schemaRegistryId)
	return c.findResourceAndDeleteApiKey(ctx, serviceAccountId, resourceId)
}

func (c *Client) CreateServiceAccountRoleBinding(ctx context.Context, serviceAccount models.ServiceAccountId, clusterId models.ClusterId) error {

	cluster, err := c.clusters.Get(clusterId)
	if err != nil {
		return err
	}

	if cluster.OrganizationId == "" ||
		cluster.EnvironmentId == "" ||
		cluster.SchemaRegistryId == "" {
		return ErrMissingSchemaRegistryIds
	}

	url := c.cloudApiAccess.ApiEndpoint + "/iam/v2/role-bindings"
	payload := fmt.Sprintf(`{
		"principal": "User:%s",
		"role_name": "DeveloperRead",
		"crn_pattern": "crn://confluent.cloud/organization=%s/environment=%s/schema-registry=%s/subject=*"
	}`, serviceAccount, cluster.OrganizationId, cluster.EnvironmentId, cluster.SchemaRegistryId)

	response, err := c.post(ctx, url, payload, c.cloudApiAccess.ApiKey())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err != nil {
		return err
	}

	roleBindingResponse := &createRoleBindingResponse{}
	derr := json.NewDecoder(response.Body).Decode(roleBindingResponse)
	if derr != nil {
		return derr
	}

	return nil
}

func (c *Client) CreateTopic(ctx context.Context, clusterId models.ClusterId, name string, partitions int, retention int64) error {
	cluster, err := c.clusters.Get(clusterId)
	if err != nil {
		return err
	}
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

func (c *Client) GetConfluentInternalUsers(ctx context.Context) ([]models.ConfluentInternalUser, error) {
	// Note: this endpoint is not documented in the Confluent Cloud API docs
	url := c.cloudApiAccess.UserApiEndpoint

	response, err := c.get(ctx, url, c.cloudApiAccess.ApiKey())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var users usersResponse
	if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users.Users, nil
}

func (c *Client) get(ctx context.Context, url string, apiKey models.ApiKey) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.SetBasicAuth(apiKey.Username, apiKey.Password)

	return c.getResponseReader(request, "")
}

func (c *Client) DeleteTopic(ctx context.Context, clusterId models.ClusterId, topicName string) error {
	cluster, _ := c.clusters.Get(clusterId)
	url := fmt.Sprintf("%s/kafka/v3/clusters/%s/topics/%s", cluster.AdminApiEndpoint, clusterId, topicName)

	response, err := c.delete(ctx, url, cluster.AdminApiKey)
	if err != nil {
		return err
	}
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
	Version    int32  `json:"version"`
	SchemaType string `json:"schemaType"`
	Schema     string `json:"schema"`
}

func (c *Client) RegisterSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version int32) error {
	cluster, err := c.clusters.Get(clusterId)

	if err != nil {
		return err
	}

	if cluster == nil || len(cluster.SchemaRegistryApiEndpoint) == 0 {
		return ErrNoSchemaRegistry
	}

	url := fmt.Sprintf("%s/subjects/%s/versions", cluster.SchemaRegistryApiEndpoint, subject)

	payload, err := json.Marshal(schemaPayload{
		Version:    version,
		SchemaType: "JSON",
		Schema:     schema,
	})
	if err != nil {
		return err
	}

	// "Content-Type: application/vnd.schemaregistry.v1+json"

	response, err := c.post(ctx, url, string(payload), cluster.SchemaRegistryApiKey)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err != nil {
		// log
	}

	return err
}

func (c *Client) DeleteSchema(ctx context.Context, clusterId models.ClusterId, subject string, schema string, version string) error {

	cluster, err := c.clusters.Get(clusterId)

	if err != nil {
		return err
	}

	if len(cluster.SchemaRegistryApiEndpoint) == 0 {
		return ErrNoSchemaRegistry
	}

	url := fmt.Sprintf("%s/subjects/%s/versions/%s", cluster.SchemaRegistryApiEndpoint, subject, version)

	_, err = c.delete(ctx, url, cluster.SchemaRegistryApiKey)

	return err
}

type ClientError struct {
	Url     string
	Status  int
	Message string
}

func (mr *ClientError) Error() string {
	return fmt.Sprintf("confluent client (%s) failed with status code %d: %s", mr.Url, mr.Status, mr.Message)
}

func NewClientError(url string, status int, message string) error {
	return &ClientError{Url: url, Status: status, Message: message}
}

var ErrNoSchemaRegistry = errors.New("no schema registry")
