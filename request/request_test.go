package request_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/kickback-app/common/mocks"
	"github.com/kickback-app/common/request"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestSuccessfulGet(t *testing.T) {
	body := `{"hello": "test", "world": 3}`
	httpClient := mocks.NewRequestMock(&mocks.NewRequestMockOpts{
		Responses: []*http.Response{
			{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(body))),
			},
		},
		Validators: []mocks.RequestValidator{
			{
				ExpectedMethod:  "GET",
				ExpectedURLPath: "mockURL/v1/path",
			},
		},
	})
	var res map[string]interface{}
	var reason interface{}
	r := request.DefaultR(httpClient).SetResult(&res).SetReason(&reason)
	resp, err := r.Get("mockURL/v1/path")
	require.Nil(t, err, "no error on get expected")
	require.False(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 1, httpClient.CallCount(), "call count")
	require.Equal(t, map[string]interface{}{
		"hello": "test",
		"world": float64(3),
	}, res, "expected output to be equal")
}

func TestFailedGet(t *testing.T) {
	httpClient := mocks.NewRequestMock(&mocks.NewRequestMockOpts{
		Responses: []*http.Response{
			{
				StatusCode: 400,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"error": "myError"}`))),
			},
		},
		Validators: []mocks.RequestValidator{
			{
				ExpectedMethod:  "GET",
				ExpectedURLPath: "mockURL/v1/path",
			},
		},
	})
	var res map[string]interface{}
	var reason map[string]interface{}
	r := request.DefaultR(httpClient).SetResult(&res).SetReason(&reason)
	resp, err := r.Get("mockURL/v1/path")
	require.Nil(t, err, "should be no request err, only IsError")
	require.IsType(t, request.BadStatusError{}, resp.Error(), "expected error type")
	require.True(t, resp.IsError(), "expected isError to be true")
	require.Equal(t, 1, httpClient.CallCount(), "call count")
	require.Equal(t, map[string]interface{}{
		"error": "myError",
	}, reason, "expected output to be equal")
}

func TestRetryGet(t *testing.T) {
	httpClient := mocks.NewRequestMock(&mocks.NewRequestMockOpts{
		Responses: []*http.Response{
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
			{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"hello": "test", "world": 3}`))),
			},
		},
		Validators: []mocks.RequestValidator{
			{
				ExpectedMethod:  "GET",
				ExpectedURLPath: "mockURL/v1/path",
			},
			{
				ExpectedMethod:  "GET",
				ExpectedURLPath: "mockURL/v1/path",
			},
		},
	})
	var res map[string]interface{}
	var reason interface{}
	r := request.DefaultR(httpClient).SetResult(&res).SetReason(&reason)
	resp, err := r.Get("mockURL/v1/path")
	require.Nil(t, err, "no error on get expected")
	require.False(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 2, httpClient.CallCount(), "call count")
	require.Equal(t, map[string]interface{}{
		"hello": "test",
		"world": float64(3),
	}, res, "expected output to be equal")
}

func TestMaxRetriesExhausted(t *testing.T) {
	httpClient := mocks.NewRequestMock(&mocks.NewRequestMockOpts{
		Responses: []*http.Response{
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
			},
		},
		Validators: []mocks.RequestValidator{
			{
				ExpectedMethod:  "GET",
				ExpectedURLPath: "mockURL/v1/path",
			},
			{
				ExpectedMethod:  "GET",
				ExpectedURLPath: "mockURL/v1/path",
			},
			{
				ExpectedMethod:  "GET",
				ExpectedURLPath: "mockURL/v1/path",
			},
		},
	})
	var res map[string]interface{}
	var reason interface{}
	r := request.DefaultR(httpClient).SetResult(&res).SetReason(&reason)
	_, err := r.Get("mockURL/v1/path")
	require.NotNil(t, err, "expected err")
	require.Equal(t, 3, httpClient.CallCount(), "call count")
	require.Equal(t, "max retries exhausted", err.Error(), "err check")
}

func TestSuccessfulPost(t *testing.T) {
	body := "test"
	returnBody := `{"yoohoo": true}`
	httpClient := mocks.NewRequestMock(&mocks.NewRequestMockOpts{
		Responses: []*http.Response{
			{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(returnBody))),
			},
		},
		Validators: []mocks.RequestValidator{
			{
				ExpectedMethod:     "POST",
				ExpectedURLPath:    "mockURL/v1/path",
				ExpectedCalledWith: body,
				Fuzzy:              true,
			},
		},
	})
	var res interface{}
	var reason interface{}
	r := request.DefaultR(httpClient).SetResult(&res).SetReason(&reason).SetBody(body)
	resp, err := r.Post("mockURL/v1/path")
	require.Nil(t, err, "no error on post expected")
	require.False(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 1, httpClient.CallCount(), "call count")
	require.Equal(t, map[string]interface{}{
		"yoohoo": true,
	}, res, "expected output to be equal")
}

func TestSuccessfulPostRetry(t *testing.T) {
	body := "test"
	returnBody := `{"yoohoo": true}`
	httpClient := mocks.NewRequestMock(&mocks.NewRequestMockOpts{
		Responses: []*http.Response{
			{
				StatusCode: 500,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			},
			{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(returnBody))),
			},
		},
		Validators: []mocks.RequestValidator{
			{
				ExpectedMethod:     "POST",
				ExpectedURLPath:    "mockURL/v1/path",
				ExpectedCalledWith: body,
				Fuzzy:              true,
			},
		},
	})
	var res interface{}
	var reason interface{}
	r := request.DefaultR(httpClient).SetResult(&res).SetReason(&reason).SetBody(body)
	resp, err := r.Post("mockURL/v1/path")
	require.Nil(t, err, "no error on post expected")
	require.False(t, resp.IsError(), "expected isError to be false")
	require.Equal(t, 2, httpClient.CallCount(), "call count")
	require.Equal(t, map[string]interface{}{
		"yoohoo": true,
	}, res, "expected output to be equal")
}
