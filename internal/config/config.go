package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Headless         bool     `mapstructure:"headless"`
	TimeoutSeconds   int      `mapstructure:"timeout_seconds"`
	Proxy            Proxy    `mapstructure:"proxy"`
	TargetAddress    string   `mapstructure:"target_address"`
	TargetCategories []string `mapstructure:"target_categories"`
	OutputFile       string   `mapstructure:"output_file"`
}

type Proxy struct {
	Enabled bool   `mapstructure:"enabled"`
	Server  string `mapstructure:"server"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
