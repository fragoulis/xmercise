// Package company defines the company aggregate and its invariants.
package company

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
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

// ParseType converts an API or database string into a company type.
func ParseType(value string) (Type, error) {
	t := Type(value)
	if !t.Valid() {
		return "", fmt.Errorf("invalid company type %q", value)
	}

	return t, nil
}

// State is the serializable company state used by adapters.
type State struct {
	ID             uuid.UUID
	Name           string
	Description    *string
	EmployeesCount int
	Registered     bool
	Type           Type
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
	state := State{
		ID:             input.ID,
		Name:           input.Name,
		Description:    cloneStringPtr(input.Description),
		EmployeesCount: input.EmployeesCount,
		Registered:     input.Registered,
		Type:           input.Type,
		CreatedAt:      utc(input.CreatedAt),
		UpdatedAt:      utc(input.CreatedAt),
	}

	c, err := FromState(state)
	if err != nil {
		return nil, CompanyCreatedEvent{}, err
	}

	return c, CompanyCreatedEvent{
		Company:    c.State(),
		OccurredAt: c.createdAt,
	}, nil
}

// FromState restores a company from trusted storage while rechecking invariants.
func FromState(state State) (*Company, error) {
	state.Description = cloneStringPtr(state.Description)
	state.CreatedAt = utc(state.CreatedAt)
	state.UpdatedAt = utc(state.UpdatedAt)

	if err := validateState(state); err != nil {
		return nil, err
	}

	return &Company{
		id:             state.ID,
		name:           state.Name,
		description:    state.Description,
		employeesCount: state.EmployeesCount,
		registered:     state.Registered,
		companyType:    state.Type,
		createdAt:      state.CreatedAt,
		updatedAt:      state.UpdatedAt,
	}, nil
}

// Update patches a company and returns the matching domain event.
func (c *Company) Update(input UpdateInput) (CompanyUpdatedEvent, error) {
	state := c.State()

	if input.Name != nil {
		state.Name = *input.Name
	}
	if input.Description.Present {
		state.Description = cloneStringPtr(input.Description.Value)
	}
	if input.EmployeesCount != nil {
		state.EmployeesCount = *input.EmployeesCount
	}
	if input.Registered != nil {
		state.Registered = *input.Registered
	}
	if input.Type != nil {
		state.Type = *input.Type
	}
	state.UpdatedAt = utc(input.UpdatedAt)

	updated, err := FromState(state)
	if err != nil {
		return CompanyUpdatedEvent{}, err
	}

	*c = *updated

	return CompanyUpdatedEvent{
		Company:    c.State(),
		OccurredAt: c.updatedAt,
	}, nil
}

// Deleted returns the domain event for a hard delete.
func (c *Company) Deleted(at time.Time) (CompanyDeletedEvent, error) {
	occurredAt := utc(at)
	if occurredAt.IsZero() {
		return CompanyDeletedEvent{}, ValidationError{
			Violations: []Violation{
				{Field: "occurred_at", Message: "is required"},
			},
		}
	}

	return CompanyDeletedEvent{
		CompanyID:  c.id,
		OccurredAt: occurredAt,
	}, nil
}

// State returns a copy of the company state.
func (c *Company) State() State {
	return State{
		ID:             c.id,
		Name:           c.name,
		Description:    cloneStringPtr(c.description),
		EmployeesCount: c.employeesCount,
		Registered:     c.registered,
		Type:           c.companyType,
		CreatedAt:      c.createdAt,
		UpdatedAt:      c.updatedAt,
	}
}

func validateState(state State) error {
	var violations []Violation

	if state.ID == uuid.Nil {
		violations = append(violations, Violation{Field: "id", Message: "is required"})
	}
	if strings.TrimSpace(state.Name) == "" {
		violations = append(violations, Violation{Field: "name", Message: "is required"})
	} else if utf8.RuneCountInString(state.Name) > MaxNameLength {
		violations = append(violations, Violation{Field: "name", Message: "must be at most 15 characters"})
	}
	if state.Description != nil && utf8.RuneCountInString(*state.Description) > MaxDescriptionLength {
		violations = append(violations, Violation{Field: "description", Message: "must be at most 3000 characters"})
	}
	if state.EmployeesCount < 0 {
		violations = append(violations, Violation{Field: "employees_count", Message: "must be greater than or equal to 0"})
	}
	if !state.Type.Valid() {
		violations = append(violations, Violation{Field: "type", Message: "is invalid"})
	}
	if state.CreatedAt.IsZero() {
		violations = append(violations, Violation{Field: "created_at", Message: "is required"})
	}
	if state.UpdatedAt.IsZero() {
		violations = append(violations, Violation{Field: "updated_at", Message: "is required"})
	}
	if !state.CreatedAt.IsZero() && !state.UpdatedAt.IsZero() && state.UpdatedAt.Before(state.CreatedAt) {
		violations = append(violations, Violation{Field: "updated_at", Message: "must not be before created_at"})
	}

	if len(violations) > 0 {
		return ValidationError{Violations: violations}
	}

	return nil
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	copy := *value
	return &copy
}

func utc(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}

	return value.UTC()
}
