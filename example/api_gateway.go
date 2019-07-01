package main

import (
	"context"
	"fmt"
	"github.com/10gic/gin-opengtracing"
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
		"api_gateway",
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
	r.Use(ginopentracing.GenSpanFromHeaders(tracer))

	r.POST("/service1", service1handler)
	r.POST("/service2", 	service2handler)
	r.Run(":8000")
}

func printHeaders(message string, header http.Header) {
    fmt.Println(message)
	for k, v := range header {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func service1handler(c *gin.Context) {
	printHeaders("Incoming Headers", c.Request.Header)

	resp, err := doHttpRequestInjectSpan(c.Request.Context(), "POST", "http://localhost:8001", nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func service2handler(c *gin.Context) {
	printHeaders("Incoming Headers", c.Request.Header)

	resp, err := doHttpRequestInjectSpan(c.Request.Context(), "POST", "http://localhost:8002", nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
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
