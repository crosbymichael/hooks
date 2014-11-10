package nsqd

import (
	"fmt"

	"github.com/bitly/go-nsq"
)

func New(addr string) (*NsqStore, error) {
	p, err := nsq.NewProducer(addr, nil)
	if err != nil {
		return nil, err
	}
	return &NsqStore{
		p: p,
	}, nil
}

type NsqStore struct {
	p *nsq.Producer
}

func (n *NsqStore) Save(table string, data []byte) error {
	return n.p.Publish(fmt.Sprintf("hooks-%s", table), data)
}

func (n *NsqStore) Close() error {
	n.p.Stop()
	return nil
}
