package serviceaccount

import (
	"context"

	"github.com/dfds/confluent-gateway/internal/models"
	proc "github.com/dfds/confluent-gateway/internal/process"
	"github.com/dfds/confluent-gateway/logging"
)

type logger interface {
	LogTrace(string, ...string)
}

type process struct {
	logger    logging.Logger
	database  models.Database
	confluent Confluent
	vault     Vault
	factory   OutboxFactory
}

func NewProcess(logger logging.Logger, database models.Database, confluent Confluent, vault Vault, factory OutboxFactory) Process {
	return &process{
		logger:    logger,
		database:  database,
		confluent: confluent,
		vault:     vault,
		factory:   factory,
	}
}

type ProcessInput struct {
	CapabilityId models.CapabilityId
	ClusterId    models.ClusterId
}

func (p *process) Process(ctx context.Context, input ProcessInput) error {
	session := p.database.NewSession(ctx)

	return proc.PrepareSteps[*StepContext]().
		Step(ensureServiceAccountStep).
		Step(ensureServiceAccountAclStep).Until(func(c *StepContext) bool { return c.HasClusterAccess() }).
		Step(ensureServiceAccountApiKeyStep).
		Step(ensureServiceAccountApiKeyAreStoredInVaultStep).
		Step(ensureServiceAccountHasSchemaRegistryAccessStep).
		Run(func(step func(*StepContext) error) error {
			return session.Transaction(func(tx models.Transaction) error {
				stepContext := p.getStepContext(ctx, tx, input)
				err := step(stepContext)
				if err != nil {
					return err
				}
				return nil
			})
		})
}

func (p *process) getStepContext(ctx context.Context, tx models.Transaction, input ProcessInput) *StepContext {
	logger := p.logger
	newAccountService := NewAccountService(ctx, p.confluent, tx)
	vault := NewVaultService(ctx, p.vault)
	outbox := p.factory(tx)

	return NewStepContext(logger, newAccountService, vault, outbox, input)
}

// region Steps

type EnsureServiceAccountStepRequirement interface {
	logger
	HasServiceAccount() bool
	CreateServiceAccount() error
}

func ensureServiceAccountStepInner(sr EnsureServiceAccountStepRequirement) error {
	sr.LogTrace("Running {Step}", "EnsureServiceAccount")
	if sr.HasServiceAccount() {
		return nil
	}

	err := sr.CreateServiceAccount()
	if err != nil {
		return err
	}

	return nil
}

func ensureServiceAccountStep(step *StepContext) error {
	return ensureServiceAccountStepInner(step)
}

type EnsureServiceAccountAclStep interface {
	logger
	HasClusterAccess() bool
	GetOrCreateClusterAccess() (*models.ClusterAccess, error)
	CreateAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error
}

func ensureServiceAccountAclStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountAclStep) error {
		step.LogTrace("Running {Step}", "EnsureServiceAccountAcl")
		if step.HasClusterAccess() {
			return nil
		}

		clusterAccess, err := step.GetOrCreateClusterAccess()
		if err != nil {
			return err
		}

		entries := clusterAccess.GetAclPendingCreation()
		if len(entries) == 0 {
			// no acl entries left => continue
			return nil

		} else {
			nextEntry := entries[0]

			return step.CreateAclEntry(clusterAccess, nextEntry)
		}
	}
	return inner(step)
}

type EnsureServiceAccountApiKeyStep interface {
	logger
	HasApiKey(clusterAccess *models.ClusterAccess) bool
	GetClusterAccess() (*models.ClusterAccess, error)
	CreateClusterApiKey(clusterAccess *models.ClusterAccess) error
}

func ensureServiceAccountApiKeyStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountApiKeyStep) error {
		step.LogTrace("Running {Step}", "EnsureServiceAccountApiKey")

		clusterAccess, err := step.GetClusterAccess()
		if err != nil {
			return err
		}

		if step.HasApiKey(clusterAccess) {
			return nil
		}

		err = step.CreateClusterApiKey(clusterAccess)
		if err != nil {
			return err
		}

		return nil
	}
	return inner(step)
}

type EnsureServiceAccountApiKeyAreStoredInVaultStep interface {
	logger
	HasApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error)
	GetClusterAccess() (*models.ClusterAccess, error)
	StoreApiKey(clusterAccess *models.ClusterAccess) error
}

func ensureServiceAccountApiKeyAreStoredInVaultStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountApiKeyAreStoredInVaultStep) error {
		step.LogTrace("Running {Step}", "EnsureServiceAccountApiKeyAreStoredInVault")

		clusterAccess, err := step.GetClusterAccess()
		if err != nil {
			return err
		}

		hasKey, err := step.HasApiKeyInVault(clusterAccess)
		if err != nil {
			return err
		}
		if hasKey {
			return nil
		}

		if err = step.StoreApiKey(clusterAccess); err != nil {
			return err
		}

		return nil
	}
	return inner(step)
}

type EnsureServiceAccountHasSchemaRegistryAccess interface {
	logger
	GetServiceAccount() (*models.ServiceAccount, error)
	EnsureSchemaRegistryApiKey(models.ServiceAccountId) error
	CreateServiceAccountRoleBinding() error
	StoreSchemaRegistryApiKey() error
}

func ensureServiceAccountHasSchemaRegistryAccessStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountHasSchemaRegistryAccess) error {
		step.LogTrace("Running {Step}", "EnsureServiceAccountHasSchemaRegistryAccess")

		account, err := step.GetServiceAccount()

		err = step.EnsureSchemaRegistryApiKey(account.Id)
		if err != nil {
			return err
		}

		err = step.CreateServiceAccountRoleBinding()
		if err != nil {
			return err
		}

		if err = step.StoreSchemaRegistryApiKey(); err != nil {
			return err
		}

		return nil
	}
	return inner(step)
}
