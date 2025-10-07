package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iteais/sdk/pkg/models"
	"github.com/iteais/sdk/pkg/utils"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type InternalFetchConfig struct {
	Method  string
	Url     string
	Body    string
	JWT     string
	TraceId string
}

func FetchUserById(id int64, traceId string, jwt string) (models.User, error) {
	resp := InternalFetch(InternalFetchConfig{
		Method:  "GET",
		Url:     fmt.Sprintf("%s/user/%d", os.Getenv("USER_SERVER"), id),
		TraceId: traceId,
		JWT:     jwt,
	})

	defer func() {
		_ = resp.Body.Close()
	}()

	type RespModel struct {
		Data  models.User `json:"data"`
		Error error       `json:"error"`
	}

	var respModel RespModel
	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &respModel)
		if err != nil {
			return models.User{}, err

		}
		return respModel.Data, err
	}

	return models.User{}, errors.New("Event microservice status code: " + resp.Status)
}

func FetchEventById(id int64, traceId string, jwt string) (models.Event, error) {
	resp := InternalFetch(InternalFetchConfig{
		Method:  "GET",
		Url:     fmt.Sprintf("%s/event/%d", os.Getenv("EVENT_SERVER"), id),
		TraceId: traceId,
		JWT:     jwt,
	})

	defer func() {
		_ = resp.Body.Close()
	}()

	type RespModel struct {
		Data  models.Event `json:"data"`
		Error error        `json:"error"`
	}

	var respModel RespModel
	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &respModel)
		if err != nil {
			return models.Event{}, err

		}
		return respModel.Data, err
	}

	return models.Event{}, errors.New("Event microservice status code: " + resp.Status)
}

func InternalFetch(config InternalFetchConfig) *http.Response {
	client := &http.Client{
		Transport: NewRetryableTransport(nil, 3, 1*time.Second, config.TraceId), // 3 retries, 1s initial delay
		Timeout:   10 * time.Second,                                             // Set a timeout for the request
	}

	req, err := http.NewRequest(config.Method, config.Url, nil)

	if config.Body != "" {
		req, err = http.NewRequest(config.Method, config.Url, strings.NewReader(config.Body))
	}

	if err != nil {
		App.Log.WithField(TraceIdContextKey, config.TraceId).Println("Error creating request:", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(TraceIdHttpHeader, config.TraceId)

	if config.JWT != "" {
		req.Header.Set(utils.AuthHeader, config.JWT)
	}

	resp, err := client.Do(req)
	if err != nil {
		App.Log.WithField(TraceIdContextKey, config.TraceId).Println("Error making "+config.Method+" request:", err)
		return nil
	}

	return resp
}

type RetryableTransport struct {
	transport    http.RoundTripper
	maxRetries   int
	initialDelay time.Duration
	traceId      string
}

// NewRetryableTransport creates a new RetryableTransport.
func NewRetryableTransport(transport http.RoundTripper, maxRetries int, initialDelay time.Duration, traceId string) *RetryableTransport {
	if transport == nil {
		transport = &http.Transport{}
	}
	return &RetryableTransport{
		transport:    transport,
		maxRetries:   maxRetries,
		initialDelay: initialDelay,
		traceId:      traceId,
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (t *RetryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		// Read and store the body for potential retries.
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset for first attempt
	}

	for i := 0; i < t.maxRetries; i++ {
		// Reset the request body for each retry if it exists.
		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err := t.transport.RoundTrip(req)

		if err == nil && !shouldRetryStatusCode(resp.StatusCode) {
			return resp, nil // Success or non-retryable status
		}

		if err != nil {
			App.Log.WithField(TraceIdContextKey, t.traceId).Printf("Attempt %d failed with error: %v\n", i+1, err)
		} else {

			App.Log.WithField(TraceIdContextKey, t.traceId).Printf("Attempt %d failed with status: %d\n", i+1, resp.StatusCode)
			drainBody(resp) // Drain body to allow connection reuse
		}

		if i < t.maxRetries-1 {
			delay := time.Duration(math.Pow(2, float64(i))) * t.initialDelay // Exponential backoff
			App.Log.WithField(TraceIdContextKey, t.traceId).Printf("Retrying in %v...\n", delay)
			time.Sleep(delay)
		}
	}
	return nil, fmt.Errorf("failed after %d retries", t.maxRetries)
}

// shouldRetryStatusCode checks if the HTTP status code indicates a retryable error.
func shouldRetryStatusCode(statusCode int) bool {
	return statusCode == http.StatusRequestTimeout ||
		statusCode == http.StatusTooManyRequests ||
		(statusCode >= 500 && statusCode != http.StatusNotImplemented) // Exclude 501 Not Implemented
}

// drainBody drains and closes the response body.
func drainBody(resp *http.Response) {
	if resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}
