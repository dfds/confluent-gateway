package httpslistener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Shopify/sarama"
	c "github.com/dfds/confluent-gateway/configuration"
	"github.com/gin-gonic/gin"
)

const (
	prodClusterName = "lkc-4npj6"
	devClusterName  = "lkc-3wqzw"
)

type HttpsListener struct {
	Server *http.Server
}

type brokerConfiguration struct {
	broker   string
	username string
	password string
}

func getBrokerConfig(config *c.Configuration, cluster string) (*brokerConfiguration, error) {
	switch cluster {
	case "prod", prodClusterName:
		return &brokerConfiguration{
			broker:   config.KafkaClusterProdBroker,
			username: config.KafkaClusterProdUserName,
			password: config.KafkaClusterProdPassword,
		}, nil
	case "dev", devClusterName:
		return &brokerConfiguration{
			broker:   config.KafkaClusterDevBroker,
			username: config.KafkaClusterDevUserName,
			password: config.KafkaClusterDevPassword,
		}, nil
	default:
		return nil, errors.New("cluster not known")
	}
}

func NewServer(config *c.Configuration) (*HttpsListener, error) {
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	router.GET("/cluster/:cluster/topic/:topic/consumers", func(c *gin.Context) {
		cluster := c.Param("cluster")
		topic := c.Param("topic")
		brokerConfig, err := getBrokerConfig(config, cluster)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
		}
		kafkaConfig := sarama.NewConfig()
		if brokerConfig.username != "" && brokerConfig.password != "" {
			kafkaConfig.Net.SASL.Enable = true
			kafkaConfig.Net.SASL.User = brokerConfig.username
			kafkaConfig.Net.SASL.Password = brokerConfig.password
		}
		fmt.Printf("Broker: %s\n", brokerConfig.broker)
		client, err := sarama.NewClient([]string{brokerConfig.broker}, kafkaConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
		}
		adm, err := sarama.NewClusterAdminFromClient(client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
		}
		defer adm.Close()

		partition_ids, err := client.Partitions(topic)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
		}

		groups, err := adm.ListConsumerGroups()
		if err != nil {
			panic(err)
		}

		consumers := []string{}
		for group := range groups {
			offsets, err := adm.ListConsumerGroupOffsets(group, map[string][]int32{topic: partition_ids})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{})
			}
			if offsets != nil {
				if offsets.Blocks[topic] != nil && offsets.Blocks[topic][0] != nil {
					if offsets.Blocks[topic][0].Offset > 0 {
						consumers = append(consumers, group)
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"topic":     topic,
			"consumers": consumers,
		})
	})

	addr := "127.0.0.1:8080"
	if config.IsProduction() {
		addr = ":443"
	}

	return &HttpsListener{
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}, nil
}

func (s *HttpsListener) Start(ctx context.Context) error {
	if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("listen: %s\n", err)
	}
	return nil
}

func (s *HttpsListener) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return s.Server.Shutdown(ctx)
}
