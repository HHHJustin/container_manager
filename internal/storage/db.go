package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func OpenDefault() (*sql.DB, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		host := getenvDefault("DB_HOST", "postgres")
		port := getenvDefault("DB_PORT", "5432")
		user := getenvDefault("DB_USER", "postgres")
		pass := getenvDefault("DB_PASSWORD", "postgres")
		name := getenvDefault("DB_NAME", "containerdb")
		url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
	}
	return sql.Open("postgres", url)
}

func Migrate(db *sql.DB) error {
    _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS containers (
    id TEXT PRIMARY KEY,
    name TEXT,
    image TEXT,
    status TEXT,
    created_at BIGINT
);
CREATE TABLE IF NOT EXISTS container_tasks (
    id TEXT PRIMARY KEY,
    container_id TEXT,
    cmd_json TEXT,
    status TEXT,
    exit_code INT,
    logs TEXT,
    created_at BIGINT,
    finished_at BIGINT
);
`)
	return err
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
