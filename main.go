package main

import (
	"fmt"
	"os"
	"webproxy/cli"
)

func main() {
	args := os.Args[1:]
	if err := cli.Run(args); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
