package exporter

import (
	"time"

	prommetrics "github.com/lisa/speedtest-go-exporter/pkg/exporter"

	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) error {
	timeoutDuration := time.Second * time.Duration(timeoutSeconds)

	metrics := prommetrics.MetricsServer{
		MetricsTimeout: timeoutDuration,
		MetricsPath:    metricsPath,
		ListenPort:     listenPort,
	}
	metrics.Listen()

	return nil
}
