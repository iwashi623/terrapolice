package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strings"
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

	// 設定ファイルから設定を読み込む
	// TODO .ymlファイル対応
	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// 実行時ログを出力するためのチャネルを作成
	outCh := make(chan outputLine)
	go func() {
		for line := range outCh {
			// 実行時ログを出力
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
			// Run terraform init
			if err := runTerraformCommand(ctx, terraformInitCommand, directory, outCh); err != nil {
				fmt.Printf("Error running terraform init in directory %s: %v\n", directory, err)
				return
			}
			// Run terraform plan
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
		buffer.WriteString(line + "\n")
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
		execErr := execNotify(ctx, command, errBuffer, true)
		if execErr != nil {
			return fmt.Errorf("error running execNotify: %w", execErr)
		}
		return fmt.Errorf("error running terraform %s: %w", command, err)
	}

	err = execNotify(ctx, command, outBuffer, false)
	if err != nil {
		return fmt.Errorf("error running execNotify: %w", err)
	}

	return nil
}

func execNotify(ctx context.Context, command string, buf *bytes.Buffer, isError bool) error {
	var statusStr string
	if isError {
		statusStr = notification.StatusError
	} else if command == terraformPlanCommand {
		statusStr = getStatusStr(buf)
	} else {
		return nil
	}

	status, err := notification.NewStatus(statusStr)
	if err != nil {
		return fmt.Errorf("error creating status: %w", err)
	}
	params := &notification.NotifyParams{
		Status: status,
		Buffer: buf,
	}
	notifier := notification.NewNotifier("slack_bot")
	notifier.Notify(ctx, params)
	return nil
}

func getStatusStr(buf *bytes.Buffer) string {
	str := notification.StatusDiffDetected

	// "No changes."という文字数が含まれている場合のみsuccessとする
	// この判定はterraform planの出力に完全に依存しているため、判別方法は要検討
	if strings.Contains(buf.String(), "No changes.") {
		str = notification.StatusSuccess
	}
	return str
}
