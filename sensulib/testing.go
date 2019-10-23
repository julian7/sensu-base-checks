package sensulib

type Errorable interface {
	Errorf(string, ...interface{})
}

func CatchExit(t Errorable, name string, want int) {
	err := recover()
	if err == nil {
		t.Errorf("%s never exited", name)
		return
	} else if code, ok := err.(exitCode); ok {
		if int(code) != want {
			t.Errorf("%s exited with %d, want %d", name, code, want)
		}
	} else {
		t.Errorf("%s received unexpected panic: %v", name, err)
	}
}
