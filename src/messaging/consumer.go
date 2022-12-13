package messaging

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"time"
)

type Consumer interface {
	Start(ctx context.Context) error
	Stop() error
}

type consumer struct {
	logger      logging.Logger
	groupId     string
	kafkaReader *kafka.Reader
	isStarted   bool
	dispatcher  Dispatcher
}

func (c *consumer) Start(ctx context.Context) error {
	c.isStarted = true

	for {
		c.logger.Trace("[START] Consumer {GroupId} is waiting for next message...", c.groupId)

		m, err := c.kafkaReader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				c.logger.Information("[START] Waiting for messages has been cancelled for consumer {GroupId}", c.groupId)
				return nil
			}

			c.logger.Error(&err, "[START] Fatal error when consumer {GroupId} is fetching next message", c.groupId)
			return err
		}

		// *************************************************************************
		// handle message here!
		json := string(m.Value)
		c.logger.Information("[START] Message received: {Message}", json)

		if err := c.dispatcher.Dispatch(ctx, RawMessage{Data: m.Value}); err != nil {
			c.logger.Error(&err, "[START] Consumer {GroupId} could not dispatch {Offset} on topic {Topic}", c.groupId, fmt.Sprint(m.Offset), m.Topic)
			return err
		}
		// *************************************************************************

		if err := c.kafkaReader.CommitMessages(ctx, m); err != nil {
			c.logger.Error(&err, "[START] Consumer {GroupId} could not commit offset {Offset} on topic {Topic}", c.groupId, fmt.Sprint(m.Offset), m.Topic)
			return err
		}

		c.logger.Debug("[START] Consumer {GroupId} has committed offset {Offset} on topic {Topic}", c.groupId, fmt.Sprint(m.Offset), m.Topic)
	}
}

func (c *consumer) Stop() error {
	c.logger.Debug("[STOP] Trying to stop consumer {GroupId}...", c.groupId)

	if !c.isStarted {
		c.logger.Debug("[STOP] Consumer {GroupId} has not been started!", c.groupId)
		return nil
	}

	c.isStarted = false

	c.logger.Trace("[STOP] Closing internal kafka reader for consumer {GroupId}", c.groupId)

	if err := c.kafkaReader.Close(); err != nil {
		c.logger.Error(&err, "[STOP] Error while closing kafka reader for consumer {GroupId}", c.groupId)
		return err
	}

	c.logger.Information("[STOP] Consumer {GroupId} has now been stopped!", c.groupId)
	return nil
}

type ConsumerCredentials struct {
	UserName string
	Password string
}

type ConsumerOptions struct {
	Broker      string
	GroupId     string
	Topics      []string
	Credentials *ConsumerCredentials
}

func NewConsumer(logger logging.Logger, dispatcher Dispatcher, options ConsumerOptions) (Consumer, error) {
	dialer := kafka.Dialer{
		Timeout:   10 * time.Second, // connection timeout could be taken in as part of options instead
		DualStack: true,
	}

	if options.Credentials != nil {
		dialer.TLS = &tls.Config{}
		dialer.SASLMechanism = plain.Mechanism{
			Username: options.Credentials.UserName,
			Password: options.Credentials.Password,
		}
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{options.Broker},
		GroupID:     options.GroupId,
		GroupTopics: options.Topics,
		Logger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Trace(fmt.Sprintf(s, i...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Error(nil, fmt.Sprintf(s, i...))
		}),
		Dialer: &dialer,
	})

	consumer := consumer{
		logger:      logger,
		groupId:     options.GroupId,
		kafkaReader: reader,
		dispatcher:  dispatcher,
	}

	return &consumer, nil
}
