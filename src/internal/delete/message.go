package delete

type TopicDeletionRequested struct {
	TopicId string `json:"kafkaTopicId"`
}

type TopicDeleted struct {
	TopicId string `json:"kafkaTopicId"`
}

func (t *TopicDeleted) PartitionKey() string {
	return t.TopicId
}
