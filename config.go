package main

import (
	"encoding/json"
	"io/ioutil"
)

// Config represents the configuration information.
type Config struct {
	Neo4jAddr    string `json:"neo4jAddr"`
	HttpUsername string `json:"httpUsername"`
	HttpPassword string `json:"httpPassword"`
}

// loadConfig reads the config file and returns the parsed Config.
// If an error is found, the function will panic.
func loadConfig() *Config {
	var conf Config
	// Get the config file.
	config_file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		Logger.Fatalf("Error loading config file: %v\n", err)
	}
	if err = json.Unmarshal(config_file, &conf); err != nil {
		Logger.Fatalf("Error loading config file: %v\n", err)
	}
	return &conf
}
