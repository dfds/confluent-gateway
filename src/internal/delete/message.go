package delete

type TopicDeletionRequested struct {
	TopicId string `json:"topicId"`
}

type TopicDeleted struct {
	TopicId string `json:"topicId"`
}

func (t *TopicDeleted) PartitionKey() string {
	return t.TopicId
}
