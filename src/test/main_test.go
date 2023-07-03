package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/configuration"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/test/mocks"
	"os"
	"testing"
)

var (
	testTopicId      = "test-topic-id"
	testCapabilityId = models.CapabilityId("test-capability-id")
	testClusterId    = models.ClusterId(string("test-cluster-id"))
)

type testerApp struct {
	logger          logging.Logger
	config          configuration.Configuration
	db              *storage.Database
	confluentClient *confluent.Client
	vaultClient     *vault.Vault
}

func newTesterApp(db *storage.Database, confluentClient *confluent.Client, vaultClient *vault.Vault) *testerApp {
	return &testerApp{db: db, confluentClient: confluentClient, vaultClient: vaultClient}
}

var tester *testerApp

func TestMain(m *testing.M) {

	logger := logging.NewLogger(logging.LoggerOptions{
		IsProduction: false,
		AppName:      "Confluent Gateway Test App",
	})

	logger.Information("Setting up tester app")
	err := setupTester(logger)
	if err != nil {
		panic(fmt.Errorf("TesterApp setup error: %s", err))
	}

	logger.Information("Starting tests")
	testRunCode := m.Run()

	logger.Information(fmt.Sprintf("Finished tests with code %d", testRunCode))
	os.Exit(testRunCode)

}

func setupTester(logger logging.Logger) error {
	config := configuration.LoadIntoWithEnvPath(&configuration.Configuration{}, "./../.env")

	db, err := storage.NewDatabase(config.DbConnectionString, logger)
	if err != nil {
		return err
	}

	clusters, err := db.GetClusters(context.Background())
	if err != nil {
		return err
	}
	confluentClient := confluent.NewClient(logger, config.CreateCloudApiAccess(), storage.NewClusterCache(clusters))

	mock := mocks.NewVaultMock()

	tester = newTesterApp(db, confluentClient, &mock)

	return nil
}
