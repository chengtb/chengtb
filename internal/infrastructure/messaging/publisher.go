package messaging

import (
	"encoding/json"
	"fmt"

	appuser "github.com/chengtb/chengtb/internal/application/user"
	apperrors "github.com/chengtb/chengtb/pkg/errors"
	"github.com/nats-io/nats.go"
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

// errEnvelope is the JSON structure the subscriber sends on error.
type errEnvelope struct {
	Error string `json:"error"`
}

// request sends a NATS request and unmarshals the response into out.
// If the response carries an error envelope the error is returned.
func (c *Client) request(subject string, payload any, out any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "marshal request", err)
	}
	msg, err := c.nc.Request(subject, data, RequestTimeout)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, fmt.Sprintf("nats request %s", subject), err)
	}

	// Check for error envelope first.
	var env errEnvelope
	if err := json.Unmarshal(msg.Data, &env); err == nil && env.Error != "" {
		return apperrors.New(apperrors.CodeInternal, env.Error)
	}

	if out != nil {
		if err := json.Unmarshal(msg.Data, out); err != nil {
			return apperrors.Wrap(apperrors.CodeInternal, "unmarshal response", err)
		}
	}
	return nil
}

// CreateUser sends a CreateUserCommand to the user service.
func (c *Client) CreateUser(cmd appuser.CreateUserCommand) (*appuser.UserDTO, error) {
	var dto appuser.UserDTO
	if err := c.request(SubjectCreateUser, cmd, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// GetUser sends a GetUserQuery to the user service.
func (c *Client) GetUser(q appuser.GetUserQuery) (*appuser.UserDTO, error) {
	var dto appuser.UserDTO
	if err := c.request(SubjectGetUser, q, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// ListUsers sends a ListUsersQuery to the user service.
func (c *Client) ListUsers() ([]*appuser.UserDTO, error) {
	var dtos []*appuser.UserDTO
	if err := c.request(SubjectListUsers, struct{}{}, &dtos); err != nil {
		return nil, err
	}
	return dtos, nil
}

// UpdateUser sends an UpdateUserCommand to the user service.
func (c *Client) UpdateUser(cmd appuser.UpdateUserCommand) (*appuser.UserDTO, error) {
	var dto appuser.UserDTO
	if err := c.request(SubjectUpdateUser, cmd, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// DeleteUser sends a DeleteUserCommand to the user service.
func (c *Client) DeleteUser(cmd appuser.DeleteUserCommand) error {
	return c.request(SubjectDeleteUser, cmd, nil)
}
