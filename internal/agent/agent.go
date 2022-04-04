package agent

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/nikolaevs92/Practicum/internal/datastorage"
)

type Config struct {
	Server         string
	PollInterval   time.Duration
	ReportInterval time.Duration
	ReportRetries  int
	Key            string
}

const (
	gaugeTypeName   string = "gauge"
	counterTypeName string = "counter"
)

type CollectorAgent struct {
	cfg Config

	stats          runtime.MemStats
	TotalMemory    uint64
	FreeMemory     uint64
	CPUutilization map[string]float64
	PollCount      uint64
	RandomValue    float64
	mu             sync.RWMutex
}

func New(config Config) *CollectorAgent {
	collector := new(CollectorAgent)
	collector.cfg = config
	collector.CPUutilization = make(map[string]float64)
	return collector
}

func (collector *CollectorAgent) Collect(t time.Time) {
	collector.mu.Lock()
	defer collector.mu.Unlock()

	log.Println("Start collect stat")

	runtime.ReadMemStats(&collector.stats)

	v, _ := mem.VirtualMemory()
	collector.TotalMemory = v.Total
	collector.FreeMemory = v.Free
	for i := 1; i <= runtime.NumCPU(); i++ {
		collector.CPUutilization[fmt.Sprintf("CPUutilization%d", i)] = rand.Float64()
	}

	collector.RandomValue = rand.Float64()
	collector.PollCount++

	log.Println("End collect stat")
}

func (collector *CollectorAgent) PostWithRetrues(url string, contentType string, body []byte) (*http.Response, error) {
	resp, err := http.Post(url, contentType, bytes.NewReader(body))
	for i := 0; i < collector.cfg.ReportRetries && err != nil; i++ {
		resp, err = http.Post(url, "application/json", bytes.NewReader(body))
	}
	return resp, err
}

func (collector *CollectorAgent) PostOneStat(metrics datastorage.Metrics) {
	log.Println("Post one stat to " + collector.cfg.Server)
	log.Println(metrics)
	url := "http://" + path.Join(collector.cfg.Server, "update")

	metrics.Hash, _ = metrics.CalcHash(collector.cfg.Key)
	body, err := metrics.MarshalJSON()
	if err != nil {
		log.Println("Error while marshal " + err.Error())
		return
	}
	resp, err := collector.PostWithRetrues(url, "application/json", body)
	if err != nil {
		log.Println("Post error" + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf(url, " status code ", resp.StatusCode)
	}
	log.Println("Post one stat: succesed")
}

func (collector *CollectorAgent) PostOneGaugeStat(metricName string, metricValue float64) {
	collector.PostOneStat(datastorage.Metrics{
		ID:    metricName,
		MType: gaugeTypeName,
		Value: metricValue,
	})
}

func (collector *CollectorAgent) PostOneCounterStat(metricName string, metricValue uint64) {
	collector.PostOneStat(datastorage.Metrics{
		ID:    metricName,
		MType: counterTypeName,
		Delta: metricValue,
	})
}

func (collector *CollectorAgent) Report(t time.Time) {
	collector.mu.RLock()
	defer collector.mu.RUnlock()

	go collector.PostOneGaugeStat("Alloc", float64(collector.stats.Alloc))
	go collector.PostOneGaugeStat("TotalAlloc", float64(collector.stats.TotalAlloc))
	go collector.PostOneGaugeStat("Frees", float64(collector.stats.Frees))
	go collector.PostOneGaugeStat("BuckHashSys", float64(collector.stats.BuckHashSys))
	go collector.PostOneGaugeStat("Frees", float64(collector.stats.Frees))
	go collector.PostOneGaugeStat("GCCPUFraction", float64(collector.stats.GCCPUFraction))
	go collector.PostOneGaugeStat("GCSys", float64(collector.stats.GCSys))
	go collector.PostOneGaugeStat("HeapAlloc", float64(collector.stats.HeapAlloc))
	go collector.PostOneGaugeStat("HeapIdle", float64(collector.stats.HeapIdle))
	go collector.PostOneGaugeStat("HeapInuse", float64(collector.stats.HeapInuse))
	go collector.PostOneGaugeStat("HeapObjects", float64(collector.stats.HeapObjects))
	go collector.PostOneGaugeStat("HeapReleased", float64(collector.stats.HeapReleased))
	go collector.PostOneGaugeStat("HeapSys", float64(collector.stats.HeapSys))
	go collector.PostOneGaugeStat("LastGC", float64(collector.stats.LastGC))
	go collector.PostOneGaugeStat("Lookups", float64(collector.stats.Lookups))
	go collector.PostOneGaugeStat("MCacheInuse", float64(collector.stats.MCacheInuse))
	go collector.PostOneGaugeStat("MCacheSys", float64(collector.stats.MCacheSys))
	go collector.PostOneGaugeStat("MSpanInuse", float64(collector.stats.MSpanInuse))
	go collector.PostOneGaugeStat("MSpanSys", float64(collector.stats.MSpanSys))
	go collector.PostOneGaugeStat("Mallocs", float64(collector.stats.Mallocs))
	go collector.PostOneGaugeStat("NextGC", float64(collector.stats.NextGC))
	go collector.PostOneGaugeStat("NumForcedGC", float64(collector.stats.NumForcedGC))
	go collector.PostOneGaugeStat("NumGC", float64(collector.stats.NumGC))
	go collector.PostOneGaugeStat("OtherSys", float64(collector.stats.OtherSys))
	go collector.PostOneGaugeStat("PauseTotalNs", float64(collector.stats.PauseTotalNs))
	go collector.PostOneGaugeStat("StackInuse", float64(collector.stats.StackInuse))
	go collector.PostOneGaugeStat("StackSys", float64(collector.stats.StackSys))
	go collector.PostOneGaugeStat("Sys", float64(collector.stats.Sys))

	go collector.PostOneCounterStat("FreeMemory", collector.FreeMemory)
	go collector.PostOneCounterStat("TotalMemory", collector.TotalMemory)
	for i := 1; i <= runtime.NumCPU(); i++ {
		metricName := fmt.Sprintf("CPUutilization%d", i)
		collector.PostOneGaugeStat(metricName, collector.CPUutilization[metricName])
	}

	go collector.PostOneGaugeStat("RandomValue", float64(collector.RandomValue))

	go collector.PostOneCounterStat("PollCount", collector.PollCount)
}

func (collector *CollectorAgent) Run(end context.Context) error {
	log.Println("Collector run started")

	collectTimer := time.NewTicker(collector.cfg.PollInterval)
	reportTimer := time.NewTicker(collector.cfg.ReportInterval)

	for {
		select {
		case t := <-collectTimer.C:
			go collector.Collect(t)
		case t := <-reportTimer.C:
			go collector.Report(t)
		case <-end.Done():
			log.Println("Collector stoped")
			return nil
		}
	}
}
