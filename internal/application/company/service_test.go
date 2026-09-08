package company_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/samber/lo"

	appcompany "github.com/fragoulis/xmercise/internal/application/company"
	"github.com/fragoulis/xmercise/internal/application/outbox"
	companydomain "github.com/fragoulis/xmercise/internal/domain/company"
)

func TestServiceCreatePersistsCompanyAndOutboxEvent(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	service := appcompany.NewService(store)
	description := "shipping"

	created, err := service.Create(ctx, appcompany.CreateCommand{
		Name:           lo.ToPtr("Acme"),
		Description:    &description,
		EmployeesCount: lo.ToPtr(7),
		Registered:     lo.ToPtr(true),
		Type:           lo.ToPtr("Corporations"),
	})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	loaded, err := store.FindCompanyByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("find company: %v", err)
	}
	if loaded.Name() != "Acme" || loaded.CreatedAt().IsZero() {
		t.Fatalf("loaded company mismatch: %#v", loaded)
	}
	payload := decodePayload(t, store.outbox[0].Payload)
	if payload.EventType != string(companydomain.EventTypeCompanyCreated) {
		t.Fatalf("payload mismatch: %#v", payload)
	}
	if payload.Data["name"] != "Acme" {
		t.Fatalf("payload data = %#v, want company after-state", payload.Data)
	}
}

func TestServiceUpdatePatchesCompanyWithOutboxEvent(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	initial := mustCompany(t, companydomain.CreateInput{
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           companydomain.TypeCorporations,
	})
	store.companies[initial.ID()] = initial
	service := appcompany.NewService(store)

	name := "Workers Coop"
	updated, err := service.Update(ctx, appcompany.UpdateCommand{
		ID:   initial.ID().String(),
		Name: &name,
		Type: lo.ToPtr("Cooperative"),
	})
	if err != nil {
		t.Fatalf("update company: %v", err)
	}

	if updated.Name() != name || updated.Type() != companydomain.TypeCooperative {
		t.Fatalf("updated company mismatch: %#v", updated)
	}
	payload := decodePayload(t, store.outbox[0].Payload)
	if payload.EventType != string(companydomain.EventTypeCompanyUpdated) || payload.Data["name"] != name {
		t.Fatalf("payload mismatch: %#v", payload)
	}
}

func TestServiceDeleteRemovesCompanyWithOutboxEvent(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	created := mustCompany(t, companydomain.CreateInput{
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           companydomain.TypeCorporations,
	})
	store.companies[created.ID()] = created
	service := appcompany.NewService(store)

	err := service.Delete(ctx, appcompany.DeleteCommand{
		ID: created.ID().String(),
	})
	if err != nil {
		t.Fatalf("delete company: %v", err)
	}

	if _, err := store.FindCompanyByID(ctx, created.ID()); !errors.Is(err, appcompany.ErrCompanyNotFound) {
		t.Fatalf("find deleted error = %v, want %v", err, appcompany.ErrCompanyNotFound)
	}
	payload := decodePayload(t, store.outbox[0].Payload)
	if payload.EventType != string(companydomain.EventTypeCompanyDeleted) || payload.Data["id"] != created.ID().String() {
		t.Fatalf("payload mismatch: %#v", payload)
	}
}

func TestServiceFindOneReturnsCompany(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	created := mustCompany(t, companydomain.CreateInput{
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           companydomain.TypeCorporations,
	})
	store.companies[created.ID()] = created
	service := appcompany.NewService(store)

	loaded, err := service.FindOne(ctx, appcompany.FindOneQuery{
		ID: created.ID().String(),
	})
	if err != nil {
		t.Fatalf("find one: %v", err)
	}
	if loaded.ID() != created.ID() {
		t.Fatalf("id = %s, want %s", loaded.ID(), created.ID())
	}
}

func TestServiceValidatesCommands(t *testing.T) {
	ctx := context.Background()
	service := appcompany.NewService(newMemoryStore())
	longDescription := strings.Repeat("a", 3001)
	longName := strings.Repeat("é", 16)

	_, err := service.Create(ctx, appcompany.CreateCommand{
		Name:           &longName,
		Description:    &longDescription,
		EmployeesCount: lo.ToPtr(-1),
		Type:           lo.ToPtr("LLC"),
	})
	assertViolations(t, err, "name", "description", "employees_count", "registered", "type")

	_, err = service.Update(ctx, appcompany.UpdateCommand{
		Name:           lo.ToPtr(" "),
		Description:    companydomain.DescriptionPatch{Present: true, Value: &longDescription},
		EmployeesCount: lo.ToPtr(-1),
		Type:           lo.ToPtr("LLC"),
	})
	assertViolations(t, err, "id", "name", "description", "employees_count", "type")

	_, err = service.FindOne(ctx, appcompany.FindOneQuery{})
	assertViolations(t, err, "id")

	_, err = service.FindOne(ctx, appcompany.FindOneQuery{ID: "1"})
	assertViolations(t, err, "id")
}

