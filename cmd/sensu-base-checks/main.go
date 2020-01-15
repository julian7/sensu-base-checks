package main

import (
	"github.com/julian7/sensulib"
	"github.com/spf13/cobra"
)

var version = "SNAPSHOT"

func rootCmd() *cobra.Command {
	app := &cobra.Command{
		Use:   "sensu-base-checks",
		Short: "Base check plugin for sensu",
		Long: `Basic system-level checks (mainly) for sensu-go, but it is usable by
any nagios-style monitoring solutions too.`,
		Version: version,
	}
	app.AddCommand(filesystemCheckCmd(), httpCmd())

	return app
}

func main() {
	defer sensulib.Recover()

	if err := rootCmd().Execute(); err != nil {
		sensulib.HandleError(err)
	}
}
