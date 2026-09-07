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

// StoredState is the persisted company state used by storage adapters.
type StoredState struct {
	ID             uuid.UUID
	Name           string
	Description    *string
	EmployeesCount int
	Registered     bool
	Type           Type
	CreatedAt      time.Time
	UpdatedAt      time.Time
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

	createdAt := time.Now()
	state := StoredState{
		ID:             id,
		Name:           input.Name,
		Description:    copyDescription(input.Description),
		EmployeesCount: input.EmployeesCount,
		Registered:     input.Registered,
		Type:           input.Type,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}
	if err := validateStoredState(state); err != nil {
		return nil, CompanyCreatedEvent{}, err
	}

	c := fromStoredState(state)

	return c, CompanyCreatedEvent{
		CompanyID:  c.id,
		OccurredAt: c.createdAt,
	}, nil
}

// NewFromDB recreates a company from persisted state.
func NewFromDB(state StoredState) *Company {
	return fromStoredState(state)
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

	state := StoredState{
		ID:             c.id,
		Name:           name,
		Description:    description,
		EmployeesCount: employeesCount,
		Registered:     registered,
		Type:           companyType,
		CreatedAt:      c.createdAt,
		UpdatedAt:      time.Now(),
	}
	if err := validateStoredState(state); err != nil {
		return CompanyUpdatedEvent{}, err
	}

	*c = *fromStoredState(state)

	return CompanyUpdatedEvent{
		CompanyID:  c.id,
		OccurredAt: c.updatedAt,
	}, nil
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

func fromStoredState(state StoredState) *Company {
	return &Company{
		id:             state.ID,
		name:           state.Name,
		description:    copyDescription(state.Description),
		employeesCount: state.EmployeesCount,
		registered:     state.Registered,
		companyType:    state.Type,
		createdAt:      state.CreatedAt,
		updatedAt:      state.UpdatedAt,
	}
}

func validateStoredState(state StoredState) error {
	var violations []Violation

	if state.ID == uuid.Nil {
		violations = append(violations, Violation{
			Field:   "id",
			Message: "is required",
		})
	}
	if strings.TrimSpace(state.Name) == "" {
		violations = append(violations, Violation{
			Field:   "name",
			Message: "is required",
		})
	} else if utf8.RuneCountInString(state.Name) > MaxNameLength {
		violations = append(violations, Violation{
			Field:   "name",
			Message: "must be at most 15 characters",
		})
	}
	if state.Description != nil && utf8.RuneCountInString(*state.Description) > MaxDescriptionLength {
		violations = append(violations, Violation{
			Field:   "description",
			Message: "must be at most 3000 characters",
		})
	}
	if state.EmployeesCount < 0 {
		violations = append(violations, Violation{
			Field:   "employees_count",
			Message: "must be greater than or equal to 0",
		})
	}
	if state.CreatedAt.IsZero() {
		violations = append(violations, Violation{
			Field:   "created_at",
			Message: "is required",
		})
	}
	if state.UpdatedAt.IsZero() {
		violations = append(violations, Violation{
			Field:   "updated_at",
			Message: "is required",
		})
	}
	if !state.CreatedAt.IsZero() && !state.UpdatedAt.IsZero() && state.UpdatedAt.Before(state.CreatedAt) {
		violations = append(violations, Violation{
			Field:   "updated_at",
			Message: "must not be before created_at",
		})
	}

	if len(violations) > 0 {
		return ValidationError{
			Violations: violations,
		}
	}

	return nil
}

func copyDescription(description *string) *string {
	if description == nil {
		return nil
	}

	return lo.ToPtr(*description)
}
