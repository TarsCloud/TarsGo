package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"OpentelemetryServer/otel"
	"OpentelemetryServer/tars-protocol/StressTest"

	"github.com/TarsCloud/TarsGo/contrib/middleware/opentelemetry"
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/util/current"
)

func main() {
	serviceNameKey := fmt.Sprintf("%s.%s", "StressTest", "OpentelemetryClient")
	tp := otel.NewTracerProvider(serviceNameKey, "")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	filter := opentelemetry.New()
	tars.UseClientFilterMiddleware(filter.BuildClientFilter())
	comm := tars.GetCommunicator()
	obj := fmt.Sprintf("StressTest.OpentelemetryServer.OpenTelemetryObj@tcp -h 127.0.0.1 -p 10028 -t 60000")
	app := new(StressTest.Opentelemetry)
	comm.StringToProxy(obj, app)
	var out, i int32
	i = 11111
	ctx := current.ContextWithClientCurrent(context.Background())
	c := make(map[string]string)
	c["a"] = "b"
	for {
		ret, err := app.AddWithContext(ctx, i, i*2, &out, c)
		if err != nil {
			fmt.Printf("error: %v", err)
			return
		}
		fmt.Println(c)
		fmt.Println(ret, out)

		ret, err = app.SubWithContext(ctx, i, i*2, &out, c)
		if err != nil {
			fmt.Printf("error: %v", err)
			return
		}
		fmt.Println(c)
		fmt.Println(ret, out)
		time.Sleep(time.Second * 5)
	}
}
