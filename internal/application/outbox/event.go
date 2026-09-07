// Package outbox defines application outbox messages.
package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event is an unpublished event stored transactionally with aggregate changes.
type Event struct {
	ID            uuid.UUID
	EventType     string
	AggregateType string
	AggregateID   uuid.UUID
	OccurredAt    time.Time
	Payload       json.RawMessage
	CreatedAt     time.Time
}
