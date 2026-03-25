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

// Response is the standard JSON envelope for every API response.
//
//   - Code    – application-level status: 0 = success, non-zero = error code.
//   - Message – human-readable description (localised for errors, "ok" for success).
//   - Data    – response payload; nil for errors or operations with no body.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}
