package config

import (
	"github.com/spf13/viper"

	"github.com/nikolaevs92/Practicum/internal/agent"
	"github.com/nikolaevs92/Practicum/internal/server"
)

const (
	DefaultPollInterval   = "2s"
	DefaultReportRetries  = 2
	DefaultReportInterval = "10s"
	DefaultServer         = "127.0.0.1:8080"
)

const (
	envPollInterval   = "POLL_INTERVAL"
	envReportInterval = "REPORT_INTERVAL"
	envReportRetries  = "REPORT_RETRIES"
	envServer         = "ADDRESS"
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
