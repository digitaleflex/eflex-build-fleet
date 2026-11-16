//go:build integration

package controlplane

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestIntegrationPostgresPing(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	// verify core tables exist
	rows, err := db.Query("SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='build_jobs'")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("no result")
	}
	var n int
	rows.Scan(&n)
	if n != 1 {
		t.Fatalf("build_jobs table missing (count=%d)", n)
	}
}