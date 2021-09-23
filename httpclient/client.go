package httpclient

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type CustomHttpClient struct{}

func NewCustomHttpClient() *CustomHttpClient {
	return &CustomHttpClient{}
}

func (c *CustomHttpClient) Head(uri string, timeoutInMs int, proxy *url.URL) error {
	client := getHttpClient(timeoutInMs, proxy)
	resp, err := client.Head(uri)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if 199 < resp.StatusCode && resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("[%s] status code does not indicate success %d", uri, resp.StatusCode))
	}

	return nil
}

func (c *CustomHttpClient) Check(uri string, timeoutInMs int, proxy *url.URL) error {
	client := getHttpClient(timeoutInMs, proxy)
	resp, err := client.Get(uri)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if 199 < resp.StatusCode && resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("[%s] status code does not indicate success %d", uri, resp.StatusCode))
	}

	_, err = ioutil.ReadAll(resp.Body)
	return err
}

func (c *CustomHttpClient) Get(uri string, timeoutInMs int, proxy *url.URL) ([]byte, error) {
	client := getHttpClient(timeoutInMs, proxy)
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if 199 < resp.StatusCode && resp.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("status code does not indicate success %d", resp.StatusCode))
	}

	return ioutil.ReadAll(resp.Body)
}

func getHttpClient(timeoutInMs int, proxy *url.URL) *http.Client {
	return &http.Client{
		Timeout:   time.Duration(timeoutInMs) * time.Millisecond,
		Transport: getTransport(proxy),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			r := via[len(via)-1]
			for _, c := range r.Cookies() {
				req.AddCookie(c)
			}

			return nil
		},
	}
}

func getTransport(proxy *url.URL) *http.Transport {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		MaxIdleConnsPerHost: 10000,
		DisableKeepAlives:   true,
	}

	if proxy != nil {
		t.Proxy = http.ProxyURL(proxy)
	}

	return t
}
