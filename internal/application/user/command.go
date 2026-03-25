// Package user contains the application-level use-case handlers for the
// User domain following the CQRS pattern.
package user

// CreateUserCommand is the command to create a new user.
type CreateUserCommand struct {
	ID    string
	Name  string
	Email string
}

// UpdateUserCommand is the command to update an existing user.
type UpdateUserCommand struct {
	ID    string
	Name  string
	Email string
}

// DeleteUserCommand is the command to delete a user.
type DeleteUserCommand struct {
	ID string
}
