package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/rethinkdb"
	"github.com/crosbymichael/hooks/store"
)

var archiveCommand = cli.Command{
	Name:  "archive",
	Usage: "archive hooks into a persistent store",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "rethink-addr", Usage: "rethinkdb address"},
		cli.StringFlag{Name: "rethink-key", Usage: "rethinkdb auth key"},
		cli.StringFlag{Name: "db", Value: "github", Usage: "rethinkdb database"},
		cli.StringFlag{Name: "table", Usage: "rethinkdb table"},
		cli.StringFlag{Name: "nsqlookupd", Usage: "nsqlookupd address"},
		cli.StringFlag{Name: "topic", Usage: "nsqd topic to listen to"},
		cli.StringFlag{Name: "channel", Value: "archive", Usage: "nsqd channel to listen to"},
	},
	Action: archiveAction,
}

type storeHandler struct {
	table string
	store store.Store
}

func (s *storeHandler) HandleMessage(m *nsq.Message) error {
	return s.store.Save(s.table, m.Body)
}

func archiveAction(context *cli.Context) {
	s := make(chan os.Signal, 64)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	r, err := rethinkdb.New(context.String("rethink-addr"), context.String("db"), context.String("rethink-key"))
	if err != nil {
		logger.Fatal(err)
	}
	defer r.Close()

	consumer, err := nsq.NewConsumer(context.String("topic"), context.String("channel"), nsq.NewConfig())
	if err != nil {
		logger.Fatal(err)
	}
	consumer.AddHandler(&storeHandler{store: r, table: context.String("table")})
	if err := consumer.ConnectToNSQLookupd(context.String("nsqlookupd")); err != nil {
		logger.Fatal(err)
	}

	for {
		select {
		case <-consumer.StopChan:
			return
		case sig := <-s:
			logger.WithField("signal", sig).Debug("received signal")
			consumer.Stop()
		}
	}
}
