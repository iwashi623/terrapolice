package main

import (
	"context"
	"log"
	"os"

	"github.com/iwashi623/terrapolice/cmd"
)

func main() {
	args, err := cmd.ParseArgs(os.Args)
	if err != nil {
		log.Fatalf("Error parsing arguments: %v", err)
	}

	cli := cmd.NewCLI(args)
	exitCode, err := cli.Run(context.Background())
	if err != nil {
		log.Fatalf("Error running terraform checks: %v", err)
	}

	os.Exit(exitCode)
}
