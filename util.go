package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
)

type QueueOpts struct {
	LookupdAddr string
	Topic       string
	Channel     string
	Concurrent  int
	Signals     []os.Signal
}

func QueueOptsFromContext(context *cli.Context) QueueOpts {
	return QueueOpts{
		Signals:     []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		LookupdAddr: context.String("nsqlookupd"),
		Topic:       context.String("topic"),
		Channel:     context.String("channel"),
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
