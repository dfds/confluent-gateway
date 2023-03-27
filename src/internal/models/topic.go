package models

import (
	"fmt"
	"time"
)

type Topic struct {
	Id           string `gorm:"primarykey"`
	CapabilityId CapabilityId
	ClusterId    ClusterId
	Name         string
	Partitions   int
	Retention    int64
	CreatedAt    time.Time
}

func NewTopic(capabilityId CapabilityId, clusterId ClusterId, topicId string, topic TopicDescription) *Topic {
	return &Topic{
		Id:           topicId,
		CapabilityId: capabilityId,
		ClusterId:    clusterId,
		Name:         topic.Name,
		Partitions:   topic.Partitions,
		Retention:    topic.RetentionInMs(),
		CreatedAt:    time.Now(),
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

type fromDuration struct{ duration time.Duration }

func (r fromDuration) apply(topic *TopicDescription) error {
	topic.Retention = r.duration
	return nil
}

func RetentionFromDuration(retention time.Duration) Retention {
	return fromDuration{duration: retention}
}

type fromMs struct{ ms int64 }

func (r fromMs) apply(topic *TopicDescription) error {
	topic.Retention = time.Duration(r.ms) * time.Millisecond
	return nil
}

func RetentionFromMs(retention int64) Retention {
	return fromMs{ms: retention}
}

const foreverString = "-1"

var stringToDuration = map[string]int64{
	"forever": -1,
	"1d":      86_400_000,
	"7d":      604_800_000,
	"31d":     2_678_400_000,
	"365d":    31_536_000_000,
}

type fromString struct{ s string }

func (r fromString) apply(topic *TopicDescription) error {
	if r.s == foreverString {
		topic.Retention = -1 * time.Millisecond
		return nil
	}

	retention, ok := stringToDuration[r.s]
	if ok {
		topic.Retention = time.Duration(retention) * time.Millisecond
		return nil
	}

	d, err := time.ParseDuration(r.s)
	if err != nil {
		return fmt.Errorf("unable to parse retention: %w", err)
	}

	topic.Retention = d

	return nil
}

func RetentionFromString(retention string) Retention {
	return fromString{s: retention}
}
