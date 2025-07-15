package pkg

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/iteais/sdk/pkg/utils"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	TraceIdContextKey = "traceId"
	TraceIdHttpHeader = "X-Trace-Id"
)

func CorsMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func JsonMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tid := c.GetHeader(TraceIdHttpHeader)

		if tid == "" {
			tid = uuid.New().String()
		}

		c.Set(TraceIdContextKey, tid)

		c.Next()
	}
}

type hmacResponse struct {
	Cnt   int
	Data  ApiAccount
	Error string
}

// HmacMiddleware Проверка подписи запроса
func HmacMiddleware(checkHost string, whiteList ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		clientIP := c.ClientIP()

		if clientIP == "::1" || clientIP == "127.0.0.1" {
			c.Next()
			return
		}

		for localIp := range utils.LocalIps() {
			if utils.CheckIpsInSameSubnet(c.ClientIP(), localIp.String()) {
				c.Next()
				return
			}
		}

		wl := append([]string{MetricsEndpoint, HealthEndpoint, ReadyEndpoint}, whiteList...)

		for _, s := range wl {
			if ok, _ := regexp.MatchString(s, c.Request.URL.Path); ok {
				c.Next()
				return
			}
		}

		key := c.Request.Header.Get("Api-Key")
		Sign := c.Request.Header.Get("Api-Sign")
		Time := c.Request.Header.Get("Api-Time")

		if key == "" || Sign == "" || Time == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Api-Key or Api-Sign or Api-Time is empty"})
			return
		}

		resp := NewInternalHttpClient("GET", checkHost+"/api/byKey/"+key, "", c.GetString(TraceIdContextKey))

		if resp == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Cant call api service"})
			return
		}

		if resp.StatusCode != http.StatusOK {
			// TODO: may be 403 or 401
			c.AbortWithStatusJSON(resp.StatusCode, gin.H{"message": "Api service return status code: " + strconv.Itoa(resp.StatusCode)})
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// TODO: may be 400
			c.AbortWithStatusJSON(resp.StatusCode, gin.H{"message": "Api service read response body error: " + err.Error()})
			return
		}

		var account hmacResponse
		err = json.Unmarshal(body, &account)
		if err != nil {
			// TODO: may be 400
			c.AbortWithStatusJSON(resp.StatusCode, gin.H{"message": "Api service unmarshal response body error: " + err.Error()})
			return
		}

		now := time.Now()

		checkTime := utils.SliceString(Time)
		unixTimestampSeconds, err := strconv.ParseInt(checkTime, 10, 64)
		requestTime := time.Unix(unixTimestampSeconds, 0)

		past := now.Add(-2 * time.Minute).Unix()
		requestTimestamp := requestTime.Unix()
		future := now.Add(2 * time.Minute).Unix()

		if past > requestTimestamp || requestTimestamp > future {
			c.AbortWithStatusJSON(419, gin.H{"message": "Request has incorrect signature"})
			return
		}

		if account.Data.CanHandleWithHash(Sign, Time) == false {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Api service not approve request"})
			return
		}

		c.Next()
	}
}
