package sensulib

import (
	"fmt"
)

const (
	OK = iota
	WARN
	CRIT
	UNKNOWN
)

// Error is a sensu-aware error message, which knows
// about its criticality
type Error struct {
	criticality int
	err         error
}

// Error returns a string representing the error and its
// criticality
func (serr *Error) Error() string {
	return fmt.Sprintf("%s: %v", serr.critString(), serr.err)
}

// Exit terminates run by returning the error
func (serr *Error) Exit() {
	fmt.Printf("%s\n", serr.Error())
	Exit(serr.criticality)
}

// NewError creates a new Error
func NewError(crit int, err error) *Error {
	return &Error{criticality: crit, err: err}
}

// Warn creates a Warning-level error
func Warn(err error) *Error {
	return NewError(WARN, err)
}

// Crit creates a Critical-level error
func Crit(err error) *Error {
	return NewError(CRIT, err)
}

// Ok creates a OK-level error
func Ok(err error) *Error {
	return NewError(OK, err)
}

// Unknown creates a Unknown-level error
func Unknown(err error) *Error {
	return NewError(UNKNOWN, err)
}

func (serr *Error) critString() string {
	switch serr.criticality {
	case OK:
		return "OK"
	case WARN:
		return "WARNING"
	case CRIT:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}
