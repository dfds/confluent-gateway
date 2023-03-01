package main

import (
	"context"
	"github.com/dfds/confluent-gateway/configuration"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/create"
	del "github.com/dfds/confluent-gateway/internal/delete"
	"github.com/dfds/confluent-gateway/internal/http/metrics"
	"github.com/dfds/confluent-gateway/internal/storage"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// load configuration from .env and/or environment files
	config := configuration.LoadInto(&Configuration{})
	logger := getLogger(config)
	db := getDatabase(config, logger)
	confluentClient := getConfluentClient(logger, config, db)
	awsClient := getVault(config, logger)

	outboxFactory := getOutboxFactory(logger,
		messaging.RegisterMessage(config.TopicNameProvisioning, "topic_provisioned", &create.TopicProvisioned{}),
		messaging.RegisterMessage(config.TopicNameProvisioning, "topic_provisioning_begun", &create.TopicProvisioningBegun{}),
		messaging.RegisterMessage(config.TopicNameProvisioning, "topic_deleted", &del.TopicDeleted{}),
	)

	createTopicProcess := create.NewProcess(logger, db, confluentClient, awsClient, func(repository create.OutboxRepository) create.Outbox { return outboxFactory(repository) })
	deleteTopicProcess := del.NewProcess(logger, db, confluentClient, func(repository del.OutboxRepository) del.Outbox { return outboxFactory(repository) })

	consumer := getConsumer(logger, config.KafkaBroker, config.KafkaGroupId,
		messaging.WithCredentials(config.CreateConsumerCredentials()),
		messaging.RegisterMessageHandler(config.TopicNameSelfService, "topic_requested", create.NewTopicRequestedHandler(createTopicProcess), &create.TopicRequested{}),
		messaging.RegisterMessageHandler(config.TopicNameSelfService, "topic_deletion_requested", del.NewTopicRequestedHandler(deleteTopicProcess), &del.TopicDeletionRequested{}),
	)

	m := NewMain(logger, config, consumer)

	logger.Information("Running")

	if err := m.Run(ctx); err != nil {
		logger.Error(err, "Exit reason {Reason}", err.Error())
		os.Exit(1)
	}

	logger.Information("Done!")
}

func getLogger(config *Configuration) logging.Logger {
	return logging.NewLogger(logging.LoggerOptions{
		IsProduction: config.IsProduction(),
		AppName:      config.ApplicationName,
	})
}

func getDatabase(config *Configuration, logger logging.Logger) *storage.Database {
	db, err := storage.NewDatabase(config.DbConnectionString, logger)
	if err != nil {
		panic(err)
	}
	return db
}

func getConfluentClient(logger logging.Logger, config *Configuration, db *storage.Database) *confluent.Client {
	clusters, err := db.GetClusters(context.TODO())
	if err != nil {
		panic(err)
	}

	cache := storage.NewClusterCache(clusters)

	return confluent.NewClient(logger, config.CreateCloudApiAccess(), cache)
}

func getVault(config *Configuration, logger logging.Logger) *vault.Vault {
	vaultCfg, err := config.CreateVaultConfig()
	if err != nil {
		panic(err)
	}

	awsClient, err := vault.NewVaultClient(logger, vaultCfg)
	if err != nil {
		panic(err)
	}
	return awsClient
}

func getOutboxFactory(logger logging.Logger, options ...messaging.OutboxOption) messaging.OutboxFactory {
	outbox, err := messaging.ConfigureOutbox(logger, options...)
	if err != nil {
		panic(err)
	}
	return outbox
}

func getConsumer(logger logging.Logger, broker string, groupId string, options ...messaging.ConsumerOption) messaging.Consumer {
	consumer, err := messaging.ConfigureConsumer(logger, broker, groupId, options...)
	if err != nil {
		panic(err)
	}
	return consumer
}

type Main struct {
	Logger        logging.Logger
	Consumer      messaging.Consumer
	MetricsServer *metrics.Server
}

func NewMain(logger logging.Logger, config *Configuration, consumer messaging.Consumer) *Main {
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
	//var once sync.Once
	//defer once.Do(cleanup)

	g.Go(func() error {
		return m.Consumer.Start(ctx)
	})

	g.Go(func() error {
		// wait until cancelled
		<-ctx.Done()

		m.Logger.Warning("!!! STOPPING CONSUMER !!!")

		//once.Do(cleanup)
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
