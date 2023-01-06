package models

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"time"
)

type Topic struct {
	Id               uuid.UUID `gorm:"type:uuid;primarykey"`
	CapabilityRootId CapabilityRootId
	ClusterId        ClusterId
	Name             string
	Partitions       int
	Retention        int64
	CreatedAt        time.Time
}

func NewTopic(capabilityRootId CapabilityRootId, clusterId ClusterId, topic TopicDescription) *Topic {
	return &Topic{
		Id:               uuid.NewV4(),
		CapabilityRootId: capabilityRootId,
		ClusterId:        clusterId,
		Name:             topic.Name,
		Partitions:       topic.Partitions,
		Retention:        topic.RetentionInMs(),
		CreatedAt:        time.Now(),
	}
}

func (*Topic) TableName() string {
	return "topic"
}

type TopicDescription struct {
	Name       string
	Partitions int
	Retention  time.Duration
}

type Retention func(*TopicDescription) error

func RetentionFromDuration(retention time.Duration) Retention {
	return func(topic *TopicDescription) error {
		topic.Retention = retention
		return nil
	}
}

func RetentionFromMs(retention int64) Retention {
	return func(topic *TopicDescription) error {
		topic.Retention = time.Duration(retention) * time.Millisecond
		return nil
	}
}

const foreverString = "-1"

func RetentionFromString(retention string) Retention {
	return func(topic *TopicDescription) error {
		if retention == foreverString {
			topic.Retention = -1 * time.Millisecond
			return nil
		}

		duration, err := time.ParseDuration(retention)
		if err != nil {
			return fmt.Errorf("unable to parse retention: %w", err)
		}

		topic.Retention = duration

		return nil
	}
}

func NewTopicDescription(topicName string, partitions int, retention Retention) (TopicDescription, error) {
	topic := TopicDescription{
		Name:       topicName,
		Partitions: partitions,
	}

	err := retention(&topic)

	return topic, err
}

func (t *TopicDescription) RetentionInMs() int64 {
	return int64(t.Retention / time.Millisecond)
}
