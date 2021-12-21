package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func MakeHandleGaugeUpdate(guageChan chan GaugeDataUpdate) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/plain; charset=utf-8")
		metricName := chi.URLParam(req, "metricName")
		metricValue := chi.URLParam(req, "metricValue")

		if metricName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Empty metric_id"))
			return
		}
		body := []byte("data is recieved")
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Error on parsing guage: " + err.Error()))
			return
		}
		guageChan <- GaugeDataUpdate{metricName, value}

		rw.WriteHeader(http.StatusOK)
		rw.Write(body)
	}
}

func MakeHandleCounterUpdate(counterChan chan CounterDataUpdate) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/plain; charset=utf-8")
		metricName := chi.URLParam(req, "metricName")
		metricValue := chi.URLParam(req, "metricValue")

		if metricName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Empty metric_id"))
			return
		}

		if metricName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Empty metric_id"))
			return
		}
		body := []byte("data is recieved")
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Error on parsing counter: " + err.Error()))
			return
		}
		counterChan <- CounterDataUpdate{metricName, value}

		rw.WriteHeader(http.StatusOK)
		rw.Write(body)
	}
}

func MakeRouter(guageChan chan GaugeDataUpdate, counterChan chan CounterDataUpdate) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
	// 	rw.Header().Set("content-type", "text/plain")
	// 	rw.WriteHeader(http.StatusNotFound)
	// 	rw.Write(nil)
	// })
	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{metricName}/{metricValue}", MakeHandleGaugeUpdate(guageChan))
		r.Post("/counter/{metricName}/{metricValue}", MakeHandleCounterUpdate(counterChan))

		r.Post("/gauge/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("content-type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(nil)
		})
		r.Post("/counter/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("content-type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(nil)
		})

		r.Post("/{metricType}/{metricName}/{metricValue}", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("content-type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusNotImplemented)
			rw.Write(nil)
		})
	})

	return r
}

type DataServer struct {
	DataHolder CollectedData
	Server     string
	Port       string
}

func (dataServer *DataServer) Initite() {
	if dataServer.Server == "" {
		dataServer.Server = "127.0.0.1:8080"
	}
	dataServer.DataHolder.Initiate()
}

func (dataServer *DataServer) RunHTTPServer(guageChan chan GaugeDataUpdate, counterChan chan CounterDataUpdate, end context.Context) {
	dataServer.Initite()
	r := MakeRouter(guageChan, counterChan)

	server := &http.Server{
		Addr:    dataServer.Server,
		Handler: r,
	}
	go func() {
		<-end.Done()
		fmt.Println("Shutting down the HTTP server...")
		server.Shutdown(end)
	}()
	log.Fatal(server.ListenAndServe())
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
