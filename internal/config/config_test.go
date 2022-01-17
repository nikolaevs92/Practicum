package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	cfg := LoadConfig()

	assert.Equal(t, cfg.Agent.PollInterval, time.Duration(DefaultPollInterval)*time.Second)
	assert.Equal(t, cfg.Agent.ReportInterval, time.Duration(DefaultReportInterval)*time.Second)
	assert.Equal(t, cfg.Agent.Server, DefaultServer)
	assert.Equal(t, cfg.Server.Server, DefaultServer)
}
