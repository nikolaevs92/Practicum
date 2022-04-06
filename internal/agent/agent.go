package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
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

	v, err := mem.VirtualMemory()
	if err != nil {
		collector.TotalMemory = v.Total
		collector.FreeMemory = v.Free
	}
	c, err := cpu.Percent(time.Millisecond, true)
	if err != nil {
		for i := 1; i <= runtime.NumCPU(); i++ {
			collector.CPUutilization[fmt.Sprintf("CPUutilization%d", i)] = c[i]
		}
	} else {
		for i := 1; i <= runtime.NumCPU(); i++ {
			collector.CPUutilization[fmt.Sprintf("CPUutilization%d", i)] = 0
		}
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

func (collector *CollectorAgent) getMetrcisSlice() []datastorage.Metrics {
	collector.mu.RLock()
	defer collector.mu.RUnlock()

	metrics := []datastorage.Metrics{
		datastorage.Metrics{
			ID:    "Alloc",
			MType: gaugeTypeName,
			Value: float64(collector.stats.Alloc),
		},
		datastorage.Metrics{
			ID:    "TotalAlloc",
			MType: gaugeTypeName,
			Value: float64(collector.stats.TotalAlloc),
		},
		datastorage.Metrics{
			ID:    "Frees",
			MType: gaugeTypeName,
			Value: float64(collector.stats.Frees),
		},
		datastorage.Metrics{
			ID:    "BuckHashSys",
			MType: gaugeTypeName,
			Value: float64(collector.stats.BuckHashSys),
		},
		datastorage.Metrics{
			ID:    "GCCPUFraction",
			MType: gaugeTypeName,
			Value: float64(collector.stats.GCCPUFraction),
		},
		datastorage.Metrics{
			ID:    "GCSys",
			MType: gaugeTypeName,
			Value: float64(collector.stats.GCSys),
		},
		datastorage.Metrics{
			ID:    "HeapAlloc",
			MType: gaugeTypeName,
			Value: float64(collector.stats.HeapAlloc),
		},
		datastorage.Metrics{
			ID:    "HeapIdle",
			MType: gaugeTypeName,
			Value: float64(collector.stats.HeapIdle),
		},
		datastorage.Metrics{
			ID:    "HeapInuse",
			MType: gaugeTypeName,
			Value: float64(collector.stats.HeapInuse),
		},
		datastorage.Metrics{
			ID:    "HeapObjects",
			MType: gaugeTypeName,
			Value: float64(collector.stats.HeapObjects),
		},
		datastorage.Metrics{
			ID:    "HeapReleased",
			MType: gaugeTypeName,
			Value: float64(collector.stats.HeapReleased),
		},
		datastorage.Metrics{
			ID:    "HeapSys",
			MType: gaugeTypeName,
			Value: float64(collector.stats.HeapSys),
		},
		datastorage.Metrics{
			ID:    "LastGC",
			MType: gaugeTypeName,
			Value: float64(collector.stats.LastGC),
		},
		datastorage.Metrics{
			ID:    "Lookups",
			MType: gaugeTypeName,
			Value: float64(collector.stats.Lookups),
		},
		datastorage.Metrics{
			ID:    "MCacheInuse",
			MType: gaugeTypeName,
			Value: float64(collector.stats.MCacheInuse),
		},
		datastorage.Metrics{
			ID:    "MCacheSys",
			MType: gaugeTypeName,
			Value: float64(collector.stats.MCacheSys),
		},
		datastorage.Metrics{
			ID:    "MSpanInuse",
			MType: gaugeTypeName,
			Value: float64(collector.stats.MSpanInuse),
		},
		datastorage.Metrics{
			ID:    "MSpanSys",
			MType: gaugeTypeName,
			Value: float64(collector.stats.MSpanSys),
		},
		datastorage.Metrics{
			ID:    "Mallocs",
			MType: gaugeTypeName,
			Value: float64(collector.stats.Mallocs),
		},
		datastorage.Metrics{
			ID:    "NextGC",
			MType: gaugeTypeName,
			Value: float64(collector.stats.NextGC),
		},
		datastorage.Metrics{
			ID:    "NumForcedGC",
			MType: gaugeTypeName,
			Value: float64(collector.stats.NumForcedGC),
		},
		datastorage.Metrics{
			ID:    "NumGC",
			MType: gaugeTypeName,
			Value: float64(collector.stats.NumGC),
		},
		datastorage.Metrics{
			ID:    "OtherSys",
			MType: gaugeTypeName,
			Value: float64(collector.stats.OtherSys),
		},
		datastorage.Metrics{
			ID:    "PauseTotalNs",
			MType: gaugeTypeName,
			Value: float64(collector.stats.PauseTotalNs),
		},
		datastorage.Metrics{
			ID:    "StackInuse",
			MType: gaugeTypeName,
			Value: float64(collector.stats.StackInuse),
		},
		datastorage.Metrics{
			ID:    "StackSys",
			MType: gaugeTypeName,
			Value: float64(collector.stats.StackSys),
		},
		datastorage.Metrics{
			ID:    "Sys",
			MType: gaugeTypeName,
			Value: float64(collector.stats.Sys),
		},

		datastorage.Metrics{
			ID:    "RandomValue",
			MType: gaugeTypeName,
			Value: float64(collector.RandomValue),
		},

		datastorage.Metrics{
			ID:    "FreeMemory",
			MType: counterTypeName,
			Delta: collector.FreeMemory,
		},
		datastorage.Metrics{
			ID:    "TotalMemory",
			MType: counterTypeName,
			Delta: collector.TotalMemory,
		},
		datastorage.Metrics{
			ID:    "PollCount",
			MType: counterTypeName,
			Delta: collector.PollCount,
		},
	}

	for i := 1; i <= runtime.NumCPU(); i++ {
		metricName := fmt.Sprintf("CPUutilization%d", i)

		metrics = append(metrics, datastorage.Metrics{
			ID:    metricName,
			MType: gaugeTypeName,
			Value: collector.CPUutilization[metricName],
		})
	}

	return metrics
}

func (collector *CollectorAgent) Report(t time.Time) {
	metrics := collector.getMetrcisSlice()

	log.Println("Post batch stats to " + collector.cfg.Server)
	log.Println(metrics)
	url := "http://" + path.Join(collector.cfg.Server, "updates")

	for i := range metrics {
		metrics[i].Hash, _ = metrics[i].CalcHash(collector.cfg.Key)
	}

	body, err := json.Marshal(metrics)
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
	log.Println("Post batch stats: succesed")
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
