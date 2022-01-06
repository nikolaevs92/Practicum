package datastorage

import (
	"context"
	"errors"
	"strconv"
)

const (
	GaugeTypeName   string = "gauge"
	CounterTypeName string = "counter"
)

type GaugeDataUpdate struct {
	Name     string
	Value    float64
	Responce chan bool
}

type CounterDataUpdate struct {
	Name     string
	Value    uint64
	Responce chan bool
}

type GasugeDataResponce struct {
	Value   float64
	Success bool
}

type CounterDataResponce struct {
	Value   uint64
	Success bool
}

type GaugeDataRequest struct {
	Name     string
	Responce chan GasugeDataResponce
}

type CounterDataRequest struct {
	Name     string
	Responce chan CounterDataResponce
}

type CollectedDataRequest struct {
	Responce chan CollectedDataResponce
}

type CollectedDataResponce struct {
	GaugeData   map[string]float64
	CounterData map[string]uint64
	Success     bool
}

type DataStorage struct {
	GaugeData          map[string]float64
	CounterData        map[string]uint64
	GaugeUpdateChan    chan GaugeDataUpdate
	CounterUpdateChan  chan CounterDataUpdate
	GaugeRequestChan   chan GaugeDataRequest
	CounterRequestChan chan CounterDataRequest
	RequestChan        chan CollectedDataRequest
}

func (data *DataStorage) Init() {
	data.GaugeData = map[string]float64{}
	data.CounterData = map[string]uint64{}
	data.GaugeUpdateChan = make(chan GaugeDataUpdate, 1024)
	data.CounterUpdateChan = make(chan CounterDataUpdate, 1024)
	data.GaugeRequestChan = make(chan GaugeDataRequest, 1024)
	data.CounterRequestChan = make(chan CounterDataRequest, 1024)
	data.RequestChan = make(chan CollectedDataRequest, 1024)
}

// func (data *DataStorage) GetCounterData() map[string]uint64 {
// 	return data.CounterData
// }

// func (data *DataStorage) GetGaugeData() map[string]float64 {
// 	return data.GaugeData
// }

func New() *DataStorage {
	dataStorage := new(DataStorage)
	dataStorage.Init()
	return dataStorage
}

func (data *DataStorage) RunReciver(end context.Context) {
	for {
		select {
		case update := <-data.GaugeUpdateChan:
			data.GaugeData[update.Name] = update.Value
			update.Responce <- true
		case update := <-data.CounterUpdateChan:
			data.CounterData[update.Name] += update.Value
			update.Responce <- true
		case request := <-data.GaugeRequestChan:
			value, ok := data.GaugeData[request.Name]
			request.Responce <- GasugeDataResponce{value, ok}
		case request := <-data.CounterRequestChan:
			value, ok := data.CounterData[request.Name]
			request.Responce <- CounterDataResponce{value, ok}
		case request := <-data.RequestChan:
			request.Responce <- CollectedDataResponce{data.GaugeData, data.CounterData, true}
		case <-end.Done():
			return
		}
	}
}

func (data *DataStorage) GetUpdate(metricType string, metricName string, metricValue string) error {
	if metricName == "" {
		return errors.New("DataStorage: GetUpdate: metricName should be not empty")
	}

	responceChan := make(chan bool, 1)

	switch metricType {
	case GaugeTypeName:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return errors.New("DataStorage: GetUpdate: error whith parsing gauge metricValue: ") // + err.GetString())
		}
		data.GaugeUpdateChan <- GaugeDataUpdate{metricName, value, responceChan}

	case CounterTypeName:
		value, err := strconv.ParseUint(metricValue, 10, 64)
		if err != nil {
			return errors.New("DataStorage: GetUpdate: error whith parsing counter metricValue: ") // + err.GetString())
		}
		data.CounterUpdateChan <- CounterDataUpdate{metricName, value, responceChan}

	default:
		return errors.New(
			"DataStorage: GetUpdate: invalid metricType value, valid values: " + GaugeTypeName + ", " + CounterTypeName)
	}

	success := <-responceChan
	if !success {
		return errors.New("DataStorage: GetUpdate: some error")
	}

	return nil
}

func (data *DataStorage) GetGaugeValue(metricName string) (float64, error) {
	if metricName == "" {
		return 0, errors.New("DataStorage: GetGaugeValue: metricName should be not empty")
	}
	responceChan := make(chan GasugeDataResponce, 1)
	data.GaugeRequestChan <- GaugeDataRequest{metricName, responceChan}

	responce := <-responceChan
	if responce.Success {
		return responce.Value, nil
	} else {
		return 0, errors.New("DataStorage: GetGaugeValue: some error")
	}
}

func (data *DataStorage) GetCounterValue(metricName string) (uint64, error) {
	if metricName == "" {
		return 0, errors.New("DataStorage: GetCounterValue: metricName should be not empty")
	}
	responceChan := make(chan CounterDataResponce, 1)
	data.CounterRequestChan <- CounterDataRequest{metricName, responceChan}

	responce := <-responceChan
	if responce.Success {
		return responce.Value, nil
	} else {
		return 0, errors.New("DataStorage: GetCounterValue: some error")
	}
}

func (data *DataStorage) GetStats() (map[string]float64, map[string]uint64, error) {
	responceChan := make(chan CollectedDataResponce, 1)
	data.RequestChan <- CollectedDataRequest{responceChan}
	responce := <-responceChan

	if responce.Success {
		return responce.GaugeData, responce.CounterData, nil
	} else {
		return nil, nil, errors.New("DataStorage: GetStats: some error")
	}
}
