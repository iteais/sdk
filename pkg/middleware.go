package pkg

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"iter"
	"net"
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

func localIps() iter.Seq[net.IP] {
	return func(yield func(net.IP) bool) {
		addrs, _ := net.InterfaceAddrs()
		for _, address := range addrs {
			// check if the address is a loopback or multicast address
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				if ipnet.IP.IsLoopback() {
					continue
				}

				yield(ipnet.IP)
			}
		}
	}
}

func checkSubnet(c *gin.Context, subnet *net.IPNet) bool {
	clientIP := c.ClientIP()

	// Проверка IP адреса клиента
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	if subnet.Contains(ip) {
		return true
	}
	return false
}

// HmacMiddleware Проверка подписи запроса
func HmacMiddleware(checkHost string, whiteList ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		for localIp := range localIps() {
			if checkSubnet(c, &net.IPNet{IP: localIp, Mask: net.CIDRMask(32, 32)}) {
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

		checkTime := sliceString(Time)
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

func sliceString(s string) string {
	n := 10

	// Convert the string to a slice of runes
	runes := []rune(s)

	// Check if the string has at least n runes
	if len(runes) > n {
		// Slice the rune slice and convert back to a string
		return string(runes[:n])
	} else {
		// If the string is shorter than n, print the whole string
		return s
	}
}
