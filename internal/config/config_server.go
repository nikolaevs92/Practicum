package config

import (
	"github.com/spf13/viper"

	"github.com/nikolaevs92/Practicum/internal/server"
)

func NewServerConfig(v *viper.Viper) *server.Config {
	v.SetDefault(envServer, DefaultServer)
	return &server.Config{
		Server: v.GetString(envServer),
	}
}
