package confluent

//type (
//	Client interface {
//		CreateServiceAccount(name string, description string) main.ServiceAccountId
//		CreateACLEntry(clusterId main.ClusterId, serviceAccountId main.ServiceAccountId, entry main.AclEntry)
//		CreateApiKey(clusterId main.ClusterId, serviceAccountId main.ServiceAccountId) ApiKey
//		CreateTopic(clusterId main.ClusterId, name string, partitions int, retention int)
//	}
//	client struct {
//		provider CredentialsProvider
//	}
//	topicConfig struct {
//		name  string `json:"name"`
//		value string `json:"value"`
//	}
//	createTopicRequest struct {
//		topicName         string        `json:"topic_name"`
//		partitionCount    int           `json:"partition_count"`
//		replicationFactor int           `json:"replication_factor"`
//		configs           []topicConfig `json:"configs"`
//	}
//)

//func CreateTopic(credentials CredentialsProvider, clusterId ClusterId, name string, partitions int, retention int) {
//	if clusterId == nil {
//		fmt.Println("cluster id is nil")
//		return
//	}
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
//}

type client struct {
}

//
//func NewConfluentClient(provider CredentialsProvider) Client {
//	return &client{provider: provider}
//}
