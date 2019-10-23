package sensulib

import "os"

var osExit = os.Exit

func exit(exitval int) {
	osExit(exitval)
}

func SetExit(exitFunc func(int)) {
	osExit = exitFunc
}
