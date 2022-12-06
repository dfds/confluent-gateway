package messaging

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewDefaultDeserializer(t *testing.T) {
	sut := NewDefaultDeserializer(&messageTypeRegistryStub{})

	assert.IsType(t, &deserializer{}, sut)
}

func TestDeserialize(t *testing.T) {
	sut := NewDefaultDeserializer(&messageTypeRegistryStub{&someMessage{}})

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
	sut := NewDefaultDeserializer(&messageTypeRegistryStub{&someMessage{}})

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
	tests := []struct {
		name     string
		registry MessageTypeRegistry
		msg      RawMessage
	}{
		{
			name:     "unknown message type",
			registry: &errorStub{errors.New("message type not found")},
			msg:      RawMessage{Data: []byte(`{"messageId":"id","type":"???","data":{"name":"new-name"}}`)},
		},
		{
			name:     "malformed json",
			registry: &messageTypeRegistryStub{&someMessage{}},
			msg:      RawMessage{Data: []byte(`{,}`)},
		},
		{
			name:     "wrong json type",
			registry: &messageTypeRegistryStub{&someMessage{}},
			msg:      RawMessage{Data: []byte(`{"messageId":1}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDefaultDeserializer(tt.registry)

			got, err := d.Deserialize(tt.msg)

			assert.Error(t, err)
			assert.Nil(t, got)
		})
	}
}

type someMessage struct {
	Name string `json:"name"`
}

type messageTypeRegistryStub struct {
	exampleMessage interface{}
}

func (s *messageTypeRegistryStub) GetMessageType(string) (reflect.Type, error) {
	return getMessageType(s.exampleMessage)
}

func (s *errorStub) GetMessageType(string) (reflect.Type, error) {
	return nil, s.error
}
