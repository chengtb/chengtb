// Package dto contains data-transfer objects shared between the HTTP layer
// and NATS messages.
package dto

// CreateUserRequest is the HTTP request body for POST /users.
type CreateUserRequest struct {
	ID    string `json:"id" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// UpdateUserRequest is the HTTP request body for PUT /users/:id.
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ErrorResponse is the standard error JSON envelope returned by the API.
// Code is the application-level numeric error code for programmatic use;
// Error carries the localised, human-readable message.
type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}
