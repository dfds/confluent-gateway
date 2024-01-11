package schema

type MessageContractRequested struct {
	MessageContractId string `json:"messageContractId"`
	TopicId           string `json:"kafkaTopicId"`
	MessageType       string `json:"messageType"`
	Description       string `json:"description"`
	Schema            string `json:"schema"`
	SchemaVersion     int32  `json:"schemaVersion"`
}

type SchemaRegistered struct {
	MessageContractId string `json:"messageContractId"`
}

func (t *SchemaRegistered) PartitionKey() string {
	return t.MessageContractId
}

type SchemaRegistrationFailed struct {
	MessageContractId string `json:"messageContractId"`
	Reason            string `json:"reason"`
}

func (t *SchemaRegistrationFailed) PartitionKey() string {
	return t.MessageContractId
}
