package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/confluent"
	"github.com/dfds/confluent-gateway/http/metrics"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
	"github.com/dfds/confluent-gateway/storage"
	"github.com/dfds/confluent-gateway/vault"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const dsn = "host=localhost user=postgres password=p dbname=db port=5432 sslmode=disable"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	logger := logging.NewLogger(logging.LoggerOptions{
		IsProduction: false,
		AppName:      "lala",
	})

	db, err := storage.NewDatabase(dsn, logger)
	if err != nil {
		panic(err)
	}

	confluentClient := confluent.NewConfluentClient(confluent.CloudApiAccess{
		ApiEndpoint: "http://localhost:5051",
		Username:    "user",
		Password:    "pass",
	}, db)

	config, err := vault.NewTestConfig("http://localhost:5051/aws-ssm-put")
	if err != nil {
		panic(err)
	}

	awsClient, err := vault.NewVaultClient(logger, config)
	if err != nil {
		panic(err)
	}

	process := models.NewTopicCreationProcess(db, confluentClient, awsClient)

	registry := messaging.NewMessageRegistry()
	deserializer := messaging.NewDefaultDeserializer(registry)
	if err := registry.RegisterMessageHandler("hello", "topic_requested", NewTopicRequestedHandler(process), &TopicRequested{}); err != nil {
		panic(err)
	}
	dispatcher := messaging.NewDispatcher(registry, deserializer)

	logger.Information("New consumer")
	consumer, _ := messaging.NewConsumer(logger, dispatcher, messaging.ConsumerOptions{
		Broker:      "localhost:9092",
		GroupId:     "test-consumer-1",
		Topics:      []string{"hello"}, //registry.GetTopics(),
		Credentials: nil,
	})

	m := NewMain(logger, consumer)

	logger.Information("Running")

	if err := m.Run(ctx); err != nil {
		logger.Error(&err, "Exit reason {Reason}", err.Error())
	}

	logger.Information("Done!")

	//r := gin.Default()
	//r.GET("/ping", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "pong",
	//	})
	//})
	//_ = r.Run() // listen and serve on 0.0.0.0:8080

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

	g.Go(func() error {
		//return errors.New("FAILED")
		<-gCtx.Done()
		fmt.Println("loop done")
		return nil
	})

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
			m.Logger.Error(&err, err.Error())
		}
	}
	if m.MetricsServer != nil {
		err := m.MetricsServer.Close()
		if err != nil {
			m.Logger.Error(&err, err.Error())
		}
	}
}

// region TopicRequestedHandler

type TopicRequestedHandler struct {
	process *models.TopicCreationProcess
}

func NewTopicRequestedHandler(process *models.TopicCreationProcess) messaging.MessageHandler {
	return &TopicRequestedHandler{process: process}
}

func (h *TopicRequestedHandler) Handle(msgContext messaging.MessageContext) error {

	switch cmd := msgContext.Message().(type) {

	case *TopicRequested:

		fmt.Printf(
			"TopicRequested:\n"+
				" Capability: %s\n"+
				" Cluster:    %s\n"+
				" Topic:      %s\n"+
				" Partitions: %d\n"+
				" Retention:  %d\n",
			cmd.CapabilityRootId, cmd.ClusterId, cmd.TopicName, cmd.Partitions, cmd.Retention)

		return h.process.ProcessLogic(context.TODO(), models.NewTopicHasBeenRequested{
			CapabilityRootId: cmd.CapabilityRootId,
			ClusterId:        cmd.ClusterId,
			TopicName:        cmd.TopicName,
			Partitions:       cmd.Partitions,
			Retention:        cmd.Retention,
		})

	default:
		log.Fatalf("Unknown message %#v", cmd)
	}

	return nil
}

type TopicRequested struct {
	CapabilityRootId string `json:"capabilityRootId"`
	ClusterId        string `json:"clusterId"`
	TopicName        string `json:"topicName"`
	Partitions       int    `json:"partitions"`
	Retention        int    `json:"retention"`
}

// endregion
