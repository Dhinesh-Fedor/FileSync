package main

import (
	"fmt"
	"os"

	"FileSync/cmd/cli"
)

func main() {

	if len(os.Args) < 3 || os.Args[1] != "fs" {
		fmt.Println("Usage: fs <command>")
		return
	}

	var err error

	switch os.Args[2] {

	case "scan":
		err = cli.RunScanner()

	case "compare":
		err = cli.RunCompare()

	case "add":
		err = cli.RunAdd()

	case "list":
		err = cli.RunList()

	case "details":
		err = cli.RunDetails()

	case "delete":
		err = cli.RunDelete()

	case "changes":
		err = cli.RunChanges()

	case "sync":
		err = cli.RunSync()

	case "report":
		err = cli.RunReport()

	default:
		fmt.Println("Unknown Command.")
		return
	}

	if err != nil {
		fmt.Println(err)
	}
}
