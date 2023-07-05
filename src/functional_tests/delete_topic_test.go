package main

import (
	"context"
	"fmt"
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

func setupDeleteTopicHttpMock() {

	// we refer to topics in confluent by name for some reason
	gock.
		New(dbSeedAdminApiEndpoint).
		Delete(fmt.Sprintf("/kafka/v3/clusters/%s/topics/%s", testClusterId, nameOfDeletedTopic)).
		BasicAuth(dbSeedAdminUser, dbSeedAdminPassword).
		Reply(200).
		BodyString("") // our code panics on empty responses
}

func TestDeleteTopicProcess(t *testing.T) {

	err := testerApp.db.CreateTopic(&models.Topic{
		Id:           testTopicId,
		CapabilityId: testCapabilityId,
		ClusterId:    testClusterId,
		Name:         nameOfDeletedTopic,
		CreatedAt:    time.Now(),
	})
	require.NoError(t, err)

	topic, err := testerApp.db.GetTopic(testTopicId)
	require.NoError(t, err)
	require.Equal(t, topic.Id, testTopicId)

	outboxFactory, err := messaging.ConfigureOutbox(testerApp.logger,
		messaging.RegisterMessage(testerApp.config.TopicNameProvisioning, "topic-deleted", &delete.TopicDeleted{}),
	)
	require.NoError(t, err)
	process := delete.NewProcess(testerApp.logger, testerApp.db, testerApp.confluentClient, func(repository delete.OutboxRepository) delete.Outbox {
		return outboxFactory(repository)
	})
	input := delete.ProcessInput{
		TopicId: testTopicId,
	}
	setupDeleteTopicHttpMock()
	err = process.Process(context.Background(), input)
	require.NoError(t, err)

	topic, err = testerApp.db.GetTopic(testTopicId)
	require.ErrorIs(t, err, storage.ErrTopicNotFound)

	entries, err := testerApp.db.GetAllOutboxEntries()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, testerApp.config.TopicNameProvisioning, entries[0].Topic)

	requirePayloadIsEqual(t, entries[0], "topic-deleted")
}