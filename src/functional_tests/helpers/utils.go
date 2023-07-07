package helpers

import (
	"encoding/json"
	"github.com/dfds/confluent-gateway/messaging"
	"github.com/stretchr/testify/require"
	"testing"
)

func RequireOutboxPayloadIsEqual(t *testing.T, outboxEntry *messaging.OutboxEntry, expectedType string) {

	type payload struct {
		Type string `json:"type"`
	}
	var data payload
	err := json.Unmarshal([]byte(outboxEntry.Payload), &data)
	require.NoError(t, err)

	require.Equal(t, expectedType, data.Type)
}
