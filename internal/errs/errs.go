package errs

import "errors"

var (
	ErrDuplicateLogin          = errors.New("this login is already used")
	ErrNotFound                = errors.New("URL is not present")
	ErrInternalServerError     = errors.New("internal server error")
	ErrNotAuthorized           = errors.New("invalid token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrUserMismatch            = errors.New("user tried to delete not owned urls")
	ErrGone                    = errors.New("url was marked as deleted")
)
