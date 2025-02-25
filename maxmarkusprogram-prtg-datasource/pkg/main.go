package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"github.com/1DeliDolu/PRTG/maxmarkusprogram/prtg/pkg/plugin"
)

func main() {
	opts := datasource.ManageOpts{
		TracingOpts: tracing.Opts{
			CustomAttributes: []attribute.KeyValue{
				attribute.String("plugin.name", "prtg"),
				attribute.String("plugin.type", "datasource"),
			},
		},
	}

	if err := datasource.Manage("grafana-prtg-datasource", plugin.NewDatasource, opts); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
