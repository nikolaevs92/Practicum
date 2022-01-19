package config

import (
	"github.com/spf13/viper"

	"github.com/nikolaevs92/Practicum/internal/agent"
)

func NewAgentConfig(v *viper.Viper) *agent.Config {
	v.SetDefault(envPollInterval, DefaultPollInterval)
	v.SetDefault(envReportInterval, DefaultReportInterval)
	v.SetDefault(envReportRetries, DefaultReportRetries)

	return &agent.Config{
		PollInterval:   v.GetDuration(envPollInterval),
		ReportInterval: v.GetDuration(envReportInterval),
		ReportRetries:  v.GetInt(envReportRetries),
		Server:         v.GetString(envServer),
	}
}
