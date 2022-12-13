package process

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/dfds/confluent-gateway/models"
	"log"
)

type TopicRequestedHandler struct {
	process *models.TopicCreationProcess
}

func NewTopicRequestedHandler(db models.Database, confluent models.Confluent, vault models.Vault) messaging.MessageHandler {
	process := models.NewTopicCreationProcess(db, confluent, vault)
	return &TopicRequestedHandler{process: process}
}

func (h *TopicRequestedHandler) Handle(ctx context.Context, msgContext messaging.MessageContext) error {

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

		return h.process.ProcessLogic(ctx, models.NewTopicHasBeenRequested{
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
