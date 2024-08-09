package errs

import "errors"

var (
	ErrUserType           = errors.New("wrong user type")
	ErrUserExists         = errors.New("user already exists")
	ErrFlatStatus         = errors.New("wrong flat status")
	ErrInvalidCredentials = errors.New("Invalid credentials")
)
