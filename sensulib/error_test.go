package sensulib

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	type fields struct {
		criticality int
		err         error
	}

	tests := []struct {
		name   string
		fields fields
		error  string
		exits  int
	}{
		{"OK test", fields{0, errors.New("ok")}, "OK: ok", 0},
		{"WARN test", fields{1, errors.New("warn")}, "WARNING: warn", 1},
		{"CRIT test", fields{2, errors.New("crit")}, "CRITICAL: crit", 2},
		{"UNKNOWN test", fields{3, errors.New("unknown")}, "UNKNOWN: unknown", 3},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			serr := &Error{
				criticality: tt.fields.criticality,
				err:         tt.fields.err,
			}
			t.Run("Error", func(t *testing.T) {
				if got := serr.Error(); got != tt.error {
					t.Errorf("Error.Error() = %v, want %v", got, tt.error)
				}
			})
			t.Run("Exit", func(t *testing.T) {
				defer CatchExit(t, "Error.Error()", tt.exits)
				serr.Exit()
			})
		})
	}
}
