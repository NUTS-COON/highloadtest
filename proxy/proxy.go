package proxy

import (
	"net/url"
)

type Queue interface {
	Enqueue(u *url.URL)
	Dequeue() *url.URL
}

type DefaultQueue struct {
	proxies []*url.URL
}

func NewDefaultQueue() *DefaultQueue {
	return &DefaultQueue{proxies: []*url.URL{}}
}

func (q *DefaultQueue) Enqueue(u *url.URL) {
	q.proxies = append(q.proxies, u)
}

func (q *DefaultQueue) Dequeue() *url.URL {
	if len(q.proxies) == 0 {
		return nil
	}

	u := q.proxies[0]
	q.proxies = q.proxies[1:]

	return u
}
