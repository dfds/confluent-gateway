package main

import (
	"context"
	"github.com/dfds/confluent-gateway/functional_tests/helpers"
	"github.com/dfds/confluent-gateway/internal/serviceaccount"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
	"testing"
)

func setupCreateServiceAccountHttpMock(processInput serviceaccount.ProcessInput) {

	payload := string(`{
		"display_name": "` + processInput.CapabilityId + `",
		"description": "` + "Created by Confluent Gateway" + `"
	}`)

	gock.New(testerApp.config.ConfluentCloudApiUrl).
		Post("/iam/v2/service-accounts").
		BodyString(payload).
		BasicAuth(testerApp.config.ConfluentCloudApiUserName, testerApp.config.ConfluentCloudApiPassword).
		Reply(200).
		JSON(map[string]string{"id": "service_account_id_1234"})

}

func TestCreateServiceAccountProcess(t *testing.T) {
	createServiceAccountVariables := helpers.NewTestVariables("create_serviceaccount_test")
	defer func() {
		testerApp.db.RemoveAllOutboxEntries()
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
	setupCreateServiceAccountHttpMock(input)
	err = process.Process(context.Background(), input)
	require.NoError(t, err)
}
