package models

type Topic struct {
	Name       string
	Partitions int
	Retention  int
}

func NewTopic(topicName string, partitions int, retention int) Topic {
	return Topic{
		Name:       topicName,
		Partitions: partitions,
		Retention:  retention,
	}
}
