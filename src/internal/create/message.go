package create

type TopicRequested struct {
	TopicId          string `json:"topicId"`          // V1
	CapabilityRootId string `json:"capabilityRootId"` // V1
	ClusterId        string `json:"clusterId"`        // V1
	TopicName        string `json:"topicName"`        // V1

	KafkaTopicId   string `json:"kafkaTopicId"`   // V2
	CapabilityId   string `json:"capabilityId"`   // V2
	KafkaClusterId string `json:"kafkaClusterId"` // V2
	KafkaTopicName string `json:"kafkaTopicName"` // V2

	Partitions int    `json:"partitions"` // V*
	Retention  string `json:"retention"`  // V*
}

func (r *TopicRequested) GetCapabilityId() string {
	if len(r.CapabilityId) > 0 {
		return r.CapabilityId
	} else {
		return r.CapabilityRootId
	}
}

func (r *TopicRequested) GetClusterId() string {
	if len(r.KafkaClusterId) > 0 {
		return r.KafkaClusterId
	} else {
		return r.ClusterId
	}
}

func (r *TopicRequested) GetTopicId() string {
	if len(r.KafkaTopicId) > 0 {
		return r.KafkaTopicId
	} else {
		return r.TopicId
	}
}

func (r *TopicRequested) GetTopicName() string {
	if len(r.KafkaTopicName) > 0 {
		return r.KafkaTopicName
	} else {
		return r.TopicName
	}
}

type TopicProvisioned struct {
	TopicId      string `json:"topicId"`
	CapabilityId string `json:"capabilityRootId"`
	ClusterId    string `json:"clusterId"`
	TopicName    string `json:"topicName"`
}

func (t *TopicProvisioned) PartitionKey() string {
	return t.TopicId
}

type TopicProvisioningBegun struct {
	TopicId      string `json:"topicId"`
	CapabilityId string `json:"capabilityRootId"`
	ClusterId    string `json:"clusterId"`
	TopicName    string `json:"topicName"`
}

func (t *TopicProvisioningBegun) PartitionKey() string {
	return t.TopicId
}
