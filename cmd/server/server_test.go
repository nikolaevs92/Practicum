package main

import (
	"context"
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

func TestStatHandler(t *testing.T) {
	tests := []struct {
		testName   string
		urlPath    string
		statusCode int
	}{
		{
			testName:   "empty_update",
			urlPath:    "/update/",
			statusCode: 404,
		},
		{
			testName:   "wrong_path_len",
			urlPath:    "/update/asdd/",
			statusCode: 404,
		},
		{
			testName:   "wrong_path_len",
			urlPath:    "/update/asd/asdasd//asd",
			statusCode: 404,
		},
		{
			testName:   "wrong_type",
			urlPath:    "/update/guaaage/fds/235",
			statusCode: 501,
		},
		{
			testName:   "empty_metric_name",
			urlPath:    "/update/gauge//343.000",
			statusCode: 400,
		},
		{
			testName:   "empty_value",
			urlPath:    "/update/gauge/asd",
			statusCode: 400,
		},
		{
			testName:   "correct_guage",
			urlPath:    "/update/gauge/asd/234.1",
			statusCode: 200,
		},
		{
			testName:   "correct_guage",
			urlPath:    "/update/gauge/asd/-1234.1",
			statusCode: 200,
		},
		{
			testName:   "correct_guage",
			urlPath:    "/update/gauge/aFFsd/0.001",
			statusCode: 200,
		},
		{
			testName:   "correct_guage",
			urlPath:    "/update/gauge/as111d/1111",
			statusCode: 200,
		},
		{
			testName:   "correct_counter",
			urlPath:    "/update/counter/as111d/1111",
			statusCode: 200,
		},
		{
			testName:   "correct_counter",
			urlPath:    "/update/counter/a/1111111",
			statusCode: 200,
		},
		{
			testName:   "correct_counter",
			urlPath:    "/update/counter/as1dD1d/0",
			statusCode: 200,
		},
	}

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-cancelChan
		cancel()
	}()

	storage := datastorage.New()
	storage.Init()
	go storage.RunReciver(ctx)

	r := server.MakeRouter(storage)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			resp, body := testRequest(t, ts, "POST", tt.urlPath)
			defer resp.Body.Close()

			if !assert.Equal(t, tt.statusCode, resp.StatusCode) {
				fmt.Println(body)
			}
			assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method string, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
