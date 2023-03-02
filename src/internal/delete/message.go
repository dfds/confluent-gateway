package create

type TopicDeletionRequested struct {
	CapabilityId string `json:"capabilityRootId"`
	ClusterId    string `json:"clusterId"`
	TopicName    string `json:"topicName"`
}

type TopicDeleted struct {
	partitionKey string
	CapabilityId string `json:"capabilityRootId"`
	ClusterId    string `json:"clusterId"`
	TopicName    string `json:"topicName"`
}

func (t *TopicDeleted) PartitionKey() string {
	return t.partitionKey
}
