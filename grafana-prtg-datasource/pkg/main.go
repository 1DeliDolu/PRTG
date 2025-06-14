package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/1DeliDolu/grafana-plugins/grafana-prtg-datasource/pkg/plugin"
	"go.opentelemetry.io/otel/attribute"
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

	if err := datasource.Manage("grafana-prtg-plugin", plugin.NewDatasource, opts); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
