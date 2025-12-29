package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     Server     `yaml:"server"`
	LoggerMode LoggerMode `yaml:"loggerMode"`
	TimeZone   TimeZone   `yaml:"timeZone"`
}

type SleepInterval struct {
	Interval time.Duration `yaml:"interval"`
}
type TimeZone struct {
	TimeZone string `yaml:"timeZone"`
}

type Server struct {
	Port        string
	Environment string
}

type BunConfig struct {
	DSN string
}

type LoggerMode struct {
	Development bool
	Prod        bool
	Level       string
}

func LoadConfig(path string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}
	return v, nil
}

func ParseConfig(v *viper.Viper) (*Config, error) {
	if v == nil {
		return nil, errors.New("nil viper instance")
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
