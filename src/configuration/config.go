package configuration

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/vault"
	"github.com/dfds/confluent-gateway/messaging"
)

type Configuration struct {
	ApplicationName                    string `env:"CG_APPLICATION_NAME"`
	Environment                        string `env:"CG_ENVIRONMENT"`
	ConfluentCloudApiUrl               string `env:"CG_CONFLUENT_CLOUD_API_URL"`
	ConfluentCloudApiUserName          string `env:"CG_CONFLUENT_CLOUD_API_USERNAME"`
	ConfluentCloudApiPassword          string `env:"CG_CONFLUENT_CLOUD_API_PASSWORD"`
	ConfluentUserApiUrl                string `env:"CG_CONFLUENT_USER_API_URL"`
	VaultApiUrl                        string `env:"CG_VAULT_API_URL"`
	KafkaBroker                        string `env:"DEFAULT_KAFKA_BOOTSTRAP_SERVERS"`
	KafkaUserName                      string `env:"DEFAULT_KAFKA_SASL_USERNAME"`
	KafkaPassword                      string `env:"DEFAULT_KAFKA_SASL_PASSWORD"`
	KafkaGroupId                       string `env:"CG_KAFKA_GROUP_ID"`
	DbConnectionString                 string `env:"CG_DB_CONNECTION_STRING"`
	TopicNameKafkaClusterAccess        string `env:"CG_TOPIC_NAME_KAFKA_CLUSTER_ACCESS"`
	TopicNameKafkaClusterAccessGranted string `env:"CG_TOPIC_NAME_KAFKA_CLUSTER_ACCESS_GRANTED"`
	TopicNameSelfService               string `env:"CG_TOPIC_NAME_SELF_SERVICE"`
	TopicNameProvisioning              string `env:"CG_TOPIC_NAME_PROVISIONING"`
	TopicNameMessageContract           string `env:"CG_TOPIC_NAME_MESSAGE_CONTRACT"`
	TopicNameSchema                    string `env:"CG_TOPIC_NAME_SCHEMA"`
	ApiHttpListenAddress               string `env:"CG_API_HTTP_LISTEN_ADDRESS"`
}

func (c *Configuration) IsProduction() bool {
	return strings.EqualFold(c.Environment, "production")
}

func (c *Configuration) CreateConsumerCredentials() *messaging.ConsumerCredentials {
	if !c.IsProduction() {
		return nil
	}

	return &messaging.ConsumerCredentials{
		UserName: c.KafkaUserName,
		Password: c.KafkaPassword,
	}
}

func (c *Configuration) CreateVaultConfig() (*aws.Config, error) {
	if c.IsProduction() {
		return vault.NewDefaultConfig()
	} else {
		return vault.NewTestConfig(c.VaultApiUrl)
	}
}

func (c *Configuration) CreateCloudApiAccess() confluent.CloudApiAccess {
	return confluent.CloudApiAccess{
		ApiEndpoint:     c.ConfluentCloudApiUrl,
		Username:        c.ConfluentCloudApiUserName,
		Password:        c.ConfluentCloudApiPassword,
		UserApiEndpoint: c.ConfluentUserApiUrl,
	}
}
