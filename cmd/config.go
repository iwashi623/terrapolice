package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultTimeout is the default timeout for running terraform plan
	DefaultTimeout = 300 * time.Second
)

type Directory struct {
	Path string `json:"path"`
}

type Config struct {
	BaseDirectory string `json:"base_directory"`
	Directories   []Directory
	Timeout       time.Duration `json:"timeout"`
	Notification  string        `json:"notification"`
}

func (cli *CLI) loadConfig(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer f.Close()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	if err := config.setTimeOut(); err != nil {
		return fmt.Errorf("error setting timeout: %w", err)
	}

	cli.Config = &config

	return nil
}

func (c *Config) setTimeOut() error {
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	if c.Timeout > 0 {
		c.Timeout = c.Timeout * time.Second
	}
	return nil
}

func (c *Config) getDirectories() []string {
	var combinedPaths []string
	for _, dir := range c.Directories {
		combinedPath := filepath.Join(c.BaseDirectory, dir.Path)
		combinedPaths = append(combinedPaths, combinedPath)
	}
	return combinedPaths
}
