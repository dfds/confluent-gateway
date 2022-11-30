package messaging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMessageContext(t *testing.T) {
	sut := NewMessageContext(map[string]string{}, &dummyMessage{})

	assert.IsType(t, &messageContext{}, sut)
}

func TestHeaders(t *testing.T) {
	headers := map[string]string{}
	sut := NewMessageContext(headers, &dummyMessage{})

	assert.Equal(t, headers, sut.Headers())

}

func TestMessage(t *testing.T) {
	message := &dummyMessage{}
	sut := NewMessageContext(map[string]string{}, message)

	assert.Equal(t, message, sut.Message())

}
