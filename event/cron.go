package event

import (
	"github.com/spf13/cobra"
)

var cronEventCmd = &cobra.Command{
	Use:   "cronEvent",
	Short: "Start an event listener for cronEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
