package config

import (
	"encoding/json"
	"os"
	"path/filepath"

)

type Config struct {
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil{
		return Config{}, err
	}

	path := filepath.Join(home, ".gatorconfig.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil

}

func (cfg *Config) SetUser(name string) error {
	cfg.CurrentUserName = name

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	path := filepath.Join(home, ".gatorconfig.json")

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)

}