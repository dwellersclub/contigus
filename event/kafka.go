package event

import (
	"github.com/spf13/cobra"
)

var kafkaEventCmd = &cobra.Command{
	Use:   "kafkaEvent",
	Short: "Start an event listener for kafkaEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
