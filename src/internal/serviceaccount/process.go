package serviceaccount

import (
	"context"
	"fmt"

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
		Step(EnsureServiceAccountStep).
		Step(ensureServiceAccountAclStep).Until(func(c *StepContext) bool { return c.HasClusterAccess() }).
		Step(ensureServiceAccountApiKeyStep).
		Step(ensureServiceAccountApiKeyAreStoredInVaultStep).
		Run(func(step func(*StepContext) error) error {
			return session.Transaction(func(tx models.Transaction) error {
				stepContext := p.getStepContext(ctx, tx, input)
				fmt.Printf("Starting Step\n")
				err := step(stepContext)
				if err != nil {
					fmt.Printf("Step: FAIL\n")
					return err
				}
				fmt.Printf("Step: SUCCESS\n")
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

func EnsureServiceAccountStep(step *StepContext) error {
	fmt.Printf("Step:1\n\n")
	inner := func(sr EnsureServiceAccountStepRequirement) error {
		sr.LogTrace("Running {Step}", "EnsureServiceAccount")
		if sr.HasServiceAccount() {
			fmt.Printf("Step 1: SKIPPED\n\n")
			return nil
		}

		err := sr.CreateServiceAccount()
		if err != nil {
			fmt.Printf("FLUTTERSHY\n")
			return err
		}

		return nil
	}
	return inner(step)
}

type EnsureServiceAccountAclStep interface {
	logger
	HasClusterAccess() bool
	GetOrCreateClusterAccess() (*models.ClusterAccess, error)
	CreateAclEntry(clusterAccess *models.ClusterAccess, nextEntry models.AclEntry) error
}

func ensureServiceAccountAclStep(step *StepContext) error {
	fmt.Printf("Step:2\n\n")
	inner := func(step EnsureServiceAccountAclStep) error {
		step.LogTrace("Running {Step}", "EnsureServiceAccountAcl")
		if step.HasClusterAccess() {
			fmt.Printf("Step 2: SKIPPED\n\n")
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
	CreateApiKey(clusterAccess *models.ClusterAccess) error
}

func ensureServiceAccountApiKeyStep(step *StepContext) error {
	fmt.Printf("Step:3\n\n")
	inner := func(step EnsureServiceAccountApiKeyStep) error {
		step.LogTrace("Running {Step}", "EnsureServiceAccountApiKey")

		clusterAccess, err := step.GetClusterAccess()
		if err != nil {
			return err
		}

		if step.HasApiKey(clusterAccess) {
			fmt.Printf("Step 3: SKIPPED\n\n")
			return nil
		}

		err = step.CreateApiKey(clusterAccess)
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
	fmt.Printf("Step:4\n\n")
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
			fmt.Printf("Step 4: SKIPPED\n\n")
			return nil
		}

		if err = step.StoreApiKey(clusterAccess); err != nil {
			return err
		}

		return nil
	}
	return inner(step)
}
