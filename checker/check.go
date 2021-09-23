package checker

import (
	"fmt"
	"highloadtest/httpclient"
	"highloadtest/logging"
	"highloadtest/search"
	"math"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	RequestTimeoutInMs         int
	MaxProxyTestAttempts       int
	InitThreadCount            int
	MaxSearchIteration         int
	TestThreadsErrThreshold    float64
	TimeoutInMsBetweenAttempts int
	MaxParallelHostTesting     int
	HostTestCount              int
}

type HostTester struct {
	cachedResult map[string]int
	dataProvider *httpclient.DataProvider
	conf         Config
	l            logging.Logger
}

type CheckResult struct {
	Item         search.ResourceItem
	MaxThread    int
	NotAvailable bool
}

func NewHostLoadTester(config Config, l logging.Logger, dataProvider *httpclient.DataProvider) *HostTester {
	return &HostTester{
		dataProvider: dataProvider,
		cachedResult: map[string]int{},
		conf:         config,
		l:            l,
	}
}

func (t *HostTester) CheckHighLoad(items []search.ResourceItem, useProxy, withoutCache bool) []CheckResult {
	res := make([]CheckResult, len(items))
	l := make(chan bool, t.conf.MaxParallelHostTesting)

	var wg sync.WaitGroup
	for i, item := range items {
		l <- true
		wg.Add(1)
		go func(i int, item search.ResourceItem) {
			res[i] = t.checkItem(item, useProxy, withoutCache)
			_ = <-l
			wg.Done()
		}(i, item)
	}

	wg.Wait()
	close(l)

	return res
}

func (t *HostTester) checkItem(item search.ResourceItem, useProxy, withoutCache bool) CheckResult {
	if !withoutCache {
		if m, ok := t.cachedResult[item.Host]; ok {
			return CheckResult{
				Item:      item,
				MaxThread: m,
			}
		}
	}

	var p *url.URL
	var data []byte
	if useProxy {
		p, data = t.dataProvider.LoadWithProxy(item.Url)
	} else {
		data = t.dataProvider.Load(item.Url)
	}

	if data == nil {
		return CheckResult{
			Item:         item,
			NotAvailable: true,
		}
	}

	maxThread := t.getMaxThread(item, p)
	if p != nil {
		t.dataProvider.EnqueueProxy(p)
	}
	t.cachedResult[item.Host] = maxThread
	return CheckResult{
		Item:      item,
		MaxThread: maxThread,
	}
}

func (t *HostTester) getMaxThread(item search.ResourceItem, proxy *url.URL) int {
	var coef float64 = 1
	var coefPrev float64 = 0
	var threadCount int
	var threadCountSuccess int
	var upperBoundFound bool
	var errCount int
	maxSearchIteration := t.conf.MaxSearchIteration

	for {
		if upperBoundFound {
			maxSearchIteration = maxSearchIteration - 1
		}

		if maxSearchIteration < 0 {
			break
		}

		threadCount = int(float64(t.conf.InitThreadCount) * coef)
		maxErrCount := int(math.Round(float64(threadCount) * t.conf.TestThreadsErrThreshold))

		for j := 0; j < t.conf.HostTestCount; j++ {
			errCount = t.testThreadsCount(item, threadCount, proxy)
			if errCount > maxErrCount {
				break
			}
		}

		coefTmp := coef
		if errCount <= maxErrCount {
			threadCountSuccess = threadCount
			if upperBoundFound {
				coef = coef + (math.Abs(coef-coefPrev) / 2)
			} else {
				coef = coef * 2
			}
		} else {
			coef = coef - (math.Abs(coef-coefPrev) / 2)
			upperBoundFound = true
		}

		coefPrev = coefTmp
		t.l.Info(fmt.Sprintf("[%s] threads %d successThreads %d errCount %d/%d upperBoundFound %t",
			item.Host, threadCount, threadCountSuccess, errCount, maxErrCount, upperBoundFound))
		time.Sleep(time.Duration(t.conf.TimeoutInMsBetweenAttempts) * time.Microsecond)
	}

	return threadCountSuccess
}

func (t *HostTester) testThreadsCount(item search.ResourceItem, count int, proxy *url.URL) int {
	var errCount int32 = 0
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			err := t.dataProvider.Check(item.Url, proxy)
			if err != nil {
				t.l.Debug(fmt.Sprintf("[%s] threads: %d err %s", item.Host, count, err.Error()))
				atomic.AddInt32(&errCount, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return int(errCount)
}
