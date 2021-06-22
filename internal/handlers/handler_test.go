package handlers

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type fakeDoer struct {
	response  *http.Response
	err       error
	sleepTime time.Duration
}

func (d fakeDoer) Do(r *http.Request) (*http.Response, error) {
	if d.err == nil {
		time.Sleep(d.sleepTime)
	}
	return d.response, d.err
}

type body struct {
}

func (b *body) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (b *body) Close() error {
	return nil
}

var checkURLTests = []struct {
	err                  error
	response             *http.Response
	url                  string
	sleepTime            time.Duration
	expectedHealthyValue float64
}{
	{
		response: &http.Response{
			StatusCode: 200,
			Body:       &body{},
		},
		sleepTime:            100 * time.Millisecond,
		err:                  nil,
		url:                  "example.com",
		expectedHealthyValue: 1,
	},
	{
		response: &http.Response{
			StatusCode: 500,
			Body:       &body{},
		},
		sleepTime:            50 * time.Millisecond,
		err:                  nil,
		url:                  "example.com",
		expectedHealthyValue: 0,
	},
	{
		sleepTime:            50 * time.Millisecond,
		err:                  fmt.Errorf("error"),
		url:                  "example.com",
		expectedHealthyValue: 0,
	},
}

func TestCheckURL(t *testing.T) {
	for _, test := range checkURLTests {
		client = &fakeDoer{
			response:  test.response,
			err:       test.err,
			sleepTime: test.sleepTime,
		}

		err := CheckURL(test.url)
		if test.err != nil {
			assert.EqualError(t, test.err, err.Error())
		} else {
			assert.NoError(t, err)
		}

		healthyValue := testutil.ToFloat64(healthy)
		assert.Equal(t, test.expectedHealthyValue, healthyValue)

		responseValue := testutil.ToFloat64(responseTime)
		assert.GreaterOrEqual(t, responseValue, float64(test.sleepTime.Milliseconds()))
	}
}
