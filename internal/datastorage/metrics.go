package datastorage

import (
	"encoding/json"
	"errors"
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

func (metrics *Metrics) MarshalJSON() ([]byte, error) {
	switch metrics.MType {
	case CounterTypeName:
		aliasValue := &struct {
			ID    string `json:"id"`   // имя метрики
			MType string `json:"type"` // параметр, принимающий значение gauge или counter
			Delta uint64 `json:"delta"`
		}{
			ID:    metrics.ID,
			MType: metrics.MType,
			Delta: metrics.Delta,
		}
		return json.Marshal(aliasValue)
	case GaugeTypeName:
		aliasValue := &struct {
			ID    string  `json:"id"`    // имя метрики
			MType string  `json:"type"`  // параметр, принимающий значение gauge или counter
			Value float64 `json:"value"` // значение метрики в случае передачи gauge
		}{
			ID:    metrics.ID,
			MType: metrics.MType,
			Value: metrics.Value,
		}
		return json.Marshal(aliasValue)
	default:
		return nil, errors.New("wrong MType")
	}
}

func (metrics Metrics) String() string {
	return fmt.Sprintf("ID:%v MType:%v Value:%f Delta:%d", metrics.ID, metrics.MType, metrics.Value, metrics.Delta)
}
