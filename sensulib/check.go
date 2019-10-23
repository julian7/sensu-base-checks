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
	if err == nil {
		Exit(0)
	}

	var exiter ErrorExiter
	if errors.As(err, &exiter) {
		exiter.Exit()
		return // tests reach this
	}

	fmt.Printf("Error: %v\n", err)
	Exit(3)
}
