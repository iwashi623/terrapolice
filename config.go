package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	// DefaultTimeout is the default timeout for running terraform plan
	DefaultTimeout = 300 * time.Second
)

type Config struct {
	BaseDirectory string `json:"base_directory"`
	Directories   []struct {
		Path string `json:"path"`
	} `json:"directories"`
	Timeout time.Duration `json:"timeout"`
}

func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, err
	}

	if err := config.setTimeOut(); err != nil {
		return nil, fmt.Errorf("error setting timeout: %w", err)
	}

	return &config, nil
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
