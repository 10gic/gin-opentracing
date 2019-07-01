package main

import (
	"fmt"
	"github.com/10gic/gin-opengtracing"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
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
		"service2",
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
	r.Run(":8002")
}

func printHeaders(message string, header http.Header) {
	fmt.Println(message)
	for k, v := range header {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func handler(c *gin.Context) {
	printHeaders("Incoming Headers", c.Request.Header)
	c.Status(http.StatusOK)
}
