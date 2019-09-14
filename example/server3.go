package main

import (
	"context"
	"fmt"
	"github.com/10gic/gin-opengtracing"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
	"net/http"
	"time"
)

func main() {
	// Configure tracing
	sender, err := jaeger.NewUDPTransport("localhost:6831", 0)
	if err != nil {
		panic(err)
	}
	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	tracer, closer := jaeger.NewTracer(
		"service3",
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

	r.POST("", handler)
	r.Run(":8003")
}

func printHeaders(message string, header http.Header) {
	fmt.Println(message)
	for k, v := range header {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func handler(c *gin.Context) {
	printHeaders("Incoming Headers", c.Request.Header)
	func1(c.Request.Context())
	c.Status(http.StatusOK)
}

func func1(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "func1")
	defer span.Finish()

	fmt.Printf("call func1\n")

	time.Sleep(50 * time.Millisecond)

	func2(ctx)

	time.Sleep(100 * time.Millisecond)
}

func func2(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "func2")
	defer span.Finish()

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("call func2\n")
}
