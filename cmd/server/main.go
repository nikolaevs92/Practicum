package main

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

type GuageDataUpdate struct {
	Name  string
	Value float64
}

type CounterDataUpdate struct {
	Name  string
	Value int64
}

type CollectedData struct {
	guageData   map[string]float64
	counterData map[string]int64
}

func (data CollectedData) RunReciver(guageChan chan GuageDataUpdate, counterChan chan CounterDataUpdate, end context.Context) {
	for {
		select {
		case update := <-guageChan:
			data.guageData[update.Name] = update.Value
		case update := <-counterChan:
			data.counterData[update.Name] += update.Value
		case <-end.Done():
			return
		}
	}
}

func MakeHandler(guageChan chan GuageDataUpdate, counterChan chan CounterDataUpdate) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "text/plain")
		body := []byte("data is recieved")
		queryPath := strings.Split(req.URL.Path, "/")[1:]

		if len(queryPath) != 3 {
			w.WriteHeader(http.StatusBadRequest)
			body = []byte("Wrong request")
		} else {
			switch queryPath[0] {
			case gaugeTypeName:
				value, err := strconv.ParseFloat(queryPath[2], 64)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					body = []byte("Error on parsing guage: " + err.Error())
				}
				guageChan <- GuageDataUpdate{queryPath[1], value}
				w.WriteHeader(http.StatusOK)
			case counterTypeName:
				value, err := strconv.ParseInt(queryPath[2], 10, 64)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					body = []byte("Error on parsing counter: " + err.Error())
				}
				counterChan <- CounterDataUpdate{queryPath[1], value}
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusBadRequest)
				body = []byte("Unsupported type: " + queryPath[0])
			}
		}
		w.Write(body)
	}
}

type dataServer struct {
	dataHolder CollectedData
	Server     string
}

func (collector *dataServer) Initite() {
	if collector.Server == "" {
		collector.Server = "127.0.0.1:8080"
	}
}

func (dataServer dataServer) RunHttpServer(guageChan chan GuageDataUpdate, counterChan chan CounterDataUpdate, end context.Context) {
	dataServer.Initite()
	http.Handle("/update/", MakeHandler(guageChan, counterChan))
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

func (server dataServer) Run(end context.Context) {
	guageChan := make(chan GuageDataUpdate, 1024)
	counterChan := make(chan CounterDataUpdate, 1024)

	dataHolderEndCtx, dataHolderCancel := context.WithCancel(end)
	defer dataHolderCancel()
	go server.dataHolder.RunReciver(guageChan, counterChan, dataHolderEndCtx)

	httpServerEndCtx, httpServerCancel := context.WithCancel(end)
	defer httpServerCancel()
	server.RunHttpServer(guageChan, counterChan, httpServerEndCtx)
}

func RunServerDefault() {
	server := new(dataServer)
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

func main() {
	RunServerDefault()
}
