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

	if company.ID() != id {
		t.Fatalf("id = %s, want %s", company.ID(), id)
	}
	if company.CreatedAt().Location() != time.UTC {
		t.Fatalf("created location = %v, want UTC", company.CreatedAt().Location())
	}
	if company.UpdatedAt() != company.CreatedAt() {
		t.Fatalf("updated_at = %s, want created_at %s", company.UpdatedAt(), company.CreatedAt())
	}
	if event.Type() != EventTypeCompanyCreated {
		t.Fatalf("event type = %s, want %s", event.Type(), EventTypeCompanyCreated)
	}
	if event.AggregateID() != id {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), id)
	}
}

func TestNewGeneratesID(t *testing.T) {
	company, event, err := New(CreateInput{
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           TypeCorporations,
		CreatedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	if company.ID() == uuid.Nil {
		t.Fatal("id was not generated")
	}
	if event.AggregateID() != company.ID() {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), company.ID())
	}
}

func TestNewRejectsInvalidCompany(t *testing.T) {
	description := strings.Repeat("a", MaxDescriptionLength+1)
	_, _, err := New(CreateInput{
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

	wantFields := []string{"name", "description", "employees_count", "created_at", "updated_at"}
	for _, field := range wantFields {
		if !hasViolation(validationErr, field) {
			t.Fatalf("missing violation for %q in %#v", field, validationErr.Violations)
		}
	}
	if hasViolation(validationErr, "type") {
		t.Fatalf("type validation belongs to the service: %#v", validationErr.Violations)
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

	if company.Name() != name {
		t.Fatalf("name = %q, want %q", company.Name(), name)
	}
	if _, ok := company.Description(); ok {
		t.Fatal("description is present, want cleared")
	}
	if company.EmployeesCount() != employeesCount {
		t.Fatalf("employees count = %d, want %d", company.EmployeesCount(), employeesCount)
	}
	if company.Registered() != registered {
		t.Fatalf("registered = %t, want %t", company.Registered(), registered)
	}
	if company.Type() != companyType {
		t.Fatalf("type = %s, want %s", company.Type(), companyType)
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

	event := company.Deleted(time.Now())

	if event.Type() != EventTypeCompanyDeleted {
		t.Fatalf("event type = %s, want %s", event.Type(), EventTypeCompanyDeleted)
	}
	if event.AggregateID() != company.ID() {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), company.ID())
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
