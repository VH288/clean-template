package event

import "clean-template/internal/domain/sample/entity"

const (
	TypeCreated = "sample.created"
	TypeUpdated = "sample.updated"
	TypeDeleted = "sample.deleted"
)

// Envelope is the shared publish/consume contract for sample domain events.
type Envelope struct {
	EventID string         `json:"event_id"`
	Type    string         `json:"type"`
	Sample  *entity.Sample `json:"sample"`
}
