package create

type TopicDeletionRequested struct {
	CapabilityId string `json:"capabilityRootId"`
	ClusterId    string `json:"clusterId"`
	TopicId      string `json:"topicId"`
	TopicName    string `json:"topicName"`
}

type TopicDeleted struct {
	CapabilityId string `json:"capabilityRootId"`
	ClusterId    string `json:"clusterId"`
	TopicId      string `json:"topicId"`
	TopicName    string `json:"topicName"`
}

func (t *TopicDeleted) PartitionKey() string {
	return t.TopicId
}
