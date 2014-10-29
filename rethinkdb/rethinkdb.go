package rethinkdb

import "github.com/dancannon/gorethink"

type RethinkStore struct {
	session *gorethink.Session
	table   string
}

type payload struct {
	Timestamp interface{} `gorethink:"timestamp"`
	Payload   interface{} `gorethink:"payload"`
}

func New(addr, db, table string) (*RethinkStore, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  addr,
		Database: db,
	})
	if err != nil {
		return nil, err
	}
	return &RethinkStore{
		session: session,
		table:   table,
	}, nil
}

func (r *RethinkStore) Save(data []byte) error {
	p := payload{
		Timestamp: gorethink.Now(),
		Payload:   gorethink.Json(string(data)),
	}

	_, err := gorethink.Table(r.table).Insert(p).RunWrite(r.session)
	return err
}
