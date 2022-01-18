package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikolaevs92/Practicum/internal/datastorage"
	"github.com/nikolaevs92/Practicum/internal/server"
)

func TestJSONHandler(t *testing.T) {
	type Query struct {
		urlPath    string
		input      datastorage.Metrics
		output     datastorage.Metrics
		statusCode int
	}

	tests := []struct {
		testName string
		queries  []Query
	}{
		{
			testName: "empty_update",
			queries: []Query{
				{
					urlPath:    "/update",
					input:      datastorage.Metrics{},
					output:     datastorage.Metrics{},
					statusCode: 404,
				},
				{
					urlPath: "/update",
					input: datastorage.Metrics{
						ID:    "some",
						MType: datastorage.GaugeTypeName,
						Value: 0.1,
					},
					output: datastorage.Metrics{
						ID:    "some",
						MType: datastorage.GaugeTypeName,
						Value: 0.1,
					},
					statusCode: 200,
				},
				{
					urlPath: "/value",
					input: datastorage.Metrics{
						ID:    "some",
						MType: datastorage.GaugeTypeName,
					},
					output: datastorage.Metrics{
						ID:    "some",
						MType: datastorage.GaugeTypeName,
						Value: 0.1,
					},
					statusCode: 200,
				},
			},
		},
	}

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-cancelChan
		cancel()
	}()

	for _, tt := range tests {

		storage := datastorage.New()
		storage.Init()
		go storage.RunReciver(ctx)

		r := server.MakeRouter(storage)
		ts := httptest.NewServer(r)

		t.Run(tt.testName, func(t *testing.T) {
			for _, tq := range tt.queries {
				resp, body := testJSONRequest(t, ts, "POST", tq.urlPath, tq.input)
				defer resp.Body.Close()

				if !assert.Equal(t, tq.statusCode, resp.StatusCode) {
					fmt.Println(body)
				}
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			}
		})
		ts.Close()
	}
}

func testJSONRequest(t *testing.T, ts *httptest.Server, method string, path string, input datastorage.Metrics) (*http.Response, datastorage.Metrics) {
	res, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader(res))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	metrics := datastorage.Metrics{}
	if err := json.Unmarshal(respBody, &metrics); err != nil {
		panic(err)
	}

	return resp, metrics
}
