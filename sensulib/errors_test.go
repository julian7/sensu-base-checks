package sensulib_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/julian7/sensu-base-checks/sensulib"
)

func TestNewErrors(t *testing.T) {
	want := &sensulib.Errors{}
	if got := sensulib.NewErrors(); !reflect.DeepEqual(got, want) {
		t.Errorf("NewErrors() = %v, want %v", got, want)
	}
}

func TestErrors_methods(t *testing.T) {
	tests := []struct {
		name   string
		errs   *sensulib.Errors
		error  string
		exits  int
		retErr bool
	}{
		{"nil", nil, "OK", 0, false},
		{"empty", &sensulib.Errors{}, "OK", 0, false},
		{"one", &sensulib.Errors{sensulib.NewError(1, errors.New("msg"))}, "WARNING: msg", 1, true},
		{"two", &sensulib.Errors{
			sensulib.NewError(0, errors.New("all's well")),
			sensulib.NewError(2, errors.New("omg")),
		}, "OK: all's well\nCRITICAL: omg", 2, true},
		{"skips nil", &sensulib.Errors{
			sensulib.NewError(2, errors.New("omg")),
			nil,
		}, "CRITICAL: omg", 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Error", func(t *testing.T) {
				if got := tt.errs.Error(); got != tt.error {
					t.Errorf("Errors.Error() = %#v, want %v", got, tt.error)
				}
			})
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
			tt.testErrorsExit(t)
		})
	}
}

func (tt *testErrorsExamples) testErrorsExit(t *testing.T) {
	t.Run("Exit", func(t *testing.T) {
		defer sensulib.CatchExit(t, "Errors.Error()", tt.exits)
		tt.errs.Exit()
	})
}
