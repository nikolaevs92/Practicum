package datastorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

func (m *Metrics) MarshalJSON() ([]byte, error) {
	switch m.MType {
	case CounterTypeName:
		aliasValue := &struct {
			ID    string  `json:"id"`              // имя метрики
			MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
			Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
			Delta uint64  `json:"delta"`
		}{
			ID:    m.ID,
			MType: m.MType,
			Delta: m.Delta,
			Value: 0,
		}
		body, _ := json.Marshal(aliasValue)
		log.Println(string(body))
		return body, nil
	case GaugeTypeName:
		aliasValue := &struct {
			ID    string  `json:"id"`   // имя метрики
			MType string  `json:"type"` // параметр, принимающий значение gauge или counter
			Delta uint64  `json:"delta,omitempty"`
			Value float64 `json:"value"` // значение метрики в случае передачи gauge
		}{
			ID:    m.ID,
			MType: m.MType,
			Value: m.Value,
			Delta: 0,
		}
		body, _ := json.Marshal(aliasValue)
		log.Println(string(body))
		return body, nil
	default:
		return nil, errors.New("wrong MType")
	}
}

func (metrics Metrics) String() string {
	return fmt.Sprintf("ID:%v MType:%v Value:%f Delta:%d", metrics.ID, metrics.MType, metrics.Value, metrics.Delta)
}
