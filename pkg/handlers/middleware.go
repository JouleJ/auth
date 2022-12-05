package handlers

import (
    "bytes"
    "fmt"
    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
    "io/ioutil"
    "net/http"
)

type LoggingMiddleware struct {
    handler http.Handler
    logger *zap.Logger
}

func NewLoggingMiddleware(handler http.Handler, logger *zap.Logger) *LoggingMiddleware {
    return &LoggingMiddleware {
        handler: handler,
        logger: logger,
    }
}

func (this *LoggingMiddleware) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
    body, err := ioutil.ReadAll(request.Body)
    if err == nil {
        request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
    }
    this.logger.Debug(
        "Serving Request",
        zap.String("body", string(body)),
        zap.Any("header", request.Header))
    this.handler.ServeHTTP(responseWriter, request)
}

type RecoveryMiddleware struct {
    handler http.Handler
    logger *zap.Logger
}

func NewRecoveryMiddleware(handler http.Handler, logger *zap.Logger) *RecoveryMiddleware {
    return &RecoveryMiddleware {
        handler: handler,
        logger: logger,
    }
}

func (this *RecoveryMiddleware) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
    defer func() {
        if err := recover(); err != nil {
            this.logger.Error(
                "Recovering from panic in http request",
                zap.String("error", fmt.Sprintf("%v", err)))
        }
    }()
    this.handler.ServeHTTP(responseWriter, request)
}

type TracerMiddleware struct {
    handler http.Handler
    tracer trace.Tracer
}

func NewTracerMiddleware(handler http.Handler, tracer trace.Tracer) *TracerMiddleware {
    return &TracerMiddleware {
        handler: handler,
        tracer: tracer,
    }
}

func (this *TracerMiddleware) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
    newCtx, span := this.tracer.Start(request.Context(), "Serve HTTP")
    defer span.End()
    this.handler.ServeHTTP(responseWriter, request.WithContext(newCtx))
}
