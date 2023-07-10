package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/functional_tests/helpers"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/serviceaccount"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
	"testing"
)

const internalConfluentUserId = 1

func setupCreateServiceAccountHttpMock(processInput serviceaccount.ProcessInput, serviceAccountId models.ServiceAccountId) {

	payload := string(`{
		"display_name": "` + processInput.CapabilityId + `",
		"description": "` + "Created by Confluent Gateway" + `"
	}`)

	gock.New(testerApp.config.ConfluentCloudApiUrl).
		Post("/iam/v2/service-accounts").
		BodyString(payload).
		BasicAuth(testerApp.config.ConfluentCloudApiUserName, testerApp.config.ConfluentCloudApiPassword).
		Reply(200).
		JSON(map[string]string{"id": fmt.Sprintf("%s", serviceAccountId)})

}

func setupGetInternalConfluentUsersHttpMock(serviceAccountId models.ServiceAccountId) {

	s := struct {
		Users []models.ConfluentInternalUser `json:"users"`
	}{
		Users: []models.ConfluentInternalUser{
			{
				Id:         internalConfluentUserId,
				ResourceID: string(serviceAccountId),
			},
		},
	}

	gock.New(testerApp.config.ConfluentCloudApiUrl).
		Get("/service_accounts").
		BasicAuth(testerApp.config.ConfluentCloudApiUserName, testerApp.config.ConfluentCloudApiPassword).
		Reply(200).
		JSON(s)

}

func setupCreateAclHttpMock(capabilityId models.CapabilityId, clusterId models.ClusterId, userAccountId models.UserAccountId) {
	definitions := models.CreateAclDefinitions(capabilityId)
	for _, definition := range definitions {
		payload := `{
			"resource_type": "` + string(definition.ResourceType) + `",
			"resource_name": "` + string(definition.ResourceName) + `",
			"pattern_type": "` + string(definition.PatternType) + `",
			"principal": "` + string(userAccountId) + `",
			"host": "*",
			"operation": "` + string(definition.OperationType) + `",
			"permission": "` + string(definition.PermissionType) + `"
		}`

		gock.New(testerApp.dbSeedVariables.AdminApiEndpoint).
			Post(fmt.Sprintf("/kafka/v3/clusters/%s/acls", clusterId)).
			BodyString(payload).
			BasicAuth(testerApp.dbSeedVariables.AdminUser, testerApp.dbSeedVariables.AdminPassword).
			Reply(200).
			BodyString("")

	}
}

func setupCreateApiKeyMock(resourceId string, serviceAccountId models.ServiceAccountId, createKeyId string, createKeySecret string) {
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

	r := struct {
		Id   string `json:"id"`
		Spec struct {
			Secret string `json:"secret"`
		} `json:"spec"`
	}{
		Id: createKeyId,
		Spec: struct {
			Secret string `json:"secret"`
		}{
			Secret: createKeySecret,
		},
	}

	gock.New(testerApp.config.ConfluentCloudApiUrl).
		Post(fmt.Sprintf("/iam/v2/api-keys")).
		BodyString(payload).
		BasicAuth(testerApp.config.ConfluentCloudApiUserName, testerApp.config.ConfluentCloudApiPassword).
		Reply(200).
		JSON(r)
}

func setupRoleBindingHTTPMock(serviceAccount string, cluster models.Cluster) {
	payload := fmt.Sprintf(`{
		"principal": "User:%s",
		"role_name": "DeveloperRead",
		"crn_pattern": "crn://confluent.cloud/organization=%s/environment=%s/schema-registry=%s/subject=*"
	}`, serviceAccount, cluster.OrganizationId, cluster.EnvironmentId, cluster.SchemaRegistryId)

	gock.New(testerApp.config.ConfluentCloudApiUrl).
		Post("/iam/v2/role-bindings").
		BodyString(payload).
		BasicAuth(testerApp.config.ConfluentCloudApiUserName, testerApp.config.ConfluentCloudApiPassword).
		Reply(200).
		JSON(map[string]string{"id": "not-used"})
}

