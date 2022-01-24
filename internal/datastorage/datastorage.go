package datastorage

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"time"
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

type StoredData struct {
	GaugeData   map[string]float64
	CounterData map[string]uint64

	storedTS time.Time
}

type DataStorage struct {
	Data               StoredData
	GaugeUpdateChan    chan GaugeDataUpdate
	CounterUpdateChan  chan CounterDataUpdate
	GaugeRequestChan   chan GaugeDataRequest
	CounterRequestChan chan CounterDataRequest
	RequestChan        chan CollectedDataRequest

	cfg StorageConfig
}

func (storage *DataStorage) Init() {
	storage.GaugeUpdateChan = make(chan GaugeDataUpdate, 1024)
	storage.CounterUpdateChan = make(chan CounterDataUpdate, 1024)
	storage.GaugeRequestChan = make(chan GaugeDataRequest, 1024)
	storage.CounterRequestChan = make(chan CounterDataRequest, 1024)
	storage.RequestChan = make(chan CollectedDataRequest, 1024)
}

func (storage *DataStorage) RestoreData() error {
	if !(storage.cfg.Restore && storage.cfg.Store) {
		log.Println("No data restoring")
		storage.Data.GaugeData = map[string]float64{}
		storage.Data.CounterData = map[string]uint64{}
		return nil
	}
	log.Println("Start restore data from: " + storage.cfg.StoreFile)

	file, err := os.OpenFile(storage.cfg.StoreFile, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&storage.Data)
	switch err {
	case io.EOF:
		storage.Data.GaugeData = map[string]float64{}
		storage.Data.CounterData = map[string]uint64{}
	default:
		return err
	}

	log.Println("Restore data: succesed")
	return nil
}

func (storage *DataStorage) StoreData(t time.Time) error {
	if !storage.cfg.Store {
		return nil
	}

	file, err := os.OpenFile(storage.cfg.StoreFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	storage.Data.storedTS = t
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(&storage.Data); err != io.EOF && err != nil {
		return err
	}

	log.Println("Store data: succesed")
	return nil
}

func New(cfg StorageConfig) *DataStorage {
	log.Println("Create Storage")
	log.Println(cfg)
	dataStorage := new(DataStorage)
	dataStorage.Init()
	dataStorage.cfg = cfg
	if err := dataStorage.RestoreData(); err != nil {
		panic(err)
	}
	return dataStorage
}

func (storage *DataStorage) RunReciver(end context.Context) {
	log.Println("Start Reciver")
	var storeTimer *time.Ticker
	if storage.cfg.StoreInterval > 0 {
		storeTimer = time.NewTicker(storage.cfg.StoreInterval)
	} else {
		storeTimer = time.NewTicker(1)
		storeTimer.Stop()
	}

	for {
		select {
		case update := <-storage.GaugeUpdateChan:
			storage.Data.GaugeData[update.Name] = update.Value
			update.Responce <- true
		case update := <-storage.CounterUpdateChan:
			storage.Data.CounterData[update.Name] += update.Value
			update.Responce <- true
		case request := <-storage.GaugeRequestChan:
			value, ok := storage.Data.GaugeData[request.Name]
			request.Responce <- GasugeDataResponce{value, ok}
		case request := <-storage.CounterRequestChan:
			value, ok := storage.Data.CounterData[request.Name]
			request.Responce <- CounterDataResponce{value, ok}
		case request := <-storage.RequestChan:
			request.Responce <- CollectedDataResponce{storage.Data.GaugeData, storage.Data.CounterData, true}
		case t := <-storeTimer.C:
			_ = storage.StoreData(t)
		case <-end.Done():
			log.Println("End Reciver")
			return
		}
	}
}

func (storage *DataStorage) GetUpdate(metricType string, metricName string, metricValue string) error {
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
		storage.GaugeUpdateChan <- GaugeDataUpdate{metricName, value, responceChan}

	case CounterTypeName:
		value, err := strconv.ParseUint(metricValue, 10, 64)
		if err != nil {
			return errors.New("DataStorage: GetUpdate: error whith parsing counter metricValue: ") // + err.GetString())
		}
		storage.CounterUpdateChan <- CounterDataUpdate{metricName, value, responceChan}

	default:
		return errors.New(
			"DataStorage: GetUpdate: invalid metricType value, valid values: " + GaugeTypeName + ", " + CounterTypeName)
	}

	success := <-responceChan
	if !success {
		return errors.New("DataStorage: GetUpdate: some error")
	}
	if storage.cfg.Synchronized {
		storage.StoreData(time.Now())
	}

	return nil
}

func (storage *DataStorage) GetJSONUpdate(jsonDump []byte) error {
	metrics := Metrics{}
	if err := json.Unmarshal(jsonDump, &metrics); err != nil {
		panic(err)
	}
	return storage.GetUpdate(metrics.MType, metrics.ID, metrics.GetStrValue())
}

func (storage *DataStorage) GetJSONValue(jsonDump []byte) ([]byte, error) {
	metrics := Metrics{}
	if err := json.Unmarshal(jsonDump, &metrics); err != nil {
		panic(err)
	}
	switch metrics.MType {
	case GaugeTypeName:
		value, err := storage.GetGaugeValue(metrics.ID)
		if err != nil {
			return jsonDump, err
		}
		metrics.Value = value
		metrics.Delta = 0

	case CounterTypeName:
		value, err := storage.GetCounterValue(metrics.ID)
		if err != nil {
			return jsonDump, err
		}
		metrics.Delta = value
		metrics.Value = 0
	default:
		return jsonDump, errors.New("Wrong MType: " + metrics.MType)
	}
	res, err := metrics.MarshalJSON()
	if err != nil {
		return jsonDump, errors.New("error on encoding json")
	}
	return res, nil
}

func (storage *DataStorage) GetGaugeValue(metricName string) (float64, error) {
	if metricName == "" {
		return 0, errors.New("DataStorage: GetGaugeValue: metricName should be not empty")
	}
	responceChan := make(chan GasugeDataResponce, 1)
	storage.GaugeRequestChan <- GaugeDataRequest{metricName, responceChan}

	responce := <-responceChan
	if responce.Success {
		return responce.Value, nil
	} else {
		return 0, errors.New("DataStorage: GetGaugeValue: some error")
	}
}

func (storage *DataStorage) GetCounterValue(metricName string) (uint64, error) {
	if metricName == "" {
		return 0, errors.New("DataStorage: GetCounterValue: metricName should be not empty")
	}
	responceChan := make(chan CounterDataResponce, 1)
	storage.CounterRequestChan <- CounterDataRequest{metricName, responceChan}

	responce := <-responceChan
	if responce.Success {
		return responce.Value, nil
	} else {
		return 0, errors.New("DataStorage: GetCounterValue: some error")
	}
}

func (storage *DataStorage) GetStats() (map[string]float64, map[string]uint64, error) {
	responceChan := make(chan CollectedDataResponce, 1)
	storage.RequestChan <- CollectedDataRequest{responceChan}
	responce := <-responceChan

	if responce.Success {
		return responce.GaugeData, responce.CounterData, nil
	} else {
		return nil, nil, errors.New("DataStorage: GetStats: some error")
	}
}
