package extensions

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpRequest struct {
	startTime                time.Time
	userAgent                string
	sourceCorrelationId      string
	parentId                 string
	operationId              string
	requestId                string
	operationName            string
	traceparent              *Traceparent
	tracestate               *Tracestate
	legacyRootId             string
	correlationContextHeader string
}

func ParseHttpRequest(c *gin.Context) *HttpRequest {
	httpRequest := &HttpRequest{}
	httpRequest.startTime = time.Now()
	httpRequest.userAgent = c.Request.UserAgent()
	httpRequest.sourceCorrelationId = GetCorrelationIdManager().GetCorrelationContextTarget(c.Request.Header.Get(RequestContextHeader))
	// tracestateHeader := c.Request.Header.Get(TraceStateHeader)   // w3c header
	// traceparentHeader := c.Request.Header.Get(TraceparentHeader) // w3c header
	requestIdHeader := c.Request.Header.Get(RequestIdHeader)      // default AI header
	legacyParentId := c.Request.Header.Get(ParentIdHeader)        // legacy AI header
	httpRequest.legacyRootId = c.Request.Header.Get(RootIdHeader) // legacy AI header
	httpRequest.correlationContextHeader = c.Request.Header.Get(CorrelationContextHeader)
	httpRequest.operationName = fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)

	if len(requestIdHeader) > 0 {
		httpRequest.parentId = requestIdHeader
		httpRequest.requestId = correlationIdManager.GenerateRequestId(requestIdHeader)
		httpRequest.operationId = correlationIdManager.GetRootId(httpRequest.requestId)
	} else {
		httpRequest.parentId = legacyParentId
		if len(httpRequest.legacyRootId) > 0 {
			httpRequest.requestId = correlationIdManager.GenerateRequestId(httpRequest.legacyRootId)
		} else {
			httpRequest.requestId = correlationIdManager.GenerateRequestId(legacyParentId)
		}
		httpRequest.operationId = correlationIdManager.GetRootId(httpRequest.requestId)
	}
	return httpRequest
}
