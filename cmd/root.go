package cmd

import (
	goflag "flag"

	"github.com/lisa/speedtest-go-exporter/cmd/exporter"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	klog "k8s.io/klog/v2"
)

var rootCmd = &cobra.Command{
	Use:   "speedtest-exporter [command]",
	Short: "Speedtest Prometheus Exporter",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	klog.InitFlags(nil)
	defer klog.Flush()

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
}

func Execute() {
	rootCmd.AddCommand(exporter.Command())

	_ = rootCmd.Execute()
	//
}
