package extensions

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	correlationIdPrefix string = "cid-v1:"
	w3cEnabled          bool   = false
	requestIdMaxLength  int    = 1024
)

type CorrelationIdManager struct {
	currentRootId uint32
	requestNumber uint32
}

type Traceparent struct {
	ParentId  string
	SpanId    string
	TraceFlag string
	TraceId   string
	Version   string
}

type Tracestate struct {
	Fieldmap []string
}

type Operation struct {
	Name     string
	Id       string
	ParentId string
	*Traceparent
	*Tracestate
}

type CorrelationContextObject struct {
	Operation
	CustomProperties map[string]string
}

type contextKey struct{}

var activeContextKey = contextKey{}

var correlationIdManager = newCorrelationIdManager()

func GetCorrelationIdManager() *CorrelationIdManager {
	return correlationIdManager
}

func newCorrelationIdManager() *CorrelationIdManager {
	manager := &CorrelationIdManager{
		rand.Uint32(),
		1,
	}
	return manager
}

func (manager *CorrelationIdManager) GetCorrelationContextTarget(requestContextHeader string) string {
	if len(requestContextHeader) > 0 {
		keyValues := strings.Split(requestContextHeader, ",")
		for i := 0; i < len(keyValues); i++ {
			keyValue := strings.Split(keyValues[i], "=")
			if len(keyValue) == 2 && keyValue[0] == RequestContextSourceKey {
				return keyValue[1]
			}
		}
	}
	return ""
}

func (manager *CorrelationIdManager) GenerateRootId() string {
	return "|" + hex.EncodeToString(newUUID().Bytes()) + "."
}

func (manager *CorrelationIdManager) GenerateRequestId(parentId string) string {
	if len(parentId) > 0 {
		if parentId[0] != '|' {
			parentId = "|" + parentId
		}

		if parentId[len(parentId)-1] != '.' {
			parentId = parentId + "."
		}
		suffix := strconv.FormatUint(uint64(atomic.AddUint32(&manager.currentRootId, 1)), 16)
		return manager.appendSuffix(parentId, suffix, "_")
	}
	return manager.GenerateRootId()
}

func (manager *CorrelationIdManager) GenerateDependencyId(parentId string) string {
	if len(parentId) == 0 {
		return ""
	}
	return fmt.Sprintf("%s%d.", parentId, atomic.AddUint32(&manager.requestNumber, 1))
}

func (manager *CorrelationIdManager) appendSuffix(parentId, suffix, delimiter string) string {
	if len(parentId)+len(suffix) < requestIdMaxLength {
		return parentId + suffix + delimiter
	}
	trimPosition := requestIdMaxLength - 9
	if len(parentId) > trimPosition {
		for ; trimPosition > 1; trimPosition-- {
			var c = parentId[trimPosition-1]
			if c == '.' || c == '_' {
				break
			}
		}
	}
	if trimPosition <= 1 {
		// parentId is not a valid ID
		return manager.GenerateRootId()
	}
	suffix = strconv.FormatUint(uint64(rand.Uint32()), 16)
	for len(suffix) < 8 {
		suffix = "0" + suffix
	}
	return parentId[0:trimPosition] + suffix + "#"
}

func (manager *CorrelationIdManager) GetRootId(id string) string {
	endIndex := strings.Index(id, ".")
	if endIndex < 0 {
		endIndex = len(id)
	}
	startIndex := 0
	if id[0] == '|' {
		startIndex = 1
	}
	return id[startIndex:endIndex]
}

func GenerateContextObject(operationId, parentId, operationName, correlationContextHeader string, traceparent *Traceparent, tracestate *Tracestate) *CorrelationContextObject {
	if len(parentId) == 0 {
		parentId = operationId
	}
	contextObject := &CorrelationContextObject{
		Operation{
			operationName,
			operationId,
			parentId,
			traceparent,
			tracestate,
		},
		make(map[string]string),
	}
	return contextObject
}

func CorrelationContextWithContext(ctx context.Context, contextObject *CorrelationContextObject) context.Context {
	if ctx == nil {
		return context.WithValue(context.Background(), activeContextKey, contextObject)
	}
	return context.WithValue(ctx, activeContextKey, contextObject)
}

func CorrelationContextFromContext(ctx context.Context) *CorrelationContextObject {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(activeContextKey)
	if ctxObj, ok := val.(*CorrelationContextObject); ok {
		return ctxObj
	}
	return nil
}
