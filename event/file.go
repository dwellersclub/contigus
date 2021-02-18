package event

import (
	"github.com/spf13/cobra"
)

var fileEventCmd = &cobra.Command{
	Use:   "fileEvent",
	Short: "Start an event listener for fileEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
