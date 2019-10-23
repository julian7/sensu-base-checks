package sensulib_test

import (
	"errors"
	"testing"

	"github.com/julian7/sensu-base-checks/sensulib"
)

type testErrorsExamples struct {
	name   string
	errs   *sensulib.Errors
	error  string
	exits  int
	retErr bool
}

func errorsBuilder(errs ...*sensulib.Error) *sensulib.Errors {
	ret := &sensulib.Errors{}
	for _, err := range errs {
		ret.Add(err)
	}
	return ret
}

func TestErrors_methods(t *testing.T) {
	tests := []testErrorsExamples{
		{"nil", nil, "OK", 0, false},
		{"empty", errorsBuilder(), "OK", 0, false},
		{"one", errorsBuilder(sensulib.NewError(1, errors.New("msg"))), "WARNING: msg", 1, true},
		{"two", errorsBuilder(
			sensulib.NewError(0, errors.New("all's well")),
			sensulib.NewError(2, errors.New("omg")),
		), "OK: all's well\nCRITICAL: omg", 2, true},
		{"skips nil", errorsBuilder(
			sensulib.NewError(2, errors.New("omg")),
			nil,
		), "CRITICAL: omg", 2, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.testErrorsError(t)
			tt.testErrorsExit(t)
			tt.testErrorsReturn(t)
		})
	}
}

func (tt *testErrorsExamples) testErrorsError(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		if got := tt.errs.Error(); got != tt.error {
			t.Errorf("Errors.Error() = %#v, want %v", got, tt.error)
		}
	})
}

func (tt *testErrorsExamples) testErrorsExit(t *testing.T) {
	t.Run("Exit", func(t *testing.T) {
		defer sensulib.CatchExit(t, "Errors.Error()", tt.exits)
		tt.errs.Exit()
	})
}

func (tt *testErrorsExamples) testErrorsReturn(t *testing.T) {
	t.Run("Return", func(t *testing.T) {
		got := tt.errs.Return(errors.New("default error"))
		if tt.retErr != (got == tt.errs) {
			t.Errorf(
				"Errors.Return() = %v, wants %v",
				got,
				map[bool]string{
					false: "default error",
					true:  "original errors",
				}[tt.retErr],
			)
		}
	})
}
