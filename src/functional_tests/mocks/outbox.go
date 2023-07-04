package mocks

import (
	"github.com/dfds/confluent-gateway/messaging"
)

type OutboxRepository interface {
	AddToOutbox(entry *messaging.OutboxEntry) error
}
