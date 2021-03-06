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

func (storage *SQLStorage) GetJSONArray(jsonDump []byte) ([]byte, error) {
	log.Println("Start butch update")

	metricsArray := []Metrics{}
	log.Println(string(jsonDump))
	if err := json.Unmarshal(jsonDump, &metricsArray); err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("json parsed")

	for _, el := range metricsArray {
		metricsHash, _ := el.CalcHash(storage.cfg.Key)
		if storage.cfg.Key != "" && metricsHash != el.Hash {
			log.Println("Wrong hash, " + metricsHash + " " + el.Hash)
			return nil, errors.New("wrong hash")
		}
	}

	tx, err := storage.DB.Begin()
	if err != nil {
		log.Println("Transaxtion didnt started: " + err.Error())
		return nil, err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback()

	// шаг 2 — готовим инструкцию
	var queryTemplate string
	switch storage.cfg.DBType {
	case "sqlite3":
		queryTemplate = "INSERT INTO statistics5 VALUES(?, ?, ?, ?) ON CONFLICT (ID, MType) DO UPDATE SET Delta = statistics5.Delta + ?, Value = ?;"
	case "postgres":
		queryTemplate = "INSERT INTO statistics5 VALUES($1, $2, $3, $4) ON CONFLICT (ID, MType) DO UPDATE SET Delta = statistics5.Delta + $5, Value = $6;"
	}
	stmt, err := tx.PrepareContext(storage.ctx, queryTemplate)
	if err != nil {
		log.Println("Context didnt prepared: " + err.Error())
		return nil, err
	}

	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer stmt.Close()

	for _, metric := range metricsArray {
		log.Println("insert metric: " + metric.String())
		if _, err = stmt.ExecContext(storage.ctx, metric.ID, metric.MType, metric.Delta, metric.Value, metric.Delta, metric.Value); err != nil {
			log.Println("Metric didnt insert: " + metric.String() + ". Error: " + err.Error())
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return metricsArray[0].MarshalJSON()
}

func (storage *SQLStorage) GetUpdate(metricType string, metricName string, metricValue string) error {
	log.Printf("Update start: ID:%v MType:%v Value:%s\n", metricName, metricType, metricValue)
	var queryTemplate string
	switch storage.cfg.DBType {
	case "sqlite3":
		queryTemplate = "INSERT INTO statistics5 VALUES(?, ?, ?, ?) ON CONFLICT (ID, MType) DO UPDATE SET Delta = statistics5.Delta + ?, Value = ?;"
	case "postgres":
		queryTemplate = "INSERT INTO statistics5 VALUES($1, $2, $3, $4) ON CONFLICT (ID, MType) DO UPDATE SET Delta = statistics5.Delta + $5, Value = $6;"
	}

	switch metricType {
	case GaugeTypeName:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			log.Println("DataStorage: GetUpdate: error whith parsing gauge metricValue: " + err.Error())
			return errors.New("DataStorage: GetUpdate: error whith parsing gauge metricValue: " + err.Error())
		}
		res, err := storage.DB.ExecContext(storage.ctx, queryTemplate, metricName, metricType, 0, value, 0, value)
		if err != nil {
			log.Println("DataStorage: GetUpdate: error whith upsert to DB: " + err.Error())
			return errors.New("DataStorage: GetUpdate: error whith upsert to DB: " + err.Error())
		}
		count, err := res.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(count)

	case CounterTypeName:
		value, err := strconv.ParseUint(metricValue, 10, 64)
		if err != nil {
			log.Println("DataStorage: GetUpdate: error whith parsing gauge metricValue: " + err.Error())
			return errors.New("DataStorage: GetUpdate: error whith parsing counter metricValue: " + err.Error())
		}
		res, err := storage.DB.ExecContext(storage.ctx, queryTemplate, metricName, metricType, value, 0, value, 0)
		if err != nil {
			log.Println("DataStorage: GetUpdate: error whith upsert to DB: " + err.Error())
			return errors.New("DataStorage: GetUpdate: error whith upsert to DB: " + err.Error())
		}
		count, err := res.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(count)
	default:
		log.Println("DataStorage: GetUpdate: invalid metricType value: " + metricType + ", valid values: " + GaugeTypeName + ", " + CounterTypeName)
		return errors.New(
			"DataStorage: GetUpdate: invalid metricType value, valid values: " + GaugeTypeName + ", " + CounterTypeName)
	}

	return nil
}

func (storage *SQLStorage) GetGaugeValue(metricName string) (float64, error) {
	var queryTemplate string
	switch storage.cfg.DBType {
	case "sqlite3":
		queryTemplate = "SELECT Value FROM statistics5 WHERE ID = ? and MType = ? limit 1;"
	case "postgres":
		queryTemplate = "SELECT Value FROM statistics5 WHERE ID = $1 and MType = $2 limit 1;"
	}

	row := storage.DB.QueryRowContext(storage.ctx, queryTemplate, metricName, "gauge")
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
		queryTemplate = "SELECT Delta FROM statistics5 WHERE ID = ? and MType = ? limit 1;"
	case "postgres":
		queryTemplate = "SELECT Delta FROM statistics5 WHERE ID = $1 and MType = $2 limit 1;"
	}

	row := storage.DB.QueryRowContext(storage.ctx, queryTemplate, metricName, "counter")
	var res uint64
	err := row.Scan(&res)
	if err != nil {
		log.Println(err)
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
	_, err = storage.DB.ExecContext(storage.ctx,
		"CREATE TABLE IF NOT EXISTS statistics5 ( ID text, MType text, Delta bigserial, Value double precision, CONSTRAINT id_pk PRIMARY KEY (ID), CONSTRAINT id_type_uq UNIQUE (ID, MType));")
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
	log.Println(string(jsonDump))
	if err := json.Unmarshal(jsonDump, &metrics); err != nil {
		log.Println(err)
		return err
	}
	log.Println("json parsed")

	log.Println("StartUpdate" + metrics.String())
	metricsHash, _ := metrics.CalcHash(storage.cfg.Key)
	if storage.cfg.Key != "" && metricsHash != metrics.Hash {
		log.Println("Wrong hash, " + metricsHash + " " + metrics.Hash)
		return errors.New("wrong hash")
	}

	return storage.GetUpdate(metrics.MType, metrics.ID, metrics.GetStrValue())

}

func (storage *SQLStorage) GetJSONValue(jsonDump []byte) ([]byte, error) {
	log.Println("Get value request" + string(jsonDump))
	metrics := Metrics{}
	if err := json.Unmarshal(jsonDump, &metrics); err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(metrics.String())

	switch metrics.MType {
	case GaugeTypeName:
		value, err := storage.GetGaugeValue(metrics.ID)
		if err != nil {
			log.Println(err)
			return jsonDump, err
		}
		metrics.Value = value
		metrics.Delta = 0

	case CounterTypeName:
		value, err := storage.GetCounterValue(metrics.ID)
		if err != nil {
			log.Println(err)
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
