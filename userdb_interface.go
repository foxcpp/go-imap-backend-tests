package backendtests

import (
	"errors"

	"github.com/emersion/go-imap/backend"
)

var (
	ErrUserAlreadyExists = errors.New("imap: user already exists")
	ErrUserDoesntExists  = errors.New("imap: user doesn't exists")
)

// UsersDB is additional backend interface that allows external code
// to perform administrative actions on backend's storage related to users.
type UsersDB interface {
	// CreateUser creates new user with specified username.
	//
	// No mailboxes are created for new user, even INBOX.  This should be done
	// manually using CreateMailbox.
	//
	// It is error to create user which already exists.  ErrUserAlreadyExists
	// will be returned in this case.
	//
	// To backend implementers: You need to only return error with same message
	// for tests to work correctly.
	CreateUser(username, password string) error

	// DeleteUser deletes user account from backend storage, along with all
	// mailboxes and messages.
	//
	// It is error to delete user which doesn't exists.  ErrUserDoesntExists
	// will be returned in this case.
	//
	// To backend implementers: You need to only return error with same message
	// for tests to work correctly.
	DeleteUser(username string) error

	// SetUserPassword updates password of existsing user.
	//
	// It is error to update user which doesn't exists.  ErrUserDoesntExists
	// will be returned in this case.
	//
	// To backend implementers: You need to only return error with same message
	// for tests to work correctly.
	SetUserPassword(username, newPassword string) error
}

type IMAPUsersDB interface {
	UsersDB

	// GetUser is same as Backend.Login but doesn't
	// performs any authentication.
	GetUser(username string) (backend.User, error)
}
