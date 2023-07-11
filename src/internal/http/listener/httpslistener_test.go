package httpslistener_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/dfds/confluent-gateway/configuration"
	httpslistener "github.com/dfds/confluent-gateway/internal/http/listener"
	"github.com/stretchr/testify/assert"
)

const (
	topic1 = "i.exist.and.people.listen.to.me"
	topic2 = "i.also.exist.and.a.thing.listens"

	consumer1 = "consumer1"
	consumer2 = "consumer2"
	consumer3 = "consumer3"
)

type mockConsumerGroupHandler struct{}

func (m *mockConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error { return nil }
func (m *mockConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error       { return nil }
func (m *mockConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		session.MarkMessage(message, "")
	}
	return nil
}

type responseData struct {
	Topic     string   `json:"topic"`
	Consumers []string `json:"consumers"`
}

func TestConsumerCount(t *testing.T) {
	/*
	** SETUP
	 */
	ctx := context.Background()

	config := configuration.LoadInto("../../../", &configuration.Configuration{})

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	client, err := sarama.NewClient([]string{config.KafkaClusterProdBroker}, kafkaConfig)
	if err != nil {
		t.Fatal(err)
	}

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	// setup topics
	admin.CreateTopic(topic1, &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 1,
	}, false)
	admin.CreateTopic(topic2, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)

	// setup consumers groups and offsets
	cg1, err := sarama.NewConsumerGroupFromClient(consumer1, client)
	if err != nil {
		t.Fatal(err)
	}
	defer cg1.Close()
	go func(ctx context.Context) {
		if err := cg1.Consume(ctx, []string{topic1}, &mockConsumerGroupHandler{}); err != nil {
			panic(err)
		}
	}(ctx)

	cg2, err := sarama.NewConsumerGroupFromClient(consumer2, client)
	if err != nil {
		t.Fatal(err)
	}
	defer cg2.Close()
	go func(ctx context.Context) {
		if err := cg2.Consume(ctx, []string{topic1}, &mockConsumerGroupHandler{}); err != nil {
			panic(err)
		}
	}(ctx)

	cg3, err := sarama.NewConsumerGroupFromClient(consumer3, client)
	if err != nil {
		t.Fatal(err)
	}
	defer cg3.Close()
	go func(ctx context.Context) {
		if err := cg3.Consume(ctx, []string{topic2}, &mockConsumerGroupHandler{}); err != nil {
			panic(err)
		}
	}(ctx)

	// setup producers and generate events on topics
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		//t.Fatal(err)
		panic(err)
	}
	defer producer.Close()
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic1,
		Value: sarama.StringEncoder("hello world"),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic2,
		Value: sarama.StringEncoder("hello person"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// give consumers time to consume
	time.Sleep(time.Second * 60)
	ctx.Done()

	/*
	** VERIFICATION
	** TODO: Test multiple clusters
	 */

	tests := []struct {
		name           string
		topic          string
		cluster        string
		want_consumers []string
	}{
		{
			name:           "Non-existent topic returns empty consumers list",
			topic:          "i.do.not.exist",
			cluster:        "prod",
			want_consumers: []string{},
		},
		{
			name:           "Topic with no consumers returns empty consumers list",
			topic:          "i.exist.but.i.am.boring",
			cluster:        "prod",
			want_consumers: []string{},
		},
		{
			name:           "Topic with consumers returns non-empty consumers list",
			topic:          topic1,
			cluster:        "prod",
			want_consumers: []string{consumer1, consumer2},
		},
		{
			name:           "Another topic with consumers returns non-empty consumers list",
			topic:          topic2,
			cluster:        "prod",
			want_consumers: []string{consumer3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// perform test
			url := fmt.Sprintf("/cluster/%s/topic/%s/consumers", tt.cluster, tt.topic)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			server, err := httpslistener.NewServer(config)
			if err != nil {
				t.Fatal(err)
			}
			server.Server.Handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status OK, but received %d", recorder.Code)
			}

			var response responseData
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.topic, response.Topic, "expected topic %s, but received %s", tt.topic, response.Topic)
			sort.Strings(tt.want_consumers)
			sort.Strings(response.Consumers)
			assert.EqualValuesf(t, tt.want_consumers, response.Consumers, "expected consumers (%s), but received %s", tt.want_consumers, response.Consumers)
		})
	}
}
