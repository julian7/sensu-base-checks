package command

import "github.com/spf13/cobra"

// Runnable is an interface command.New can take to generate a new cobra command
type Runnable interface {
	Run(*cobra.Command, []string) error
}

// New returns a now *cobra.Command suitable for self-contained configuration
func New(runnable Runnable, use, short, long string) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Short:         short,
		Long:          long,
		RunE:          runnable.Run,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}
