package messaging

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/segmentio/kafka-go"
)

type Producer interface {
	Produce(context context.Context, msg RawOutgoingMessage) error
}

type RawOutgoingMessage struct {
	Topic        string
	PartitionKey string
	Headers      map[string]string
	Payload      string
}

func convertToTransportMessage(msg RawOutgoingMessage) kafka.Message {
	var headers []kafka.Header
	for key, value := range msg.Headers {
		headers = append(
			headers,
			kafka.Header{
				Key:   key,
				Value: []byte(value),
			},
		)
	}

	return kafka.Message{
		Topic:   msg.Topic,
		Key:     []byte(msg.PartitionKey),
		Value:   []byte(msg.Payload),
		Headers: headers,
	}
}

func (p *realProducer) Produce(ctx context.Context, msg RawOutgoingMessage) error {
	p.logger.Trace("Producing outgoing message {OutgoingMessage}", fmt.Sprintf("%v", msg))
	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.broker),
		RequiredAcks: kafka.RequireAll,
	}
	defer writer.Close()

	return writer.WriteMessages(ctx, convertToTransportMessage(msg))
}

type realProducer struct {
	logger logging.Logger
	broker string
}

type ProducerOptions struct {
	Broker string
}

func NewProducer(logger logging.Logger, options ProducerOptions) Producer {
	return &realProducer{
		logger: logger,
		broker: options.Broker,
	}
}
