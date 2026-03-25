package user_test

import (
	"context"
	"testing"

	appuser "github.com/chengtb/chengtb/internal/application/user"
	"github.com/chengtb/chengtb/internal/infrastructure/persistence"
)

func newService() *appuser.Service {
	return appuser.NewService(persistence.NewInMemoryUserRepository())
}

func TestCreateUser(t *testing.T) {
	svc := newService()
	dto, err := svc.CreateUser(context.Background(), appuser.CreateUserCommand{
		ID:    "1",
		Name:  "Alice",
		Email: "alice@example.com",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if dto.ID != "1" || dto.Name != "Alice" {
		t.Errorf("unexpected dto: %+v", dto)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	svc := newService()
	cmd := appuser.CreateUserCommand{ID: "1", Name: "Alice", Email: "alice@example.com"}
	if _, err := svc.CreateUser(context.Background(), cmd); err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}
	cmd.ID = "2"
	if _, err := svc.CreateUser(context.Background(), cmd); err == nil {
		t.Fatal("expected conflict error for duplicate email")
	}
}

func TestGetUser(t *testing.T) {
	svc := newService()
	_, _ = svc.CreateUser(context.Background(), appuser.CreateUserCommand{
		ID:    "1",
		Name:  "Alice",
		Email: "alice@example.com",
	})
	dto, err := svc.GetUser(context.Background(), appuser.GetUserQuery{ID: "1"})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if dto.ID != "1" {
		t.Errorf("expected id=1, got %s", dto.ID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	svc := newService()
	_, err := svc.GetUser(context.Background(), appuser.GetUserQuery{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestUpdateUser(t *testing.T) {
	svc := newService()
	_, _ = svc.CreateUser(context.Background(), appuser.CreateUserCommand{
		ID:    "1",
		Name:  "Alice",
		Email: "alice@example.com",
	})
	dto, err := svc.UpdateUser(context.Background(), appuser.UpdateUserCommand{
		ID:    "1",
		Name:  "Alice Updated",
		Email: "updated@example.com",
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if dto.Name != "Alice Updated" {
		t.Errorf("expected updated name, got %s", dto.Name)
	}
}

func TestDeleteUser(t *testing.T) {
	svc := newService()
	_, _ = svc.CreateUser(context.Background(), appuser.CreateUserCommand{
		ID:    "1",
		Name:  "Alice",
		Email: "alice@example.com",
	})
	if err := svc.DeleteUser(context.Background(), appuser.DeleteUserCommand{ID: "1"}); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	_, err := svc.GetUser(context.Background(), appuser.GetUserQuery{ID: "1"})
	if err == nil {
		t.Fatal("expected not-found after delete")
	}
}

func TestListUsers(t *testing.T) {
	svc := newService()
	for _, cmd := range []appuser.CreateUserCommand{
		{ID: "1", Name: "Alice", Email: "alice@example.com"},
		{ID: "2", Name: "Bob", Email: "bob@example.com"},
	} {
		if _, err := svc.CreateUser(context.Background(), cmd); err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
	}
	dtos, err := svc.ListUsers(context.Background(), appuser.ListUsersQuery{})
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(dtos) != 2 {
		t.Errorf("expected 2 users, got %d", len(dtos))
	}
}
