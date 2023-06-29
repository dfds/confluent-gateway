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
	CreateSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error
	CreateServiceAccountRoleBinding(clusterAccess *models.ClusterAccess) error
	CountClusterApiKeys(clusterAccess *models.ClusterAccess) (int, error)
	CountSchemaRegistryApiKeys(clusterAccess *models.ClusterAccess) (int, error)
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

func (c *StepContext) LogDebug(format string, args ...string) {
	c.logger.Debug(format, args...)
}

func (c *StepContext) LogError(err error, format string, args ...string) {
	c.logger.Error(err, format, args...)
}

func (c *StepContext) HasServiceAccount() bool {
	account, err := c.account.GetServiceAccount(c.input.CapabilityId)
	if err != nil {
		c.LogError(err, fmt.Sprintf("encountered error when checking if ServiceAccount exists for CapabilityId %s", c.input.CapabilityId))
		return false
	}
	return account != nil
}

func (c *StepContext) CreateServiceAccount() error {
	return c.account.CreateServiceAccount(c.input.CapabilityId, c.input.ClusterId)
}

func (c *StepContext) GetInputCapabilityId() models.CapabilityId {
	return c.input.CapabilityId
}

func (c *StepContext) HasClusterAccessWithValidAcls() bool {
	serviceAccount, _ := c.account.GetServiceAccount(c.input.CapabilityId)
	if serviceAccount != nil {
		clusterAccess, hasClusterAccess := serviceAccount.TryGetClusterAccess(c.input.ClusterId)
		if !hasClusterAccess {
			return false
		}
		for _, entry := range clusterAccess.Acl {
			if !entry.IsValid() {
				return false
			}
		}
		return true
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

func (c *StepContext) HasClusterApiKey(clusterAccess *models.ClusterAccess) bool {
	count, err := c.account.CountClusterApiKeys(clusterAccess)
	return count > 0 && err == nil
}

func (c *StepContext) HasClusterApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error) {
	return c.vault.QueryClusterApiKey(c.input.CapabilityId, clusterAccess)
}

func (c *StepContext) CreateClusterApiKey(clusterAccess *models.ClusterAccess) error {
	return c.account.CreateClusterApiKey(clusterAccess)
}

func (c *StepContext) StoreClusterApiKey(clusterAccess *models.ClusterAccess) error {
	return c.vault.StoreClusterApiKey(c.input.CapabilityId, clusterAccess)
}

func (c *StepContext) GetServiceAccount() (*models.ServiceAccount, error) {
	return c.account.GetServiceAccount(c.input.CapabilityId)
}

func (c *StepContext) EnsureHasSchemaRegistryApiKey(access *models.ClusterAccess) error {

	keyCount, err := c.account.CountSchemaRegistryApiKeys(access)
	if err != nil {
		return err
	}
	if keyCount > 0 {
		c.logger.Information("found SchemaRegistry api key, skipping creation")
		return nil
	}

	return c.account.CreateSchemaRegistryApiKey(access)
}

func (c *StepContext) HasSchemaRegistryApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error) {
	return c.vault.QuerySchemaRegistryApiKey(c.input.CapabilityId, clusterAccess)
}

func (c *StepContext) CreateServiceAccountRoleBinding(*models.ClusterAccess) error {
	access, err := c.GetClusterAccess()
	if err != nil {
		return err
	}
	return c.account.CreateServiceAccountRoleBinding(access)
}

func (c *StepContext) StoreSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error {
	return c.vault.StoreSchemaRegistryApiKey(c.input.CapabilityId, clusterAccess)
}

func (c *StepContext) RaiseServiceAccountAccessGranted() error {
	event := &ServiceAccountAccessGranted{
		CapabilityId:   string(c.input.CapabilityId),
		KafkaClusterId: string(c.input.ClusterId),
	}
	return c.outbox.Produce(event)
}
