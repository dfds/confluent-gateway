package serviceaccount

import (
	"context"
	"errors"
	"github.com/dfds/confluent-gateway/internal/confluent"

	"github.com/dfds/confluent-gateway/internal/models"
	proc "github.com/dfds/confluent-gateway/internal/process"
	"github.com/dfds/confluent-gateway/logging"
)

type logger interface {
	LogDebug(string, ...string)
	LogError(error, string, ...string)
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
		Step(ensureServiceAccountAclStep).Until(func(c *StepContext) bool { return c.HasClusterAccessWithValidAcls() }).
		Step(ensureServiceAccountApiKeyStep).
		Step(ensureServiceAccountApiKeyAreStoredInVaultStep).
		Step(ensureServiceAccountHasSchemaRegistryAccessStep).
		Step(raiseServiceAccountAccessGrantedStep).
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
	GetInputCapabilityId() models.CapabilityId
}

func ensureServiceAccountStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountStepRequirement) error {
		step.LogDebug("Running {Step}", "EnsureServiceAccount")
		if step.HasServiceAccount() {
			step.LogDebug("found existing service account for CapabilityId {CapabilityId}", string(step.GetInputCapabilityId()))
			return nil
		}

		err := step.CreateServiceAccount()
		if err != nil {
			return err
		}

		return nil
	}
	return inner(step)
}

type EnsureServiceAccountAclStep interface {
	logger
	HasClusterAccessWithValidAcls() bool
	GetInputCapabilityId() models.CapabilityId
	GetOrCreateClusterAccess() (*models.ClusterAccess, error)
	CreateAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error
}

func ensureServiceAccountAclStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountAclStep) error {
		step.LogDebug("Running {Step}", "EnsureServiceAccountAcl")

		if step.HasClusterAccessWithValidAcls() {
			step.LogDebug("skipping step: ServiceAccount {ServiceAccount} already has ClusterAccess", string(step.GetInputCapabilityId()))
			return nil
		}

		clusterAccess, err := step.GetOrCreateClusterAccess()
		if err != nil {
			return err
		}

		entries := clusterAccess.GetAclPendingCreation()
		if len(entries) == 0 {
			step.LogDebug("found no ACL pending creation")
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
	HasClusterApiKey(clusterAccess *models.ClusterAccess) (bool, error)
	GetClusterAccess() (*models.ClusterAccess, error)
	CreateClusterApiKey(clusterAccess *models.ClusterAccess) error
}

func ensureServiceAccountApiKeyStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountApiKeyStep) error {
		step.LogDebug("Running {Step}", "EnsureServiceAccountApiKey")

		clusterAccess, err := step.GetClusterAccess()
		if err != nil {
			return err
		}

		hasKey, err := step.HasClusterApiKey(clusterAccess)
		if err != nil {
			return err
		}
		if hasKey {
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
	HasClusterApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error)
	GetClusterAccess() (*models.ClusterAccess, error)
	StoreClusterApiKey(clusterAccess *models.ClusterAccess) error
}

func ensureServiceAccountApiKeyAreStoredInVaultStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountApiKeyAreStoredInVaultStep) error {
		step.LogDebug("Running {Step}", "EnsureServiceAccountApiKeyAreStoredInVault")

		clusterAccess, err := step.GetClusterAccess()
		if err != nil {
			return err
		}

		hasKey, err := step.HasClusterApiKeyInVault(clusterAccess)
		if err != nil {
			return err
		}
		if hasKey {
			return nil
		}

		if err = step.StoreClusterApiKey(clusterAccess); err != nil {
			return err
		}
		return nil
	}
	return inner(step)
}

type EnsureServiceAccountHasSchemaRegistryAccess interface {
	logger
	GetClusterAccess() (*models.ClusterAccess, error)
	CreateServiceAccountRoleBinding(*models.ClusterAccess) error
	EnsureHasSchemaRegistryApiKey(*models.ClusterAccess) error
	HasSchemaRegistryApiKeyInVault(*models.ClusterAccess) (bool, error)
	StoreSchemaRegistryApiKey(*models.ClusterAccess) error
}

func ensureServiceAccountHasSchemaRegistryAccessStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountHasSchemaRegistryAccess) error {
		step.LogDebug("Running {Step}", "EnsureServiceAccountHasSchemaRegistryAccess")

		clusterAccess, err := step.GetClusterAccess()
		if err != nil {
			return err
		}

		err = step.CreateServiceAccountRoleBinding(clusterAccess)
		if err != nil {
			if errors.Is(err, confluent.ErrMissingSchemaRegistryIds) {
				step.LogError(err, "unable to setup schema registry access")
				return nil // fallback: setup cluster without schema registry access
			}
			return err
		}
		err = step.EnsureHasSchemaRegistryApiKey(clusterAccess)
		if err != nil {
			if errors.Is(err, confluent.ErrSchemaRegistryIdIsEmpty) {
				step.LogError(err, "unable to setup schema registry access")
				return nil // fallback: setup cluster without schema registry access
			}
			return err
		}

		hasKey, err := step.HasSchemaRegistryApiKeyInVault(clusterAccess)
		if err != nil {
			return err
		}
		if hasKey {
			return nil
		}

		if err = step.StoreSchemaRegistryApiKey(clusterAccess); err != nil {
			return err
		}

		return nil
	}
	return inner(step)
}

func raiseServiceAccountAccessGrantedStep(step *StepContext) error {
	step.LogDebug("Running {Step}", "raiseServiceAccountAccessGrantedStep")
	return step.RaiseServiceAccountAccessGranted()
}
