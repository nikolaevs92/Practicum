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

func NewServerConfigWithDefaults(
	v *viper.Viper, adress string, stroreInterval time.Duration, storeFile string, restore bool, key string, dataBaseDSN string) *server.Config {

	v.SetDefault(envServer, adress)
	v.SetDefault(envStoreInterval, stroreInterval)
	v.SetDefault(envStoreFile, storeFile)
	v.SetDefault(envRestore, restore)
	v.SetDefault(envKey, key)
	v.SetDefault(envDataBaseDSN, dataBaseDSN)

	return &server.Config{
		Server: v.GetString(envServer),
		StorageConfig: datastorage.StorageConfig{
			StoreInterval: v.GetDuration(envStoreInterval),
			StoreFile:     v.GetString(envStoreFile),
			Restore:       v.GetBool(envRestore),
			Store:         v.GetString(envStoreFile) != "",
			Synchronized:  v.GetDuration(envStoreInterval) == time.Duration(0),
			Key:           v.GetString(envKey),
			DataBaseDSN:   v.GetString(envDataBaseDSN),
		},
	}
}
