package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nikolaevs92/Practicum/internal/datastorage"
)

type DataBase interface {
	GetUpdate(string, string, string) error
	GetGaugeValue(string) (float64, error)
	GetCounterValue(string) (uint64, error)
	GetStats() (map[string]float64, map[string]uint64, error)
	Init()
	RunReciver(context.Context)
	GetJSONUpdate([]byte) error
	GetJSONValue([]byte) ([]byte, error)
}

func MakeHandlerJSONUpdate(data DataBase) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "application/json; charset=utf-8")
		body, err := io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		err = data.GetJSONUpdate(body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
		}
		rw.Write(body)
	}
}

func MakeHandlerJSONValue(data DataBase) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "application/json; charset=utf-8")
		body, err := io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		respBody, err := data.GetJSONValue(body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
		}
		rw.Write(respBody)
	}
}

func MakeHandlerUpdate(data DataBase) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/plain; charset=utf-8")
		metricType := chi.URLParam(req, "metricType")
		metricName := chi.URLParam(req, "metricName")
		metricValue := chi.URLParam(req, "metricValue")

		if metricType != datastorage.GaugeTypeName && metricType != datastorage.CounterTypeName {
			rw.WriteHeader(http.StatusNotImplemented)
			rw.Write([]byte("Wrong metric type"))
			return
		}

		if metricName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Empty metric_id"))
			return
		}
		body := []byte("data is recieved")

		err := data.GetUpdate(metricType, metricName, metricValue)

		if err == nil {
			rw.WriteHeader(http.StatusOK)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
		}
		rw.Write(body)
	}
}

func MakeHandleGaugeValue(data DataBase) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/plain; charset=utf-8")
		metricName := chi.URLParam(req, "metricName")

		if metricName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Empty metric_id"))
			return
		}

		value, err := data.GetGaugeValue(metricName)

		if err == nil {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
		} else {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("metric not found"))
		}
	}
}

func MakeHandleCounterValue(data DataBase) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/plain; charset=utf-8")
		metricName := chi.URLParam(req, "metricName")

		if metricName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Empty metric_id"))
			return
		}

		value, err := data.GetCounterValue(metricName)

		if err == nil {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(strconv.FormatUint(value, 10)))
		} else {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("metric not found"))
		}
	}
}

func MakeGetHomeHandler(dataStorage DataBase) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("content-type", "text/html; charset=utf-8")

		gaugeData, counterData, _ := dataStorage.GetStats()

		metrics := map[string]string{}

		for key, value := range counterData {
			metrics[key] = strconv.Itoa(int(value))
		}
		for key, value := range gaugeData {
			metrics[key] = strconv.FormatFloat(value, 'f', -1, 64)
		}

		t, err := template.ParseFiles("../../template/home_page.html")
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

func MakeRouter(dataStorage DataBase) chi.Router {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", MakeGetHomeHandler(dataStorage))

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{metricName}", MakeHandleGaugeValue(dataStorage))
		r.Get("/counter/{metricName}", MakeHandleCounterValue(dataStorage))
		r.Post("/", MakeHandlerJSONValue(dataStorage))

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
		r.Post("/{metricType}/{metricName}/{metricValue}", MakeHandlerUpdate(dataStorage))

		r.Post("/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("content-type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(nil)
		})
		r.Post("/", MakeHandlerJSONUpdate(dataStorage))
	})

	return r
}

type Config struct {
	Server string
}

type DataServer struct {
	DataHolder DataBase
	Config
}

func (dataServer *DataServer) Init() {
	dataServer.DataHolder.Init()
}

func New(config Config) *DataServer {
	server := new(DataServer)
	server.Server = config.Server
	server.DataHolder = datastorage.New()
	server.Init()
	return server
}

func (dataServer *DataServer) RunHTTPServer(end context.Context) {

	dataServer.Init()
	r := MakeRouter(dataServer.DataHolder)

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

	DataHolderEndCtx, DataHolderCancel := context.WithCancel(end)
	defer DataHolderCancel()
	go dataServer.DataHolder.RunReciver(DataHolderEndCtx)

	httpServerEndCtx, httpServerCancel := context.WithCancel(end)
	defer httpServerCancel()
	dataServer.RunHTTPServer(httpServerEndCtx)
}
