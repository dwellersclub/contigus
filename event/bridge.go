package event

import (
	nats "github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Start an event listener for bridge",
	Run:   func(cmd *cobra.Command, args []string) {},
}

type eventPersister struct {
}

func (evp *eventPersister) init() {
	nc, err := nats.Connect(nats.DefaultURL)

	ch := make(chan *nats.Msg, 64)
	sub, err := nc.ChanSubscribe("foo", ch)
	msg := <-ch
}
