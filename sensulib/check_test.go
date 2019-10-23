package sensulib_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/julian7/sensu-base-checks/sensulib"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		exits int
		msg   string
	}{
		{"OK", nil, 0, "OK: ok"},
		{"sensu error", sensulib.Crit(errors.New("omg")), 2, "CRITICAL: omg"},
		{"unknown", errors.New("unknown"), 3, "Error: unknown"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var reader *os.File

			defer func(origStdout *os.File) { os.Stdout = origStdout }(os.Stdout)
			reader, os.Stdout, _ = os.Pipe()
			captureChan := make(chan string)
			go func(reader *os.File, outchan chan<- string) {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, reader)
				outchan <- buf.String()
			}(reader, captureChan)
			defer sensulib.CatchExit(t, "HandleError()", tt.exits)
			sensulib.HandleError(tt.err)
			os.Stdout.Close()
			captured := <-captureChan
			if captured != tt.msg {
				t.Errorf("HandleError() printed `%v`, want `%v`", captured, tt.msg)
			}
		})
	}
}
