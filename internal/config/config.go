package config

import (
	"fmt"
	"os"
	"encoding/json"
)

const configFileName = "/gatorconfig.json"

type Config struct {
	Url string	`json:"db_url"`
	User string	`json:"current_user_name"`
}

func (c *Config) SetUser(username string) error {

	c.User = username
	return write(c)
}

func Read(path string) (Config, error) {

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("Error getting config file path: %w\n", err)
	}

	file, err := os.Open(configFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("Error reading config file: %w", err)
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, fmt.Errorf("Error reading config file: %w", err)
	}

	return config, nil
}


func getConfigFilePath() (string, error) {
	path, err := os.Getwd()

	if err != nil {
		return "", err
	}

	path += "/gatorconfig.json"
	return path, nil
}

func write(c *Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	os.Create(configFilePath)
	json := "{\n\t\"db_url\": \"" + c.Url + "\",\n\t\"current_user_name\": " + "\"" + c.User + "\"\n}"

	os.WriteFile(configFilePath, []byte(json), 0666)
	if err != nil {
		return fmt.Errorf("Error writing to file: %w", err)
	}

	return nil
}



