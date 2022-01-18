package datastorage

import (
	"fmt"
	"strconv"
)

type Metrics struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta uint64  `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (metrics *Metrics) GetStrValue() string {
	switch metrics.MType {
	case GaugeTypeName:
		return strconv.FormatFloat(metrics.Value, 'g', 20, 64)
	case CounterTypeName:
		return strconv.FormatUint(metrics.Delta, 10)
	default:
		return ""
	}
}

func (metrics Metrics) String() string {
	return fmt.Sprintf("ID:%v MType:%v Value:%f Delta:%d)", metrics.ID, metrics.MType, metrics.Value, metrics.Delta)
}
