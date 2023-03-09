package create

type MessageContractRequested struct {
	MessageContractId string `json:"messageContractId"`
	TopicId           string `json:"kafkaTopicId"`
	MessageType       string `json:"messageType"`
	Description       string `json:"description"`
	Schema            string `json:"schema"`
}

type SchemaRegistered struct {
	MessageContractId string `json:"messageContractId"`
}

func (t *SchemaRegistered) PartitionKey() string {
	return t.MessageContractId
}
