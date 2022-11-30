package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/dfds/confluent-gateway/messaging"
	"log"
)

const dsn = "host=localhost user=postgres password=p dbname=db port=5432 sslmode=disable"

func main() {
	//db, err := database.NewDatabase(dsn)
	//if err != nil {
	//	panic(err)
	//}
	//
	//process := models.NewTopicCreationProcess(db, &mocks.MockClient{
	//	ServiceAccountId: "sa-some",
	//	ApiKey: models.ApiKey{
	//		Username: "u-some",
	//		Password: "p-another",
	//	},
	//}, &mocks.MockAwsClient{})
	//
	//if err := process.ProcessLogic(context.TODO(), models.NewTopicHasBeenRequested{
	//	CapabilityRootId: "new-cap",
	//	ClusterId:        "cluster-id",
	//	TopicName:        "new-topic",
	//	Partitions:       1,
	//	Retention:        -1,
	//}); err != nil {
	//	panic(err)
	//}

	//r := gin.Default()
	//r.GET("/ping", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "pong",
	//	})
	//})
	//_ = r.Run() // listen and serve on 0.0.0.0:8080

	registry := messaging.NewMessageRegistry()
	deserializer := messaging.NewDefaultDeserializer(registry)
	if err := registry.RegisterMessageHandler("topic_requested", NewTopicRequestedHandler(), &TopicRequested{}); err != nil {
		log.Fatal(err)
	}
	dispatcher := messaging.NewDispatcher(registry, deserializer)

	j := `
{
	"messageId": "message-id",
	"type": "topic_requested",
	"data": {
		"capabilityRootId": "new-cap",
		"clusterId": "cluster-id",
		"topicName": "new-topic",
		"partitions": 1,
		"retention": -1
	}
}
`
	data := []byte(j)

	if err := dispatcher.Dispatch(messaging.RawMessage{Data: data}); err != nil {
		panic(err)
	}

	log := logging.NewLogger(logging.LoggerOptions{
		IsProduction: false,
		AppName:      "lala",
	})

	consumer, _ := messaging.NewConsumer(log, messaging.ConsumerOptions{
		Broker:      "localhost:9092",
		GroupId:     "test-consumer-1",
		Topics:      []string{"hello"},
		Credentials: nil,
	})

	log.Information("Starting consumer...")
	consumer.Start(context.Background())
}

// region TopicRequestedHandler

type TopicRequestedHandler struct{}

func NewTopicRequestedHandler() messaging.MessageHandler {
	return &TopicRequestedHandler{}
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
		return nil

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