type memoryStore struct {
	companies map[uuid.UUID]*companydomain.Company
	outbox    []outbox.Event
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		companies: make(map[uuid.UUID]*companydomain.Company),
	}
}

func (s *memoryStore) FindCompanyByID(ctx context.Context, id uuid.UUID) (*companydomain.Company, error) {
	_ = ctx
	found, ok := s.companies[id]
	if !ok {
		return nil, appcompany.ErrCompanyNotFound
	}

	return cloneCompany(found), nil
}

func (s *memoryStore) WithinTx(ctx context.Context, fn func(ctx context.Context, tx appcompany.Tx) error) error {
	tx := &memoryTx{
		companies: cloneCompanies(s.companies),
		outbox:    append([]outbox.Event(nil), s.outbox...),
	}
	if err := fn(ctx, tx); err != nil {
		return err
	}

	s.companies = tx.companies
	s.outbox = tx.outbox
	return nil
}

type memoryTx struct {
	companies map[uuid.UUID]*companydomain.Company
	outbox    []outbox.Event
}

func (tx *memoryTx) InsertCompany(ctx context.Context, c *companydomain.Company) error {
	_ = ctx
	for _, existing := range tx.companies {
		if existing.Name() == c.Name() {
			return appcompany.ErrCompanyNameTaken
		}
	}
	tx.companies[c.ID()] = cloneCompany(c)
	return nil
}

func (tx *memoryTx) FindCompanyByIDForUpdate(ctx context.Context, id uuid.UUID) (*companydomain.Company, error) {
	_ = ctx
	found, ok := tx.companies[id]
	if !ok {
		return nil, appcompany.ErrCompanyNotFound
	}

	return cloneCompany(found), nil
}

func (tx *memoryTx) UpdateCompany(ctx context.Context, c *companydomain.Company) error {
	_ = ctx
	if _, ok := tx.companies[c.ID()]; !ok {
		return appcompany.ErrCompanyNotFound
	}
	tx.companies[c.ID()] = cloneCompany(c)
	return nil
}

func (tx *memoryTx) DeleteCompany(ctx context.Context, id uuid.UUID) error {
	_ = ctx
	if _, ok := tx.companies[id]; !ok {
		return appcompany.ErrCompanyNotFound
	}
	delete(tx.companies, id)
	return nil
}

func (tx *memoryTx) InsertOutboxEvent(ctx context.Context, event outbox.Event) error {
	_ = ctx
	tx.outbox = append(tx.outbox, event)
	return nil
}

type decodedPayload struct {
	EventType string         `json:"event_type"`
	Data      map[string]any `json:"data"`
}

func decodePayload(t *testing.T, payload []byte) decodedPayload {
	t.Helper()
	var decoded decodedPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	return decoded
}

func mustCompany(t *testing.T, input companydomain.CreateInput) *companydomain.Company {
	t.Helper()
	created, _ := companydomain.New(input)
	return created
}

func cloneCompanies(companies map[uuid.UUID]*companydomain.Company) map[uuid.UUID]*companydomain.Company {
	cloned := make(map[uuid.UUID]*companydomain.Company, len(companies))
	for id, c := range companies {
		cloned[id] = cloneCompany(c)
	}

	return cloned
}

func cloneCompany(c *companydomain.Company) *companydomain.Company {
	return companydomain.NewFromDB(
		c.ID(),
		c.Name(),
		c.Description(),
		c.EmployeesCount(),
		c.Registered(),
		c.Type(),
		c.CreatedAt(),
		c.UpdatedAt(),
	)
}

func assertViolations(t *testing.T, err error, fields ...string) {
	t.Helper()
	var validationErr appcompany.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %T, want ValidationError", err)
	}
	for _, field := range fields {
		found := false
		for _, violation := range validationErr.Violations {
			if violation.Field == field {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing violation for %q in %#v", field, validationErr.Violations)
		}
	}
}
