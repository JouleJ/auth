package main

import (
    "auth/pkg/config"
    "auth/pkg/handlers"
    "auth/pkg/log"
)

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
    "context"
    "fmt"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer(cfg *config.Config) (trace.Tracer, error) {
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.JaegerAddr)))
    if err != nil {
        return nil, err
    }

    tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.TracerName))))

    otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

    return otel.Tracer(cfg.TracerName), nil
}

func main() {
    config, err := config.Load()
    if err != nil {
        fmt.Printf("Failed to load config: %v\n", err)
        return
    }

    logger, err := log.New(config)
    if err != nil {
        fmt.Printf("Failed to initialize logger: %v\n", err)
        return
    }
    defer logger.Sync()

    signals := make(chan os.Signal, 1)
    finishedShutdown := make(chan struct{}, 1)
    signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

    tracer, err := initTracer(config)
    if err != nil {
        logger.Error("Failed to create tracer", zap.String("error", err.Error()))
        return
    }

    handler := http.NewServeMux()
    handler.Handle("/login", handlers.NewLoginHandler(config))
    handler.Handle("/verify", handlers.NewVerifyHandler(config))

    decoratedHandler := http.Handler(handler)
    decoratedHandler = handlers.NewLoggingMiddleware(decoratedHandler, logger)
    decoratedHandler = handlers.NewRecoveryMiddleware(decoratedHandler, logger)
    decoratedHandler = handlers.NewTracerMiddleware(decoratedHandler, tracer)

    server := http.Server{Handler: decoratedHandler, Addr: fmt.Sprintf(":%v", config.Port)}
    logger.Info("Started server", zap.Int("port", config.Port))

    go func() {
        defer close(finishedShutdown)

        <-signals
        err := server.Shutdown(context.Background())
        fmt.Printf("server.Shutdown() returned %v\n", err)
    }()

    err = server.ListenAndServe()
    if err != http.ErrServerClosed {
        fmt.Printf("server.ListenAndServe() returned %v\n", err)
        close(signals)
    }

    <-finishedShutdown
}
