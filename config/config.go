package config

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

type Server struct {
	ListenAddress string `yaml:"listen_address"`
}

type Config struct {
	Server               Server   `yaml:"server"`
	Database             Database `yaml:"database"`
	InvestConfigFilePath string   `yaml:"invest_config_file_path"`
}

type Database struct {
	Dialect string `yaml:"dialect"`
	Path    string `yaml:"path"` // if sqlite
	Debug   bool   `yaml:"debug"`
}

func New(filepath string) (*Config, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	d := yaml.NewDecoder(bytes.NewReader(content))
	if err = d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
