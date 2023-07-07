package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/configuration"
	"github.com/dfds/confluent-gateway/functional_tests/helpers"
	"github.com/dfds/confluent-gateway/functional_tests/mocks"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SeedVariables from db/seed/cluster.csv TODO: figure out how to set it up in a better way
type SeedVariables struct {
	AdminUser                   string
	AdminPassword               string
	SchemaRegistryAdminUser     string
	SchemaRegistryAdminPassword string
	AdminApiEndpoint            string
	SchemaRegistryApiEndpoint   string
	ProductionClusterId         models.ClusterId
	DevelopmentClusterId        models.ClusterId
}

type TesterApp struct {
	logger          logging.Logger
	config          *configuration.Configuration
	db              *mocks.Database
	confluentClient *confluent.Client
	vaultClient     *vault.Vault
	dbSeedVariables *SeedVariables
}

func newTesterApp(logger logging.Logger, config *configuration.Configuration, db *mocks.Database, confluentClient *confluent.Client, vaultClient *vault.Vault, seedVariables *SeedVariables) *TesterApp {
	return &TesterApp{logger: logger,
		config:          config,
		db:              db,
		confluentClient: confluentClient,
		vaultClient:     vaultClient,
		dbSeedVariables: seedVariables,
	}
}

func CreateAndSetupTester(logger logging.Logger) (*TesterApp, error) {
	config := configuration.LoadInto(&configuration.Configuration{})

	db, err := storage.NewDatabase(config.DbConnectionString, logger)
	if err != nil {
		return nil, err
	}
	// TODO: Figure out a way to not having to open db connection twice
	rawDb, err := gorm.Open(postgres.Open(config.DbConnectionString))
	if err != nil {
		return nil, err
	}

	mockDb := mocks.NewDatabase(db, rawDb)

	clusters, err := mockDb.GetClusters(context.Background())
	if err != nil {
		return nil, err
	}
	confluentClient := confluent.NewClient(logger, config.CreateCloudApiAccess(), storage.NewClusterCache(clusters))

	mockVault := mocks.NewVaultMock()

	seedVariables :=
		&SeedVariables{
			AdminUser:                   "admin_user",
			AdminPassword:               "admin_pass",
			SchemaRegistryAdminUser:     "admin_user",
			SchemaRegistryAdminPassword: "admin_pass",
			AdminApiEndpoint:            "http://localhost:5051",
			SchemaRegistryApiEndpoint:   "http://localhost:5051",
			ProductionClusterId:         "abc-1234",
			DevelopmentClusterId:        "def-5678",
		}

	return newTesterApp(logger, config, mockDb, confluentClient, &mockVault, seedVariables), nil
}

func (t *TesterApp) FullTearDown() {

	var errors helpers.ErrorList

	t.logger.Information("Performing full tear down of test environment (db wipe minus cluster table)")
	errors.AppendIfErr(t.db.RemoveAllOutboxEntries())
	errors.AppendErrors(t.db.RemoveAllServiceAccounts())
	errors.AppendIfErr(t.db.RemoveAllCreateProcesses())
	errors.AppendIfErr(t.db.RemoveAllDeleteProcesses())
	errors.AppendIfErr(t.db.RemoveAllTopics())

	if errors.HasErrors() {
		t.logger.Warning(fmt.Sprintf("Tearing down produced errors:\n%s", errors.String()))
	}
}
