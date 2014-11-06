package handler

type Store interface {
	Save(table string, data []byte) error
}
