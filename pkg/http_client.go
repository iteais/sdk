package pkg

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

func NewInternalHttpClient(method string, url string, requestBody string, traceId string) *http.Response {
	client := &http.Client{
		Transport: NewRetryableTransport(nil, 3, 1*time.Second, traceId), // 3 retries, 1s initial delay
		Timeout:   10 * time.Second,                                      // Set a timeout for the request
	}

	req, err := http.NewRequest(method, url, strings.NewReader(requestBody))
	if err != nil {
		App.Log.WithField(TraceIdContextKey, traceId).Println("Error creating request:", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(TraceIdHttpHeader, traceId)

	resp, err := client.Do(req)
	if err != nil {
		App.Log.WithField(TraceIdContextKey, traceId).Println("Error making "+method+" request:", err)
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
