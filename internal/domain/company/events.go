package company

import (
	"time"

	"github.com/google/uuid"
)

const aggregateTypeCompany = "company"

// EventType identifies a company domain event type.
type EventType string

const (
	// EventTypeCompanyCreated is emitted when a company is created.
	EventTypeCompanyCreated EventType = "company.created"
	// EventTypeCompanyUpdated is emitted when a company is updated.
	EventTypeCompanyUpdated EventType = "company.updated"
	// EventTypeCompanyDeleted is emitted when a company is deleted.
	EventTypeCompanyDeleted EventType = "company.deleted"
)

// Event is the common shape of company domain events.
type Event interface {
	Type() EventType
	AggregateType() string
	AggregateID() uuid.UUID
	Occurred() time.Time
}

// CompanyCreatedEvent describes a created company.
type CompanyCreatedEvent struct {
	CompanyID  uuid.UUID
	OccurredAt time.Time
}

// Type returns the event type.
func (e CompanyCreatedEvent) Type() EventType {
	return EventTypeCompanyCreated
}

// AggregateType returns the aggregate type.
func (e CompanyCreatedEvent) AggregateType() string {
	return aggregateTypeCompany
}

// AggregateID returns the aggregate ID.
func (e CompanyCreatedEvent) AggregateID() uuid.UUID {
	return e.CompanyID
}

// Occurred returns when the event happened.
func (e CompanyCreatedEvent) Occurred() time.Time {
	return e.OccurredAt
}

// CompanyUpdatedEvent describes an updated company.
type CompanyUpdatedEvent struct {
	CompanyID  uuid.UUID
	OccurredAt time.Time
}

// Type returns the event type.
func (e CompanyUpdatedEvent) Type() EventType {
	return EventTypeCompanyUpdated
}

// AggregateType returns the aggregate type.
func (e CompanyUpdatedEvent) AggregateType() string {
	return aggregateTypeCompany
}

// AggregateID returns the aggregate ID.
func (e CompanyUpdatedEvent) AggregateID() uuid.UUID {
	return e.CompanyID
}

// Occurred returns when the event happened.
func (e CompanyUpdatedEvent) Occurred() time.Time {
	return e.OccurredAt
}

// CompanyDeletedEvent describes a deleted company.
type CompanyDeletedEvent struct {
	CompanyID  uuid.UUID
	OccurredAt time.Time
}

// Type returns the event type.
func (e CompanyDeletedEvent) Type() EventType {
	return EventTypeCompanyDeleted
}

// AggregateType returns the aggregate type.
func (e CompanyDeletedEvent) AggregateType() string {
	return aggregateTypeCompany
}

// AggregateID returns the aggregate ID.
func (e CompanyDeletedEvent) AggregateID() uuid.UUID {
	return e.CompanyID
}

// Occurred returns when the event happened.
func (e CompanyDeletedEvent) Occurred() time.Time {
	return e.OccurredAt
}
