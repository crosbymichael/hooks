package handler

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/hooks/store"
	"github.com/gorilla/mux"
)

func NewGithubHandler(store store.Store, secret string, logger *logrus.Logger) http.Handler {
	return &Github{
		store:  store,
		logger: logger,
		secret: secret,
	}
}

type Github struct {
	store  store.Store
	secret string
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

	if !h.validateSignature(fields, r, data) {
		h.logger.WithFields(fields).Warn("github webhook signature verification failed")
		http.Error(w, http.StatusText(404), 404)
		return
	}

	if err := h.store.Save(repo, data); err != nil {
		h.logger.WithFields(fields).Errorf("save %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Github) validateSignature(fields logrus.Fields, r *http.Request, payload []byte) bool {
	// if we don't have a secret to validate then just return true
	// because the user does not care about security
	if h.secret == "" {
		return true
	}
	actual := r.Header.Get("X-Hub-Signature")
	fields["gh_signature"] = actual
	expected, err := getExpectedSignature([]byte(h.secret), payload)
	if err != nil {
		h.logger.WithFields(fields).Errorf("expected signature %s", err)
		return false
	}
	fields["signature"] = expected
	h.logger.WithFields(fields).Debugf("github request signature")
	return hmac.Equal([]byte(expected), []byte(actual))
}

func getExpectedSignature(raw, payload []byte) (string, error) {
	mac := hmac.New(sha1.New, raw)
	if _, err := mac.Write(payload); err != nil {
		return "", nil
	}
	return fmt.Sprintf("sha1=%s", hex.EncodeToString(mac.Sum(nil))), nil
}
