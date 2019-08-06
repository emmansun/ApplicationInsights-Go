package extensions

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"

	"github.com/gin-gonic/gin"
)

const CorrelationContextObjectKey = "correlationContextObject"

var client appinsights.TelemetryClient

func SetupGin(engine *gin.Engine, config *InsightsConfig) {
	engine.Use(InsightsWithConfig(config))
}

func GetCurrentClient() appinsights.TelemetryClient {
	return client
}

func GetCorrelationContext(c *gin.Context) context.Context {
	if c.Keys != nil {
		ctx, ok := c.Keys[CorrelationContextObjectKey].(context.Context)
		if ok {
			return ctx
		}
	}
	return context.Background()
}

func GetCorrelationContextObject(c *gin.Context) *CorrelationContextObject {
	ctx := GetCorrelationContext(c)
	return CorrelationContextFromContext(ctx)
}

func resolveScheme(r *http.Request) string {
	switch {
	case r.URL.Scheme == "https":
		return "https"
	case r.TLS != nil:
		return "https"
	case strings.HasPrefix(r.Proto, "HTTPS"):
		return "https"
	default:
		return "http"
	}
}

func resolveHost(r *http.Request) (host string) {
	switch {
	case r.Header.Get("X-Host") != "":
		return r.Header.Get("X-Host")
	case r.Host != "":
		return r.Host
	case r.URL.Host != "":
		return r.URL.Host
	default:
		return "localhost"
	}
}

func InsightsWithConfig(config *InsightsConfig) gin.HandlerFunc {
	client = createTelemetryClient(config)
	env := os.Getenv("GOENV")
	if len(env) == 0 {
		env = "DEV"
	}
	client.Context().CommonProperties["environment"] = env

	return func(c *gin.Context) {

		httpRequest := ParseHttpRequest(c)
		correlationContextObject := GenerateContextObject(httpRequest.operationId, httpRequest.requestId, httpRequest.operationName, httpRequest.correlationContextHeader, nil, nil)
		if c.Keys == nil {
			c.Keys = make(map[string]interface{})
		}
		ctx := CorrelationContextWithContext(context.Background(), correlationContextObject)
		c.Keys[CorrelationContextObjectKey] = ctx
		url := c.Request.URL.String()

		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		if path == "/readiness" || path == "/liveness" {
			return
		}

		if url == path || strings.HasPrefix(url, path) {
			url = resolveScheme(c.Request) + "://" + resolveHost(c.Request) + url
		}

		end := time.Now()
		duration := end.Sub(start)

		clientIP := c.ClientIP()
		if len(clientIP) > 2 && clientIP[0:2] == "::" {
			clientIP = "127.0.0.1"
		}
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()
		//errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		bodySize := c.Writer.Size()

		telemetry := appinsights.NewRequestTelemetry(method, url, duration, strconv.Itoa(statusCode))
		telemetry.Id = httpRequest.requestId
		telemetry.Name = fmt.Sprintf("%s %s", method, path)
		telemetry.Properties["user-agent"] = userAgent
		if len(httpRequest.sourceCorrelationId) > 0 {
			telemetry.Source = httpRequest.sourceCorrelationId
		} else {
			telemetry.Source = clientIP
		}
		telemetry.Tags.Location().SetIp(clientIP)
		telemetry.Tags.Operation().SetName(httpRequest.operationName)
		telemetry.Tags.Operation().SetId(httpRequest.operationId)
		telemetry.Tags.Operation().SetParentId(httpRequest.parentId)
		telemetry.Timestamp = end
		telemetry.MarkTime(start, end)

		telemetry.Success = statusCode >= 200 && statusCode < 400

		telemetry.Measurements["body-size"] = float64(bodySize)

		client.Track(telemetry)
	}
}
