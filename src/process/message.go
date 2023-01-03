package process

type TopicRequested struct {
	CapabilityRootId string `json:"capabilityRootId"`
	ClusterId        string `json:"clusterId"`
	TopicName        string `json:"topicName"`
	Partitions       int    `json:"partitions"`
	Retention        string `json:"retention"`
}

type TopicProvisioned struct {
	partitionKey     string
	CapabilityRootId string `json:"capabilityRootId"`
	ClusterId        string `json:"clusterId"`
	TopicName        string `json:"topicName"`
}

func (t *TopicProvisioned) PartitionKey() string {
	return t.partitionKey
}

type TopicProvisioningBegun struct {
	partitionKey     string
	CapabilityRootId string `json:"capabilityRootId"`
	ClusterId        string `json:"clusterId"`
	TopicName        string `json:"topicName"`
}

func (t *TopicProvisioningBegun) PartitionKey() string {
	return t.partitionKey
}
