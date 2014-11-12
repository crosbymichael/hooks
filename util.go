package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
	"github.com/dancannon/gorethink"
)

type QueueOpts struct {
	LookupdAddr string
	Topic       string
	Channel     string
	Concurrent  int
	Signals     []os.Signal
}

func QueueOptsFromContext(channel, topic string) QueueOpts {
	return QueueOpts{
		Signals:     []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		LookupdAddr: config.Lookupd,
		Topic:       topic,
		Channel:     channel,
		Concurrent:  1,
	}
}

func ProcessQueue(handler nsq.Handler, opts QueueOpts) error {
	if opts.Concurrent == 0 {
		opts.Concurrent = 1
	}
	s := make(chan os.Signal, 64)
	signal.Notify(s, opts.Signals...)

	consumer, err := nsq.NewConsumer(opts.Topic, opts.Channel, nsq.NewConfig())
	if err != nil {
		return err
	}
	consumer.AddConcurrentHandlers(handler, opts.Concurrent)
	if err := consumer.ConnectToNSQLookupd(opts.LookupdAddr); err != nil {
		return err
	}

	for {
		select {
		case <-consumer.StopChan:
			return nil
		case sig := <-s:
			logger.WithField("signal", sig).Debug("received signal")
			consumer.Stop()
		}
	}
	return nil
}

func NewRethinkdbSession() (*gorethink.Session, error) {
	return gorethink.Connect(gorethink.ConnectOpts{
		Database: config.RethinkdbDatabase,
		AuthKey:  config.RethinkdbKey,
		Address:  config.RethinkdbAddress,
	})
}
