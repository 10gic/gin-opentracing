package main

import (
	"context"
	"fmt"
	"github.com/10gic/opengtracing-gin"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
	"io"
	"net/http"
)

func main() {
	// Configure tracing
	sender, err := jaeger.NewUDPTransport("localhost:6831", 0)
	if err != nil {
		panic(err)
	}
	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	tracer, closer := jaeger.NewTracer(
		"service1",
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(sender),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.ZipkinSharedRPCSpan(true),
	)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	// Set up routes
	r := gin.Default()
	r.Use(otgin.GenSpanFromHeaders(tracer))

	r.POST("", handler)
	r.Run(":8001")
}

func printHeaders(message string, header http.Header) {
	fmt.Println(message)
	for k, v := range header {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func handler(c *gin.Context) {
	printHeaders("Incoming Headers", c.Request.Header)

	resp, err := doHttpRequestInjectSpan(c.Request.Context(), "POST", "http://localhost:8003", nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected response from service3")
	}
	c.Status(http.StatusOK)
}

func doHttpRequestInjectSpan(ctx context.Context, method string, url string, body io.Reader) (*http.Response, error) {
	httpClient := &http.Client{}
	httpReq, _ := http.NewRequest(method, url, body)

	if span := opentracing.SpanFromContext(ctx); span != nil {
		if err := opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(httpReq.Header)); err != nil {
			fmt.Printf("inject span error: %s", err)
		}
	} else {
		fmt.Println("no span found")
	}

	return httpClient.Do(httpReq)
}
