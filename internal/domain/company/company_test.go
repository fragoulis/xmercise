package company

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewCreatesCompanyAndEvent(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("offset", 2*60*60))
	description := "shipping"
	id := uuid.New()

	company, event, err := New(CreateInput{
		ID:             id,
		Name:           "Acme",
		Description:    &description,
		EmployeesCount: 7,
		Registered:     true,
		Type:           TypeCorporations,
		CreatedAt:      now,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	state := company.State()
	if state.ID != id {
		t.Fatalf("id = %s, want %s", state.ID, id)
	}
	if state.CreatedAt.Location() != time.UTC {
		t.Fatalf("created location = %v, want UTC", state.CreatedAt.Location())
	}
	if state.UpdatedAt != state.CreatedAt {
		t.Fatalf("updated_at = %s, want created_at %s", state.UpdatedAt, state.CreatedAt)
	}
	if event.Type() != EventTypeCompanyCreated {
		t.Fatalf("event type = %s, want %s", event.Type(), EventTypeCompanyCreated)
	}
	if event.AggregateID() != id {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), id)
	}
}

func TestNewRejectsInvalidCompany(t *testing.T) {
	description := strings.Repeat("a", MaxDescriptionLength+1)
	_, _, err := New(CreateInput{
		ID:             uuid.Nil,
		Name:           strings.Repeat("a", MaxNameLength+1),
		Description:    &description,
		EmployeesCount: -1,
		Registered:     false,
		Type:           Type("LLC"),
		CreatedAt:      time.Time{},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %T, want ValidationError", err)
	}

	wantFields := []string{"id", "name", "description", "employees_count", "type", "created_at", "updated_at"}
	for _, field := range wantFields {
		if !hasViolation(validationErr, field) {
			t.Fatalf("missing violation for %q in %#v", field, validationErr.Violations)
		}
	}
}

func TestUpdatePatchesCompany(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	company, _, err := New(CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           TypeCorporations,
		CreatedAt:      createdAt,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	name := "Workers Coop"
	employeesCount := 11
	registered := false
	companyType := TypeCooperative
	updatedAt := createdAt.Add(time.Hour)

	event, err := company.Update(UpdateInput{
		Name:           &name,
		Description:    DescriptionPatch{Present: true, Value: nil},
		EmployeesCount: &employeesCount,
		Registered:     &registered,
		Type:           &companyType,
		UpdatedAt:      updatedAt,
	})
	if err != nil {
		t.Fatalf("update company: %v", err)
	}

	state := company.State()
	if state.Name != name {
		t.Fatalf("name = %q, want %q", state.Name, name)
	}
	if state.Description != nil {
		t.Fatalf("description = %q, want nil", *state.Description)
	}
	if state.EmployeesCount != employeesCount {
		t.Fatalf("employees count = %d, want %d", state.EmployeesCount, employeesCount)
	}
	if state.Registered != registered {
		t.Fatalf("registered = %t, want %t", state.Registered, registered)
	}
	if state.Type != companyType {
		t.Fatalf("type = %s, want %s", state.Type, companyType)
	}
	if event.Type() != EventTypeCompanyUpdated {
		t.Fatalf("event type = %s, want %s", event.Type(), EventTypeCompanyUpdated)
	}
}

func TestDeletedReturnsEvent(t *testing.T) {
	company, _, err := New(CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           TypeCorporations,
		CreatedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	event, err := company.Deleted(time.Now())
	if err != nil {
		t.Fatalf("deleted event: %v", err)
	}

	if event.Type() != EventTypeCompanyDeleted {
		t.Fatalf("event type = %s, want %s", event.Type(), EventTypeCompanyDeleted)
	}
	if event.AggregateID() != company.State().ID {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), company.State().ID)
	}
}

func hasViolation(err ValidationError, field string) bool {
	for _, violation := range err.Violations {
		if violation.Field == field {
			return true
		}
	}

	return false
}
