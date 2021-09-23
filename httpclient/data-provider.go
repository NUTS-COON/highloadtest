package httpclient

import (
	"fmt"
	"highloadtest/logging"
	"highloadtest/proxy"
	"net/url"
)

type Config struct {
	MaxProxyTestAttempts int
	RequestTimeoutInMs   int
}

type DataProvider struct {
	client     *CustomHttpClient
	proxyQueue proxy.Queue
	conf       Config
	l          logging.Logger
}

func NewDataProvider(client *CustomHttpClient, l logging.Logger, proxyQueue proxy.Queue, conf Config) *DataProvider {
	return &DataProvider{client: client, proxyQueue: proxyQueue, conf: conf, l: l}
}

func (t *DataProvider) LoadWithProxy(uri string) (*url.URL, []byte) {
	for i := 0; i < t.conf.MaxProxyTestAttempts; i++ {
		p := t.proxyQueue.Dequeue()
		if p == nil {
			t.l.Error("[DataProvider] no proxy")
			return nil, nil
		}

		data, err := t.client.Get(uri, t.conf.RequestTimeoutInMs, p)
		if err == nil {
			return p, data
		}

		if err != nil {
			t.l.Debug(fmt.Sprintf("[DataProvider] [%s] proxy %s attempt %d err %s", uri, p, i+1, err.Error()))
		}
	}

	return nil, nil
}

func (t *DataProvider) Load(uri string) []byte {
	data, err := t.client.Get(uri, t.conf.RequestTimeoutInMs, nil)

	if err != nil {
		t.l.Debug(fmt.Sprintf("[DataProvider] [%s] err %s", uri, err.Error()))
		return nil
	}

	return data
}

func (t *DataProvider) Check(uri string, proxy *url.URL) error {
	return t.client.Check(uri, t.conf.RequestTimeoutInMs, proxy)
}

func (t *DataProvider) EnqueueProxy(u *url.URL) {
	if u != nil {
		t.proxyQueue.Enqueue(u)
	}
}
