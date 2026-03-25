package user_test

import (
	"testing"

	domain "github.com/chengtb/chengtb/internal/domain/user"
)

func TestNew_ValidUser(t *testing.T) {
	u, err := domain.New("1", "Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.ID() != "1" {
		t.Errorf("expected id=1, got %s", u.ID())
	}
	if u.Name() != "Alice" {
		t.Errorf("expected name=Alice, got %s", u.Name())
	}
	if u.Email() != "alice@example.com" {
		t.Errorf("expected email=alice@example.com, got %s", u.Email())
	}
}

func TestNew_MissingID(t *testing.T) {
	_, err := domain.New("", "Alice", "alice@example.com")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestNew_MissingName(t *testing.T) {
	_, err := domain.New("1", "", "alice@example.com")
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestNew_MissingEmail(t *testing.T) {
	_, err := domain.New("1", "Alice", "")
	if err == nil {
		t.Fatal("expected error for missing email")
	}
}

func TestRename(t *testing.T) {
	u, _ := domain.New("1", "Alice", "alice@example.com")
	if err := u.Rename("Bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name() != "Bob" {
		t.Errorf("expected name=Bob, got %s", u.Name())
	}
}

func TestRename_EmptyName(t *testing.T) {
	u, _ := domain.New("1", "Alice", "alice@example.com")
	if err := u.Rename(""); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestChangeEmail(t *testing.T) {
	u, _ := domain.New("1", "Alice", "alice@example.com")
	if err := u.ChangeEmail("newalice@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email() != "newalice@example.com" {
		t.Errorf("expected newalice@example.com, got %s", u.Email())
	}
}

func TestChangeEmail_Empty(t *testing.T) {
	u, _ := domain.New("1", "Alice", "alice@example.com")
	if err := u.ChangeEmail(""); err == nil {
		t.Fatal("expected error for empty email")
	}
}
