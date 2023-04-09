package request

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// HTTPClient describes an http client
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type request struct {
	client          HTTPClient
	headers         map[string]string
	numRetries      int
	retryInterval   time.Duration
	retryPolicy     func(*http.Response, error) bool
	currAttempt     int
	resultContainer interface{}
	reasonContainer interface{}
	body            interface{}
}

type response struct {
	resp     *http.Response
	err      error
	hasError bool
}

func (r *request) SetHeader(key, value string) {
	r.headers[key] = value
}

func (r *request) SetBasicAuth(username, password string) {
	r.headers["Authorization"] = "Basic " + basicAuth(username, password)
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (r *response) Response() *http.Response {
	return r.resp
}

func (r *response) IsError() bool {
	return r.hasError
}

func (r *response) Error() error {
	return r.err
}

func (r *response) StatusCode() int {
	return r.resp.StatusCode
}

func (r *request) Get(url string) (*response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
	resp, err := r.Do(req)
	hasError := err != nil
	respErr := err
	if resp != nil {
		if resp.StatusCode > 399 {
			hasError = true
			e := BadStatusError{code: resp.StatusCode}
			reasonBytes, _ := json.Marshal(r.resultContainer)
			_ = json.Unmarshal(reasonBytes, r.reasonContainer)
			respErr = e
		}
	}
	return &response{
		hasError: hasError,
		resp:     resp,
		err:      respErr, // this is a post request error; i.e. we successfully made the request but the status code is bad etc.
	}, err
}

func (r *request) Post(url string) (*response, error) {
	if r.body == nil {
		r.body = `{}`
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(r.body)
	req, err := http.NewRequest(http.MethodPost, url, b)
	if err != nil {
		return nil, err
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
	resp, err := r.Do(req)
	hasError := err != nil
	respErr := err
	if resp != nil {
		if resp.StatusCode > 399 {
			hasError = true
			e := BadStatusError{code: resp.StatusCode}
			reasonBytes, _ := json.Marshal(r.resultContainer)
			_ = json.Unmarshal(reasonBytes, r.reasonContainer)
			respErr = e
		}
	}
	return &response{
		hasError: hasError,
		resp:     resp,
		err:      respErr, // this is a post request error; i.e. we successfully made the request but the status code is bad etc.
	}, err
}

func R() *request {
	return &request{}
}

func DefaultR(client HTTPClient) *request {
	return &request{
		client: client,
		headers: map[string]string{
			"Content-Type": "application/json",
		},
		numRetries:    2,
		retryInterval: 2 * time.Second,
		retryPolicy: func(resp *http.Response, err error) bool {
			return resp.StatusCode >= 500
		},
	}
}

func (r *request) SetResult(container interface{}) *request {
	r.resultContainer = container
	return r
}

func (r *request) SetReason(container interface{}) *request {
	r.reasonContainer = container
	return r
}

func (r *request) SetBody(body interface{}) *request {
	r.body = body
	return r
}

func (r *request) Do(req *http.Request) (*http.Response, error) {
	r.currAttempt = 0
	for r.currAttempt < (r.numRetries + 1) {
		resp, err := r.client.Do(req)
		shouldRetry := r.retryPolicy(resp, err)
		if shouldRetry {
			r.currAttempt++
			time.Sleep(r.retryInterval)
			// refill request body
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(r.body)
			req.Body = io.NopCloser(b)
			continue
		}
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to read response body: %v", err)
		}
		err = json.Unmarshal(body, r.resultContainer)
		return resp, err
	}
	return nil, errors.New("max retries exhausted")
}

type BadStatusError struct {
	code int
}

func (e BadStatusError) Error() string {
	return fmt.Sprintf("bad status code: %v", e.code)
}

func (e BadStatusError) Code() int {
	return e.code
}
