package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/nikolaevs92/Practicum/internal/agent"
	"github.com/nikolaevs92/Practicum/internal/server"
)

const (
	DefaultPollInterval   = time.Second * 2
	DefaultReportRetries  = 2
	DefaultReportInterval = time.Second * 10
	DefaultStoreInterval  = time.Second * 300
	DefaultStoreFile      = "/tmp/devops-metrics-db.json"
	DefaultRestore        = true
	DefaultServer         = "127.0.0.1:8080"
)

const (
	envPollInterval   = "POLL_INTERVAL"
	envReportInterval = "REPORT_INTERVAL"
	envStoreInterval  = "STORE_INTERVAL"
	envStoreFile      = "STORE_FILE"
	envRestore        = "RESTORE"
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

	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	conf := &Config{
		Viper:  v,
		Agent:  NewAgentConfig(v),
		Server: NewServerConfig(v),
	}

	return conf
}
