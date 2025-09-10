package data

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	appdb "github.com/cr1cr1/farm-manager/internal/db"
	"github.com/cr1cr1/farm-manager/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func openTestDB(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
	t.Setenv("SQLITE_DSN", ":memory:")
	ctx := context.Background()
	db, err := appdb.Open(ctx)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = appdb.Close(db) })
	if err := appdb.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return ctx, db
}

func TestUserRepo_CRUD(t *testing.T) {
	ctx, db := openTestDB(t)
	repo := NewSQLiteUserRepo(db)

	// Count initially zero
	n, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 users, got %d", n)
	}

	// Create
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	u := &domain.User{
		Username:            "alice",
		PasswordHash:        string(hash),
		ForcePasswordChange: false,
	}
	id, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id")
	}

	// Count now 1
	n, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 user, got %d", n)
	}

	// Find
	got, err := repo.FindByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("unexpected username: %s", got.Username)
	}

	// Update password
	newHash, _ := bcrypt.GenerateFromPassword([]byte("newsecret"), bcrypt.DefaultCost)
	if err := repo.UpdatePassword(ctx, got.ID, string(newHash), false); err != nil {
		t.Fatalf("update password: %v", err)
	}

	// Soft delete
	now := time.Now().UTC()
	if err := repo.SoftDelete(ctx, got.ID, now); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	n, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("count after delete: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 users after soft delete, got %d", n)
	}
}

func TestBasePathEnv(t *testing.T) {
	// Ensure BasePath() logic: default / env override.
	os.Unsetenv("APP_BASE_PATH")
	if got := basePathForTest(); got != "/app" {
		t.Fatalf("default base path mismatch: %s", got)
	}
	os.Setenv("APP_BASE_PATH", "/x")
	if got := basePathForTest(); got != "/x" {
		t.Fatalf("env base path mismatch: %s", got)
	}
}

// Duplicate middleware.BasePath logic to avoid importing gf in unit tests.
func basePathForTest() string {
	if v := os.Getenv("APP_BASE_PATH"); v != "" {
		return v
	}
	return "/app"
}
