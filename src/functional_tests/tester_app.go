package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/configuration"
	"github.com/dfds/confluent-gateway/functional_tests/mocks"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var (
	testServiceAccountId = models.ServiceAccountId("test-service-account-id")
	testTopicId          = "test-topic-id"
	testUserAccountId    = models.UserAccountId("test-user")
	testCapabilityId     = models.CapabilityId("test-capability-id")
	testClusterId        = models.ClusterId("abc-1234")
)

type TesterBlackboard struct {
	serviceAccount *models.ServiceAccount
}

type TesterApp struct {
	logger          logging.Logger
	config          *configuration.Configuration
	db              *mocks.Database
	confluentClient *confluent.Client
	vaultClient     *vault.Vault
	blackboard      *TesterBlackboard
}

func newTesterApp(logger logging.Logger, config *configuration.Configuration, db *mocks.Database, confluentClient *confluent.Client, vaultClient *vault.Vault) *TesterApp {
	return &TesterApp{logger: logger,
		config:          config,
		db:              db,
		confluentClient: confluentClient,
		vaultClient:     vaultClient,
		blackboard:      &TesterBlackboard{}}
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

	return newTesterApp(logger, config, mockDb, confluentClient, &mockVault), nil
}

func (t *TesterApp) AddMockServiceAccountWithClusterAccess() {

	clusterAccess := []models.ClusterAccess{
		*models.NewClusterAccess(testServiceAccountId, testUserAccountId, testClusterId, testCapabilityId),
	}

	newServiceAccount := &models.ServiceAccount{
		Id:              testServiceAccountId,
		UserAccountId:   testUserAccountId,
		CapabilityId:    testCapabilityId,
		ClusterAccesses: clusterAccess,
		CreatedAt:       time.Now(),
	}
	t.blackboard.serviceAccount = newServiceAccount

	err := t.db.CreateServiceAccount(newServiceAccount)
	if err != nil {
		panic(err)
	}
	account, err := t.db.GetServiceAccount(newServiceAccount.CapabilityId)
	if err != nil {
		return
	}
	for _, access := range account.ClusterAccesses {
		for _, entry := range access.Acl {
			err := t.db.UpdateAclEntry(&entry)
			if err != nil {
				return
			}
		}
	}

}

func (t *TesterApp) RemoveMockServiceAccount() error {
	if t.blackboard.serviceAccount == nil {
		return fmt.Errorf("attempted to remove mock service account that hasn't been noted on the blackboard")
	}
	err := t.db.RemoveServiceAccount(t.blackboard.serviceAccount)
	if err != nil {
		return err
	}
	return nil
}

func (t *TesterApp) TearDown() {

	var errors []error

	appendIfErr := func(slice []error, err error) []error {
		if err != nil {
			slice = append(slice, err)
		}
		return slice
	}

	t.logger.Information("Tearing down test environment")
	errors = appendIfErr(errors, t.db.DeleteTopic(testTopicId))
	errors = appendIfErr(errors, t.db.RemoveAllOutboxEntries())
	errors = appendIfErr(errors, t.db.RemoveServiceAccount(t.blackboard.serviceAccount))

	if len(errors) != 0 {
		errorList := ""
		for _, err := range errors {
			errorList += fmt.Sprintf("%s\n", err)
		}
		t.logger.Warning(fmt.Sprintf("Tearing down produced errors:\n%s", errorList))
	}
}
