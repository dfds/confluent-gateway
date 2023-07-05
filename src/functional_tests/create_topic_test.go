package main

import (
	"context"
	"encoding/json"
	"github.com/dfds/confluent-gateway/internal/create"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
	"time"
)

func TestCreateTopicProcess(t *testing.T) {

	outboxFactory, err := messaging.ConfigureOutbox(testerApp.logger,
		messaging.RegisterMessage(testerApp.config.TopicNameProvisioning, "topic_provisioned", &create.TopicProvisioned{}),
		messaging.RegisterMessage(testerApp.config.TopicNameProvisioning, "topic_provisioning_begun", &create.TopicProvisioningBegun{}),
	)
	require.NoError(t, err)

	process := create.NewProcess(testerApp.logger, testerApp.db, testerApp.confluentClient, func(repository create.OutboxRepository) create.Outbox {
		return outboxFactory(repository)
	})

	topic, err := testerApp.db.GetTopic(testTopicId)
	require.ErrorIs(t, err, storage.ErrTopicNotFound)

	topicDescription := models.TopicDescription{
		Name:       "topic-name-1234",
		Partitions: 1,
		Retention:  time.Hour * 24,
	}

	input := create.ProcessInput{
		TopicId:      testTopicId,
		CapabilityId: testCapabilityId,
		ClusterId:    testClusterId,
		Topic:        topicDescription,
	}

	// first we try without having a valid service account
	err = process.Process(context.Background(), input)
	require.ErrorIs(t, err, create.ErrMissingServiceAccount)

	// let's try again after getting a service account
	testerApp.AddMockServiceAccountWithClusterAccess()
	defer testerApp.RemoveMockServiceAccount()
	err = process.Process(context.Background(), input)
	require.NoError(t, err)

	topic, err = testerApp.db.GetTopic(testTopicId)
	require.NoError(t, err)

	require.Equal(t, topic.Id, testTopicId)
	require.Equal(t, topic.Name, topicDescription.Name)
	require.Equal(t, topic.Partitions, topicDescription.Partitions)
	require.Equal(t, topic.Retention, topicDescription.Retention.Milliseconds())

	entries, err := testerApp.db.GetAllOutboxEntries()
	require.NoError(t, err)
	// 3 messages: 1 from unsuccessful run  and 2 from a successful run
	require.Equal(t, 3, len(entries))
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].OccurredUtc.Before(entries[j].OccurredUtc)
	})
	require.Equal(t, testerApp.config.TopicNameProvisioning, entries[0].Topic)
	require.Equal(t, testerApp.config.TopicNameProvisioning, entries[1].Topic)
	require.Equal(t, testerApp.config.TopicNameProvisioning, entries[2].Topic)

	requirePayloadIsEqual(t, entries[0], "topic_provisioning_begun")
	requirePayloadIsEqual(t, entries[1], "topic_provisioning_begun")
	requirePayloadIsEqual(t, entries[2], "topic_provisioned")

}

func requirePayloadIsEqual(t *testing.T, outboxEntry *messaging.OutboxEntry, expectedType string) {

	type payload struct {
		Type string `json:"type"`
	}
	var data payload
	err := json.Unmarshal([]byte(outboxEntry.Payload), &data)
	require.NoError(t, err)

	require.Equal(t, expectedType, data.Type)
}
