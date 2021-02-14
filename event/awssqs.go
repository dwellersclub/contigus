package event

import (
	"github.com/spf13/cobra"
)

var awssqsCmd = &cobra.Command{
	Use:   "awssqs",
	Short: "Start a event listener for awssqs",
	Run:   func(cmd *cobra.Command, args []string) {},
}
