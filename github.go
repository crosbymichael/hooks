package main

import (
	"net/http"

	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/server"
	"github.com/gorilla/mux"
)

var serveCommand = cli.Command{
	Name:  "github",
	Usage: "handle github webhooks",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "nsqd", Usage: "nsqd address"},
		cli.StringFlag{Name: "secret", Usage: "github secret for the webhook"},
		cli.StringFlag{Name: "addr", Value: ":8001", Usage: "HTTP address to serve api on"},
	},
	Action: serveAction,
}

func serveAction(context *cli.Context) {
	producer, err := nsq.NewProducer(context.String("nsqd"), nsq.NewConfig())
	if err != nil {
		logger.Fatal(err)
	}
	defer producer.Stop()
	r := mux.NewRouter()
	r.Handle(server.ROUTE, server.New(producer, context.String("secret"), logger)).Methods("POST")
	if err := http.ListenAndServe(context.String("addr"), r); err != nil {
		logger.Fatal(err)
	}
}
