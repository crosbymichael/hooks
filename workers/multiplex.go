package workers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/dancannon/gorethink"
)

// Payload is the message body that is sent to the MultiplexWorker for sending a
// webhook payload to external urls
type Payload struct {
	// ID is the id of the key within the database where the data lives
	ID string `json:"id"`
	// URL is the url of the client where the payload should be sent
	URL string `json:"url"`
	// Table is the name of the table to fetch the raw data from
	Table string `json:"table"`
}

// NewMultiplexWorker returns a nsq.Handler that will process messages for calling external webook urls
// with a specified timeout.  It requires a session to rethinkdb so retreive the data for posting to the
// enternal urls.
func NewMultiplexWorker(session *gorethink.Session, timeout time.Duration, logger *logrus.Logger) *MultiplexWorker {
	return &MultiplexWorker{
		session: session,
		logger:  logger,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

type MultiplexWorker struct {
	client  *http.Client
	session *gorethink.Session
	logger  *logrus.Logger
}

func (w *MultiplexWorker) Close() error {
	return w.session.Close()
}

func (w *MultiplexWorker) HandleMessage(m *nsq.Message) error {
	var p *Payload
	if err := json.Unmarshal(m.Body, &p); err != nil {
		return err
	}
	request, err := w.newRequest(p)
	if err != nil {
		return err
	}

	resp, err := w.client.Do(request)
	if err != nil {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		w.logger.WithFields(logrus.Fields{
			"url":           p.URL,
			"error":         err,
			"response_code": code,
		}).Error("issue request")
		// do not return an error here because it's probably client code and we don't want to requeue
		return nil
	}
	w.logger.WithFields(logrus.Fields{
		"url":           p.URL,
		"response_code": resp.StatusCode,
	}).Debug("issue request")
	return nil
}

// newRequest creates a new http request to the payload's URL.  The body
// of the request is fetched from rethinkdb with the payload's ID as the
// rethinkdb document id.
func (w *MultiplexWorker) newRequest(p *Payload) (*http.Request, error) {
	hook, err := w.fetchPayload(p)
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

// fetchPayload returns the webhook's body in raw bytes.  It strips
// out the 'sha' field from the body so that it is not sent to the external user.
func (w *MultiplexWorker) fetchPayload(p *Payload) ([]byte, error) {
	r, err := gorethink.Table(p.Table).Get(p.ID).Field("payload").Without("sha").Run(w.session)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var i map[string]interface{}
	if err := r.One(&i); err != nil {
		return nil, err
	}
	return json.Marshal(i)
}
