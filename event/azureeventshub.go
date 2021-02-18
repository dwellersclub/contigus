package event

import (
	"github.com/spf13/cobra"
)

var azureEventCmd = &cobra.Command{
	Use:   "azureEvent",
	Short: "Start an event listener for azureEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
