package messaging

import (
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertToTransportMessageAssignsExpectedTopic(t *testing.T) {
	result := convertToTransportMessage(OutgoingMessage{Topic: "foo"})
	assert.Equal(t, "foo", result.Topic)
}

func TestConvertToTransportMessageAssignsExpectedPartitionKey(t *testing.T) {
	result := convertToTransportMessage(OutgoingMessage{PartitionKey: "foo"})
	assert.Equal(t, []byte("foo"), result.Key)
}

func TestConvertToTransportMessageAssignsExpectedHeadersWhenEmpty(t *testing.T) {
	result := convertToTransportMessage(OutgoingMessage{})
	assert.Empty(t, result.Headers)
}

func TestConvertToTransportMessageAssignsExpectedHeadersWhenContainingSingle(t *testing.T) {
	result := convertToTransportMessage(OutgoingMessage{Headers: map[string]string{"foo": "bar"}})
	expected := []kafka.Header{{Key: "foo", Value: []byte("bar")}}
	assert.Equal(t, expected, result.Headers)
}

func TestConvertToTransportMessageAssignsExpectedHeadersWhenContainingMultiple(t *testing.T) {
	result := convertToTransportMessage(OutgoingMessage{Headers: map[string]string{"foo": "bar", "baz": "qux"}})
	expected := []kafka.Header{
		{Key: "foo", Value: []byte("bar")},
		{Key: "baz", Value: []byte("qux")},
	}
	assert.Equal(t, expected, result.Headers)
}

func TestConvertToTransportMessageAssignsExpectedValue(t *testing.T) {
	result := convertToTransportMessage(OutgoingMessage{Payload: "foo"})
	assert.Equal(t, []byte("foo"), result.Value)
}
