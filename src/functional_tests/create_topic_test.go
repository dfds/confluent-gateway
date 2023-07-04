package main

import (
	"context"
	"github.com/dfds/confluent-gateway/internal/create"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreateTopic(t *testing.T) {

	outboxFactory, err := messaging.ConfigureOutbox(testerApp.logger,
		messaging.RegisterMessage(testerApp.config.TopicNameProvisioning, "topic_provisioned", &create.TopicProvisioned{}),
		messaging.RegisterMessage(testerApp.config.TopicNameProvisioning, "topic_provisioning_begun", &create.TopicProvisioningBegun{}),
	)
	if err != nil {
		return
	}
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

	err = process.Process(context.Background(), input)
	require.NoError(t, err)

	topic, err = testerApp.db.GetTopic(testTopicId)
	require.NoError(t, err)

	require.Equal(t, topic.Id, testTopicId)
	require.Equal(t, topic.Name, topicDescription.Name)
	require.Equal(t, topic.Partitions, topicDescription.Partitions)
	require.Equal(t, topic.Retention, topicDescription.Retention)
}