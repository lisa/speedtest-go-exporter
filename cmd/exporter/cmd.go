package exporter

import (
	"github.com/spf13/cobra"
)

var (
	timeoutSeconds int64
	listenPort     int
	metricsPath    string
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exporter",
		Short: "Runs the exporter",
		RunE:  Run,
	}
	cmd.Flags().Int64VarP(&timeoutSeconds, "timeoutSec", "t", 120, "Timeout in seconds for each speedtest. Tests exceeding this timeout will be terminated with no results recorded.")
	cmd.Flags().IntVarP(&listenPort, "port", "P", 8080, "Port to listen for metrics on")
	cmd.Flags().StringVarP(&metricsPath, "path", "p", "/metrics", "Metrics path to listen on")

	return cmd
}
