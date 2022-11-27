package confluent

import (
	"github.com/dfds/confluent-gateway/models"
)

type createTopicRequest struct {
	TopicName         string `json:"topic_name"`
	PartitionCount    int    `json:"partition_count"`
	ReplicationFactor int    `json:"replication_factor"`
	Configs           []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"configs"`
}

type client struct {
	cloudApiAccess    *models.CloudApiAccess
	clusterRepository *models.ClusterRepository
}

func (c *client) CreateServiceAccount(name string, description string) models.ServiceAccountId {
	//TODO implement me
	panic("implement me")
}

func (c *client) CreateACLEntry(clusterId models.ClusterId, serviceAccountId models.ServiceAccountId, entry models.AclDefinition) {
	//TODO implement me
	panic("implement me")
}

func (c *client) CreateApiKey(clusterId models.ClusterId, serviceAccountId models.ServiceAccountId) models.ApiKey {
	//TODO implement me
	panic("implement me")
}

func (c *client) CreateTopic(clusterId models.ClusterId, name string, partitions int, retention int) {

	//if clusterId = nil {
	//	fmt.Println("cluster id is nil")
	//	return
	//}
	//
	//	url := fmt.Sprintf("/kafka/v3/clusters/%s/topics", clusterId.String())
	//
	//	payload, _ := json.Marshal(createTopicRequest{
	//		topicName:         name,
	//		partitionCount:    partitions,
	//		replicationFactor: 3,
	//		configs: []topicConfig{{
	//			name:  "retention.ms",
	//			value: string(retention),
	//		}},
	//	})
	//
	//	//_, _ := http.Post(
	//	//	url,
	//	//	"application/json",
	//	//	bytes.NewBuffer(payload),
	//	//)
}

func NewConfluentClient(cloudApiAccess *models.CloudApiAccess, clusterRepository *models.ClusterRepository) models.ConfluentClient {
	return &client{cloudApiAccess: cloudApiAccess, clusterRepository: clusterRepository}
}
