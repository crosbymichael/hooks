package main

import (
	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/workers"
)

var broadcastCommand = cli.Command{
	Name:   "broadcast",
	Usage:  "broadcast is a command that accepts jobs off of a queue and sends a hook to third party services",
	Action: broadcastAction,
}

func broadcastAction(context *cli.Context) {
	session, err := NewRethinkdbSession()
	if err != nil {
		logger.Fatal(err)
	}
	handler := workers.NewMultiplexWorker(session, config.Broadcast.Timeout.Duration, logger)
	defer handler.Close()
	if err := ProcessQueue(handler, QueueOptsFromContext(config.Broadcast.Topic, config.Broadcast.Channel)); err != nil {
		logger.Fatal(err)
	}
}
