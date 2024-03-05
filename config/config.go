package config

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/bataloff/tiknkoff/pkg/database"
)

type Server struct {
	ListenAddress string `yaml:"listen_address"`
}

type Config struct {
	Server               Server        `yaml:"server"`
	Database             *database.Opt `yaml:"database"`
	InvestConfigFilePath string        `yaml:"invest_config_file_path"`
	PingInvest           bool          `yaml:"ping_invest"`
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
