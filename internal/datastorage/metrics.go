package datastorage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	Hash  string  `json:"hash,omitempty"`  // значение хеш-функции
}

func (metrics *Metrics) CalcHash(key string) (string, error) {
	if key == "" {
		return "", nil
	}
	h := hmac.New(sha256.New, []byte(key))

	switch metrics.MType {
	case GaugeTypeName:
		h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metrics.ID, metrics.Value)))
	case CounterTypeName:
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", metrics.ID, metrics.Delta)))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
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
			Hash  string `json:"hash,omitempty"`
		}{
			ID:    metrics.ID,
			MType: metrics.MType,
			Delta: metrics.Delta,
			Hash:  metrics.Hash,
		}
		return json.Marshal(aliasValue)
	case GaugeTypeName:
		aliasValue := &struct {
			ID    string  `json:"id"`    // имя метрики
			MType string  `json:"type"`  // параметр, принимающий значение gauge или counter
			Value float64 `json:"value"` // значение метрики в случае передачи gauge
			Hash  string  `json:"hash,omitempty"`
		}{
			ID:    metrics.ID,
			MType: metrics.MType,
			Value: metrics.Value,
			Hash:  metrics.Hash,
		}
		return json.Marshal(aliasValue)
	default:
		return nil, errors.New("wrong MType")
	}
}

func (metrics Metrics) String() string {
	return fmt.Sprintf("ID:%v MType:%v Value:%f Delta:%d", metrics.ID, metrics.MType, metrics.Value, metrics.Delta)
}
