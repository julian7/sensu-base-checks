package sensulib_test

import (
	"errors"
	"testing"

	"github.com/julian7/sensu-base-checks/sensulib"
)

func TestHandleError(t *testing.T) {
	var exited bool
	var got int
	testExit := func(exitval int) {
		exited = true
		got = exitval
	}
	sensulib.SetExit(testExit)
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
			exited = false
			sensulib.HandleError(tt.err)
			if exited != true {
				t.Errorf("HandleError() never exited")
			} else if got != tt.exits {
				t.Errorf("HandleError() exited with = %v, want %v", got, tt.exits)
			}
		})
	}
}
