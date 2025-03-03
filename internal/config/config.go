package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	return write(*c)
}

func write(cfg Config) error {
	filePath, err := getConfigFilePath()
	if err != nil {
        return err
    }

	data, err := json.Marshal(cfg)
	if err != nil {
        return err
    }

	return os.WriteFile(filePath, data, 0644)
}

func getConfigFilePath() (string, error) {
	dirname, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
	return filepath.Join(dirname, configFileName), nil
}

func Read() (Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
        return Config{}, err
    }

	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
