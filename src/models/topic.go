package models

import "time"

type Topic struct {
	Name       string
	Partitions int
	Retention  time.Duration
}

func NewTopic(topicName string, partitions int, retention time.Duration) Topic {
	return Topic{
		Name:       topicName,
		Partitions: partitions,
		Retention:  retention,
	}
}
