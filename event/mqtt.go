package event

import (
	"github.com/spf13/cobra"
)

var mqttEventCmd = &cobra.Command{
	Use:   "mqttEvent",
	Short: "Start an event listener for mqttEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
