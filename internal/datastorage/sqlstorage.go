package datastorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/lib/pq"
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
	var queryTemplate string
	switch storage.cfg.DBType {
	case "sqlite3":
		queryTemplate = "INSERT INTO data VALUES(?, ?, ?, ?) ON CONFLICT (ID) DO UPDATE SET Delta = ?, Value = ?;"
	case "postgres":
		queryTemplate = "INSERT INTO data VALUES($N, $N, $N, $N) ON CONFLICT (ID) DO UPDATE SET Delta = $N, Value = $N;"
	}

	switch metricType {
	case GaugeTypeName:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return errors.New("DataStorage: GetUpdate: error whith parsing gauge metricValue: ") // + err.GetString())
		}
		res, err := storage.DB.ExecContext(storage.ctx, queryTemplate, metricName, metricType, 0, value, 0, value)
		if err != nil {
			return errors.New("DataStorage: GetUpdate: error whith upsert to DB: ") // + err.GetString())
		}
		count, err := res.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(count)

	case CounterTypeName:
		value, err := strconv.ParseUint(metricValue, 10, 64)
		if err != nil {
			return errors.New("DataStorage: GetUpdate: error whith parsing counter metricValue: ") // + err.GetString())
		}
		res, err := storage.DB.ExecContext(storage.ctx, queryTemplate, metricName, metricType, value, 0, value, 0)
		if err != nil {
			return errors.New("DataStorage: GetUpdate: error whith upsert to DB: ") // + err.GetString())
		}
		count, err := res.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(count)
	default:
		return errors.New(
			"DataStorage: GetUpdate: invalid metricType value, valid values: " + GaugeTypeName + ", " + CounterTypeName)
	}
	// row := storage.DB.QueryRowContext(storage.ctx, "SELECT Value, Delta from data where ID = ? limit 1;", metricName)
	// var value float64
	// var delta uint64
	// err := row.Scan(&value, &delta)
	// if err == nil {
	// 	log.Println("all ok")
	// } else {
	// 	log.Println(err.Error())
	// }
	return nil
}

func (storage *SQLStorage) GetGaugeValue(metricName string) (float64, error) {
	var queryTemplate string
	switch storage.cfg.DBType {
	case "sqlite3":
		queryTemplate = "SELECT Value FROM data WHERE ID = ? and MType = \"gauge\" limit 1;"
	case "postgres":
		queryTemplate = "SELECT Value FROM data WHERE ID = $N and MType = \"gauge\" limit 1;"
	}

	row := storage.DB.QueryRowContext(storage.ctx, queryTemplate, metricName)
	var res float64
	err := row.Scan(&res)
	if err != nil {
		return 0, errors.New("no data")
	}

	return res, nil
}

func (storage *SQLStorage) GetCounterValue(metricName string) (uint64, error) {
	var queryTemplate string
	switch storage.cfg.DBType {
	case "sqlite3":
		queryTemplate = "SELECT Delta FROM data WHERE ID = ? and MType = \"counter\" limit 1;"
	case "postgres":
		queryTemplate = "SELECT Delta FROM data WHERE ID = $N and MType = \"counter\" limit 1;"
	}

	row := storage.DB.QueryRowContext(storage.ctx, queryTemplate, metricName)
	var res uint64
	err := row.Scan(&res)
	if err != nil {
		return 0, errors.New("no data")
	}

	return res, nil
}

func (storage *SQLStorage) GetStats() (map[string]float64, map[string]uint64, error) {
	return map[string]float64{}, map[string]uint64{}, nil
}

func (storage *SQLStorage) Init() {
}

func (storage *SQLStorage) RunReciver(end context.Context) {
	storage.ctx = end

	db, err := sql.Open(storage.cfg.DBType, storage.cfg.DataBaseDSN)
	storage.DB = db
	if err != nil {
		log.Println("sql arent opened")
		log.Println(err)
		return
	}
	defer db.Close()

	// create table
	_, err = storage.DB.ExecContext(storage.ctx, "CREATE TABLE IF NOT EXISTS data ( ID text PRIMARY KEY, MType text, Delta integer, Value double precision )")
	if err != nil {
		log.Println("table arent created")
		log.Println(err)
		return
	}
	<-storage.ctx.Done()
}

func (storage *SQLStorage) Ping() bool {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	err := storage.DB.PingContext(ctx)
	return err == nil
}

func (storage *SQLStorage) GetJSONUpdate(jsonDump []byte) error {
	metrics := Metrics{}
	if err := json.Unmarshal(jsonDump, &metrics); err != nil {
		return err
	}

	metricsHash, _ := metrics.CalcHash(storage.cfg.Key)
	if storage.cfg.Key != "" && metricsHash != metrics.Hash {
		log.Println("Wrong hash, " + metricsHash + " " + metrics.Hash)
		return errors.New("wrong hash")
	}

	return storage.GetUpdate(metrics.MType, metrics.ID, metrics.GetStrValue())
}

func (storage *SQLStorage) GetJSONValue(jsonDump []byte) ([]byte, error) {
	metrics := Metrics{}
	if err := json.Unmarshal(jsonDump, &metrics); err != nil {
		return nil, err
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

	metrics.Hash, _ = metrics.CalcHash(storage.cfg.Key)
	res, err := metrics.MarshalJSON()
	if err != nil {
		return jsonDump, errors.New("error on encoding json")
	}
	return res, nil
}
