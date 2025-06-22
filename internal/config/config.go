package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name,omitempty"`
}

const configFileName = "gatorconfig.json"

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserConfigDir()
	if err != nil {
		return "", errors.New("Could not locate home dir")
	}
	return filepath.Join(homeDir, configFileName), nil
}

func ReadConfigFile() (Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, errors.New("failed to read config file")
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return Config{}, errors.New("failed to unmarshal")
	}

	return cfg, nil
}

func (cfg *Config) setUser(username string) error {
	cfg.CurrentUserName = username
	return write(*cfg)
}

func write(cfg Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
