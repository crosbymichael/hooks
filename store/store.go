package store

import "io"

// Store provides a way to persist the raw webhook
type Store interface {
	io.Closer
	Save(table string, data []byte) error
}
