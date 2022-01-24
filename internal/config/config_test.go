package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	cfg := LoadConfig()

	assert.Equal(t, cfg.Agent.PollInterval, 2*time.Second)
	assert.Equal(t, cfg.Agent.ReportInterval, 10*time.Second)
	assert.Equal(t, cfg.Agent.Server, DefaultServer)
	assert.Equal(t, cfg.Server.Server, DefaultServer)
}
