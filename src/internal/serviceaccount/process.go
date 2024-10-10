package serviceaccount

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	proc "github.com/dfds/confluent-gateway/internal/process"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/logging"
)

type logger interface {
	LogDebug(string, ...string)
	LogWarning(string, ...string)
	LogError(error, string, ...string)
}

type process struct {
	logger    logging.Logger
	database  models.Database
	confluent Confluent
	vault     vault.Vault
	factory   OutboxFactory
}

func NewProcess(logger logging.Logger, database models.Database, confluent Confluent, vault vault.Vault, factory OutboxFactory) Process {
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
		Step(ensureServiceAccountAclStep).
		Step(ensureServiceAccountClusterAccessStep).
		Step(ensureServiceAccountSchemaRegistryAccessStep).
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
	vaultService := NewVaultService(ctx, p.vault)
	outbox := p.factory(tx)

	return NewStepContext(logger, newAccountService, vaultService, outbox, input)
}

// region Steps

type EnsureServiceAccountStepRequirement interface {
	logger
	HasServiceAccount() bool
	CreateServiceAccount() error
	CreateServiceAccountDbLink() error
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
		if errors.Is(confluent.ErrFoundExistingServiceAccount, err) {
			step.LogWarning("found dangling service account for CapabilityId {CapabilityId}, attempting to connect confluent cloud service account with db service account", string(step.GetInputCapabilityId()))
			return step.CreateServiceAccountDbLink()
		}
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
		}

		for _, entry := range entries {
			err = step.CreateAclEntry(clusterAccess, entry)
			if err != nil {
				return fmt.Errorf("unable to create ACL entry with definition %s, error: %w", entry.AclDefinition, err)
			}
		}
		return nil
	}
	return inner(step)
}

type EnsureServiceAccountClusterAccessStep interface {
	logger
	GetClusterAccess() (*models.ClusterAccess, error)
	HasClusterApiKey(clusterAccess *models.ClusterAccess) (bool, error)
	HasClusterApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error)
	CreateClusterApiKeyAndStoreInVault(clusterAccess *models.ClusterAccess, shouldOverwriteKey bool) error
	DeleteClusterApiKey(clusterAccess *models.ClusterAccess) error
}

func ensureServiceAccountClusterAccessStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountClusterAccessStep) error {
		step.LogDebug("Running {Step}", "EnsureServiceAccountClusterAccessStep")
		clusterAccess, err := step.GetClusterAccess()
		if err != nil {
			return err
		}

		hasKeyInConfluent, err := step.HasClusterApiKey(clusterAccess)
		if err != nil {
			return err
		}
		hasKeyInVault, err := step.HasClusterApiKeyInVault(clusterAccess)
		if err != nil {
			return err
		}
		if hasKeyInVault && hasKeyInConfluent {
			return nil
		}

		recreateKey := false
		if hasKeyInConfluent && !hasKeyInVault {
			step.LogWarning("found existing api key in Confluent, but not in Parameter Store. Deleting key and creating again.")
			err = step.DeleteClusterApiKey(clusterAccess)
			if err != nil {
				return err
			}
		} else if !hasKeyInConfluent && hasKeyInVault { // not sure if this can happen
			step.LogWarning("found existing key in Parameter Store, but not in Confluent. Creating new key and updating Parameter Store.")
			recreateKey = true
		}

		return step.CreateClusterApiKeyAndStoreInVault(clusterAccess, recreateKey)
	}
	return inner(step)
}

type EnsureServiceAccountSchemaRegistryAccessStep interface {
	logger
	GetClusterAccess() (*models.ClusterAccess, error)
	HasSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) (bool, error)
	HasSchemaRegistryApiKeyInVault(clusterAccess *models.ClusterAccess) (bool, error)
	CreateServiceAccountRoleBinding(*models.ClusterAccess) error
	CreateSchemaRegistryApiKeyAndStoreInVault(clusterAccess *models.ClusterAccess, shouldOverwriteKey bool) error
	DeleteSchemaRegistryApiKey(clusterAccess *models.ClusterAccess) error
}

func ensureServiceAccountSchemaRegistryAccessStep(step *StepContext) error {
	inner := func(step EnsureServiceAccountSchemaRegistryAccessStep) error {
		step.LogDebug("Running {Step}", "EnsureServiceAccountSchemaRegistryAccessStep")

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

		hasKeyInConfluent, err := step.HasSchemaRegistryApiKey(clusterAccess)
		if err != nil {
			return err
		}
		hasKeyInVault, err := step.HasSchemaRegistryApiKeyInVault(clusterAccess)
		if err != nil {
			return err
		}

		if hasKeyInVault && hasKeyInConfluent {
			return nil
		}

		recreateKey := false
		if hasKeyInConfluent && !hasKeyInVault {
			step.LogWarning("found existing api key in Confluent, but not in Parameter Store. Deleting key and creating again.")
			err = step.DeleteSchemaRegistryApiKey(clusterAccess)
			if err != nil {
				return err
			}
		} else if !hasKeyInConfluent && hasKeyInVault { // not sure if this can happen
			step.LogWarning("found existing key in Parameter Store, but not in Confluent. Creating new key and updating Parameter Store.")
			recreateKey = true
		}

		err = step.CreateSchemaRegistryApiKeyAndStoreInVault(clusterAccess, recreateKey)
		if err != nil {
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
