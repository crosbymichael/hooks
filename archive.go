package main

import (
	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/workers"
)

var archiveCommand = cli.Command{
	Name:   "archive",
	Usage:  "archive hooks into a rethinkdb for processing",
	Action: archiveAction,
}

func archiveAction(context *cli.Context) {
	session, err := NewRethinkdbSession()
	if err != nil {
		logger.Fatal(err)
	}
	defer session.Close()
	producer, err := nsq.NewProducer(config.NSQD, nsq.NewConfig())
	if err != nil {
		logger.Fatal(err)
	}
	defer producer.Stop()
	handler := workers.NewArchiveWorker(session, config.Archive.ArchiveTable, config.Archive.SubscribersTable, config.Archive.BroadcastChannel, producer)
	if err := ProcessQueue(handler, QueueOptsFromContext(config.Archive.HooksChannel, "archive")); err != nil {
		logger.Fatal(err)
	}
}
