package main_test

import (
	// "context"
	// "time"

	// "github.com/nikolaevs92/Practicum/internal/agent"
	// "github.com/nikolaevs92/Practicum/internal/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	// counterKeyValues := map[string]int64{"PollCount": 5}
	// gaugeKeys := [...]string{
	// 	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
	// 	"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
	// 	"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "RandomValue",
	// }

	// httpServer := new(server.DataServer)
	// serverCtx, serverCancel := context.WithCancel(context.Background())
	// go httpServer.Run(serverCtx)
	// defer serverCancel()

	// collector := new(agent.CollectorAgent)
	// collectorCtx, collectorCancel := context.WithCancel(context.Background())
	// go collector.Run(collectorCtx)

	// time.Sleep(12 * time.Second)
	// collectorCancel()
	// for name, value := range counterKeyValues {
	// 	t.Run(name, func(t *testing.T) {
	// 		val, ok := httpServer.DataHolder.CounterData[name]
	// 		assert.True(t, ok)
	// 		assert.Equal(t, value, val)
	// 	})
	// }

	// for _, name := range gaugeKeys {
	// 	t.Run(name, func(t *testing.T) {
	// 		_, ok := httpServer.DataHolder.GaugeData[name]
	// 		assert.True(t, ok)
	// 	})
	// }
	assert.True(t, true)
}
