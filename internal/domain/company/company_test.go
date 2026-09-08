package company_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/fragoulis/xmercise/internal/domain/company"
)

func TestNewCreatesCompanyAndEvent(t *testing.T) {
	description := "shipping"
	id := uuid.New()

	created, event := company.New(company.CreateInput{
		ID:             id,
		Name:           "Acme",
		Description:    &description,
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})

	if created.ID() != id {
		t.Fatalf("id = %s, want %s", created.ID(), id)
	}
	if created.CreatedAt().IsZero() {
		t.Fatal("created_at is zero")
	}
	if created.UpdatedAt() != created.CreatedAt() {
		t.Fatalf("updated_at = %s, want created_at %s", created.UpdatedAt(), created.CreatedAt())
	}
	if event.Type() != company.EventTypeCompanyCreated {
		t.Fatalf("event type = %s, want %s", event.Type(), company.EventTypeCompanyCreated)
	}
	if event.AggregateID() != id {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), id)
	}
}

func TestNewGeneratesID(t *testing.T) {
	created, event := company.New(company.CreateInput{
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})

	if created.ID() == uuid.Nil {
		t.Fatal("id was not generated")
	}
	if event.AggregateID() != created.ID() {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), created.ID())
	}
}

func TestNewFromDBRecreatesPersistedValues(t *testing.T) {
	description := "shipping"
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("offset", 2*60*60))
	updatedAt := createdAt.Add(time.Hour)
	id := uuid.New()

	created := company.NewFromDB(
		id,
		"Acme",
		&description,
		7,
		true,
		company.TypeCorporations,
		createdAt,
		updatedAt,
	)
	loadedDescription := created.Description()
	if loadedDescription == nil || *loadedDescription != description {
		t.Fatalf("description = %v; want %q", loadedDescription, description)
	}
	if created.CreatedAt() != createdAt {
		t.Fatalf("created_at = %s, want %s", created.CreatedAt(), createdAt)
	}
	if created.UpdatedAt() != updatedAt {
		t.Fatalf("updated_at = %s, want %s", created.UpdatedAt(), updatedAt)
	}
}

func TestUpdatePatchesCompany(t *testing.T) {
	created, _ := company.New(company.CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})

	name := "Workers Coop"
	employeesCount := 11
	registered := false
	companyType := company.TypeCooperative

	event := created.Update(company.UpdateInput{
		Name: &name,
		Description: company.DescriptionPatch{
			Present: true,
			Value:   nil,
		},
		EmployeesCount: &employeesCount,
		Registered:     &registered,
		Type:           &companyType,
	})

	if created.Name() != name {
		t.Fatalf("name = %q, want %q", created.Name(), name)
	}
	if created.Description() != nil {
		t.Fatal("description is present, want cleared")
	}
	if created.EmployeesCount() != employeesCount {
		t.Fatalf("employees count = %d, want %d", created.EmployeesCount(), employeesCount)
	}
	if created.Registered() != registered {
		t.Fatalf("registered = %t, want %t", created.Registered(), registered)
	}
	if created.Type() != companyType {
		t.Fatalf("type = %s, want %s", created.Type(), companyType)
	}
	if event.Type() != company.EventTypeCompanyUpdated {
		t.Fatalf("event type = %s, want %s", event.Type(), company.EventTypeCompanyUpdated)
	}
}

func TestDeletedReturnsEvent(t *testing.T) {
	created, _ := company.New(company.CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})

	event := created.Deleted()

	if event.Type() != company.EventTypeCompanyDeleted {
		t.Fatalf("event type = %s, want %s", event.Type(), company.EventTypeCompanyDeleted)
	}
	if event.AggregateID() != created.ID() {
		t.Fatalf("event aggregate id = %s, want %s", event.AggregateID(), created.ID())
	}
}
