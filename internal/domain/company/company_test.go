package company_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/fragoulis/xmercise/internal/domain/company"
)

func TestNewCreatesCompanyAndEvent(t *testing.T) {
	description := "shipping"
	id := uuid.New()

	c, event, err := company.New(company.CreateInput{
		ID:             id,
		Name:           "Acme",
		Description:    &description,
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	if c.ID() != id {
		t.Fatalf("id = %s, want %s", c.ID(), id)
	}
	if c.CreatedAt().IsZero() {
		t.Fatal("created_at is zero")
	}
	if c.UpdatedAt() != c.CreatedAt() {
		t.Fatalf("updated_at = %s, want created_at %s", c.UpdatedAt(), c.CreatedAt())
	}
	if event.Type() != company.EventTypeCompanyCreated {
		t.Fatalf("event type = %s, want %s", event.Type(), company.EventTypeCompanyCreated)
	}
	if event.AggregateID() != id {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), id)
	}
}

func TestNewGeneratesID(t *testing.T) {
	c, event, err := company.New(company.CreateInput{
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	if c.ID() == uuid.Nil {
		t.Fatal("id was not generated")
	}
	if event.AggregateID() != c.ID() {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), c.ID())
	}
}

func TestNewRejectsInvalidCompany(t *testing.T) {
	description := strings.Repeat("a", company.MaxDescriptionLength+1)
	_, _, err := company.New(company.CreateInput{
		Name:           strings.Repeat("a", company.MaxNameLength+1),
		Description:    &description,
		EmployeesCount: -1,
		Registered:     false,
		Type:           company.Type("LLC"),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr company.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %T, want ValidationError", err)
	}

	wantFields := []string{"name", "description", "employees_count"}
	for _, field := range wantFields {
		if !hasViolation(validationErr, field) {
			t.Fatalf("missing violation for %q in %#v", field, validationErr.Violations)
		}
	}
	if hasViolation(validationErr, "type") {
		t.Fatalf("type validation belongs to the service: %#v", validationErr.Violations)
	}
}

func TestNewFromDBRecreatesPersistedState(t *testing.T) {
	description := "shipping"
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("offset", 2*60*60))
	updatedAt := createdAt.Add(time.Hour)
	id := uuid.New()

	c := company.NewFromDB(company.StoredState{
		ID:             id,
		Name:           "Acme",
		Description:    &description,
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	})
	description = "changed"

	loadedDescription, ok := c.Description()
	if !ok || loadedDescription != "shipping" {
		t.Fatalf("description = %q, %t; want shipping, true", loadedDescription, ok)
	}
	if c.CreatedAt() != createdAt {
		t.Fatalf("created_at = %s, want %s", c.CreatedAt(), createdAt)
	}
	if c.UpdatedAt() != updatedAt {
		t.Fatalf("updated_at = %s, want %s", c.UpdatedAt(), updatedAt)
	}
}

func TestUpdatePatchesCompany(t *testing.T) {
	c, _, err := company.New(company.CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	name := "Workers Coop"
	employeesCount := 11
	registered := false
	companyType := company.TypeCooperative

	event, err := c.Update(company.UpdateInput{
		Name:           &name,
		Description:    company.DescriptionPatch{Present: true, Value: nil},
		EmployeesCount: &employeesCount,
		Registered:     &registered,
		Type:           &companyType,
	})
	if err != nil {
		t.Fatalf("update company: %v", err)
	}

	if c.Name() != name {
		t.Fatalf("name = %q, want %q", c.Name(), name)
	}
	if _, ok := c.Description(); ok {
		t.Fatal("description is present, want cleared")
	}
	if c.EmployeesCount() != employeesCount {
		t.Fatalf("employees count = %d, want %d", c.EmployeesCount(), employeesCount)
	}
	if c.Registered() != registered {
		t.Fatalf("registered = %t, want %t", c.Registered(), registered)
	}
	if c.Type() != companyType {
		t.Fatalf("type = %s, want %s", c.Type(), companyType)
	}
	if event.Type() != company.EventTypeCompanyUpdated {
		t.Fatalf("event type = %s, want %s", event.Type(), company.EventTypeCompanyUpdated)
	}
}

func TestDeletedReturnsEvent(t *testing.T) {
	c, _, err := company.New(company.CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	event := c.Deleted()

	if event.Type() != company.EventTypeCompanyDeleted {
		t.Fatalf("event type = %s, want %s", event.Type(), company.EventTypeCompanyDeleted)
	}
	if event.AggregateID() != c.ID() {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), c.ID())
	}
}

func hasViolation(err company.ValidationError, field string) bool {
	for _, violation := range err.Violations {
		if violation.Field == field {
			return true
		}
	}

	return false
}
