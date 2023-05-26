package main

import (
	"context"
	"log"
	"os"

	"github.com/iwashi623/terrapolice/cmd"
)

func main() {
	cli, err := cmd.NewCLI(cmd.ParseArgs)
	if err != nil {
		log.Fatalf("Error creating CLI: %v", err)
	}

	ctx := context.Background()
	exitCode, err := cli.Run(ctx)
	if err != nil {
		log.Fatalf("Error running CLI: %v", err)
	}

	os.Exit(exitCode)
}
