package event

import (
	"github.com/spf13/cobra"
)

var fileEventCmd = &cobra.Command{
	Use:   "fileEvent",
	Short: "Start a event listener for fileEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
