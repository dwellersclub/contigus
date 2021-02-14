package event

import (
	"github.com/spf13/cobra"
)

var routingEventCmd = &cobra.Command{
	Use:   "routingEvent",
	Short: "Start a event listener for routingEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
