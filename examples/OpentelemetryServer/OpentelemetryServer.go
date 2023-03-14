package main

import (
	"context"
	"fmt"
	"log"

	"OpentelemetryServer/tars-protocol/StressTest"
	"OpentelemetryServer/tracer"

	"github.com/TarsCloud/TarsGo/contrib/middleware/opentelemetry"
	"github.com/TarsCloud/TarsGo/tars"
)

func main() {
	cfg := tars.GetServerConfig()
	serviceNameKey := fmt.Sprintf("%s.%s", cfg.App, cfg.Server)
	tp := tracer.NewTracerProvider(serviceNameKey, "")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	filter := opentelemetry.New()
	tars.UseServerFilterMiddleware(filter.BuildServerFilter())
	tars.UseClientFilterMiddleware(filter.BuildClientFilter())
	imp := new(OpentelemetryImp)                                               //New Imp
	app := new(StressTest.ContextTest)                                         //New init the A Tars
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".OpenTelemetryObj") //Register Servant
	tars.Run()
}
