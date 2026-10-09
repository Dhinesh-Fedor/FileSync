package main

import (
	"os"

	"FileSync/cmd/cli"
)

func main() {
	if err := cli.Execute(os.Args[1:]); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
	}
}
