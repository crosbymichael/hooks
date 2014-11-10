package main

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/crosbymichael/hooks/handler"
	"github.com/crosbymichael/hooks/nsqd"
	"github.com/crosbymichael/hooks/rethinkdb"
	"github.com/crosbymichael/hooks/store"
	"github.com/gorilla/mux"
)

var serveCommand = cli.Command{
	Name:  "github",
	Usage: "handle github webhooks",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "addr", Value: ":8001", Usage: "HTTP address to serve api on"},
		cli.StringFlag{Name: "rethink-addr", Usage: "rethinkdb address"},
		cli.StringFlag{Name: "nsqd-addr", Usage: "nsqd address"},
		cli.StringFlag{Name: "db", Value: "github", Usage: "rethinkdb database"},
		cli.StringFlag{Name: "secret", Usage: "github secret for the webhook"},
	},
	Action: serveAction,
}

func newStore(context *cli.Context) (store.Store, error) {
	if addr := context.String("rethink-addr"); addr != "" {
		return rethinkdb.New(addr, context.String("db"))
	}
	if addr := context.String("nsqd-addr"); addr != "" {
		return nsqd.New(addr)
	}
	return nil, fmt.Errorf("no backend store to connect to. specify --rethink-addr || --nsqd-addr.")
}

func serveAction(context *cli.Context) {
	r := mux.NewRouter()
	store, err := newStore(context)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Close()
	r.Handle("/{user:.*}/{repo:.*}/", handler.NewGithubHandler(store, context.String("secret"), logger)).Methods("POST")
	if err := http.ListenAndServe(context.String("addr"), r); err != nil {
		logger.Fatal(err)
	}
}
