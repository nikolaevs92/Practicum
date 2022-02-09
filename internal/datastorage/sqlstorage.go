package datastorage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLStorage struct {
	cfg StorageConfig
	ctx context.Context
	DB  *sql.DB
}

func NewSQLStorage(cfg StorageConfig) *SQLStorage {
	dataStorage := new(SQLStorage)
	dataStorage.cfg = cfg
	return dataStorage
}

func (storage *SQLStorage) GetUpdate(metricType string, metricName string, metricValue string) error {
	return nil
}

func (storage *SQLStorage) GetGaugeValue(metricName string) (float64, error) {
	return 0, nil
}

func (storage *SQLStorage) GetCounterValue(metricName string) (uint64, error) {
	return 0, nil
}

func (storage *SQLStorage) GetStats() (map[string]float64, map[string]uint64, error) {
	return map[string]float64{}, map[string]uint64{}, nil
}

func (storage *SQLStorage) Init() {
}

func (storage *SQLStorage) RunReciver(end context.Context) {
	storage.ctx = end

	db, err := sql.Open("sqlite3", "db.db")
	if err != nil {
		panic(err)
	}
	storage.DB = db

	defer db.Close()
	<-storage.ctx.Done()
}

func (storage *SQLStorage) Ping() bool {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	err := storage.DB.PingContext(ctx)
	return err == nil
}

func (storage *SQLStorage) GetJSONUpdate(jsonDump []byte) error {
	return nil
}

func (storage *SQLStorage) GetJSONValue(jsonDump []byte) ([]byte, error) {
	return []byte{}, nil
}
