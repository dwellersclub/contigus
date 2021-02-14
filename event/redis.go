package event

import (
	"github.com/spf13/cobra"
)

var redisEventCmd = &cobra.Command{
	Use:   "redisEvent",
	Short: "Start a event listener for redisEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
