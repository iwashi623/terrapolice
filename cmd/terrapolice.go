package cmd

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/iwashi623/terrapolice/notification"
)

const (
	terraformInitCmd = "init"
	terraformPlanCmd = "plan"
	ExitCodeOK       = 0
	ExitCodeError    = 1
	maxConcurrency   = 10
)

type CLI struct {
	Args   *Args
	Config *Config
}

type Args struct {
	ConfigPath string
}

type outputLine struct {
	source string
	line   string
}

type CLIParseFunc func([]string) (*Args, error)

func NewCLI(parseArgs CLIParseFunc) (*CLI, error) {
	args, err := parseArgs(os.Args)
	if err != nil {
		return nil, fmt.Errorf("parsing args: %v", err)
	}
	return &CLI{
		Args: args,
	}, nil
}

func (cli *CLI) Run(ctx context.Context) (int, error) {
	if err := cli.loadConfig(cli.Args.ConfigPath); err != nil {
		return ExitCodeError, fmt.Errorf("loading config: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, cli.Config.Timeout)
	defer cancel()

	// Run terraform checks
	exitCode, err := cli.run(ctx)
	if err != nil {
		return ExitCodeError, fmt.Errorf("running terraform checks: %v", err)
	}

	// All done
	return exitCode, nil
}

func ParseArgs(args []string) (*Args, error) {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	var configPath string
	flags.StringVar(&configPath, "f", "terrapolice.json", "Path to the configuration file")

	err := flags.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	return &Args{ConfigPath: configPath}, nil
}

func (cli *CLI) run(ctx context.Context) (int, error) {
	// 実行時ログを出力するためのチャネルを作成
	outCh := make(chan outputLine)
	go func() {
		for line := range outCh {
			// terraformコマンドの実行時ログを出力
			fmt.Printf("%s: %s\n", line.source, line.line)
		}
	}()

	var wg sync.WaitGroup

	directories := cli.Config.getDirectories()
	numDirectories := len(directories)

	dirCh := make(chan string, numDirectories)
	for _, dir := range directories {
		dirCh <- dir
	}
	close(dirCh)

	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dir := range dirCh {
				// Run terraform init
				if err := cli.runTerraformCommand(ctx, terraformInitCmd, dir, outCh); err != nil {
					color.Red("terraform init failed in directory %s: %v", dir, err)
					return
				}
				// Run terraform plan
				if err := cli.runTerraformCommand(ctx, terraformPlanCmd, dir, outCh); err != nil {
					color.Red("terraform plan failed in directory %s: %v", dir, err)
				}
			}
		}()
	}

	wg.Wait()
	close(outCh)

	return ExitCodeOK, nil
}

func (cli *CLI) runTerraformCommand(ctx context.Context, command, directory string, ch chan<- outputLine) error {
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
		execErr := cli.execNotify(ctx, command, directory, errBuffer, true)
		if execErr != nil {
			return fmt.Errorf("error running execNotify: %w", execErr)
		}
		return fmt.Errorf("error running terraform %s: %w", command, err)
	}

	err = cli.execNotify(ctx, command, directory, outBuffer, false)
	if err != nil {
		return fmt.Errorf("error running execNotify: %w", err)
	}

	return nil
}

func (cli *CLI) execNotify(ctx context.Context, command string, directory string, buf *bytes.Buffer, isError bool) error {
	var statusStr string
	if isError {
		statusStr = notification.StatusError
	} else if command == terraformPlanCmd {
		statusStr = getStatusStr(buf)
	} else {
		return nil
	}

	status, err := notification.NewStatus(statusStr)
	if err != nil {
		return fmt.Errorf("error creating status: %w", err)
	}

	notifier, err := notification.NewNotifier(cli.Config.Notification)
	if err != nil {
		return fmt.Errorf("error creating notifier: %w", err)
	}

	params := &notification.NotifyParams{
		Status:    status,
		Buffer:    buf,
		Command:   command,
		Directory: directory,
	}
	notifier.Notify(ctx, params)
	return nil
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

func getStatusStr(buf *bytes.Buffer) string {
	str := notification.StatusDiffDetected

	// "No changes."という文字数が含まれている場合のみsuccessとする
	// この判定はterraform planの出力に完全に依存しているため、判別方法は要検討
	if strings.Contains(buf.String(), "No changes.") {
		str = notification.StatusSuccess
	}
	return str
}
