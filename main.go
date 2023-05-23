package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/iwashi623/terrapolice/notification"
)

const (
	terraformInitCommand = "init"
	terraformPlanCommand = "plan"
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

	outCh := make(chan outputLine)
	go func() {
		for line := range outCh {
			fmt.Printf("%s: %s\n", line.source, line.line)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	directories := config.getDirectories()

	var wg sync.WaitGroup

	for _, dir := range directories {
		wg.Add(1)
		go func(directory string) {
			defer wg.Done()
			if err := runTerraformCommand(ctx, terraformInitCommand, directory, outCh); err != nil {
				fmt.Printf("Error running terraform init in directory %s: %v\n", directory, err)
				return
			}
			if err := runTerraformCommand(ctx, terraformPlanCommand, directory, outCh); err != nil {
				fmt.Printf("Error running terraform plan in directory %s: %v\n", directory, err)
			}
		}(dir)
	}

	wg.Wait()
	close(outCh)
}

func readOutput(ctx context.Context, source string, r io.Reader, ch chan<- outputLine, buffer *bytes.Buffer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		select {
		case <-ctx.Done():
			return
		case ch <- outputLine{source, line}:
		}
		buffer.WriteString(line)
		buffer.WriteString("\n") // preserve newline
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

	outBuffer := &bytes.Buffer{}
	errBuffer := &bytes.Buffer{}
	go readOutput(ctx, directory+" [stdout]", stdout, ch, outBuffer)
	go readOutput(ctx, directory+" [stderr]", stderr, ch, errBuffer)

	err = cmd.Wait()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("terraform %s in directory %s timed out", command, directory)
		}
		return fmt.Errorf("error running terraform %s: %w", command, err)
	}

	if command == terraformPlanCommand {
		notifier := notification.CreateNotifier("slack_bot")
		params := notification.NotifyParams{
			Status: "success",
			Buffer: outBuffer,
		}
		notifier.Notify(params)
	}

	return nil
}
