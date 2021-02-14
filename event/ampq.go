package event

import (
	"github.com/spf13/cobra"
)

var ampqCmd = &cobra.Command{
	Use:   "ampq",
	Short: "Start a event listener for ampq",
	Run:   func(cmd *cobra.Command, args []string) {},
}
