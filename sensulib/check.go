package sensulib

import (
	"errors"
	"fmt"
)

type ErrorExiter interface {
	error
	Exit()
}

func HandleError(err error) {
	var exiter ErrorExiter
	if errors.As(err, &exiter) {
		exiter.Exit()
		return // tests reach this
	}
	fmt.Printf("Error: %v\n", err)
	exit(3)
}
