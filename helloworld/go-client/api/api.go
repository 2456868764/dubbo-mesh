package api

import (
	"context"
	"dubbo-mesh/helloworld/go-client/utils"
	greet "dubbo-mesh/helloworld/proto"
	"dubbo.apache.org/dubbo-go/v3/client"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var defaultTraceHeaders = []string{
	// All applications should propagate x-request-id. This header is
	// included in access log statements and is used for consistent trace
	// sampling and log sampling decisions in Istio.
	"X-Request-Id",

	// Lightstep tracing header. Propagate this if you use lightstep tracing
	// in Istio (see
	// https://istio.io/latest/docs/tasks/observability/distributed-tracing/lightstep/)
	// Note: this should probably be changed to use B3 or W3C TRACE_CONTEXT.
	// Lightstep recommends using B3 or TRACE_CONTEXT and most application
	// libraries from lightstep do not support x-ot-span-context.
	"X-Ot-Span-Context",

	// Datadog tracing header. Propagate these headers if you use Datadog
	// tracing.
	"x-datadog-trace-id",
	"x-datadog-parent-id",
	"x-datadog-sampling-priority",

	// b3 trace headers. Compatible with Zipkin, OpenCensusAgent, and
	// Stackdriver Istio configurations. Commented out since they are
	// propagated by the OpenTracing tracer above.
	"X-B3-TraceId", "X-B3-SpanId", "X-B3-ParentSpanId", "X-B3-Sampled", "X-B3-Flags",

	// Jager
	"uber-trace-id",

	// Grpc binary trace context. Compatible with OpenCensusAgent nad
	// Stackdriver Istio configurations.
	"grpc-trace-bin",

	// W3C Trace Context. Compatible with OpenCensusAgent and Stackdriver Istio
	// configurations.
	"traceparent",
	"tracestate",

	// Cloud trace context. Compatible with OpenCensusAgent and Stackdriver Istio
	// configurations.
	"x-cloud-trace-context",

	// SkyWalking trace headers.
	"sw8",

	// Context and session specific headers
	"cookie", "jwt", "Authorization",

	// Application-specific headers to forward.
	"end-user",
	"user-agent",

	// httpbin headers
	"X-Httpbin-Trace-Host",
	"X-Httpbin-Trace-Service",
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

func Greet(c *gin.Context) {
	url := utils.GetDUBBOServerUrl()
	cli, err := client.NewClient(
		client.WithClientURL(url),
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}

	svc, err := greet.NewGreetService(cli)
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}
	ctx := context.Background()
	resp, err := svc.Greet(ctx, &greet.GreetRequest{Name: "hello world"})
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}
	attachments := make(map[string]interface{})
	if ctx.Value(constant.AttachmentKey) != nil {
		attachments = ctx.Value(constant.AttachmentKey).(map[string]interface{})
	}
	response := make(map[string]any)
	response["attachments"] = attachments
	response["body"] = resp
	request := NewResponseFromContext(c)
	responseAny := make(map[string]any, 0)
	responseAny["request"] = request
	responseAny["response"] = response
	c.JSON(http.StatusOK, responseAny)
}

func Headers(c *gin.Context) {
	headers := c.Request.Header
	response := make(map[string]string, len(headers))
	for hk, hv := range headers {
		response[hk] = strings.Join(hv, ",")
	}
	c.JSON(http.StatusOK, response)
}
