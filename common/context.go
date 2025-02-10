package common

import (
	"crypto/rand"

	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const TraceID contextKey = "traceID"
const SpanID contextKey = "spanID"

func GenerateTraceID() string {
	traceID := trace.TraceID{}
	_, err := rand.Read(traceID[:])
	if err != nil {
		return ""
	}
	return traceID.String()
}

func GenerateSpanID() string {
	spanID := trace.SpanID{}
	_, err := rand.Read(spanID[:])
	if err != nil {
		return ""
	}
	return spanID.String()
}
