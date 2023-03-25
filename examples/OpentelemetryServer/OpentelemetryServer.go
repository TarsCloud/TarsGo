package main

import (
	"context"
	"fmt"
	"log"

	"OpentelemetryServer/otel"
	"OpentelemetryServer/tars-protocol/StressTest"

	"github.com/TarsCloud/TarsGo/contrib/middleware/opentelemetry"
	"github.com/TarsCloud/TarsGo/tars"
)

func main() {
	cfg := tars.GetServerConfig()
	serviceNameKey := fmt.Sprintf("%s.%s", cfg.App, cfg.Server)
	tp := otel.NewTracerProvider(serviceNameKey, "")
	mp := otel.NewMeterProvider(serviceNameKey, "")
	defer func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down metrics provider: %v", err)
		}

		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down traces provider: %v", err)
		}
	}()

	filter := opentelemetry.New()
	tars.UseServerFilterMiddleware(filter.BuildServerFilter())
	tars.UseClientFilterMiddleware(filter.BuildClientFilter())
	imp := new(OpentelemetryImp)                                               //New Imp
	app := new(StressTest.Opentelemetry)                                       //New init the A Tars
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".OpenTelemetryObj") //Register Servant
	tars.Run()
}
