package main

import (
	"net/http"

	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/server"
	"github.com/gorilla/mux"
)

var githubCommand = cli.Command{
	Name:   "github",
	Usage:  "handle github webhooks by pushing them onto a queue names hooks-{reponame}",
	Action: githubAction,
}

func githubAction(context *cli.Context) {
	producer, err := nsq.NewProducer(config.NSQD, nsq.NewConfig())
	if err != nil {
		logger.Fatal(err)
	}
	defer producer.Stop()
	r := mux.NewRouter()
	r.Handle(server.ROUTE, server.New(producer, config.Github.Secret, logger)).Methods("POST")
	if err := http.ListenAndServe(config.Github.Listen, r); err != nil {
		logger.Fatal(err)
	}
}
