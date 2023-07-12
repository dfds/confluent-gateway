package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/functional_tests/helpers"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/h2non/gock"
	"testing"
	"time"

	delete "github.com/dfds/confluent-gateway/internal/delete"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/stretchr/testify/require"
)

const nameOfDeletedTopic = "my-cool-topic-name"

func setupDeleteTopicHttpMock(clusterId models.ClusterId, seedVariables *SeedVariables) {

	gock.
		New(seedVariables.AdminApiEndpoint).
		Delete(fmt.Sprintf("/kafka/v3/clusters/%s/topics/%s", clusterId, nameOfDeletedTopic)).
		BasicAuth(seedVariables.AdminUser, seedVariables.AdminPassword).
		Reply(200).
		BodyString("") // our code panics on empty responses
}

func TestDeleteTopicProcess(t *testing.T) {

	deleteTopicVariables := helpers.NewTestVariables("delete_topic_test")
	// cleanup function
	defer func() {
		testerApp.db.DeleteTopic(deleteTopicVariables.TopicId)
		// TODO: outbox messages tied to this test instead of all
		testerApp.db.RemoveAllOutboxEntries()
		testerApp.db.RemoveDeleteProcessesWithTopicId(deleteTopicVariables.TopicId)
	}()

	err := testerApp.db.CreateTopic(&models.Topic{
		Id:           deleteTopicVariables.TopicId,
		CapabilityId: deleteTopicVariables.CapabilityId,
		ClusterId:    testerApp.dbSeedVariables.DevelopmentClusterId,
		Name:         nameOfDeletedTopic,
		CreatedAt:    time.Now(),
	})
	require.NoError(t, err)

	topic, err := testerApp.db.GetTopic(deleteTopicVariables.TopicId)
	require.NoError(t, err)
	require.Equal(t, topic.Id, deleteTopicVariables.TopicId)

	outboxFactory, err := messaging.ConfigureOutbox(testerApp.logger,
		messaging.RegisterMessage(testerApp.config.TopicNameProvisioning, "topic-deleted", &delete.TopicDeleted{}),
	)
	require.NoError(t, err)
	process := delete.NewProcess(testerApp.logger, testerApp.db, testerApp.confluentClient, func(repository delete.OutboxRepository) delete.Outbox {
		return outboxFactory(repository)
	})
	input := delete.ProcessInput{
		TopicId: deleteTopicVariables.TopicId,
	}
	setupDeleteTopicHttpMock(testerApp.dbSeedVariables.DevelopmentClusterId, testerApp.dbSeedVariables)
	err = process.Process(context.Background(), input)
	require.NoError(t, err)

	topic, err = testerApp.db.GetTopic(deleteTopicVariables.TopicId)
	require.ErrorIs(t, err, storage.ErrTopicNotFound)

	entries, err := testerApp.db.GetAllOutboxEntries()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, testerApp.config.TopicNameProvisioning, entries[0].Topic)

	helpers.RequireOutboxPayloadIsEqual(t, entries[0], "topic-deleted")

	helpers.RequireNoUnmatchedGockMocks(t)
}
