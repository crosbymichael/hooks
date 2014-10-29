package handler

type Store interface {
	Save([]byte) error
}
