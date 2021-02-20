package event

import (
	"github.com/spf13/cobra"
)

var nsqEventCmd = &cobra.Command{
	Use:   "nsqEvent",
	Short: "Start an event listener for nsqEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
