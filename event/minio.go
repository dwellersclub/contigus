package event

import (
	"github.com/spf13/cobra"
)

var minioEventCmd = &cobra.Command{
	Use:   "minioEvent",
	Short: "Start a event listener for minioEvent",
	Run:   func(cmd *cobra.Command, args []string) {},
}
