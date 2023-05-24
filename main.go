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

	config, err := cmd.LoadConfig(args.ConfigPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	cli := cmd.NewCLI(args)
	exitCode, err := cli.Run(ctx, config)
	if err != nil {
		log.Fatalf("Error running terraform checks: %v", err)
	}

	os.Exit(exitCode)
}
