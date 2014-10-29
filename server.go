package main

import (
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/handler"
	"github.com/crosbymichael/hooks/rethinkdb"
	"github.com/gorilla/mux"
)

var serveCommand = cli.Command{
	Name:  "serve",
	Usage: "handle github webhooks",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "addr", Value: ":8001", Usage: "HTTP address to serve api on"},
		cli.StringFlag{Name: "rethink-addr", Value: "127.0.0.1:28015", Usage: "rethinkdb address"},
		cli.StringFlag{Name: "db", Value: "github", Usage: "rethinkdb database"},
		cli.StringFlag{Name: "table", Usage: "rethinkdb table to store data"},
	},
	Action: serveAction,
}

func serveAction(context *cli.Context) {
	r := mux.NewRouter()

	store, err := rethinkdb.New(context.String("rethink-addr"), context.String("db"), context.String("table"))
	if err != nil {
		logger.Fatal(err)
	}

	r.Handle("/{user:.*}/{repo:.*}/", handler.NewGithubHandler(store, logger)).Methods("POST")
	if err := http.ListenAndServe(context.String("addr"), r); err != nil {
		logger.Fatal(err)
	}
}
