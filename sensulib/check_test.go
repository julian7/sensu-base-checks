package sensulib_test

import (
	"errors"
	"testing"

	"github.com/julian7/sensu-base-checks/sensulib"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		exits int
	}{
		{"OK", nil, 0},
		{"sensu error", sensulib.Crit(errors.New("omg")), 2},
		{"unknown", errors.New("unknown"), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer sensulib.CatchExit(t, "HandleError()", tt.exits)
			sensulib.HandleError(tt.err)
			}
		})
	}
}
