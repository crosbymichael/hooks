package main

import (
	"time"

	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/workers"
)

var multiplexCommand = cli.Command{
	Name:  "multiplex",
	Usage: "multiple is a command that accepts jobs off of a queue and sends a hook to third party services",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "rethink-addr", Usage: "rethinkdb address"},
		cli.StringFlag{Name: "rethink-key", Usage: "rethinkdb auth key"},
		cli.StringFlag{Name: "db", Value: "github", Usage: "rethinkdb database"},
		cli.StringFlag{Name: "table", Usage: "rethinkdb table"},
		cli.StringFlag{Name: "nsqlookupd", Usage: "nsqlookupd address"},
		cli.StringFlag{Name: "topic", Usage: "nsqd topic to listen to"},
		cli.StringFlag{Name: "channel", Value: "archive", Usage: "nsqd channel to listen to"},
		cli.DurationFlag{Name: "timeout", Value: 5 * time.Second, Usage: "timeout for the external webhook endpoint to respond before terminating the connection"},
	},
	Action: multiplexAction,
}

func multiplexAction(context *cli.Context) {
	session, err := NewRethinkdbSession(context)
	if err != nil {
		logger.Fatal(err)
	}
	handler := workers.NewMultiplexWorker(session, context.Duration("timeout"), logger)
	defer handler.Close()
	if err := ProcessQueue(handler, QueueOptsFromContext(context)); err != nil {
		logger.Fatal(err)
	}
}
