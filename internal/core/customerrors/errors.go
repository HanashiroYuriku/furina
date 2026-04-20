package customerrors

import "errors"

var (
	ErrDataNotFound           = errors.New("Data Not Found")
	ErrInvalidPassword        = errors.New("Invalid Password")
	ErrTokenExpired           = errors.New("Token Expired")
	ErrDataConflict           = errors.New("Data Conflict")
	ErrBadRequest             = errors.New("Bad Request")
	ErrColldownActive         = errors.New("Please Wait 5 Minutes Before Sending Again")
	ErrInvalidCredentials     = errors.New("Invalid Username / Email")
	ErrAccountInactive        = errors.New("Account Not Verified")
	ErrAccountAlreadyVerified = errors.New("Account Already Verified")
)

type ValidationError struct {
	Detail string
}

func (e *ValidationError) Error() string {
	return "Validation Failed"
}

func NewValidationError(detail string) error {
	return &ValidationError{
		Detail: detail,
	}
}