func setupListKeysHTTPMock(resourceId string, serviceAccountId models.ServiceAccountId, totalSize int) {

	response := struct {
		Metadata struct {
			TotalSize int `json:"total_size"`
		} `json:"metadata"`
	}{
		Metadata: struct {
			TotalSize int `json:"total_size"`
		}{
			TotalSize: totalSize,
		},
	}
	//http://localhost:5051/iam/v2/api-keys?spec.owner=create_serviceaccount_test-service-account-id&spec.resource=def-5678
	gock.New(testerApp.config.ConfluentCloudApiUrl).
		MatchParam("spec.owner", string(serviceAccountId)).
		MatchParam("spec.resource", resourceId).
		Get("/iam/v2/api-keys").
		BasicAuth(testerApp.config.ConfluentCloudApiUserName, testerApp.config.ConfluentCloudApiPassword).
		Reply(200).
		JSON(response)
}

func TestCreateServiceAccountProcess(t *testing.T) {
	createServiceAccountVariables := helpers.NewTestVariables("create_serviceaccount_test")
	userAccountId := models.MakeUserAccountId(internalConfluentUserId)
	defer func() {
		testerApp.db.RemoveAllOutboxEntries()
		testerApp.db.RemoveAllServiceAccounts()
	}()
	outboxFactory, err := messaging.ConfigureOutbox(testerApp.logger,
		messaging.RegisterMessage(testerApp.config.TopicNameKafkaClusterAccessGranted, "cluster-access-granted", &serviceaccount.ServiceAccountAccessGranted{}),
	)
	require.NoError(t, err)

	process := serviceaccount.NewProcess(testerApp.logger,
		testerApp.db,
		testerApp.confluentClient,
		*testerApp.vaultClient,
		func(repository serviceaccount.OutboxRepository) serviceaccount.Outbox {
			return outboxFactory(repository)
		})

	input := serviceaccount.ProcessInput{
		CapabilityId: createServiceAccountVariables.CapabilityId,
		ClusterId:    testerApp.dbSeedVariables.DevelopmentClusterId,
	}

	createdClusterApiKey := models.ApiKey{
		Username: "new_cluster_apikey_username",
		Password: "new_cluster_apikey_password",
	}
	createdSchemaRegistryApiKey := models.ApiKey{
		Username: "new_schema_registry_apikey_username",
		Password: "new_schema_registry_apikey_password",
	}

	// a lot of mocking going on here, might indicate that this process is doing too much

	//ensureServiceAccountStep:
	setupCreateServiceAccountHttpMock(input, createServiceAccountVariables.ServiceAccountId) // First we create a service account
	setupGetInternalConfluentUsersHttpMock(createServiceAccountVariables.ServiceAccountId)   // Then we check if we have a matching user for that service account

	//ensureServiceAccountAclStep:
	setupCreateAclHttpMock(createServiceAccountVariables.CapabilityId, input.ClusterId, userAccountId) // Then we create ACLs for the service account

	//ensureServiceAccountApiKeyStep
	setupListKeysHTTPMock(string(input.ClusterId), createServiceAccountVariables.ServiceAccountId, 1) // Then we check if the API key was created

	//ensureServiceAccountApiKeyAreStoredInVaultStep
	setupCreateApiKeyMock(string(input.ClusterId), createServiceAccountVariables.ServiceAccountId, createdClusterApiKey.Username, createdClusterApiKey.Password) // Then we create an API key for the cluster

	//ensureServiceAccountHasSchemaRegistryAccessStep
	setupListKeysHTTPMock(string(testerApp.dbSeedVariables.DevelopmentSchemaRegistryId), createServiceAccountVariables.ServiceAccountId, 0)                                                                          // Check if the api key has already been created
	setupCreateApiKeyMock(string(testerApp.dbSeedVariables.DevelopmentSchemaRegistryId), createServiceAccountVariables.ServiceAccountId, createdSchemaRegistryApiKey.Username, createdSchemaRegistryApiKey.Password) // Then we create an API key for the schema registry
	setupRoleBindingHTTPMock(string(createServiceAccountVariables.ServiceAccountId), testerApp.dbSeedVariables.GetDevelopmentClusterValues())                                                                        // Then we create a role binding for the service account

	err = process.Process(context.Background(), input)
	require.NoError(t, err)

	// check outbox
	outboxEntries, err := testerApp.db.GetAllOutboxEntries()
	require.NoError(t, err)
	require.Len(t, outboxEntries, 1)
	require.Equal(t, outboxEntries[0].Topic, testerApp.config.TopicNameKafkaClusterAccessGranted)
}
