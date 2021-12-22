package server

import (
	"context"
	"fmt"
	"html/template"
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

type GasugeDataResponce struct {
	Value   float64
	Success bool
}

type CounterDataResponce struct {
	Value   int64
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
	Responce chan CollectedData
}

type CollectedData struct {
	GaugeData   map[string]float64
	CounterData map[string]int64
}

func (data *CollectedData) Initiate() {
	data.GaugeData = map[string]float64{}
	data.CounterData = map[string]int64{}
}

func (data *CollectedData) RunReciver(
	gaugeUpdateChan chan GaugeDataUpdate, counterUpdateChan chan CounterDataUpdate,
	gaugeRequestChan chan GaugeDataRequest, counterRequestChan chan CounterDataRequest,
	requestChan chan CollectedDataRequest, end context.Context) {
	for {
		select {
		case update := <-gaugeUpdateChan:
			data.GaugeData[update.Name] = update.Value
		case update := <-counterUpdateChan:
			data.CounterData[update.Name] += update.Value
		case request := <-gaugeRequestChan:
			value, ok := data.GaugeData[request.Name]
			request.Responce <- GasugeDataResponce{value, ok}
		case request := <-counterRequestChan:
			value, ok := data.CounterData[request.Name]
			request.Responce <- CounterDataResponce{value, ok}
		case request := <-requestChan:
			request.Responce <- *data
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

func MakeHandleGaugeValue(gaugeRequestChan chan GaugeDataRequest) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/plain; charset=utf-8")
		metricName := chi.URLParam(req, "metricName")

		if metricName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Empty metric_id"))
			return
		}
		responce := make(chan GasugeDataResponce, 1)
		gaugeRequestChan <- GaugeDataRequest{metricName, responce}

		res := <-responce
		if res.Success {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(strconv.FormatFloat(res.Value, 'f', -1, 64)))
		} else {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("metric not found"))
		}
	}
}

func MakeHandleCounterValue(counterRequestChan chan CounterDataRequest) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/plain; charset=utf-8")
		metricName := chi.URLParam(req, "metricName")

		if metricName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Empty metric_id"))
			return
		}
		responce := make(chan CounterDataResponce, 1)
		counterRequestChan <- CounterDataRequest{metricName, responce}

		res := <-responce
		if res.Success {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(strconv.Itoa(int(res.Value))))
		} else {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("metric not found"))
		}
	}
}

func MakeGetHomeHandler(requestChan chan CollectedDataRequest) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/html; charset=utf-8")

		responce := make(chan CollectedData, 1)
		requestChan <- CollectedDataRequest{responce}
		data := <-responce

		metrics := map[string]string{}

		for key, value := range data.CounterData {
			metrics[key] = strconv.Itoa(int(value))
		}
		for key, value := range data.GaugeData {
			metrics[key] = strconv.FormatFloat(value, 'f', -1, 64)
		}

		t, err := template.ParseFiles("home_page.html")
		if err != nil {
			fmt.Println("Could not parse template:", err)
			return
		}
		err = t.Execute(rw, metrics)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func MakeRouter(
	gaugeUpdateChan chan GaugeDataUpdate, counterUpdateChan chan CounterDataUpdate,
	gaugeRequestChan chan GaugeDataRequest, counterRequestChan chan CounterDataRequest,
	requestChan chan CollectedDataRequest) chi.Router {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", MakeGetHomeHandler(requestChan))

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{metricName}", MakeHandleGaugeValue(gaugeRequestChan))
		r.Get("/counter/{metricName}", MakeHandleCounterValue(counterRequestChan))

		r.Post("/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("content-type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusNotImplemented)
			rw.Write(nil)
		})

		r.Post("/gauge", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("content-type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(nil)
		})
		r.Post("/counter", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("content-type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(nil)
		})
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{metricName}/{metricValue}", MakeHandleGaugeUpdate(gaugeUpdateChan))
		r.Post("/counter/{metricName}/{metricValue}", MakeHandleCounterUpdate(counterUpdateChan))

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

func (dataServer *DataServer) RunHTTPServer(
	guageUpdateChan chan GaugeDataUpdate, counterUpdateChan chan CounterDataUpdate,
	gaugeRequestChan chan GaugeDataRequest, counterRequestChan chan CounterDataRequest,
	requestChan chan CollectedDataRequest, end context.Context) {

	dataServer.Initite()
	r := MakeRouter(guageUpdateChan, counterUpdateChan, gaugeRequestChan, counterRequestChan, requestChan)

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
	gaugeUpdateChan := make(chan GaugeDataUpdate, 1024)
	counterUpdateChan := make(chan CounterDataUpdate, 1024)
	gaugeRequestChan := make(chan GaugeDataRequest, 1024)
	counterRequestChan := make(chan CounterDataRequest, 1024)
	requestChan := make(chan CollectedDataRequest, 1024)

	DataHolderEndCtx, DataHolderCancel := context.WithCancel(end)
	defer DataHolderCancel()
	go dataServer.DataHolder.RunReciver(
		gaugeUpdateChan, counterUpdateChan, gaugeRequestChan, counterRequestChan, requestChan, DataHolderEndCtx)

	httpServerEndCtx, httpServerCancel := context.WithCancel(end)
	defer httpServerCancel()
	dataServer.RunHTTPServer(
		gaugeUpdateChan, counterUpdateChan, gaugeRequestChan, counterRequestChan, requestChan, httpServerEndCtx)
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
