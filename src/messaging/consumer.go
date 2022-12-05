package messaging

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Consumer interface {
	Start(ctx context.Context)
	Stop()
}

type consumer struct {
	logger           logging.Logger
	groupId          string
	kafkaReader      *kafka.Reader
	isStarted        bool
	cancellation     *context.CancelFunc
	interruptChannel chan os.Signal
}

func (c *consumer) Start(ctx context.Context) {
	cancellationContext, cancel := context.WithCancel(ctx)
	c.cancellation = &cancel
	c.isStarted = true

	for c.isStarted {
		c.logger.Trace("Consumer {GroupId} is waiting for next message...", c.groupId)

		m, err := c.kafkaReader.FetchMessage(cancellationContext)
		if err != nil {
			if !c.isStarted {
				c.logger.Debug("Waiting for messages has been cancelled for consumer {GroupId}", c.groupId)
				c.isStarted = false
				break
			}
			c.logger.Error(&err, "Fatal error when consumer {GroupId} is fetching next message", c.groupId)
		}

		// *************************************************************************
		// handle message here!
		json := string(m.Value)
		c.logger.Information("Message received: {Message}", json)
		// *************************************************************************

		err = c.kafkaReader.CommitMessages(cancellationContext, m)
		if err != nil {
			c.logger.Error(&err, "Consumer {GroupId} could not commit offset {Offset} on topic {Topic}", c.groupId, fmt.Sprint(m.Offset), m.Topic)
			break
		}

		c.logger.Debug("Consumer {GroupId} has committed offset {Offset} on topic {Topic}", c.groupId, fmt.Sprint(m.Offset), m.Topic)
	}

	c.logger.Debug("Consumer {GroupId} has ended its consumer loop!", c.groupId)
}

func (c *consumer) Stop() {
	c.logger.Debug("Trying to stop consumer {GroupId}...", c.groupId)

	if c.isStarted {
		c.isStarted = false

		c.logger.Trace("Cancelling the consumer loop context for consumer {GroupId}", c.groupId)
		(*c.cancellation)()

		c.logger.Trace("Closing internal kafka reader for consumer {GroupId}", c.groupId)
		c.kafkaReader.Close()

		c.logger.Information("Consumer {GroupId} has now been stopped!", c.groupId)
	} else {
		c.logger.Debug("Consumer {GroupId} has not been started!", c.groupId)
	}
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

func NewConsumer(logger logging.Logger, options ConsumerOptions) (Consumer, error) {
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
			logger.Trace(fmt.Sprintf(s, i))
		}),
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Error(nil, fmt.Sprintf(s, i))
		}),
		Dialer: &dialer,
	})

	consumer := consumer{
		logger:           logger,
		groupId:          options.GroupId,
		kafkaReader:      reader,
		interruptChannel: make(chan os.Signal, 1),
	}

	signal.Notify(consumer.interruptChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-consumer.interruptChannel
		logger.Information("Caught signal {Signal}, stopping consumer {GroupId}...", sig.String(), consumer.groupId)
		consumer.Stop()
	}()

	return &consumer, nil
}
