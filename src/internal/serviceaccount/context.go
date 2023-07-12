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
	CreateClusterApiKey(*models.ClusterAccess) (models.ApiKey, error)
	CreateSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) (models.ApiKey, error)
	CreateServiceAccountRoleBinding(clusterAccess *models.ClusterAccess) error
	CountClusterApiKeys(clusterAccess *models.ClusterAccess) (int, error)
	CountSchemaRegistryApiKeys(clusterAccess *models.ClusterAccess) (int, error)
	DeleteClusterApiKey(clusterAccess *models.ClusterAccess) error
	DeleteSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error
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

func (c *StepContext) LogWarning(format string, args ...string) {
	c.logger.Warning(format, args...)
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

func (c *StepContext) HasClusterApiKey(clusterAccess *models.ClusterAccess) (bool, error) {
	count, err := c.account.CountClusterApiKeys(clusterAccess)
	return count > 0, err
}

func (c *StepContext) HasSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) (bool, error) {
	count, err := c.account.CountSchemaRegistryApiKeys(clusterAccess)
	return count > 0, err
}

func (c *StepContext) HasClusterApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error) {
	return c.vault.QueryClusterApiKey(c.input.CapabilityId, clusterAccess.ClusterId)
}

func (c *StepContext) HasSchemaRegistryApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error) {
	return c.vault.QuerySchemaRegistryApiKey(c.input.CapabilityId, clusterAccess.ClusterId)
}

func createApiKeyAndStoreInVault(capabilityId models.CapabilityId, clusterAccess *models.ClusterAccess,
	creationFunc func(access *models.ClusterAccess) (models.ApiKey, error),
	storeFunc func(capabilityId models.CapabilityId, clusterId models.ClusterId, key models.ApiKey) error) error {
	key, err := creationFunc(clusterAccess)

	if err != nil {
		return err
	}

	err = storeFunc(capabilityId, clusterAccess.ClusterId, key)
	if err != nil {
		return err
	}
	return nil
}

func (c *StepContext) CreateClusterApiKeyAndStoreInVault(clusterAccess *models.ClusterAccess) error {
	return createApiKeyAndStoreInVault(c.input.CapabilityId, clusterAccess, c.account.CreateClusterApiKey, c.vault.StoreClusterApiKey)

}
func (c *StepContext) CreateSchemaRegistryApiKeyAndStoreInVault(clusterAccess *models.ClusterAccess) error {
	return createApiKeyAndStoreInVault(c.input.CapabilityId, clusterAccess, c.account.CreateSchemaRegistryApiKey, c.vault.StoreSchemaRegistryApiKey)
}

func (c *StepContext) DeleteClusterApiKey(clusterAccess *models.ClusterAccess) error {
	return c.account.DeleteClusterApiKey(clusterAccess)
}
func (c *StepContext) DeleteClusterApiKeyInVault(access *models.ClusterAccess) error {
	return c.vault.DeleteClusterApiKey(c.input.CapabilityId, access.ClusterId)
}

func (c *StepContext) DeleteSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error {
	return c.account.DeleteSchemaRegistryApiKey(clusterAccess)
}

func (c *StepContext) GetServiceAccount() (*models.ServiceAccount, error) {
	return c.account.GetServiceAccount(c.input.CapabilityId)
}

func (c *StepContext) CreateServiceAccountRoleBinding(clusterAccess *models.ClusterAccess) error {
	return c.account.CreateServiceAccountRoleBinding(clusterAccess)
}

func (c *StepContext) RaiseServiceAccountAccessGranted() error {
	event := &ServiceAccountAccessGranted{
		CapabilityId:   string(c.input.CapabilityId),
		KafkaClusterId: string(c.input.ClusterId),
	}
	return c.outbox.Produce(event)
}
