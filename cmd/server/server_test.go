package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			urlPath:    "/",
			statusCode: 400,
		},
		{
			testName:   "wrong_path_len",
			urlPath:    "/asdd/",
			statusCode: 400,
		},
		{
			testName:   "wrong_path_len",
			urlPath:    "/asd/asdasd//asd",
			statusCode: 400,
		},
		{
			testName:   "wrong_type",
			urlPath:    "/guaaage/fds/235",
			statusCode: 400,
		},
		{
			testName:   "empty_metric_name",
			urlPath:    "/gauge//343.000",
			statusCode: 400,
		},
		{
			testName:   "empty_value",
			urlPath:    "/gauge/asd/",
			statusCode: 400,
		},
		{
			testName:   "correct_guage",
			urlPath:    "/gauge/asd/234.1",
			statusCode: 200,
		},
		{
			testName:   "correct_guage",
			urlPath:    "/gauge/asd/-1234.1",
			statusCode: 200,
		},
		{
			testName:   "correct_guage",
			urlPath:    "/gauge/aFFsd/0.001",
			statusCode: 200,
		},
		{
			testName:   "correct_guage",
			urlPath:    "/gauge/as111d/1111",
			statusCode: 200,
		},
		{
			testName:   "correct_counter",
			urlPath:    "/counter/as111d/1111",
			statusCode: 200,
		},
		{
			testName:   "correct_counter",
			urlPath:    "/counter/a/1111111",
			statusCode: 200,
		},
		{
			testName:   "correct_counter",
			urlPath:    "/counter/as1dD1d/0",
			statusCode: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.urlPath, nil)
			w := httptest.NewRecorder()
			guageChan := make(chan server.GaugeDataUpdate, 1024)
			counterChan := make(chan server.CounterDataUpdate, 1024)
			// h := http.HandlerFunc(UserViewHandler(tt.users))
			h := server.MakeHandler(guageChan, counterChan, "/")
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.statusCode, result.StatusCode)
			assert.Equal(t, "text/plain", result.Header.Get("Content-Type"))

			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}
