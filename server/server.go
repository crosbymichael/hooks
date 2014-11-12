package server

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
)

const ROUTE = "/{user:.*}/{repo:.*}/"

// New returns a new http.Handler that handles github webhooks from the github API.
// After receiving a hook the handler will push the message onto the specified NSQ Queue.
//
// producer is the connection to the NSQD instance.
// secret is the secret provided when you register the webhook in the github UI.
// logger is the standard logger for the application
func New(producer *nsq.Producer, secret string, logger *logrus.Logger) http.Handler {
	return &Server{
		producer: producer,
		secret:   secret,
		logger:   logger,
	}
}

// Server handles github webhooks and pushes the messages onto a specified
// queue under the repositories.  The queue name will the be repository name
// prepended with hoosk-{reponame}
type Server struct {
	producer *nsq.Producer
	secret   string
	logger   *logrus.Logger
}

func (h *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		repo       = parseRepo(r)
		requestLog = h.logger.WithFields(newFields(r, repo))
	)
	requestLog.Debug("web hook received")

	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		requestLog.WithField("error", err).Error("read request body")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !validateSignature(requestLog, r, h.secret, data) {
		requestLog.Warn("signature verification failed")
		// return a generic NOTFOUND for auth/verification errors
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err := h.producer.Publish(fmt.Sprintf("hooks-%s", repo.Name), data); err != nil {
		requestLog.WithField("error", err).Error("publish payload onto queue")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
