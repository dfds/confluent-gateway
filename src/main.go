package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dfds/confluent-gateway/configuration"
	"github.com/dfds/confluent-gateway/confluent"
	"github.com/dfds/confluent-gateway/http/metrics"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/process"
	"github.com/dfds/confluent-gateway/storage"
	"github.com/dfds/confluent-gateway/vault"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Configuration struct {
	ApplicationName           string `env:"CG_APPLICATION_NAME"`
	Environment               string `env:"CG_ENVIRONMENT"`
	ConfluentCloudApiUrl      string `env:"CG_CONFLUENT_CLOUD_API_URL"`
	ConfluentCloudApiUserName string `env:"CG_CONFLUENT_CLOUD_API_USERNAME"`
	ConfluentCloudApiPassword string `env:"CG_CONFLUENT_CLOUD_API_PASSWORD"`
	ConfluentUserApiUrl       string `env:"CG_CONFLUENT_USER_API_URL"`
	VaultApiUrl               string `env:"CG_VAULT_API_URL"`
	KafkaBroker               string `env:"DEFAULT_KAFKA_BOOTSTRAP_SERVERS"`
	KafkaUserName             string `env:"DEFAULT_KAFKA_SASL_USERNAME"`
	KafkaPassword             string `env:"DEFAULT_KAFKA_SASL_PASSWORD"`
	KafkaGroupId              string `env:"CG_KAFKA_GROUP_ID"`
	DbConnectionString        string `env:"CG_DB_CONNECTION_STRING"`
	TopicNameSelfService      string `env:"CG_TOPIC_NAME_SELF_SERVICE"`
	TopicNameProvisioning     string `env:"CG_TOPIC_NAME_PROVISIONING"`
}

// region configuration helper functions

func (c *Configuration) IsProduction() bool {
	return strings.EqualFold(c.Environment, "production")
}

func (c *Configuration) CreateConsumerCredentials() *messaging.ConsumerCredentials {
	if !c.IsProduction() {
		return nil
	}

	return &messaging.ConsumerCredentials{
		UserName: c.KafkaUserName,
		Password: c.KafkaPassword,
	}
}

func (c *Configuration) CreateVaultConfig() (*aws.Config, error) {
	if c.IsProduction() {
		return vault.NewDefaultConfig()
	} else {
		return vault.NewTestConfig(c.VaultApiUrl)
	}
}

// endregion

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// load configuration from .env and/or environment files
	var config = Configuration{}
	configuration.LoadInto(&config)

	logger := logging.NewLogger(logging.LoggerOptions{
		IsProduction: config.IsProduction(),
		AppName:      config.ApplicationName,
	})

	db, err := storage.NewDatabase(config.DbConnectionString, logger)
	if err != nil {
		panic(err)
	}

	confluentClient := confluent.NewClient(logger, confluent.CloudApiAccess{
		ApiEndpoint:     config.ConfluentCloudApiUrl,
		Username:        config.ConfluentCloudApiUserName,
		Password:        config.ConfluentCloudApiPassword,
		UserApiEndpoint: config.ConfluentUserApiUrl,
	}, db)

	vaultCfg, err := config.CreateVaultConfig()
	if err != nil {
		panic(err)
	}

	awsClient, err := vault.NewVaultClient(logger, vaultCfg)
	if err != nil {
		panic(err)
	}

	outgoingRegistry := messaging.NewOutgoingMessageRegistry()

	registration := outgoingRegistry.
		RegisterMessage("cloudengineering.confluentgateway.provisioning", "topic_provisioned", &process.TopicProvisioned{}).
		RegisterMessage("cloudengineering.confluentgateway.provisioning", "topic_provisioning_begun", &process.TopicProvisioningBegun{})

	if err := registration.Error; err != nil {
		panic(err)
	}

	newTopic := process.NewCreateTopicProcess(logger, db, confluentClient, awsClient, outgoingRegistry)

	registry := messaging.NewMessageRegistry()
	deserializer := messaging.NewDefaultDeserializer(registry)
	if err := registry.RegisterMessageHandler(config.TopicNameSelfService, "topic_requested", process.NewTopicRequestedHandler(newTopic), &process.TopicRequested{}).Error; err != nil {
		panic(err)
	}
	dispatcher := messaging.NewDispatcher(registry, deserializer)

	consumerOptions := messaging.ConsumerOptions{
		Broker:      config.KafkaBroker,
		GroupId:     config.KafkaGroupId,
		Credentials: config.CreateConsumerCredentials(),
		Topics:      registry.GetTopics(),
	}

	consumer, err := messaging.NewConsumer(logger, dispatcher, consumerOptions)
	if err != nil {
		panic(err)
	}

	m := NewMain(logger, consumer)

	logger.Information("Running")

	if err := m.Run(ctx); err != nil {
		logger.Error(err, "Exit reason {Reason}", err.Error())
		os.Exit(1)
	}

	logger.Information("Done!")
}

type Main struct {
	Logger        logging.Logger
	Consumer      messaging.Consumer
	MetricsServer *metrics.Server
}

func NewMain(logger logging.Logger, consumer messaging.Consumer) *Main {
	return &Main{
		Logger:        logger,
		Consumer:      consumer,
		MetricsServer: metrics.NewServer(logger),
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
