// Package tracing provides requests tracing functional using opentracing specification.
//
// See https://github.com/opentracing/opentracing-go for more information
package otgin

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// GenSpanFromHeaders returns gin.HandlerFunc (middleware) that extracts parent span data from HTTP headers and
// starts a new span referenced to span found in HTTP headers.
//
// It calls ctx.Next() to measure execution time of all following handlers.
func GenSpanFromHeaders(tracer opentracing.Tracer, advancedOpts ...opentracing.StartSpanOption) gin.HandlerFunc {

	return func(ginCtx *gin.Context) {

		operationName := "HTTP " + ginCtx.Request.Method + " " + ginCtx.Request.URL.String()

		var span opentracing.Span

		wireSpanContext, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(ginCtx.Request.Header))
		if err != nil {
			// If no span found in HTTP header, just start new span with no parent
			span = tracer.StartSpan(operationName, advancedOpts...)
		} else {
			// If span found in HTTP header, start new span, and set the new span to be a child of the found span
			opts := append([]opentracing.StartSpanOption{opentracing.ChildOf(wireSpanContext)}, advancedOpts...)
			span = tracer.StartSpan(operationName, opts...)
		}

		// Modify ginCtx.Request, attach the span info.
		// We can get back span info from ginCtx like this:
		// ctx := ginCtx.Request.Context()
		// if span := opentracing.SpanFromContext(ctx); span != nil {
		//     .......
		// }
		ginCtx.Request = ginCtx.Request.WithContext(opentracing.ContextWithSpan(ginCtx, span))

		defer span.Finish()

		ext.HTTPMethod.Set(span, ginCtx.Request.Method)
		ext.HTTPUrl.Set(span, ginCtx.Request.URL.String())

		ginCtx.Next()

		ext.HTTPStatusCode.Set(span, uint16(ginCtx.Writer.Status()))
		ext.Error.Set(span, ginCtx.Writer.Status() >= 400 || len(ginCtx.Errors) > 0)
	}
}
