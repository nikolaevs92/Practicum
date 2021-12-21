package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	gaugeTypeName   string = "gauge"
	counterTypeName string = "counter"
)

type GaugeDataUpdate struct {
	Name  string
	Value float64
}

type CounterDataUpdate struct {
	Name  string
	Value int64
}

type CollectedData struct {
	GaugeData   map[string]float64
	CounterData map[string]int64
}

func (data *CollectedData) Initiate() {
	data.GaugeData = map[string]float64{}
	data.CounterData = map[string]int64{}
}

func (data *CollectedData) RunReciver(guageChan chan GaugeDataUpdate, counterChan chan CounterDataUpdate, end context.Context) {
	for {
		select {
		case update := <-guageChan:
			data.GaugeData[update.Name] = update.Value
		case update := <-counterChan:
			data.CounterData[update.Name] += update.Value
		case <-end.Done():
			return
		}
	}
}

func MakeHandler(guageChan chan GaugeDataUpdate, counterChan chan CounterDataUpdate, entryPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "text/plain")
		body := []byte("data is recieved")
		queryPath := strings.Split(strings.TrimPrefix(req.URL.Path, entryPath), "/")

		if len(queryPath) != 3 {
			w.WriteHeader(http.StatusBadRequest)
			body = []byte("Wrong request")
		} else if queryPath[1] == "" {
			w.WriteHeader(http.StatusBadRequest)
			body = []byte("Empty metric_id")
		} else {
			switch queryPath[0] {
			case gaugeTypeName:
				value, err := strconv.ParseFloat(queryPath[2], 64)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					body = []byte("Error on parsing guage: " + err.Error())
				} else {
					guageChan <- GaugeDataUpdate{queryPath[1], value}
					w.WriteHeader(http.StatusOK)
				}
			case counterTypeName:
				value, err := strconv.ParseInt(queryPath[2], 10, 64)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					body = []byte("Error on parsing counter: " + err.Error())
				} else {
					counterChan <- CounterDataUpdate{queryPath[1], value}
					w.WriteHeader(http.StatusOK)
				}
			default:
				w.WriteHeader(http.StatusBadRequest)
				body = []byte("Unsupported type: " + queryPath[0])
			}
		}
		w.Write(body)
	}
}

type DataServer struct {
	DataHolder CollectedData
	Server     string
}

func (dataServer *DataServer) Initite() {
	if dataServer.Server == "" {
		dataServer.Server = "127.0.0.1:8080"
	}
	dataServer.DataHolder.Initiate()
}

func (dataServer *DataServer) RunHTTPServer(guageChan chan GaugeDataUpdate, counterChan chan CounterDataUpdate, end context.Context) {
	dataServer.Initite()
	http.Handle("/update/", MakeHandler(guageChan, counterChan, "/update/"))
	server := &http.Server{
		Addr: dataServer.Server,
	}
	go func() {
		<-end.Done()
		fmt.Println("Shutting down the HTTP server...")
		server.Shutdown(end)
	}()
	server.ListenAndServe()
}

func (dataServer *DataServer) Run(end context.Context) {
	guageChan := make(chan GaugeDataUpdate, 1024)
	counterChan := make(chan CounterDataUpdate, 1024)

	DataHolderEndCtx, DataHolderCancel := context.WithCancel(end)
	defer DataHolderCancel()
	go dataServer.DataHolder.RunReciver(guageChan, counterChan, DataHolderEndCtx)

	httpServerEndCtx, httpServerCancel := context.WithCancel(end)
	defer httpServerCancel()
	dataServer.RunHTTPServer(guageChan, counterChan, httpServerEndCtx)
}

func RunServerDefault() {
	server := new(DataServer)
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-cancelChan
		cancel()
	}()

	server.Run(ctx)

	fmt.Println("Program end")
}
