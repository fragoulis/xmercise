// Package company defines the company aggregate.
package company

import (
	"time"

	"github.com/google/uuid"
)

// Type is a company legal type.
type Type string

const (
	// TypeCorporations identifies corporation companies.
	TypeCorporations Type = "Corporations"
	// TypeNonProfit identifies non-profit companies.
	TypeNonProfit Type = "NonProfit"
	// TypeCooperative identifies cooperative companies.
	TypeCooperative Type = "Cooperative"
	// TypeSoleProprietorship identifies sole proprietorship companies.
	TypeSoleProprietorship Type = "Sole Proprietorship"
)

// CreateInput contains data required to create a company.
type CreateInput struct {
	ID             uuid.UUID
	Name           string
	Description    *string
	EmployeesCount int
	Registered     bool
	Type           Type
}

// UpdateInput contains patch data for an existing company.
type UpdateInput struct {
	Name           *string
	Description    DescriptionPatch
	EmployeesCount *int
	Registered     *bool
	Type           *Type
}

// DescriptionPatch represents an omitted, cleared, or replaced description.
type DescriptionPatch struct {
	Present bool
	Value   *string
}

// Company is the company aggregate root.
type Company struct {
	id             uuid.UUID
	name           string
	description    *string
	employeesCount int
	registered     bool
	companyType    Type
	createdAt      time.Time
	updatedAt      time.Time
}

// New creates a company and returns the matching domain event.
func New(input CreateInput) (*Company, CompanyCreatedEvent) {
	id := input.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	createdAt := time.Now()
	company := &Company{
		id:             id,
		name:           input.Name,
		description:    input.Description,
		employeesCount: input.EmployeesCount,
		registered:     input.Registered,
		companyType:    input.Type,
		createdAt:      createdAt,
		updatedAt:      createdAt,
	}

	return company, CompanyCreatedEvent{
		CompanyID:  company.id,
		OccurredAt: company.createdAt,
	}
}

// NewFromDB recreates a company from persisted values.
func NewFromDB(
	id uuid.UUID,
	name string,
	description *string,
	employeesCount int,
	registered bool,
	companyType Type,
	createdAt time.Time,
	updatedAt time.Time,
) *Company {
	return &Company{
		id:             id,
		name:           name,
		description:    description,
		employeesCount: employeesCount,
		registered:     registered,
		companyType:    companyType,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// Update patches a company and returns the matching domain event.
func (c *Company) Update(input UpdateInput) CompanyUpdatedEvent {
	if input.Name != nil {
		c.name = *input.Name
	}
	if input.Description.Present {
		c.description = input.Description.Value
	}
	if input.EmployeesCount != nil {
		c.employeesCount = *input.EmployeesCount
	}
	if input.Registered != nil {
		c.registered = *input.Registered
	}
	if input.Type != nil {
		c.companyType = *input.Type
	}
	c.updatedAt = time.Now()

	return CompanyUpdatedEvent{
		CompanyID:  c.id,
		OccurredAt: c.updatedAt,
	}
}

// Deleted returns the domain event for a hard delete.
func (c *Company) Deleted() CompanyDeletedEvent {
	return CompanyDeletedEvent{
		CompanyID:  c.id,
		OccurredAt: time.Now(),
	}
}

// ID returns the company ID.
func (c *Company) ID() uuid.UUID {
	return c.id
}

// Name returns the company name.
func (c *Company) Name() string {
	return c.name
}

// Description returns the optional description.
func (c *Company) Description() *string {
	return c.description
}

// EmployeesCount returns the number of employees.
func (c *Company) EmployeesCount() int {
	return c.employeesCount
}

// Registered returns whether the company is registered.
func (c *Company) Registered() bool {
	return c.registered
}

// Type returns the company type.
func (c *Company) Type() Type {
	return c.companyType
}

// CreatedAt returns the creation time.
func (c *Company) CreatedAt() time.Time {
	return c.createdAt
}

// UpdatedAt returns the latest update time.
func (c *Company) UpdatedAt() time.Time {
	return c.updatedAt
}
