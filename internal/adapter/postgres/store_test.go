package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/fragoulis/xmercise/internal/adapter/postgres"
	appcompany "github.com/fragoulis/xmercise/internal/application/company"
	"github.com/fragoulis/xmercise/internal/application/outbox"
	"github.com/fragoulis/xmercise/internal/domain/company"
)

func TestStorePersistsCompanyAndOutboxInTransaction(t *testing.T) {
	ctx := context.Background()
	pool := newPostgresPool(t, ctx)
	store := postgres.NewStore(pool)

	description := "shipping"
	created, event, err := company.New(company.CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		Description:    &description,
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	err = store.WithinTx(ctx, func(ctx context.Context, tx appcompany.Tx) error {
		if err := tx.InsertCompany(ctx, created); err != nil {
			return err
		}

		return tx.InsertOutboxEvent(ctx, outbox.Event{
			ID:            uuid.New(),
			EventType:     string(event.Type()),
			AggregateType: event.AggregateType(),
			AggregateID:   event.AggregateID(),
			OccurredAt:    event.Occurred(),
			Payload:       json.RawMessage(`{"event_type":"company.created"}`),
			CreatedAt:     event.Occurred(),
		})
	})
	if err != nil {
		t.Fatalf("insert company and outbox: %v", err)
	}

	loaded, err := store.FindCompanyByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("find company: %v", err)
	}
	if loaded.ID() != created.ID() || loaded.Name() != "Acme" || loaded.EmployeesCount() != 7 {
		t.Fatalf("loaded company mismatch: %#v", loaded)
	}
	loadedDescription, ok := loaded.Description()
	if !ok || loadedDescription != description {
		t.Fatalf("description = %q, %t; want %q, true", loadedDescription, ok, description)
	}

	var outboxCount int
	err = pool.
		QueryRow(ctx, `SELECT count(*) FROM outbox_events WHERE aggregate_id = $1`, created.ID()).
		Scan(&outboxCount)
	if err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	if outboxCount != 1 {
		t.Fatalf("outbox count = %d, want 1", outboxCount)
	}
}

func TestStoreUpdatesDeletesAndMapsErrors(t *testing.T) {
	ctx := context.Background()
	pool := newPostgresPool(t, ctx)
	store := postgres.NewStore(pool)

	created, _, err := company.New(company.CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}
	if err := store.InsertCompany(ctx, created); err != nil {
		t.Fatalf("insert company: %v", err)
	}

	duplicate, _, err := company.New(company.CreateInput{
		ID:             uuid.New(),
		Name:           "acme",
		EmployeesCount: 1,
		Registered:     false,
		Type:           company.TypeCooperative,
	})
	if err != nil {
		t.Fatalf("new duplicate company: %v", err)
	}
	if err := store.InsertCompany(ctx, duplicate); !errors.Is(err, postgres.ErrCompanyNameTaken) {
		t.Fatalf("duplicate error = %v, want %v", err, postgres.ErrCompanyNameTaken)
	}

	newName := "Beta"
	err = store.WithinTx(ctx, func(ctx context.Context, tx appcompany.Tx) error {
		locked, err := tx.FindCompanyByIDForUpdate(ctx, created.ID())
		if err != nil {
			return err
		}

		if _, err := locked.Update(company.UpdateInput{
			Name: &newName,
		}); err != nil {
			return err
		}

		return tx.UpdateCompany(ctx, locked)
	})
	if err != nil {
		t.Fatalf("update company: %v", err)
	}

	loaded, err := store.FindCompanyByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("find updated company: %v", err)
	}
	if loaded.Name() != newName {
		t.Fatalf("name = %q, want %q", loaded.Name(), newName)
	}

	if err := store.DeleteCompany(ctx, created.ID()); err != nil {
		t.Fatalf("delete company: %v", err)
	}
	if _, err := store.FindCompanyByID(ctx, created.ID()); !errors.Is(err, postgres.ErrCompanyNotFound) {
		t.Fatalf("find deleted error = %v, want %v", err, postgres.ErrCompanyNotFound)
	}
}

func TestStoreRollsBackTransaction(t *testing.T) {
	ctx := context.Background()
	pool := newPostgresPool(t, ctx)
	store := postgres.NewStore(pool)

	created, _, err := company.New(company.CreateInput{
		ID:             uuid.New(),
		Name:           "Acme",
		EmployeesCount: 7,
		Registered:     true,
		Type:           company.TypeCorporations,
	})
	if err != nil {
		t.Fatalf("new company: %v", err)
	}

	rollbackErr := errors.New("rollback")
	err = store.WithinTx(ctx, func(ctx context.Context, tx appcompany.Tx) error {
		if err := tx.InsertCompany(ctx, created); err != nil {
			return err
		}

		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("transaction error = %v, want %v", err, rollbackErr)
	}

	if _, err := store.FindCompanyByID(ctx, created.ID()); !errors.Is(err, postgres.ErrCompanyNotFound) {
		t.Fatalf("find rolled back error = %v, want %v", err, postgres.ErrCompanyNotFound)
	}
}

func newPostgresPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	skipIfDockerUnavailable(t)

	runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	container, err := tcpostgres.Run(
		runCtx,
		"docker.io/postgres:16-alpine",
		tcpostgres.WithDatabase("companies"),
		tcpostgres.WithUsername("companies"),
		tcpostgres.WithPassword("companies"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(time.Minute),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	connString, err := container.ConnectionString(runCtx, "sslmode=disable")
	if err != nil {
		t.Fatalf("postgres connection string: %v", err)
	}

	pool, err := pgxpool.New(runCtx, connString)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := applyMigrations(runCtx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return pool
}

func skipIfDockerUnavailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker is unavailable")
	}
	if os.Getenv("DOCKER_HOST") == "" {
		command := exec.Command("docker", "context", "inspect", "-f", "{{.Endpoints.docker.Host}}")
		if output, err := command.Output(); err == nil {
			_ = os.Setenv("DOCKER_HOST", strings.TrimSpace(string(output)))
		}
	}
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	command := exec.Command("docker", "info")
	if output, err := command.CombinedOutput(); err != nil {
		t.Skipf("docker is unavailable: %v: %s", err, output)
	}
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationDir := filepath.Join(repoRoot(), "db", "migrations")
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i int, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationDir, entry.Name()))
		if err != nil {
			return err
		}
		up := strings.SplitN(string(content), "-- +goose Down", 2)[0]
		if _, err := pool.Exec(ctx, up); err != nil {
			return err
		}
	}

	return nil
}

func repoRoot() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Clean(filepath.Join(workingDir, "..", "..", ".."))
}
