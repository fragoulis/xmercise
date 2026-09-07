// Package company defines the company aggregate and its invariants.
package company

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

// MaxNameLength is the maximum company name length in Unicode code points.
const MaxNameLength = 15

// MaxDescriptionLength is the maximum company description length in Unicode code points.
const MaxDescriptionLength = 3000

// Type is the allowed company legal type.
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

// Valid reports whether the type is one of the supported values.
func (t Type) Valid() bool {
	switch t {
	case TypeCorporations, TypeNonProfit, TypeCooperative, TypeSoleProprietorship:
		return true
	default:
		return false
	}
}

// CreateInput contains data required to create a company.
type CreateInput struct {
	ID             uuid.UUID
	Name           string
	Description    *string
	EmployeesCount int
	Registered     bool
	Type           Type
	CreatedAt      time.Time
}

// UpdateInput contains patch data for an existing company.
type UpdateInput struct {
	Name           *string
	Description    DescriptionPatch
	EmployeesCount *int
	Registered     *bool
	Type           *Type
	UpdatedAt      time.Time
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
func New(input CreateInput) (*Company, CompanyCreatedEvent, error) {
	id := input.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	createdAt := utc(input.CreatedAt)
	c, err := Restore(
		id,
		input.Name,
		copyDescription(input.Description),
		input.EmployeesCount,
		input.Registered,
		input.Type,
		createdAt,
		createdAt,
	)
	if err != nil {
		return nil, CompanyCreatedEvent{}, err
	}

	return c, CompanyCreatedEvent{
		CompanyID:  c.id,
		OccurredAt: c.createdAt,
	}, nil
}

// Restore recreates a company from persistence.
func Restore(
	id uuid.UUID,
	name string,
	description *string,
	employeesCount int,
	registered bool,
	companyType Type,
	createdAt time.Time,
	updatedAt time.Time,
) (*Company, error) {
	description = copyDescription(description)
	createdAt = utc(createdAt)
	updatedAt = utc(updatedAt)

	if err := validate(id, name, description, employeesCount, createdAt, updatedAt); err != nil {
		return nil, err
	}

	return &Company{
		id:             id,
		name:           name,
		description:    description,
		employeesCount: employeesCount,
		registered:     registered,
		companyType:    companyType,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}, nil
}

// Update patches a company and returns the matching domain event.
func (c *Company) Update(input UpdateInput) (CompanyUpdatedEvent, error) {
	name := c.name
	description := copyDescription(c.description)
	employeesCount := c.employeesCount
	registered := c.registered
	companyType := c.companyType

	if input.Name != nil {
		name = *input.Name
	}
	if input.Description.Present {
		description = copyDescription(input.Description.Value)
	}
	if input.EmployeesCount != nil {
		employeesCount = *input.EmployeesCount
	}
	if input.Registered != nil {
		registered = *input.Registered
	}
	if input.Type != nil {
		companyType = *input.Type
	}

	updated, err := Restore(
		c.id,
		name,
		description,
		employeesCount,
		registered,
		companyType,
		c.createdAt,
		input.UpdatedAt,
	)
	if err != nil {
		return CompanyUpdatedEvent{}, err
	}

	*c = *updated

	return CompanyUpdatedEvent{
		CompanyID:  c.id,
		OccurredAt: c.updatedAt,
	}, nil
}

// Deleted returns the domain event for a hard delete.
func (c *Company) Deleted(at time.Time) CompanyDeletedEvent {
	return CompanyDeletedEvent{
		CompanyID:  c.id,
		OccurredAt: utc(at),
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
func (c *Company) Description() (string, bool) {
	if c.description == nil {
		return "", false
	}

	return *c.description, true
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

func validate(
	id uuid.UUID,
	name string,
	description *string,
	employeesCount int,
	createdAt time.Time,
	updatedAt time.Time,
) error {
	var violations []Violation

	if id == uuid.Nil {
		violations = append(violations, Violation{Field: "id", Message: "is required"})
	}
	if strings.TrimSpace(name) == "" {
		violations = append(violations, Violation{Field: "name", Message: "is required"})
	} else if utf8.RuneCountInString(name) > MaxNameLength {
		violations = append(violations, Violation{Field: "name", Message: "must be at most 15 characters"})
	}
	if description != nil && utf8.RuneCountInString(*description) > MaxDescriptionLength {
		violations = append(violations, Violation{Field: "description", Message: "must be at most 3000 characters"})
	}
	if employeesCount < 0 {
		violations = append(violations, Violation{Field: "employees_count", Message: "must be greater than or equal to 0"})
	}
	if createdAt.IsZero() {
		violations = append(violations, Violation{Field: "created_at", Message: "is required"})
	}
	if updatedAt.IsZero() {
		violations = append(violations, Violation{Field: "updated_at", Message: "is required"})
	}
	if !createdAt.IsZero() && !updatedAt.IsZero() && updatedAt.Before(createdAt) {
		violations = append(violations, Violation{Field: "updated_at", Message: "must not be before created_at"})
	}

	if len(violations) > 0 {
		return ValidationError{Violations: violations}
	}

	return nil
}

func copyDescription(description *string) *string {
	if description == nil {
		return nil
	}

	return lo.ToPtr(*description)
}

func utc(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}

	return value.UTC()
}
