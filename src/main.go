package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
	"log"
)

const dsn = "host=localhost user=postgres password=p dbname=db port=5432 sslmode=disable"

func main() {
	logger := logging.NewLogger(logging.LoggerOptions{
		IsProduction: false,
		AppName:      "lala",
	})

	//db, err := database.NewDatabase(dsn, logger)
	//if err != nil {
	//	panic(err)
	//}
	//
	//confluentClient := confluent.NewConfluentClient(models.CloudApiAccess{
	//	ApiEndpoint: "http://localhost:5051",
	//	Username:    "user",
	//	Password:    "pass",
	//}, db)
	//
	//awsClient := &mocks.MockAwsClient{}
	//
	//process := models.NewTopicCreationProcess(db, confluentClient, awsClient)
	//
	////r := gin.Default()
	////r.GET("/ping", func(c *gin.Context) {
	////	c.JSON(200, gin.H{
	////		"message": "pong",
	////	})
	////})
	////_ = r.Run() // listen and serve on 0.0.0.0:8080
	//
	//registry := messaging.NewMessageRegistry()
	//deserializer := messaging.NewDefaultDeserializer(registry)
	//if err := registry.RegisterMessageHandler("hello", "topic_requested", NewTopicRequestedHandler(process), &TopicRequested{}); err != nil {
	//	panic(err)
	//}
	//dispatcher := messaging.NewDispatcher(registry, deserializer)
	//
	//consumer, _ := messaging.NewConsumer(logger, dispatcher, messaging.ConsumerOptions{
	//	Broker:      "localhost:9092",
	//	GroupId:     "test-consumer-1",
	//	Topics:      registry.GetTopics(),
	//	Credentials: nil,
	//})
	//
	//logger.Information("Starting consumer...")
	//consumer.Start(context.Background())

	logger.Information("DONE!")
	
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
