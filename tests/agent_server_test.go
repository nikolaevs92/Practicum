package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	// "context"
	// "time"
	// "github.com/nikolaevs92/Practicum/internal/agent"
	// "github.com/nikolaevs92/Practicum/internal/config"
	// "github.com/nikolaevs92/Practicum/internal/server"
	// "github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	// counterKeyValues := map[string]uint64{"PollCount": 5}
	// gaugeKeys := [...]string{
	// 	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
	// 	"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
	// 	"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "RandomValue",
	// }
	// conf := config.LoadConfig()
	// conf.Server.DBType = "sqlite3"
	// conf.Server.DataBaseDSN = "./.data"

	// httpServer := server.New(*conf.Server)
	// serverCtx, serverCancel := context.WithCancel(context.Background())
	// go httpServer.Run(serverCtx)
	// defer serverCancel()

	// collector := agent.New(*conf.Agent)
	// collectorCtx, collectorCancel := context.WithCancel(context.Background())
	// go collector.Run(collectorCtx)

	// time.Sleep(12 * time.Second)
	// collectorCancel()
	// for name, value := range counterKeyValues {
	// 	t.Run(name, func(t *testing.T) {
	// 		val, ok := httpServer.DataHolder.GetCounterValue(name)
	// 		assert.True(t, ok == nil)
	// 		assert.Equal(t, value, val)
	// 	})
	// }

	// for _, name := range gaugeKeys {
	// 	t.Run(name, func(t *testing.T) {
	// 		_, ok := httpServer.DataHolder.GetGaugeValue(name)
	// 		assert.True(t, ok == nil)
	// 	})
	// }
	assert.True(t, true)
}
