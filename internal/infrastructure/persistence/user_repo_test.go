package persistence_test

import (
	"context"
	"testing"

	domain "github.com/chengtb/chengtb/internal/domain/user"
	"github.com/chengtb/chengtb/internal/infrastructure/persistence"
)

func newUser(t *testing.T, id, name, email string) *domain.User {
	t.Helper()
	u, err := domain.New(id, name, email)
	if err != nil {
		t.Fatalf("domain.New: %v", err)
	}
	return u
}

func TestSaveAndFindByID(t *testing.T) {
	repo := persistence.NewInMemoryUserRepository()
	ctx := context.Background()

	u := newUser(t, "1", "Alice", "alice@example.com")
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, "1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID() != u.ID() || got.Email() != u.Email() {
		t.Errorf("unexpected user: %+v", got)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	repo := persistence.NewInMemoryUserRepository()
	_, err := repo.FindByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestFindAll(t *testing.T) {
	repo := persistence.NewInMemoryUserRepository()
	ctx := context.Background()

	_ = repo.Save(ctx, newUser(t, "1", "Alice", "alice@example.com"))
	_ = repo.Save(ctx, newUser(t, "2", "Bob", "bob@example.com"))

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 users, got %d", len(all))
	}
}

func TestDelete(t *testing.T) {
	repo := persistence.NewInMemoryUserRepository()
	ctx := context.Background()

	_ = repo.Save(ctx, newUser(t, "1", "Alice", "alice@example.com"))
	if err := repo.Delete(ctx, "1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "1")
	if err == nil {
		t.Fatal("expected not-found error after delete")
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := persistence.NewInMemoryUserRepository()
	if err := repo.Delete(context.Background(), "nonexistent"); err == nil {
		t.Fatal("expected not-found error")
	}
}
