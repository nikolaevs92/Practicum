package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/nikolaevs92/Practicum/internal/agent"
)

func NewAgentConfig(v *viper.Viper) *agent.Config {
	v.SetDefault(envPollInterval, DefaultPollInterval)
	v.SetDefault(envReportInterval, DefaultReportInterval)
	v.SetDefault(envReportRetries, DefaultReportRetries)
	v.SetDefault(envServer, DefaultServer)
	return &agent.Config{
		PollInterval:   time.Duration(v.GetInt64(envPollInterval)) * time.Second,
		ReportInterval: time.Duration(v.GetInt64(envReportInterval)) * time.Second,
		ReportRetries:  v.GetInt(envReportRetries),
		Server:         v.GetString(envServer),
	}
}
