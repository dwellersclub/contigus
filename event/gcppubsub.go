package event

import (
	"github.com/spf13/cobra"
)

var pubSubEventCmd = &cobra.Command{
	Use:   "pubSubEvent",
	Short: "Start a event listener for pubSubEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
