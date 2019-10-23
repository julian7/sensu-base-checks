package sensulib

import (
	"os"
)

type exitCode int

// Recover allows to convert panic back to os.Exit call
func Recover() {
	if err := recover(); err != nil {
		if code, ok := err.(exitCode); ok {
			os.Exit(int(code))
			return
		}

		panic(err)
	}
}

// Exit behaves like os.Exit, but instead of exiting right away,
// it panic()s instead. This allows test code to capture exit
// calls, while the program can act as intended, by deferring
// Recover() at the beginning.
func Exit(code int) {
	panic(exitCode(code))
}
