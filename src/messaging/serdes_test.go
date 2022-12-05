package messaging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDefaultDeserializer(t *testing.T) {
	sut := NewDefaultDeserializer(NewMessageRegistry())

	assert.IsType(t, &deserializer{}, sut)
}

func TestDeserialize(t *testing.T) {
	registry := NewMessageRegistry()
	_ = registry.RegisterMessageHandler("some_topic", "event", &dummyMessageHandler{}, &someMessage{})
	sut := NewDefaultDeserializer(registry)

	got, err := sut.Deserialize(RawMessage{
		Data:    []byte(`{"messageId":"id","type":"event","someHeader": "someValue","data":{"name":"some-name"}}`),
		Headers: map[string]string{"anotherHeader": "anotherValue"},
	})

	assert.NoError(t, err)
	assert.Equal(t, &IncomingMessage{
		MessageId: "id",
		Type:      "event",
		Headers:   map[string]string{"someHeader": "someValue", "anotherHeader": "anotherValue"},
		Message: &someMessage{
			Name: "some-name",
		},
	}, got)
}

func TestDeserializeOverrideRawMessageHeaders(t *testing.T) {
	registry := NewMessageRegistry()
	_ = registry.RegisterMessageHandler("some_topic", "event", &dummyMessageHandler{}, &someMessage{})
	sut := NewDefaultDeserializer(registry)

	got, err := sut.Deserialize(RawMessage{
		Data:    []byte(`{"messageId":"id","type":"event","h": "val","data":{"name":"new-name"}}`),
		Headers: map[string]string{"h": "old"},
	})

	assert.NoError(t, err)
	assert.Equal(t, &IncomingMessage{
		MessageId: "id",
		Type:      "event",
		Headers:   map[string]string{"h": "val"},
		Message: &someMessage{
			Name: "new-name",
		},
	}, got)

}

func TestDeserializeWithError(t *testing.T) {
	registry := NewMessageRegistry()
	_ = registry.RegisterMessageHandler("some_topic", "event", &dummyMessageHandler{}, &someMessage{})

	tests := []struct {
		name string
		msg  RawMessage
	}{
		{
			name: "unknown message type",
			msg:  RawMessage{Data: []byte(`{"messageId":"id","type":"???","data":{"name":"new-name"}}`)},
		},
		{
			name: "malformed json",
			msg:  RawMessage{Data: []byte(`{,}`)},
		},
		{
			name: "wrong json type",
			msg:  RawMessage{Data: []byte(`{"messageId":1}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDefaultDeserializer(registry)

			got, err := d.Deserialize(tt.msg)

			assert.Error(t, err)
			assert.Nil(t, got)
		})
	}
}

type someMessage struct {
	Name string `json:"name"`
}
