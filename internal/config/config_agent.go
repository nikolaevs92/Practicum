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
		PollInterval:   v.GetDuration(envPollInterval),
		ReportInterval: v.GetDuration(envReportInterval),
		ReportRetries:  v.GetInt(envReportRetries),
		Server:         v.GetString(envServer),
	}
}

func NewAgentConfigWithDefaults(v *viper.Viper, server string, pollInterval time.Duration, reportInterval time.Duration, key string) *agent.Config {
	v.SetDefault(envPollInterval, pollInterval)
	v.SetDefault(envReportInterval, pollInterval)
	v.SetDefault(envReportRetries, DefaultReportRetries)
	v.SetDefault(envServer, server)
	v.SetDefault(envKey, key)

	return &agent.Config{
		PollInterval:   v.GetDuration(envPollInterval),
		ReportInterval: v.GetDuration(envReportInterval),
		ReportRetries:  v.GetInt(envReportRetries),
		Server:         v.GetString(envServer),
		Key:            v.GetString(envKey),
	}
}
