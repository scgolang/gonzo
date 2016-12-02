package nsm

// Code is the type of an error code.
type Code int

// Error codes.
const (
	ErrGeneral         Code = -1
	ErrIncompatibleAPI Code = -2
	ErrBlacklisted     Code = -3
	ErrLaunchFailed    Code = -4
	ErrNoSuchFile      Code = -5
	ErrSoSessionOpen   Code = -6
	ErrUnsavedChanges  Code = -7
	ErrNotNow          Code = -8
	ErrBadProject      Code = -9
	ErrCreateFailed    Code = -10
)

// Error extends the builtin error interface to add an nsm error code.
type Error interface {
	error
	Code() Code
}

// NewError creates an error with an nsm-specific error code.
func NewError(code Code, msg string) Error {
	return nsmError{msg: msg, code: code}
}

// nsmError represents an nsm-specific error.
type nsmError struct {
	msg  string
	code Code
}

// Error returns an error message.
func (e nsmError) Error() string {
	return e.msg
}

// Code returns an nsm-specific error code.
func (e nsmError) Code() Code {
	return e.code
}
