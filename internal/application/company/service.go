// Package company implements company use cases.
package company

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/fragoulis/xmercise/internal/application/outbox"
	companydomain "github.com/fragoulis/xmercise/internal/domain/company"
)

var (
	// ErrCompanyNotFound means the requested company does not exist.
	ErrCompanyNotFound = errors.New("company not found")
	// ErrCompanyNameTaken means another company already uses the same name ignoring case.
	ErrCompanyNameTaken = errors.New("company name taken")
)

// Store provides read and transactional write access for company use cases.
type Store interface {
	FindCompanyByID(ctx context.Context, id uuid.UUID) (*companydomain.Company, error)
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

// Tx provides transaction-scoped persistence for company mutations.
type Tx interface {
	InsertCompany(ctx context.Context, c *companydomain.Company) error
	FindCompanyByIDForUpdate(ctx context.Context, id uuid.UUID) (*companydomain.Company, error)
	UpdateCompany(ctx context.Context, c *companydomain.Company) error
	DeleteCompany(ctx context.Context, id uuid.UUID) error
	InsertOutboxEvent(ctx context.Context, event outbox.Event) error
}

// CreateCommand contains data for creating a company.
type CreateCommand struct {
	Name           string
	Description    *string
	EmployeesCount int
	Registered     bool
	Type           string
}

// UpdateCommand contains patch data for a company.
type UpdateCommand struct {
	ID             string
	Name           *string
	Description    companydomain.DescriptionPatch
	EmployeesCount *int
	Registered     *bool
	Type           *string
}

// DeleteCommand identifies a company to delete.
type DeleteCommand struct {
	ID string
}

// FindOneQuery identifies a company to retrieve.
type FindOneQuery struct {
	ID string
}

// Service coordinates company use cases.
type Service struct {
	store Store
}

// NewService creates a company service.
func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// Create creates a company and stores its outbox event in the same transaction.
func (s *Service) Create(ctx context.Context, command CreateCommand) (*companydomain.Company, error) {
	if err := validateCreateCommand(command); err != nil {
		return nil, err
	}

	created, event := companydomain.New(companydomain.CreateInput{
		Name:           command.Name,
		Description:    command.Description,
		EmployeesCount: command.EmployeesCount,
		Registered:     command.Registered,
		Type:           companydomain.Type(command.Type),
	})

	if err := s.store.WithinTx(ctx, func(ctx context.Context, tx Tx) error {
		if err := tx.InsertCompany(ctx, created); err != nil {
			return mapStoreError(err)
		}

		outboxEvent, err := buildOutboxEvent(event, companyDataFromCompany(created))
		if err != nil {
			return err
		}

		return tx.InsertOutboxEvent(ctx, outboxEvent)
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// Update patches a company and stores its outbox event in the same transaction.
func (s *Service) Update(ctx context.Context, command UpdateCommand) (*companydomain.Company, error) {
	id, err := validateUpdateCommand(command)
	if err != nil {
		return nil, err
	}
	var companyType *companydomain.Type
	if command.Type != nil {
		companyType = lo.ToPtr(companydomain.Type(*command.Type))
	}

	var updated *companydomain.Company
	if err := s.store.WithinTx(ctx, func(ctx context.Context, tx Tx) error {
		loaded, err := tx.FindCompanyByIDForUpdate(ctx, id)
		if err != nil {
			return mapStoreError(err)
		}

		event := loaded.Update(companydomain.UpdateInput{
			Name:           command.Name,
			Description:    command.Description,
			EmployeesCount: command.EmployeesCount,
			Registered:     command.Registered,
			Type:           companyType,
		})
		if err := tx.UpdateCompany(ctx, loaded); err != nil {
			return mapStoreError(err)
		}

		outboxEvent, err := buildOutboxEvent(event, companyDataFromCompany(loaded))
		if err != nil {
			return err
		}
		if err := tx.InsertOutboxEvent(ctx, outboxEvent); err != nil {
			return err
		}

		updated = loaded
		return nil
	}); err != nil {
		return nil, err
	}

	return updated, nil
}

// Delete hard-deletes a company and stores its outbox event in the same transaction.
func (s *Service) Delete(ctx context.Context, command DeleteCommand) error {
	id, err := validateID(command.ID)
	if err != nil {
		return err
	}

	return s.store.WithinTx(ctx, func(ctx context.Context, tx Tx) error {
		loaded, err := tx.FindCompanyByIDForUpdate(ctx, id)
		if err != nil {
			return mapStoreError(err)
		}

		event := loaded.Deleted()
		if err := tx.DeleteCompany(ctx, id); err != nil {
			return mapStoreError(err)
		}

		outboxEvent, err := buildOutboxEvent(event, deleteData{ID: id})
		if err != nil {
			return err
		}

		return tx.InsertOutboxEvent(ctx, outboxEvent)
	})
}

// FindOne returns one company by ID.
func (s *Service) FindOne(ctx context.Context, query FindOneQuery) (*companydomain.Company, error) {
	id, err := validateID(query.ID)
	if err != nil {
		return nil, err
	}

	found, err := s.store.FindCompanyByID(ctx, id)
	if err != nil {
		return nil, mapStoreError(err)
	}

	return found, nil
}

func mapStoreError(err error) error {
	switch {
	case errors.Is(err, ErrCompanyNotFound):
		return ErrCompanyNotFound
	case errors.Is(err, ErrCompanyNameTaken):
		return ErrCompanyNameTaken
	default:
		return err
	}
}

func buildOutboxEvent(domainEvent companydomain.Event, data any) (outbox.Event, error) {
	eventID := uuid.New()
	payload, err := json.Marshal(eventPayload{
		EventID:    eventID,
		EventType:  string(domainEvent.Type()),
		OccurredAt: domainEvent.Occurred(),
		CompanyID:  domainEvent.AggregateID(),
		Data:       data,
	})
	if err != nil {
		return outbox.Event{}, err
	}

	return outbox.Event{
		ID:            eventID,
		EventType:     string(domainEvent.Type()),
		AggregateType: domainEvent.AggregateType(),
		AggregateID:   domainEvent.AggregateID(),
		OccurredAt:    domainEvent.Occurred(),
		Payload:       payload,
		CreatedAt:     domainEvent.Occurred(),
	}, nil
}

type eventPayload struct {
	EventID    uuid.UUID `json:"event_id"`
	EventType  string    `json:"event_type"`
	OccurredAt time.Time `json:"occurred_at"`
	CompanyID  uuid.UUID `json:"company_id"`
	Data       any       `json:"data"`
}

type companyData struct {
	ID             uuid.UUID          `json:"id"`
	Name           string             `json:"name"`
	Description    *string            `json:"description"`
	EmployeesCount int                `json:"employees_count"`
	Registered     bool               `json:"registered"`
	Type           companydomain.Type `json:"type"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

type deleteData struct {
	ID uuid.UUID `json:"id"`
}

func companyDataFromCompany(c *companydomain.Company) companyData {
	return companyData{
		ID:             c.ID(),
		Name:           c.Name(),
		Description:    c.Description(),
		EmployeesCount: c.EmployeesCount(),
		Registered:     c.Registered(),
		Type:           c.Type(),
		CreatedAt:      c.CreatedAt(),
		UpdatedAt:      c.UpdatedAt(),
	}
}
