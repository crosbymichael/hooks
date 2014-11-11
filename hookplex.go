package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/codegangsta/cli"
	"github.com/dancannon/gorethink"
)

var multiplexCommand = cli.Command{
	Name:  "multiplex",
	Usage: "multiple is a command that accepts jobs off of a queue and sends a hook to third party services",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "rethink-addr", Usage: "rethinkdb address"},
		cli.StringFlag{Name: "rethink-key", Usage: "rethinkdb auth key"},
		cli.StringFlag{Name: "db", Value: "github", Usage: "rethinkdb database"},
		cli.StringFlag{Name: "table", Usage: "rethinkdb table"},
		cli.StringFlag{Name: "nsqlookupd", Usage: "nsqlookupd address"},
		cli.StringFlag{Name: "topic", Usage: "nsqd topic to listen to"},
		cli.StringFlag{Name: "channel", Value: "archive", Usage: "nsqd channel to listen to"},
	},
	Action: multiplexAction,
}

func newMultiplexHandler(context *cli.Context) (*multiplexHandler, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  context.String("rethink-addr"),
		AuthKey:  context.String("rethink-key"),
		Database: context.String("db"),
	})
	if err != nil {
		return nil, err
	}
	return &multiplexHandler{
		table:   context.String("table"),
		session: session,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

type payload struct {
	// ID is the id of the key within the database where the data lives
	ID string `json:"id"`
	// URL is the url of the client where the payload should be sent
	URL string `json:"url"`
}

type multiplexHandler struct {
	client  *http.Client
	session *gorethink.Session
	table   string
}

func (h *multiplexHandler) Close() error {
	return h.session.Close()
}

func (h *multiplexHandler) HandleMessage(m *nsq.Message) error {
	var p *payload
	if err := json.Unmarshal(m.Body, &p); err != nil {
		return err
	}
	request, err := h.newRequest(p)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(request)
	if err != nil {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		logger.WithFields(logrus.Fields{
			"url":           p.URL,
			"error":         err,
			"response_code": code,
		}).Error("issue request")
		// do not return an error here because it's probably client code and we don't want to requeue
		return nil
	}
	logger.WithFields(logrus.Fields{
		"url":           p.URL,
		"response_code": resp.StatusCode,
	}).Debug("issue request")
	return nil
}

func (h *multiplexHandler) fetchPayload(id string) ([]byte, error) {
	r, err := gorethink.Table(h.table).Get(id).Field("payload").Without("sha").Run(h.session)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var data []byte
	if err := r.One(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (h *multiplexHandler) newRequest(p *payload) (*http.Request, error) {
	hook, err := h.fetchPayload(p.ID)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", p.URL, bytes.NewBuffer(hook))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	return request, err
}

func multiplexAction(context *cli.Context) {
	handler, err := newMultiplexHandler(context)
	if err != nil {
		logger.Fatal(err)
	}
	defer handler.Close()
	if err := ProcessQueue(handler, QueueOptsFromContext(context)); err != nil {
		logger.Fatal(err)
	}
}
