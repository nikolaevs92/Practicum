package agent

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	gaugeTypeName   string = "gauge"
	counterTypeName string = "counter"
)

type CollectorAgent struct {
	Server         string
	PollInterval   int64
	ReportInterval int64

	PollCount   uint64
	RandomValue float64
	stats       runtime.MemStats
}

func (collector *CollectorAgent) Initite() {
	if collector.Server == "" {
		collector.Server = "127.0.0.1:8080"
	}
	if collector.PollInterval == 0 {
		collector.PollInterval = 2
	}
	if collector.ReportInterval == 0 {
		collector.ReportInterval = 10
	}
}

func NewCollectorAgent(server string, pollInterval int64, reportInterval int64) *CollectorAgent {
	collector := new(CollectorAgent)
	collector.Server = server
	collector.PollInterval = pollInterval
	collector.ReportInterval = reportInterval
	return collector
}

func (collector *CollectorAgent) Collect(t time.Time) {
	runtime.ReadMemStats(&collector.stats)
	collector.RandomValue = rand.Float64()
	collector.PollCount++
}

func PostOneGaugeStat(server string, metricName string, metricValue float64) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%f", server, gaugeTypeName, metricName, metricValue)
	resp, err := http.Post(url, "text/plain", strings.NewReader("body"))
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	// fmt.Println(url)
}

func PostOneCounterStat(server string, metricName string, metricValue uint64) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%d", server, counterTypeName, metricName, metricValue)
	resp, err := http.Post(url, "text/plain", strings.NewReader("body"))
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	// fmt.Println(url)
}

func (collector *CollectorAgent) Report(t time.Time) {
	go PostOneGaugeStat(collector.Server, "Alloc", float64(collector.stats.Alloc))
	go PostOneGaugeStat(collector.Server, "BuckHashSys", float64(collector.stats.BuckHashSys))
	go PostOneGaugeStat(collector.Server, "Frees", float64(collector.stats.Frees))
	go PostOneGaugeStat(collector.Server, "GCCPUFraction", float64(collector.stats.GCCPUFraction))
	go PostOneGaugeStat(collector.Server, "GCSys", float64(collector.stats.GCSys))
	go PostOneGaugeStat(collector.Server, "HeapAlloc", float64(collector.stats.HeapAlloc))
	go PostOneGaugeStat(collector.Server, "HeapIdle", float64(collector.stats.HeapIdle))
	go PostOneGaugeStat(collector.Server, "HeapInuse", float64(collector.stats.HeapInuse))
	go PostOneGaugeStat(collector.Server, "HeapObjects", float64(collector.stats.HeapObjects))
	go PostOneGaugeStat(collector.Server, "HeapReleased", float64(collector.stats.HeapReleased))
	go PostOneGaugeStat(collector.Server, "HeapSys", float64(collector.stats.HeapSys))
	go PostOneGaugeStat(collector.Server, "LastGC", float64(collector.stats.LastGC))
	go PostOneGaugeStat(collector.Server, "Lookups", float64(collector.stats.Lookups))
	go PostOneGaugeStat(collector.Server, "MCacheInuse", float64(collector.stats.MCacheInuse))
	go PostOneGaugeStat(collector.Server, "MCacheSys", float64(collector.stats.MCacheSys))
	go PostOneGaugeStat(collector.Server, "MSpanInuse", float64(collector.stats.MSpanInuse))
	go PostOneGaugeStat(collector.Server, "MSpanSys", float64(collector.stats.MSpanSys))
	go PostOneGaugeStat(collector.Server, "Mallocs", float64(collector.stats.Mallocs))
	go PostOneGaugeStat(collector.Server, "NextGC", float64(collector.stats.NextGC))
	go PostOneGaugeStat(collector.Server, "NumForcedGC", float64(collector.stats.NumForcedGC))
	go PostOneGaugeStat(collector.Server, "NumGC", float64(collector.stats.NumGC))
	go PostOneGaugeStat(collector.Server, "OtherSys", float64(collector.stats.OtherSys))
	go PostOneGaugeStat(collector.Server, "PauseTotalNs", float64(collector.stats.PauseTotalNs))
	go PostOneGaugeStat(collector.Server, "StackInuse", float64(collector.stats.StackInuse))
	go PostOneGaugeStat(collector.Server, "StackSys", float64(collector.stats.StackSys))
	go PostOneGaugeStat(collector.Server, "Sys", float64(collector.stats.Sys))

	go PostOneGaugeStat(collector.Server, "RandomValue", float64(collector.RandomValue))

	go PostOneCounterStat(collector.Server, "PollCount", collector.PollCount)
}

func (collector *CollectorAgent) Run(end context.Context) {
	collector.Initite()
	collectTimer := time.NewTicker(time.Duration(collector.PollInterval) * time.Second)
	reportTimer := time.NewTicker(time.Duration(collector.ReportInterval) * time.Second)
	for {
		select {
		case t := <-collectTimer.C:
			collector.Collect(t)
		case t := <-reportTimer.C:
			collector.Report(t)
		case <-end.Done():
			return
		}
	}
}

func RunAgentDefault() {
	collector := new(CollectorAgent)
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-cancelChan
		cancel()
	}()

	collector.Run(ctx)

	fmt.Println("Program end")
}
