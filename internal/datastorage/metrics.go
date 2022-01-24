package datastorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Metrics struct {
	ID    string  `json:"id"`    // имя метрики
	MType string  `json:"type"`  // параметр, принимающий значение gauge или counter
	Delta uint64  `json:"delta"` // значение метрики в случае передачи counter
	Value float64 `json:"value"` // значение метрики в случае передачи gauge
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
	type MetricsAlias Metrics

	switch m.MType {
	case CounterTypeName:
		aliasValue := &struct {
			*MetricsAlias
			// переопределяем поле внутри анонимной структуры
			Value float64 `json:"value,omitempty"`
		}{
			// задаём указатель на целевой объект
			MetricsAlias: (*MetricsAlias)(m),
		}
		return json.Marshal(aliasValue)
	case GaugeTypeName:
		aliasValue := &struct {
			*MetricsAlias
			// переопределяем поле внутри анонимной структуры
			Delta uint64 `json:"delta,omitempty"`
		}{
			// задаём указатель на целевой объект
			MetricsAlias: (*MetricsAlias)(m),
		}
		return json.Marshal(aliasValue)
	default:
		return nil, errors.New("wrong MType")
	}
}

func (metrics Metrics) String() string {
	return fmt.Sprintf("ID:%v MType:%v Value:%f Delta:%d", metrics.ID, metrics.MType, metrics.Value, metrics.Delta)
}
