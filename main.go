package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var startedAt = time.Now()
var verbose bool

func main() {
	flag.BoolVar(&verbose, "v", false, "enable verbose logging")
	flag.Parse()

	for _, arg := range flag.Args() {
		proj, err := LoadProject(arg)
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
