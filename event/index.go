package event

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hugo",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
				  love by spf13 and friends in Go.
				  Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {

	comds := []*cobra.Command{
		ampqCmd, awssqsCmd, azureEventCmd,
		cronEventCmd, pubSubEventCmd, kafkaEventCmd,
		minioEventCmd, mqttEventCmd, natsEventCmd,
		nsqEventCmd, redisEventCmd, fileEventCmd,
	}

	for _, comd := range comds {
		rootCmd.AddCommand(comd)
	}

}

type baseEventListener struct{}

type baseEventSource struct{}

func (bes *baseEventSource) Forward() error {
	return nil
}
