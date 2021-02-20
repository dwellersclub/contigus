package event

import (
	"github.com/spf13/cobra"
)

var natsEventCmd = &cobra.Command{
	Use:   "natsEvent",
	Short: "Start an event listener for natsEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
