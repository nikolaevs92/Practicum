package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/nikolaevs92/Practicum/internal/datastorage"
	"github.com/nikolaevs92/Practicum/internal/server"
)

func NewServerConfig(v *viper.Viper) *server.Config {
	v.SetDefault(envServer, DefaultServer)
	v.SetDefault(envStoreInterval, DefaultStoreInterval)
	v.SetDefault(envStoreFile, DefaultStoreFile)
	v.SetDefault(envRestore, DefaultRestore)

	return &server.Config{
		Server: v.GetString(envServer),
		StorageConfig: datastorage.StorageConfig{
			StoreInterval: v.GetDuration(envStoreInterval),
			StoreFile:     v.GetString(envStoreFile),
			Restore:       v.GetBool(envRestore),
			Store:         v.GetString(envStoreFile) != "",
			Synchronized:  v.GetDuration(envStoreInterval) == time.Duration(0),
		},
	}
}
