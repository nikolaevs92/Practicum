package datastorage

import (
	"fmt"
	"time"
)

type StorageConfig struct {
	StoreInterval time.Duration
	StoreFile     string
	DataBaseDSN   string
	DBType        string
	Restore       bool
	Store         bool
	Synchronized  bool
	Key           string
}

func (cfg StorageConfig) String() string {
	if cfg.Store {
		return fmt.Sprintf(
			"Store:%t Restore:%t StoreInterval:%ds StoreFile:%s",
			cfg.Store, cfg.Restore, int(cfg.StoreInterval.Seconds()), cfg.StoreFile)
	} else {
		return fmt.Sprintf("Store:%t Restore:%t", cfg.Store, cfg.Restore)
	}
}
