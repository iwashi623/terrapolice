package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sync"
)

type outputLine struct {
	source string
	line   string
}

func main() {
	configPath := flag.String("f", "terrapolice.json", "Path to the configuration file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	directories := combineBaseDirectory(config)

	var wg sync.WaitGroup

	outCh := make(chan outputLine)
	go func() {
		for line := range outCh {
			fmt.Printf("%s: %s\n", line.source, line.line)
		}
	}()

	for _, dir := range directories {
		wg.Add(1)
		go func(directory string) {
			defer wg.Done()
			if err := runTerraformCommand(ctx, "init", directory, outCh); err != nil {
				fmt.Printf("Error running terraform init in directory %s: %v\n", directory, err)
				return
			}
			if err := runTerraformCommand(ctx, "plan", directory, outCh); err != nil {
				fmt.Printf("Error running terraform plan in directory %s: %v\n", directory, err)
			}
		}(dir)
	}

	wg.Wait()
	close(outCh)
}

func combineBaseDirectory(config *Config) []string {
	var combinedPaths []string
	for _, dir := range config.Directories {
		combinedPath := filepath.Join(config.BaseDirectory, dir.Path)
		combinedPaths = append(combinedPaths, combinedPath)
	}
	return combinedPaths
}

func readOutput(ctx context.Context, source string, r io.Reader, ch chan<- outputLine) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case ch <- outputLine{source, scanner.Text()}:
		}
	}
}

func runTerraformCommand(ctx context.Context, command, directory string, ch chan<- outputLine) error {
	cmd := exec.CommandContext(ctx, "terraform", command)
	cmd.Dir = directory

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	go readOutput(ctx, directory+" [stdout]", stdout, ch)
	go readOutput(ctx, directory+" [stderr]", stderr, ch)

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("terraform %s in directory %s timed out", command, directory)
		}
		return fmt.Errorf("error running terraform %s: %w", command, err)
	}

	return nil
}
