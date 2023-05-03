package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	BaseDirectory string `json:"base_directory"`
	Directories   []struct {
		Path string `json:"path"`
	} `json:"directories"`
}

func main() {
	configPath := flag.String("f", "terrapolice.json", "Path to the configuration file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		panic(err)
	}

	combinedPaths := combineBaseDirectory(config)
	for _, path := range combinedPaths {
		fmt.Println(path)
	}
}

func loadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		// fmt.Print("hogehoge")
		return nil, err
	}

	return &config, nil
}

func combineBaseDirectory(config *Config) []string {
	var combinedPaths []string
	for _, dir := range config.Directories {
		combinedPath := filepath.Join(config.BaseDirectory, dir.Path)
		combinedPaths = append(combinedPaths, combinedPath)
	}
	return combinedPaths
}
