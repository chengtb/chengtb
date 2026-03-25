package messaging

import (
	"fmt"

	appuser "github.com/chengtb/chengtb/internal/application/user"
	"github.com/chengtb/chengtb/internal/pb"
	apperrors "github.com/chengtb/chengtb/pkg/errors"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"time"
)

// Client wraps the NATS connection and exposes typed request-reply helpers
// used by the API layer to communicate with the User micro-service.
type Client struct {
	nc *nats.Conn
}

// NewClient constructs a Client.
func NewClient(nc *nats.Conn) *Client {
	return &Client{nc: nc}
}

// request sends a NATS request with a proto-encoded payload and decodes the
// reply envelope.  On success the raw data bytes inside the envelope are
// unmarshalled into out (when out is non-nil).
func (c *Client) request(subject string, req proto.Message, out proto.Message) error {
	data, err := proto.Marshal(req)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "marshal request", err)
	}

	msg, err := c.nc.Request(subject, data, RequestTimeout)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, fmt.Sprintf("nats request %s", subject), err)
	}

	var reply pb.Reply
	if err := proto.Unmarshal(msg.Data, &reply); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "unmarshal reply envelope", err)
	}
	if reply.ErrorCode != 0 {
		return apperrors.New(apperrors.Code(reply.ErrorCode), reply.ErrorMessage)
	}

	if out != nil {
		if err := proto.Unmarshal(reply.Data, out); err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "unmarshal reply payload", err)
		}
	}
	return nil
}

// userFromProto converts a pb.User to an appuser.UserDTO.
func userFromProto(u *pb.User) *appuser.UserDTO {
	return &appuser.UserDTO{
		ID:        u.Id,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: time.Unix(0, u.CreatedAtUnix).UTC(),
		UpdatedAt: time.Unix(0, u.UpdatedAtUnix).UTC(),
	}
}

// CreateUser sends a CreateUserCommand to the user service.
func (c *Client) CreateUser(cmd appuser.CreateUserCommand) (*appuser.UserDTO, error) {
	var resp pb.User
	err := c.request(SubjectCreateUser, &pb.CreateUserRequest{
		Id:    cmd.ID,
		Name:  cmd.Name,
		Email: cmd.Email,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return userFromProto(&resp), nil
}

// GetUser sends a GetUserQuery to the user service.
func (c *Client) GetUser(q appuser.GetUserQuery) (*appuser.UserDTO, error) {
	var resp pb.User
	err := c.request(SubjectGetUser, &pb.GetUserRequest{Id: q.ID}, &resp)
	if err != nil {
		return nil, err
	}
	return userFromProto(&resp), nil
}

// ListUsers sends a ListUsersQuery to the user service.
func (c *Client) ListUsers() ([]*appuser.UserDTO, error) {
	var resp pb.UserList
	if err := c.request(SubjectListUsers, &pb.ListUsersRequest{}, &resp); err != nil {
		return nil, err
	}
	dtos := make([]*appuser.UserDTO, 0, len(resp.Users))
	for _, u := range resp.Users {
		dtos = append(dtos, userFromProto(u))
	}
	return dtos, nil
}

// UpdateUser sends an UpdateUserCommand to the user service.
func (c *Client) UpdateUser(cmd appuser.UpdateUserCommand) (*appuser.UserDTO, error) {
	var resp pb.User
	err := c.request(SubjectUpdateUser, &pb.UpdateUserRequest{
		Id:    cmd.ID,
		Name:  cmd.Name,
		Email: cmd.Email,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return userFromProto(&resp), nil
}

// DeleteUser sends a DeleteUserCommand to the user service.
func (c *Client) DeleteUser(cmd appuser.DeleteUserCommand) error {
	return c.request(SubjectDeleteUser, &pb.DeleteUserRequest{Id: cmd.ID}, nil)
}
