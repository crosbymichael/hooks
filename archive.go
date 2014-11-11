package main

import (
	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/workers"
)

var archiveCommand = cli.Command{
	Name:  "archive",
	Usage: "archive hooks into a rethinkdb for processing",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "rethink-addr", Usage: "rethinkdb address"},
		cli.StringFlag{Name: "rethink-key", Usage: "rethinkdb auth key"},
		cli.StringFlag{Name: "db", Value: "github", Usage: "rethinkdb database"},
		cli.StringFlag{Name: "table", Usage: "rethinkdb table"},
		cli.StringFlag{Name: "nsqlookupd", Usage: "nsqlookupd address"},
		cli.StringFlag{Name: "topic", Usage: "nsqd topic to listen to"},
		cli.StringFlag{Name: "channel", Value: "archive", Usage: "nsqd channel to listen to"},
		cli.StringFlag{Name: "nsqd", Usage: "nsqd address"},
		cli.StringFlag{Name: "ext-urls", Usage: "rethinkdb table for external urls to publish hooks to"},
		cli.StringFlag{Name: "hook-queue", Usage: "queue that has workers on it to publish webhooks"},
	},
	Action: archiveAction,
}

func archiveAction(context *cli.Context) {
	var (
		producer                           *nsq.Producer
		externalURLTable, hookPublishQueue string
	)
	session, err := NewRethinkdbSession(context)
	if err != nil {
		logger.Fatal(err)
	}
	defer session.Close()
	if nsqd := context.String("nsqd"); nsqd != "" {
		if producer, err = nsq.NewProducer(nsqd, nsq.NewConfig()); err != nil {
			logger.Fatal(err)
		}
		defer producer.Stop()
		externalURLTable = context.String("ext-urls")
		hookPublishQueue = context.String("hook-queue")
	}

	handler := workers.NewArchiveWorker(session, context.String("table"), externalURLTable, hookPublishQueue, producer)
	if err := ProcessQueue(handler, QueueOptsFromContext(context)); err != nil {
		logger.Fatal(err)
	}
}
