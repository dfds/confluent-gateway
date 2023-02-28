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

func NewTopicDescription(topicName string, partitions int, retention Retention) (TopicDescription, error) {
	topic := TopicDescription{
		Name:       topicName,
		Partitions: partitions,
	}

	err := retention.apply(&topic)

	return topic, err
}

func (t *TopicDescription) RetentionInMs() int64 {
	return int64(t.Retention / time.Millisecond)
}

type Retention interface {
	apply(*TopicDescription) error
}

type fromDuration struct{ retention time.Duration }

func (r fromDuration) apply(topic *TopicDescription) error {
	topic.Retention = r.retention
	return nil
}

func RetentionFromDuration(retention time.Duration) Retention {
	return fromDuration{retention: retention}
}

type fromMs struct{ retention int64 }

func (r fromMs) apply(topic *TopicDescription) error {
	topic.Retention = time.Duration(r.retention) * time.Millisecond
	return nil

}

func RetentionFromMs(retention int64) Retention {
	return fromMs{retention: retention}
}

const foreverString = "-1"

type fromString struct{ retention string }

func (r fromString) apply(topic *TopicDescription) error {
	if r.retention == foreverString {
		topic.Retention = -1 * time.Millisecond
		return nil
	}

	d, err := time.ParseDuration(r.retention)
	if err != nil {
		return fmt.Errorf("unable to parse retention: %w", err)
	}

	topic.Retention = d

	return nil
}

func RetentionFromString(retention string) Retention {
	return fromString{retention: retention}
}
