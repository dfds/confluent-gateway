package main

import (
	"context"
	"github.com/dfds/confluent-gateway/configuration"
	"github.com/dfds/confluent-gateway/functional_tests/mocks"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/logging"
)

var (
	testTopicId      = "test-topic-id"
	testCapabilityId = models.CapabilityId("test-capability-id")
	testClusterId    = models.ClusterId(string("test-cluster-id"))
)

type TesterApp struct {
	logger          logging.Logger
	config          configuration.Configuration
	db              *storage.Database
	confluentClient *confluent.Client
	vaultClient     *vault.Vault
}

func CreateAndSetupTester(logger logging.Logger) (*TesterApp, error) {
	config := configuration.LoadInto(&configuration.Configuration{})

	db, err := storage.NewDatabase(config.DbConnectionString, logger)
	if err != nil {
		return nil, err
	}

	clusters, err := db.GetClusters(context.Background())
	if err != nil {
		return nil, err
	}
	confluentClient := confluent.NewClient(logger, config.CreateCloudApiAccess(), storage.NewClusterCache(clusters))

	mock := mocks.NewVaultMock()

	return &TesterApp{db: db, confluentClient: confluentClient, vaultClient: &mock}, nil
}
