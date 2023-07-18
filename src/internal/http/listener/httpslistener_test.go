package httpslistener_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
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

	consumer1 = "consumer1-prod"
	consumer2 = "consumer2-prod"
	consumer3 = "consumer3-prod"
	consumer4 = "consumer4-dev"
)

type channelBuffer struct {
	cin  chan bool
	cout chan bool
}

func (cb *channelBuffer) Run() {
	go func() {
		<-cb.cin
		cb.cout <- true
	}()
}

type mockConsumerGroupHandler struct{}

func (m *mockConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error { return nil }
func (m *mockConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error       { return nil }
func (m *mockConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		session.MarkMessage(message, "")
		//session.Commit()
	}
	return nil
}

type responseData struct {
	Topic     string   `json:"topic"`
	Consumers []string `json:"consumers"`
}

func must[R interface{}](result R, err error) R {
	if err != nil {
		panic(err)
	}
	return result
}

func runConsumer(ctx context.Context, group sarama.ConsumerGroup, topic string) {
	go func() {
		if err := group.Consume(ctx, []string{topic}, &mockConsumerGroupHandler{}); err != nil {
			panic(err)
		}
	}()
}

func TestConsumerCount(t *testing.T) {
	/*
	** SETUP
	 */
	config := configuration.LoadInto("../../../", &configuration.Configuration{})

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	prodclient := must(sarama.NewClient([]string{config.KafkaClusterProdBroker}, kafkaConfig))
	prodadmin := must(sarama.NewClusterAdminFromClient(prodclient))
	defer prodadmin.Close()
	devclient := must(sarama.NewClient([]string{config.KafkaClusterDevBroker}, kafkaConfig))
	devadmin := must(sarama.NewClusterAdminFromClient(devclient))
	defer devadmin.Close()

	// setup topics
	prodadmin.CreateTopic(topic1, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)

	prodadmin.CreateTopic(topic2, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	devadmin.CreateTopic(topic2, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)

	//wg := &sync.WaitGroup{}
	ctx := context.Background()

	// setup consumers groups and offsets
	cg1 := must(sarama.NewConsumerGroupFromClient(consumer1, prodclient))
	defer cg1.Close()
	runConsumer(ctx, cg1, topic1)

	cg2 := must(sarama.NewConsumerGroupFromClient(consumer2, prodclient))
	defer cg2.Close()
	runConsumer(ctx, cg2, topic1)

	cg3 := must(sarama.NewConsumerGroupFromClient(consumer3, prodclient))
	defer cg3.Close()
	runConsumer(ctx, cg3, topic2)

	cg4 := must(sarama.NewConsumerGroupFromClient(consumer4, devclient))
	defer cg4.Close()
	runConsumer(ctx, cg4, topic2)

	// setup producers and generate events on topics
	prodproducer := must(sarama.NewSyncProducerFromClient(prodclient))
	defer prodproducer.Close()
	devproducer := must(sarama.NewSyncProducerFromClient(devclient))
	defer devproducer.Close()
	_, _, err := prodproducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic1,
		Value: sarama.StringEncoder("hello world"),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = prodproducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic2,
		Value: sarama.StringEncoder("hello person"),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = devproducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic2,
		Value: sarama.StringEncoder("hello development person"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// give consumers time to consume
	time.Sleep(time.Second * 120)
	ctx.Done()

	/*
	** VERIFICATION
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
		{
			name:           "A topic with consumers and duplicate name on another cluster returns non-empty consumers list",
			topic:          topic2,
			cluster:        "dev",
			want_consumers: []string{consumer4},
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
