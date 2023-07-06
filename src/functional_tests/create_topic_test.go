package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/create"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
	"sort"
	"strconv"
	"testing"
	"time"
)

func setupCreateTopicHttpMock(input create.ProcessInput) {

	payload := `{
		"topic_name": "` + input.Topic.Name + `",
		"partitions_count": ` + strconv.Itoa(input.Topic.Partitions) + `,
		"replication_factor": 3,
		"configs": [{
			"name": "retention.ms",
			"value": "` + strconv.FormatInt(input.Topic.Retention.Milliseconds(), 10) + `"
		}]
	}`
	gock.
		New(dbSeedAdminApiEndpoint).
		Post(fmt.Sprintf("/kafka/v3/clusters/%s/topics", input.ClusterId)).
		BasicAuth(dbSeedAdminUser, dbSeedAdminPassword).
		BodyString(payload).
		Reply(200).
		BodyString("") // our code panics on empty responses
}

func TestCreateTopicProcess(t *testing.T) {

	// cleanup function
	defer func() {
		testerApp.db.DeleteTopic(testTopicId)
		// TODO: outbox messages tied to this test instead of all
		testerApp.db.RemoveAllOutboxEntries()
		testerApp.RemoveMockServiceAccount()
	}()

	// sanity check
	topic, err := testerApp.db.GetTopic(testTopicId)
	require.ErrorIs(t, err, storage.ErrTopicNotFound)

	outboxFactory, err := messaging.ConfigureOutbox(testerApp.logger,
		messaging.RegisterMessage(testerApp.config.TopicNameProvisioning, "topic_provisioned", &create.TopicProvisioned{}),
		messaging.RegisterMessage(testerApp.config.TopicNameProvisioning, "topic_provisioning_begun", &create.TopicProvisioningBegun{}),
	)
	require.NoError(t, err)

	process := create.NewProcess(testerApp.logger, testerApp.db, testerApp.confluentClient, func(repository create.OutboxRepository) create.Outbox {
		return outboxFactory(repository)
	})
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

	setupCreateTopicHttpMock(input)

	// first we try without having a valid service account
	err = process.Process(context.Background(), input)
	require.ErrorIs(t, err, create.ErrMissingServiceAccount)

	// let's try again after getting a service account
	testerApp.AddMockServiceAccountWithClusterAccess()

	err = process.Process(context.Background(), input)
	require.NoError(t, err)

	topic, err = testerApp.db.GetTopic(testTopicId)
	require.NoError(t, err)

	require.Equal(t, topic.Id, testTopicId)
	require.Equal(t, topic.Name, topicDescription.Name)
	require.Equal(t, topic.Partitions, topicDescription.Partitions)
	require.Equal(t, topic.Retention, topicDescription.Retention.Milliseconds())

	createProcess, err := testerApp.db.GetCreateProcessState(testCapabilityId, testClusterId, topic.Name)
	require.NoError(t, err)
	require.NotNil(t, createProcess.CompletedAt)

	entries, err := testerApp.db.GetAllOutboxEntries()
	require.NoError(t, err)

	//I get 3 entries locally but 2 when run on azure pipelines....

	//3 messages: 1 from unsuccessful run  and 2 from a successful run
	require.Equal(t, 3, len(entries))
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].OccurredUtc.Before(entries[j].OccurredUtc)
	})
	require.Equal(t, testerApp.config.TopicNameProvisioning, entries[0].Topic)
	require.Equal(t, testerApp.config.TopicNameProvisioning, entries[1].Topic)
	require.Equal(t, testerApp.config.TopicNameProvisioning, entries[2].Topic)

	requireOutboxPayloadIsEqual(t, entries[0], "topic_provisioning_begun")
	requireOutboxPayloadIsEqual(t, entries[1], "topic_provisioning_begun")
	requireOutboxPayloadIsEqual(t, entries[2], "topic_provisioned")

}
