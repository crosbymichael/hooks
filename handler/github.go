package handler

import (
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func NewGithubHandler(store Store, logger *logrus.Logger) http.Handler {
	return &Github{
		store:  store,
		logger: logger,
	}
}

type Github struct {
	store  Store
	logger *logrus.Logger
}

func (h *Github) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user, repo := vars["user"], vars["repo"]
	fields := logrus.Fields{
		"host":    r.Host,
		"user":    user,
		"repo":    repo,
		"handler": "github",
	}
	h.logger.WithFields(fields).Debug("web hook received")

	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		h.logger.WithFields(fields).Errorf("read all %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.store.Save(repo, data); err != nil {
		h.logger.WithFields(fields).Errorf("save %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
