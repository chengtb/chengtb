package user

import "time"

// GetUserQuery fetches a single user by ID.
type GetUserQuery struct {
	ID string
}

// ListUsersQuery fetches all users.
type ListUsersQuery struct{}

// UserDTO is the read model returned by queries.
type UserDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
