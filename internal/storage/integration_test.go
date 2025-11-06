//go:build integration

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func withPostgres(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()
	pg, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	// Build DSN after ensuring mapped port is available
	var host string
	var port string
	for i := 0; i < 60; i++ { // up to ~30s
		h, herr := pg.Host(ctx)
		mp, perr := pg.MappedPort(ctx, "5432/tcp")
		if herr == nil && perr == nil {
			host = h
			port = mp.Port()
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if host == "" || port == "" {
		t.Fatalf("postgres not ready: mapped 5432/tcp not found")
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "postgres", "postgres", host, port, "testdb")

	var db *sql.DB
	for i := 0; i < 60; i++ { // wait up to ~30s
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			if pingErr := db.Ping(); pingErr == nil {
				break
			} else {
				_ = db.Close()
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db, func() { _ = db.Close(); _ = pg.Terminate(ctx) }
}

func TestRepository_WithRealPostgres(t *testing.T) {
	db, cleanup := withPostgres(t)
	defer cleanup()

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := NewContainerRepository(db)
	rec := ContainerRecord{ID: "id1", Name: "n1", Image: "alpine", Status: "created", CreatedAt: time.Now().Unix()}
	if err := repo.Create(rec); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.UpdateStatus("id1", "running"); err != nil {
		t.Fatalf("update: %v", err)
	}
}
