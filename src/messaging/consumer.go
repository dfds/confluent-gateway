package messaging

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/dfds/confluent-gateway/logging"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"time"
)

func ConfigureConsumer(logger logging.Logger, broker string, groupId string, options ...ConsumerOption) (Consumer, error) {
	registry := NewMessageRegistry()
	deserializer := NewDefaultDeserializer(registry)

	cfg := &consumerConfig{
		options:  ConsumerOptions{},
		registry: registry,
	}

	options = append([]ConsumerOption{
		setBroker(broker),
		setGroupId(groupId),
	}, options...)

	for _, option := range options {
		if err := option.apply(cfg); err != nil {
			return nil, err
		}
	}

	dispatcher := NewDispatcher(registry, deserializer)

	consumerOptions := cfg.options
	consumerOptions.Topics = registry.GetTopics()

	consumer, err := NewConsumer(logger, dispatcher, consumerOptions)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

type consumerConfig struct {
	options  ConsumerOptions
	registry MessageRegistry
}

type ConsumerOption interface {
	apply(cfg *consumerConfig) error
}

type brokerOption struct{ broker string }

func (o brokerOption) apply(cfg *consumerConfig) error {
	if len(o.broker) > 0 {
		cfg.options.Broker = o.broker
		return nil
	}
	return ErrNoBroker
}

func setBroker(broker string) ConsumerOption {
	return brokerOption{broker: broker}
}

var ErrNoBroker = errors.New("no bootstrap.servers specified")

type groupIdOption struct{ groupId string }

func (o groupIdOption) apply(cfg *consumerConfig) error {
	if len(o.groupId) > 0 {
		cfg.options.GroupId = o.groupId
		return nil
	}
	return ErrNoGroupId
}

func setGroupId(groupId string) ConsumerOption {
	return groupIdOption{groupId: groupId}
}

var ErrNoGroupId = errors.New("no group.id specified")

type credentialsOption struct{ credentials *ConsumerCredentials }

func (o credentialsOption) apply(cfg *consumerConfig) error {
	cfg.options.Credentials = o.credentials
	return nil
}

func WithCredentials(credentials *ConsumerCredentials) ConsumerOption {
	return credentialsOption{credentials: credentials}
}

type messageHandlerOption struct {
	topicName string
	eventType string
	handler   MessageHandler
	message   interface{}
}

func (o messageHandlerOption) apply(cfg *consumerConfig) error {
	return cfg.registry.RegisterMessageHandler(o.topicName, o.eventType, o.handler, o.message)
}

type nopHandler struct {
	logger logging.Logger
}

func NewNopHandler(logger logging.Logger) MessageHandler {
	return &nopHandler{logger: logger}
}

func (h *nopHandler) Handle(ctx context.Context, msgContext MessageContext) error {
	h.logger.Information("Ignoring unregistered message")
	return nil
}

type Nop struct {
}

func RegisterMessageHandler(topicName string, eventType string, handler MessageHandler, message interface{}) ConsumerOption {
	return messageHandlerOption{
		topicName: topicName,
		eventType: eventType,
		handler:   handler,
		message:   message,
	}
}

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

			c.logger.Error(err, "[START] Fatal error when consumer {GroupId} is fetching next message", c.groupId)
			return err
		}

		// *************************************************************************
		// handle message here!
		json := string(m.Value)
		c.logger.Information("[START] Message received: {Message}", json)

		if err := c.dispatcher.Dispatch(ctx, RawMessage{Data: m.Value}); err != nil {
			c.logger.Error(err, "[START] Consumer {GroupId} could not dispatch {Offset} on topic {Topic}", c.groupId, fmt.Sprint(m.Offset), m.Topic)
			return err
		}
		// *************************************************************************

		if err := c.kafkaReader.CommitMessages(ctx, m); err != nil {
			c.logger.Error(err, "[START] Consumer {GroupId} could not commit offset {Offset} on topic {Topic}", c.groupId, fmt.Sprint(m.Offset), m.Topic)
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
		c.logger.Error(err, "[STOP] Error while closing kafka reader for consumer {GroupId}", c.groupId)
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
