package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/nikolaevs92/Practicum/internal/agent"
	"github.com/nikolaevs92/Practicum/internal/server"
)

const (
	DefaultPollInterval   = 2 * time.Second
	DefaultReportInterval = 2 * time.Second
	DefaultServer         = "127.0.0.1:8080"
)

const (
	envPollInterval   = "POOL_INTERVAL"
	envReportInterval = "REPORT_INTERVAL"
	envServer         = "SERVER"
)

type Config struct {
	Viper  *viper.Viper   `json:"viper"`
	Agent  *agent.Config  `json:"agent"`
	Server *server.Config `json:"server"`
}

func LoadConfig() *Config {
	v := viper.New()
	v.AutomaticEnv()

	conf := &Config{
		Viper:  v,
		Agent:  NewAgentConfig(v),
		Server: NewServerConfig(v),
	}

	return conf
}
