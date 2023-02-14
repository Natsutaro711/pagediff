package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Natsutaro711/pagediff/cmd/diff"
	"github.com/Natsutaro711/pagediff/cmd/screenshot"
)

func main() {
	screenshotCmd := flag.NewFlagSet("screenshot", flag.ExitOnError)
	screenshotUrllist := screenshotCmd.String("urllist", "list.csv", "option")
	screenshotBrowser := screenshotCmd.String("browser", "Chromium", "option")

	diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
	diffFrom := diffCmd.String("from", "", "directory")
	diffTo := diffCmd.String("to", "", "directory")

	if len(os.Args) < 2 {
		fmt.Println("Usage: pagediff <command> [arguments]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "screenshot":
		screenshotCmd.Parse(os.Args[2:])
		err := screenshot.ScreenShot(*screenshotUrllist, *screenshotBrowser)
		if err != nil {
			fmt.Printf("failed to screenshot : %v\n", err)
		}

	case "diff":
		diffCmd.Parse(os.Args[2:])
		err := diff.Diff(*diffFrom, *diffTo)
		if err != nil {
			fmt.Printf("failed to diff : %v\n", err)
		}

	default:
		fmt.Println("Unknown command: ", os.Args[1])
		os.Exit(1)
	}
}
