package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dfds/confluent-gateway/configuration"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/create"
	del "github.com/dfds/confluent-gateway/internal/delete"
	"github.com/dfds/confluent-gateway/internal/http/metrics"
	schema "github.com/dfds/confluent-gateway/internal/schema"
	"github.com/dfds/confluent-gateway/internal/serviceaccount"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// load configuration from .env and/or environment files
	config := configuration.LoadInto("", &configuration.Configuration{})
	logger := logging.NewLogger(logging.LoggerOptions{IsProduction: config.IsProduction(), AppName: config.ApplicationName})
	db := Must(storage.NewDatabase(config.DbConnectionString, logger))
	clusters := Must(db.GetClusters(ctx))
	confluentClient := confluent.NewClient(logger, config.CreateCloudApiAccess(), storage.NewClusterCache(clusters))
	awsClient := Must(vault.NewVaultClient(logger, Must(config.CreateVaultConfig())))

	outboxFactory := Must(messaging.ConfigureOutbox(logger,
		// TODO -- fix inconsistency in message type
		messaging.RegisterMessage(config.TopicNameProvisioning, "topic_provisioned", &create.TopicProvisioned{}),
		messaging.RegisterMessage(config.TopicNameProvisioning, "topic_provisioning_begun", &create.TopicProvisioningBegun{}),
		messaging.RegisterMessage(config.TopicNameProvisioning, "topic-deleted", &del.TopicDeleted{}),
		messaging.RegisterMessage(config.TopicNameSchema, "schema-registered", &schema.SchemaRegistered{}),
		messaging.RegisterMessage(config.TopicNameSchema, "schema-registration-failed", &schema.SchemaRegistrationFailed{}),
		messaging.RegisterMessage(config.TopicNameKafkaClusterAccessGranted, "cluster-access-granted", &serviceaccount.ServiceAccountAccessGranted{}),
	))
	createTopicProcess := create.NewProcess(logger, db, confluentClient, func(repository create.OutboxRepository) create.Outbox { return outboxFactory(repository) })
	createServiceAccountProcess := serviceaccount.NewProcess(logger, db, confluentClient, awsClient, func(repository serviceaccount.OutboxRepository) serviceaccount.Outbox {
		return outboxFactory(repository)
	})
	deleteTopicProcess := del.NewProcess(logger, db, confluentClient, func(repository del.OutboxRepository) del.Outbox { return outboxFactory(repository) })
	addSchemaProcess := schema.NewProcess(logger, db, confluentClient, awsClient, func(repository schema.OutboxRepository) schema.Outbox { return outboxFactory(repository) })
	consumer := Must(messaging.ConfigureConsumer(logger, config.KafkaBroker, config.KafkaGroupId,
		messaging.WithCredentials(config.CreateConsumerCredentials()),
		messaging.RegisterMessageHandler(config.TopicNameSelfService, "topic_requested", create.NewTopicRequestedHandler(createTopicProcess), &create.TopicRequested{}),
		messaging.RegisterMessageHandler(config.TopicNameSelfService, "topic-requested", create.NewTopicRequestedHandler(createTopicProcess), &create.TopicRequested{}),
		messaging.RegisterMessageHandler(config.TopicNameSelfService, "topic-deleted", del.NewTopicRequestedHandler(deleteTopicProcess), &del.TopicDeletionRequested{}),
		messaging.RegisterMessageHandler(config.TopicNameMessageContract, "message-contract-requested", schema.NewSchemaAddedHandler(addSchemaProcess), &schema.MessageContractRequested{}),
		messaging.RegisterMessageHandler(config.TopicNameMessageContract, "message-contract-provisioned", messaging.NewNopHandler(logger), &messaging.Nop{}),
		messaging.RegisterMessageHandler(config.TopicNameKafkaClusterAccess, "cluster-access-requested", serviceaccount.NewAccessRequestedHandler(createServiceAccountProcess), &serviceaccount.ServiceAccountAccessRequested{}),
	))

	m := NewMain(logger, config, consumer)

	logger.Information("Running")

	if err := m.Run(ctx); err != nil {
		logger.Error(err, "Exit reason {Reason}", err.Error())
		os.Exit(1)
	}

	logger.Information("Done!")
}

func Must[T any](any T, err error) T {
	if err != nil {
		panic(err)
	}
	return any
}

type Main struct {
	Logger        logging.Logger
	Consumer      messaging.Consumer
	MetricsServer *metrics.Server
}

func NewMain(logger logging.Logger, config *configuration.Configuration, consumer messaging.Consumer) *Main {
	return &Main{
		Logger:        logger,
		Consumer:      consumer,
		MetricsServer: metrics.NewServer(logger, config.IsProduction()),
	}
}

func (m *Main) Run(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	m.RunMetricsServer(g, gCtx)
	m.RunConsumer(g, gCtx)

	// wait for context or all go routines to finish
	return g.Wait()
}

func (m *Main) RunMetricsServer(g *errgroup.Group, ctx context.Context) {
	g.Go(m.MetricsServer.Open)

	g.Go(func() error {
		// wait until cancelled
		<-ctx.Done()

		return m.MetricsServer.Close()
	})
}

func (m *Main) RunConsumer(g *errgroup.Group, ctx context.Context) {
	cleanup := func() {
		log.Println("Stopping consumer")
		err := m.Consumer.Stop()
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		log.Println("Consumer stopped")
	}

	g.Go(func() error {
		return m.Consumer.Start(ctx)
	})

	g.Go(func() error {
		// wait until cancelled
		<-ctx.Done()

		m.Logger.Warning("!!! STOPPING CONSUMER !!!")

		cleanup()

		m.Logger.Warning("!!! STOPPED CONSUMER !!!")

		return nil
	})
}

func (m *Main) Close() {
	if m.Consumer != nil {
		err := m.Consumer.Stop()
		if err != nil {
			m.Logger.Error(err, err.Error())
		}
	}
	if m.MetricsServer != nil {
		err := m.MetricsServer.Close()
		if err != nil {
			m.Logger.Error(err, err.Error())
		}
	}
}
