package config

import (
	"encoding/json"
	"fmt"
	"os"

	l4g "github.com/alecthomas/log4go"
)

var Cfg *Config = &Config{}

type Config struct {
	ServiceSettings  ServiceSettings
	LogSettings      LogSettings
	DatabaseSettings DatabaseSettings
}

type ServiceSettings struct {
	ListenAddress string
}

type LogSettings struct {
	EnableConsole bool
}

type DatabaseSettings struct {
	DatabaseUsername   string
	DatabasePassword   string
	DriverName         string
	MaxIdleConns       int
	MaxOpenConns       int
	ServerName         string
	DatabaseName       string
	DataSource         string
	DataSourceReplicas []string
	Trace              bool
}

func LoadConfig(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic("Error opening config file " + filePath + "\nerror: " + err.Error())
	}
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		panic("Error decoding config file " + filePath + "\nerror: " + err.Error())
	}
	l4g.Info("Successfully loaded configs")

	fmt.Println("DEBUGconfig!", config)

	Cfg = &config
}
