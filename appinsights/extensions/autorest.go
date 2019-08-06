package extensions

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"

	"github.com/Azure/go-autorest/autorest"
)

type contextKey struct{}
type contextValue struct {
	start   time.Time
	handled bool
}

var requestTimeContextKey = contextKey{}

// SetupAutorestClient setup autorest.Client with request/response inspectors to track dependency.
func SetupAutorestClient(client *autorest.Client) {
	if GetCurrentClient() != nil {
		client.RequestInspector = func(p autorest.Preparer) autorest.Preparer {
			return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
				r, err := p.Prepare(r)
				if err == nil {
					value := &contextValue{
						time.Now(),
						false,
					}
					reqCtx := context.WithValue(r.Context(), requestTimeContextKey, value)
					r = r.WithContext(reqCtx)
				}
				return r, err
			})
		}
		client.ResponseInspector = func(r autorest.Responder) autorest.Responder {
			return autorest.ResponderFunc(func(resp *http.Response) error {
				request := resp.Request
				value := request.Context().Value(requestTimeContextKey)
				if value != nil {
					cv, ok := value.(*contextValue)
					if ok && !cv.handled {
						end := time.Now()
						url := request.URL.String()
						path := request.URL.Path
						if url == path || strings.HasPrefix(url, path) {
							url = resolveScheme(request) + "://" + resolveHost(request) + url
						}
						statusCode := resp.StatusCode

						name := fmt.Sprintf("%s %s", request.Method, path)
						success := statusCode < 400
						target := resolveHost(request)
						dependency := appinsights.NewRemoteDependencyTelemetry(name, "HTTP", target, success)
						aiCtx := CorrelationContextFromContext(ctx)
						if aiCtx != nil {
							dependency.Tags.Operation().SetName(aiCtx.Operation.Name)
							dependency.Tags.Operation().SetId(aiCtx.Operation.Id)
							dependency.Tags.Operation().SetParentId(aiCtx.Operation.ParentId)
						}
						dependency.Id = GetCorrelationIdManager().GenerateDependencyId(dependency.Tags.Operation().GetParentId())
						dependency.ResultCode = strconv.Itoa(statusCode)
						dependency.Data = url
						dependency.Timestamp = end
						dependency.MarkTime(cv.start, end)
						GetCurrentClient().Track(dependency)
						cv.handled = true
					}
				}
				return r.Respond(resp)
			})
		}
	}
}
