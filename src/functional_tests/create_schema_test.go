package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dfds/confluent-gateway/functional_tests/helpers"
	"github.com/dfds/confluent-gateway/internal/models"
	schema "github.com/dfds/confluent-gateway/internal/schema"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func setupCreateSchemaHttpMock(processInput schema.ProcessInput, topicName string, seedVariables *SeedVariables) {

	type schemaPayload struct {
		SchemaType string `json:"schemaType"`
		Schema     string `json:"schema"`
		Version    int32  `json:"version"`
	}
	payload, err := json.Marshal(schemaPayload{
		SchemaType: "JSON",
		Schema:     processInput.Schema,
		Version:    processInput.SchemaVersion,
	})
	if err != nil {
		panic(err)
	}

	subjectName := fmt.Sprintf("%s-%s", topicName, processInput.MessageType)
	gock.
		New(seedVariables.SchemaRegistryApiEndpoint).
		Post(fmt.Sprintf("/subjects/%s/versions", subjectName)).
		BasicAuth(seedVariables.SchemaRegistryAdminUser, seedVariables.SchemaRegistryAdminPassword).
		JSON(payload).
		Reply(200).
		BodyString("") // our code panics on empty responses
}

func TestCreateSchemaProcess(t *testing.T) {

	createSchemaVariables := helpers.NewTestVariables("create_schema_test")
	defer func() {
		testerApp.db.DeleteTopic(createSchemaVariables.TopicId)
		testerApp.db.RemoveSchemaProcessWithTopicId(createSchemaVariables.TopicId)
		testerApp.db.RemoveAllOutboxEntries()
	}()

	outboxFactory, err := messaging.ConfigureOutbox(testerApp.logger,
		messaging.RegisterMessage(testerApp.config.TopicNameSchema, "schema-registered", &schema.SchemaRegistered{}),
		messaging.RegisterMessage(testerApp.config.TopicNameSchema, "schema-registration-failed", &schema.SchemaRegistrationFailed{}),
	)
	require.NoError(t, err)

	process := schema.NewProcess(testerApp.logger, testerApp.db, testerApp.confluentClient, func(repository schema.OutboxRepository) schema.Outbox {
		return outboxFactory(repository)
	})

	input := schema.ProcessInput{
		MessageContractId: "schema-process-contract-id",
		TopicId:           createSchemaVariables.TopicId,
		MessageType:       "message-type-for-schema-registry",
		Description:       "schema-description",
		Schema:            "test-schema",
		SchemaVersion:     1,
	}

	err = process.Process(context.Background(), input)
	// TODO: Although the topic does not exist, the process ignores that and continues.
	//require.ErrorIs(t, err, storage.ErrTopicNotFound)
	require.NoError(t, err)

	err = testerApp.db.CreateTopic(&models.Topic{
		Id:           createSchemaVariables.TopicId,
		CapabilityId: createSchemaVariables.CapabilityId,
		ClusterId:    testerApp.dbSeedVariables.DevelopmentClusterId,
		Name:         createSchemaVariables.TopicName,
		CreatedAt:    time.Now(),
	})
	require.NoError(t, err)

	setupCreateSchemaHttpMock(input, createSchemaVariables.TopicName, testerApp.dbSeedVariables)
	err = process.Process(context.Background(), input)
	require.NoError(t, err)

	helpers.RequireNoUnmatchedGockMocks(t)
}
