package main

import (
	"fmt"
	"os"
	"time"
)

var startedAt = time.Now()

func main() {
	for i := 1; i < len(os.Args); i++ {
		proj, err := LoadProject(os.Args[i])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		sess, err := NewSession(proj)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}

		err = sess.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	}
}
