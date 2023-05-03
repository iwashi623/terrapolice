package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
		panic(err)
	}

	fmt.Println(config)
}

func loadConfig(filename string) (*Config, error) {
	fmt.Println("Loading config from", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
