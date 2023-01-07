package create

type TopicDeletionRequested struct {
	CapabilityRootId string `json:"capabilityRootId"`
	ClusterId        string `json:"clusterId"`
	TopicName        string `json:"topicName"`
}

type TopicDeleted struct {
	partitionKey     string
	CapabilityRootId string `json:"capabilityRootId"`
	ClusterId        string `json:"clusterId"`
	TopicName        string `json:"topicName"`
}

func (t *TopicDeleted) PartitionKey() string {
	return t.partitionKey
}
