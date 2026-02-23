package main

import (
	"fmt"
	"os"

	"github.com/erfnzdeh/git-retime/cmd"
)

func main() {
	if err := cmd.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s\n", err)
		os.Exit(1)
	}
}
