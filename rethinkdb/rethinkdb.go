package rethinkdb

import "github.com/dancannon/gorethink"

type RethinkStore struct {
	session *gorethink.Session
}

type payload struct {
	Timestamp interface{} `gorethink:"timestamp"`
	Payload   interface{} `gorethink:"payload"`
}

func New(addr, db string) (*RethinkStore, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  addr,
		Database: db,
	})
	if err != nil {
		return nil, err
	}
	return &RethinkStore{
		session: session,
	}, nil
}

func (r *RethinkStore) Save(table string, data []byte) error {
	p := payload{
		Timestamp: gorethink.Now(),
		Payload:   gorethink.Json(string(data)),
	}
	_, err := gorethink.Table(table).Insert(p).RunWrite(r.session)
	return err
}
