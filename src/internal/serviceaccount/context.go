package serviceaccount

import (
	"fmt"

	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
)

type StepContext struct {
	//EnsureServiceAccountStepRequirement

	logger  logging.Logger
	account AccountService
	vault   VaultService
	outbox  Outbox
	input   ProcessInput
}

func NewStepContext(logger logging.Logger, account AccountService, vault VaultService, outbox Outbox, input ProcessInput) *StepContext {
	return &StepContext{logger: logger, account: account, vault: vault, outbox: outbox, input: input}
}

type AccountService interface {
	GetServiceAccount(models.CapabilityId) (*models.ServiceAccount, error)
	CreateServiceAccount(models.CapabilityId, models.ClusterId) error
	GetOrCreateClusterAccess(models.CapabilityId, models.ClusterId) (*models.ClusterAccess, error)
	GetClusterAccess(models.CapabilityId, models.ClusterId) (*models.ClusterAccess, error)
	CreateAclEntry(models.ClusterId, models.UserAccountId, *models.AclEntry) error
	CreateClusterApiKey(*models.ClusterAccess) error
	CreateSchemaRegistryApiKey(clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) error
	CreateServiceAccountRoleBinding(clusterAccess *models.ClusterAccess) error
	CountApiKeys(clusterAccess *models.ClusterAccess) (int, error)
}

type VaultService interface {
	StoreClusterApiKey(models.CapabilityId, *models.ClusterAccess) error
	QueryClusterApiKey(models.CapabilityId, *models.ClusterAccess) (bool, error)
	StoreSchemaRegistryApiKey(models.CapabilityId, *models.ClusterAccess) error
	QuerySchemaRegistryApiKey(models.CapabilityId, *models.ClusterAccess) (bool, error)
}

type Outbox interface {
	Produce(msg messaging.OutgoingMessage) error
}

type OutboxRepository interface {
	AddToOutbox(entry *messaging.OutboxEntry) error
}

type OutboxFactory func(repository OutboxRepository) Outbox

func (c *StepContext) LogTrace(format string, args ...string) {
	c.logger.Trace(format, args...)
}

func (c *StepContext) LogError(err error, format string, args ...string) {
	c.logger.Error(err, format, args...)
}

func (c *StepContext) HasServiceAccount() bool {
	account, err := c.account.GetServiceAccount(c.input.CapabilityId)
	fmt.Printf("Service account for CapabilityId: %s\n\tFound: %t\n\tError: %t\n", c.input.CapabilityId, account != nil, err != nil)
	if err != nil {
		return false
	}
	return account != nil
}

func (c *StepContext) CreateServiceAccount() error {
	return c.account.CreateServiceAccount(c.input.CapabilityId, c.input.ClusterId)
}

func (c *StepContext) HasClusterAccess() bool {
	serviceAccount, _ := c.account.GetServiceAccount(c.input.CapabilityId)
	if serviceAccount != nil {
		_, HasClusterAccess := serviceAccount.TryGetClusterAccess(c.input.ClusterId)
		return HasClusterAccess
	}
	return false
}

func (c *StepContext) GetOrCreateClusterAccess() (*models.ClusterAccess, error) {
	return c.account.GetOrCreateClusterAccess(c.input.CapabilityId, c.input.ClusterId)
}

func (c *StepContext) CreateAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error {
	return c.account.CreateAclEntry(c.input.ClusterId, clusterAccess.UserAccountId, &nextEntry)
}

func (c *StepContext) GetClusterAccess() (*models.ClusterAccess, error) {
	return c.account.GetClusterAccess(c.input.CapabilityId, c.input.ClusterId)
}

func (c *StepContext) HasApiKey(clusterAccess *models.ClusterAccess) bool {
	count, err := c.account.CountApiKeys(clusterAccess)
	return count > 0 && err == nil
}

func (c *StepContext) HasApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error) {
	return c.vault.QueryClusterApiKey(c.input.CapabilityId, clusterAccess)
}

func (c *StepContext) CreateClusterApiKey(clusterAccess *models.ClusterAccess) error {
	return c.account.CreateClusterApiKey(clusterAccess)
}

func (c *StepContext) StoreApiKey(clusterAccess *models.ClusterAccess) error {
	return c.vault.StoreClusterApiKey(c.input.CapabilityId, clusterAccess)
}

func (c *StepContext) GetServiceAccount() (*models.ServiceAccount, error) {
	return c.account.GetServiceAccount(c.input.CapabilityId)
}

func (c *StepContext) EnsureSchemaRegistryApiKey(serviceAccountId models.ServiceAccountId) error {
	return c.account.CreateSchemaRegistryApiKey(c.input.ClusterId, serviceAccountId)
}

func (c *StepContext) CreateServiceAccountRoleBinding() error {
	access, err := c.GetClusterAccess()
	if err != nil {
		return err
	}
	return c.account.CreateServiceAccountRoleBinding(access)
}

func (c *StepContext) StoreSchemaRegistryApiKey() error {
	access, err := c.GetClusterAccess()
	if err != nil {
		return err
	}
	return c.vault.StoreClusterApiKey(c.input.CapabilityId, access)
}
