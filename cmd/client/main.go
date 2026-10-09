package main

import (
	"fmt"
	"log"
	"os"

	"FileSync/internal/client"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <register|sync>")
		return
	}

	switch os.Args[1] {
	case "register":
		if len(os.Args) < 5 {
			fmt.Println("Usage: client register <name> <path> <server-url>")
			return
		}
		folder, err := client.RegisterFolder(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(folder.ID)
	case "sync":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client sync <folder-id>")
			return
		}
		report, err := client.SyncFolder(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("synced: added=%d modified=%d deleted=%d\n", report.Added, report.Modified, report.Deleted)
	default:
		fmt.Println("Usage: client <register|sync>")
	}
}
