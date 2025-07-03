package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

func NewHttpClientFromContext(method string, url string, requestBody string, c *gin.Context) *http.Response {
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a timeout for the request
	}

	req, err := http.NewRequest(method, url, strings.NewReader(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(traceIdHttpHeader, c.GetString(traceIdContextKey))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making "+method+" request:", err)
		return nil
	}

	return resp
}
