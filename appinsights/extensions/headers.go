package extensions

const (
	/**
	* Request-Context header
	 */
	RequestContextHeader = "request-context"
	/**
	* Source instrumentation header that is added by an application while making http
	* requests and retrieved by the other application when processing incoming requests.
	 */
	RequestContextSourceKey = "appId"
	/**
	* Target instrumentation header that is added to the response and retrieved by the
	* calling application when processing incoming responses.
	 */
	RequestContextTargetKey = "appId"
	/**
	* Request-Id header
	 */
	RequestIdHeader = "request-id"
	/**
	* Legacy Header containing the id of the immidiate caller
	 */
	ParentIdHeader = "x-ms-request-id"
	/**
	* Legacy Header containing the correlation id that kept the same for every telemetry item
	* accross transactions
	 */
	RootIdHeader = "x-ms-request-root-id"
	/**
	* Correlation-Context header
	*
	* Not currently actively used, but the contents should be passed from incoming to outgoing requests
	 */
	CorrelationContextHeader = "correlation-context"
	/**
	* W3C distributed tracing protocol header
	 */
	TraceparentHeader = "traceparent"
	/**
	* W3C distributed tracing protocol state header
	 */
	TraceStateHeader = "tracestate"
)
