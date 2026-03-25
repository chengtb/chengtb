package messaging

import (
	"testing"
	"time"

	appuser "github.com/chengtb/chengtb/internal/application/user"
	"github.com/chengtb/chengtb/internal/pb"
	"google.golang.org/protobuf/proto"
)

// TestUserProtoRoundTrip verifies that converting a UserDTO → pb.User → UserDTO
// preserves all fields, including timestamp precision.
func TestUserProtoRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)

	original := &appuser.UserDTO{
		ID:        "u1",
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
	}

	pbUser := userToProto(original)
	got := userFromProto(pbUser)

	if got.ID != original.ID {
		t.Errorf("ID: got %q, want %q", got.ID, original.ID)
	}
	if got.Name != original.Name {
		t.Errorf("Name: got %q, want %q", got.Name, original.Name)
	}
	if got.Email != original.Email {
		t.Errorf("Email: got %q, want %q", got.Email, original.Email)
	}
	if !got.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, original.CreatedAt)
	}
	if !got.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", got.UpdatedAt, original.UpdatedAt)
	}
}

// TestReplyEnvelope_Success verifies that a success Reply envelope marshals and
// unmarshals correctly with error_code == 0 and the embedded payload intact.
func TestReplyEnvelope_Success(t *testing.T) {
	payload := &pb.User{Id: "u2", Name: "Bob", Email: "bob@example.com"}
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	envelope := &pb.Reply{ErrorCode: 0, Data: payloadBytes}
	raw, err := proto.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	var got pb.Reply
	if err := proto.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if got.ErrorCode != 0 {
		t.Errorf("ErrorCode: got %d, want 0", got.ErrorCode)
	}

	var gotUser pb.User
	if err := proto.Unmarshal(got.Data, &gotUser); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if gotUser.Id != "u2" || gotUser.Name != "Bob" {
		t.Errorf("payload mismatch: %+v", &gotUser)
	}
}

// TestReplyEnvelope_Error verifies that an error Reply encodes code and message.
func TestReplyEnvelope_Error(t *testing.T) {
	envelope := &pb.Reply{ErrorCode: 1, ErrorMessage: "resource not found"}
	raw, err := proto.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	var got pb.Reply
	if err := proto.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if got.ErrorCode != 1 {
		t.Errorf("ErrorCode: got %d, want 1", got.ErrorCode)
	}
	if got.ErrorMessage != "resource not found" {
		t.Errorf("ErrorMessage: got %q, want %q", got.ErrorMessage, "resource not found")
	}
	if len(got.Data) != 0 {
		t.Errorf("Data should be empty on error, got %d bytes", len(got.Data))
	}
}

// TestRequestMessages verifies that the five request message types
// marshal and unmarshal without data loss.
func TestRequestMessages(t *testing.T) {
	t.Run("CreateUserRequest", func(t *testing.T) {
		req := &pb.CreateUserRequest{Id: "1", Name: "Alice", Email: "alice@example.com"}
		roundtrip(t, req, &pb.CreateUserRequest{})
	})
	t.Run("GetUserRequest", func(t *testing.T) {
		roundtrip(t, &pb.GetUserRequest{Id: "2"}, &pb.GetUserRequest{})
	})
	t.Run("ListUsersRequest", func(t *testing.T) {
		roundtrip(t, &pb.ListUsersRequest{}, &pb.ListUsersRequest{})
	})
	t.Run("UpdateUserRequest", func(t *testing.T) {
		roundtrip(t, &pb.UpdateUserRequest{Id: "3", Name: "Bob", Email: "bob@example.com"}, &pb.UpdateUserRequest{})
	})
	t.Run("DeleteUserRequest", func(t *testing.T) {
		roundtrip(t, &pb.DeleteUserRequest{Id: "4"}, &pb.DeleteUserRequest{})
	})
}

// roundtrip is a generic helper that marshals src and unmarshals into dst,
// then compares them with proto.Equal.
func roundtrip(t *testing.T, src, dst proto.Message) {
	t.Helper()
	raw, err := proto.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := proto.Unmarshal(raw, dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !proto.Equal(src, dst) {
		t.Errorf("roundtrip mismatch:\n  src=%v\n  dst=%v", src, dst)
	}
}
