package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertToMessageReturnsExpectedFields(t *testing.T) {
	message := convertToMessage("1={one} 2={two}", "1", "2")
	expectedFields := []field{
		{
			name:  "one",
			value: "1",
		},
		{
			name:  "two",
			value: "2",
		},
	}
	assert.Equal(t, expectedFields, (*message).fields)
}

func TestConvertToMessageReturnsExpectedText(t *testing.T) {
	message := convertToMessage("1={one} 2={two}", "1", "2")
	assert.Equal(t, "1=1 2=2", message.text)
}
