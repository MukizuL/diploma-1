package errs

import "errors"

var (
	ErrConflictLogin           = errors.New("this login is already used by other user")
	ErrUserNotFound            = errors.New("user is not found")
	ErrOrderNotFound           = errors.New("order is not found")
	ErrInternalServerError     = errors.New("internal server error")
	ErrNotAuthorized           = errors.New("invalid token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrWrongOrderFormat        = errors.New("invalid order number format")
	ErrConflictOrder           = errors.New("this order has already been uploaded by other user")
	ErrDuplicateOrder          = errors.New("this order has already been uploaded by this user")
	ErrNoStatus                = errors.New("such status does not exist")
)
