// Package postgres contains PostgreSQL persistence adapters.
package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	appcompany "github.com/fragoulis/xmercise/internal/application/company"
	"github.com/fragoulis/xmercise/internal/application/outbox"
	"github.com/fragoulis/xmercise/internal/domain/company"
)

const uniqueViolationCode = "23505"

var (
	// ErrCompanyNotFound means the requested company row does not exist.
	ErrCompanyNotFound = appcompany.ErrCompanyNotFound
	// ErrCompanyNameTaken means another company already uses the same name ignoring case.
	ErrCompanyNameTaken = appcompany.ErrCompanyNameTaken
)

// Store persists application data in PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

// Tx is a PostgreSQL transaction-scoped adapter.
type Tx struct {
	tx pgx.Tx
}

type executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// NewStore creates a PostgreSQL store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
	}
}

// WithinTx runs fn inside a database transaction.
func (s *Store) WithinTx(ctx context.Context, fn func(ctx context.Context, tx appcompany.Tx) error) error {
	dbTx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		_ = dbTx.Rollback(ctx)
	}()

	if err := fn(ctx, &Tx{
		tx: dbTx,
	}); err != nil {
		return err
	}

	return dbTx.Commit(ctx)
}

// InsertCompany inserts a company row.
func (s *Store) InsertCompany(ctx context.Context, c *company.Company) error {
	return insertCompany(ctx, s.pool, c)
}

// FindCompanyByID finds a company by ID.
func (s *Store) FindCompanyByID(ctx context.Context, id uuid.UUID) (*company.Company, error) {
	return findCompanyByID(ctx, s.pool, id, "")
}

// UpdateCompany updates a company row.
func (s *Store) UpdateCompany(ctx context.Context, c *company.Company) error {
	return updateCompany(ctx, s.pool, c)
}

// DeleteCompany deletes a company row.
func (s *Store) DeleteCompany(ctx context.Context, id uuid.UUID) error {
	return deleteCompany(ctx, s.pool, id)
}

// InsertCompany inserts a company row inside the transaction.
func (t *Tx) InsertCompany(ctx context.Context, c *company.Company) error {
	return insertCompany(ctx, t.tx, c)
}

// FindCompanyByID finds a company by ID inside the transaction.
func (t *Tx) FindCompanyByID(ctx context.Context, id uuid.UUID) (*company.Company, error) {
	return findCompanyByID(ctx, t.tx, id, "")
}

// FindCompanyByIDForUpdate finds and locks a company row inside the transaction.
func (t *Tx) FindCompanyByIDForUpdate(ctx context.Context, id uuid.UUID) (*company.Company, error) {
	return findCompanyByID(ctx, t.tx, id, " FOR UPDATE")
}

// UpdateCompany updates a company row inside the transaction.
func (t *Tx) UpdateCompany(ctx context.Context, c *company.Company) error {
	return updateCompany(ctx, t.tx, c)
}

// DeleteCompany deletes a company row inside the transaction.
func (t *Tx) DeleteCompany(ctx context.Context, id uuid.UUID) error {
	return deleteCompany(ctx, t.tx, id)
}

// InsertOutboxEvent inserts an unpublished outbox event inside the transaction.
func (t *Tx) InsertOutboxEvent(ctx context.Context, event outbox.Event) error {
	_, err := t.tx.Exec(ctx, `
		INSERT INTO outbox_events (
			id, event_type, aggregate_type, aggregate_id, occurred_at, payload, created_at, attempts
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 0)
	`,
		event.ID,
		event.EventType,
		event.AggregateType,
		event.AggregateID,
		event.OccurredAt,
		event.Payload,
		event.CreatedAt,
	)
	return err
}

func insertCompany(ctx context.Context, db executor, c *company.Company) error {
	_, err := db.Exec(ctx, `
		INSERT INTO companies (
			id, name, description, employees_count, registered, type, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		c.ID(),
		c.Name(),
		companyDescription(c),
		c.EmployeesCount(),
		c.Registered(),
		c.Type(),
		c.CreatedAt(),
		c.UpdatedAt(),
	)
	return mapCompanyWriteError(err)
}

func findCompanyByID(ctx context.Context, db executor, id uuid.UUID, lockClause string) (*company.Company, error) {
	row := db.QueryRow(ctx, `
		SELECT id, name, description, employees_count, registered, type, created_at, updated_at
		FROM companies
		WHERE id = $1`+lockClause, id)

	return scanCompany(row)
}

func updateCompany(ctx context.Context, db executor, c *company.Company) error {
	tag, err := db.Exec(ctx, `
		UPDATE companies
		SET name = $2,
			description = $3,
			employees_count = $4,
			registered = $5,
			type = $6,
			created_at = $7,
			updated_at = $8
		WHERE id = $1
	`,
		c.ID(),
		c.Name(),
		companyDescription(c),
		c.EmployeesCount(),
		c.Registered(),
		c.Type(),
		c.CreatedAt(),
		c.UpdatedAt(),
	)
	if err != nil {
		return mapCompanyWriteError(err)
	}
	if tag.RowsAffected() != 1 {
		return ErrCompanyNotFound
	}

	return nil
}

func deleteCompany(ctx context.Context, db executor, id uuid.UUID) error {
	tag, err := db.Exec(ctx, `DELETE FROM companies WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrCompanyNotFound
	}

	return nil
}

func scanCompany(row pgx.Row) (*company.Company, error) {
	var state company.StoredState

	err := row.Scan(
		&state.ID,
		&state.Name,
		&state.Description,
		&state.EmployeesCount,
		&state.Registered,
		&state.Type,
		&state.CreatedAt,
		&state.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCompanyNotFound
	}
	if err != nil {
		return nil, err
	}

	return company.NewFromDB(state), nil
}

func companyDescription(c *company.Company) *string {
	description, ok := c.Description()
	if !ok {
		return nil
	}

	return &description
}

func mapCompanyWriteError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return ErrCompanyNameTaken
	}

	return err
}
