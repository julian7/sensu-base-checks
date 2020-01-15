package main

import "github.com/spf13/cobra"

func metricsCmd() *cobra.Command {
	app := &cobra.Command{
		Use:     "metrics",
		Aliases: []string{"metric", "m"},
		Short:   "Metric commands",
	}
	app.AddCommand(filesystemMetricsCmd())

	return app
}
