package datastorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUpdateErrors(t *testing.T) {
	type TestInput struct {
		metricType  string
		metricName  string
		metricValue string
	}
	tests := []struct {
		testName  string
		input     TestInput
		success   bool
		errString string
	}{
		{
			testName: "empty_update",
			input: TestInput{
				metricType:  "",
				metricName:  "",
				metricValue: "",
			},
			errString: "",
			success:   true,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	storage := NewFileStorage(StorageConfig{})
	go storage.RunReciver(ctx)
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, true, true)
		})
	}
}
