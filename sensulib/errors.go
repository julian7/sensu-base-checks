package sensulib

import (
	"fmt"
	"strings"
)

// Errors is a slice of Error
type Errors []*Error

// NewErrors is a new, empty Errors slice
func NewErrors() *Errors {
	errs := []*Error{}
	return (*Errors)(&errs)
}

// Error returns all the errors, with all criticalities
func (errs *Errors) Error() string {
	if errs == nil {
		return "OK"
	}

	errors := make([]string, 0, len(*errs))

	for _, err := range []*Error(*errs) {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) < 1 {
		return "OK"
	}

	return strings.Join(errors, "\n")
}

// Exit terminates run by returning error
func (errs *Errors) Exit() {
	maxCrit := 0

	if errs == nil || len(*errs) == 0 {
		Exit(0)
		return // testing goes here
	}

	for _, err := range *errs {
		if err != nil {
			if err.criticality > maxCrit {
				maxCrit = err.criticality
			}

			fmt.Println(err.Error())
		}
	}

	Exit(maxCrit)
}

// Return provides return value for aggregates
func (errs *Errors) Return(err error) error {
	if errs == nil || len(*errs) == 0 {
		return err
	}

	return errs
}

// Add appends error to the slice, if it's not nil
func (errs *Errors) Add(err *Error) {
	if err == nil {
		return
	}

	*errs = append([]*Error(*errs), err)
}
