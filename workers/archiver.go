package workers

import (
	"encoding/json"

	"github.com/bitly/go-nsq"
	"github.com/dancannon/gorethink"
)

// ExternalURL is the scheme for when users register external urls that are to be
// called so that they can subscribe to events on the repo.
type ExternalURL struct {
	URL string `gorethink:"url"`
}

// archivePayload defines the serialization scheme for saving a raw webhook
// into rethinkdb.  This is done to allow us the flexibility to define new top
// level fields without poluting the raw webhook payload.  i.e. the automaticly
// generated id for the document provided by rethinkdb.
type archivePayload struct {
	Timestamp interface{} `gorethink:"timestamp"`
	Payload   interface{} `gorethink:"payload"`
}

func NewArchiveWorker(session *gorethink.Session, table, subscribers, topic string, producer *nsq.Producer) *ArchiveWorker {
	return &ArchiveWorker{
		session:     session,
		table:       table,
		producer:    producer,
		subscribers: subscribers,
		topic:       topic,
	}
}

type ArchiveWorker struct {
	table       string
	subscribers string
	topic       string
	session     *gorethink.Session
	producer    *nsq.Producer
}

func (a *ArchiveWorker) HandleMessage(m *nsq.Message) error {
	resp, err := gorethink.Table(a.table).Insert(archivePayload{
		Timestamp: gorethink.Now(),
		Payload:   gorethink.Json(string(m.Body)),
	}).RunWrite(a.session)
	if err != nil {
		return err
	}
	// if a producer is set then we are to push the resulting data for each of the webhhooks
	// onto a new queue.
	if a.producer != nil {
		return a.pushPayload(resp.GeneratedKeys[0])
	}
	return nil
}

func (a *ArchiveWorker) pushPayload(id string) error {
	urls, err := a.fetchExternalHookURLs()
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		return nil
	}
	var (
		batch    [][]byte
		template = Payload{
			ID:    id,
			Table: a.table,
		}
	)
	for _, u := range urls {
		template.URL = u
		data, err := json.Marshal(template)
		if err != nil {
			return err
		}
		batch = append(batch, data)
	}
	return a.producer.MultiPublish(a.topic, batch)
}

func (a *ArchiveWorker) fetchExternalHookURLs() ([]string, error) {
	resp, err := gorethink.Table(a.subscribers).Run(a.session)
	if err != nil {
		return nil, err
	}
	var out []ExternalURL
	if err := resp.All(&out); err != nil {
		return nil, err
	}
	return urls(out), nil
}

func urls(ext []ExternalURL) (out []string) {
	for _, e := range ext {
		out = append(out, e.URL)
	}
	return out
}
