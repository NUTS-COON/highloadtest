package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"highloadtest/checker"
	"highloadtest/httpclient"
	"highloadtest/logging"
	"highloadtest/proxy"
	"highloadtest/search"
	"highloadtest/service"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	confPath string
	logger   logging.Logger
)

func main() {
	rootCMD := &cobra.Command{
		Use:           "root",
		Run:           Run,
		PreRunE:       PreRunE,
		PostRun:       PostRun,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCMD.Flags().StringVar(&confPath, "confPath", ".", "path config file")

	if err := rootCMD.Execute(); err != nil {
		log.Fatal(err)
	}
}

func readServiceConfig() service.Config {
	c := service.Config{}
	c.ListenAddr = viper.GetString("service.listen")
	c.UseProxy = viper.GetBool("service.use_proxy")

	return c
}

func readDataProviderConfig() httpclient.Config {
	c := httpclient.Config{}
	c.RequestTimeoutInMs = viper.GetInt("data_provider.timeout_in_ms")
	c.MaxProxyTestAttempts = viper.GetInt("data_provider.proxy_test_attempt")

	return c
}

func readHostTesterConfig() checker.Config {
	c := checker.Config{}
	c.InitThreadCount = viper.GetInt("checker.init_thread_count")
	c.MaxSearchIteration = viper.GetInt("checker.max_search_iteration")
	c.TestThreadsErrThreshold = viper.GetFloat64("checker.test_threads_err_threshold")
	c.TimeoutInMsBetweenAttempts = viper.GetInt("checker.timeout_in_ms_between_attempts")
	c.MaxParallelHostTesting = viper.GetInt("checker.max_parallel_host_testing")
	c.HostTestCount = viper.GetInt("checker.host_test_count")

	return c
}

func Run(c *cobra.Command, args []string) {
	serviceConfig := readServiceConfig()
	dataProviderConfig := readDataProviderConfig()
	hostTesterConfig := readHostTesterConfig()

	proxyQueue := proxy.NewDefaultQueue()
	if serviceConfig.UseProxy {
		initProxyQueue(proxyQueue)
	}

	baseYandexUrl := viper.GetString("service.base_yandex_url")
	if baseYandexUrl == "" {
		logger.Error("service.base_yandex_url is empty")
	}

	httpClient := httpclient.NewCustomHttpClient()
	dataProvider := httpclient.NewDataProvider(httpClient, logger, proxyQueue, dataProviderConfig)
	searchProvider := search.NewYandexSearchProvider(baseYandexUrl, dataProvider)
	siteChecker := checker.NewHostLoadTester(hostTesterConfig, logger, dataProvider)

	server := service.NewService(serviceConfig, logger, searchProvider, siteChecker)

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM)
	complete := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		err := server.Start()
		if err != nil {
			logger.Error(err.Error())
			close(complete)
		}
	}()

	go func() {
		defer wg.Done()

		select {
		case <-sign:
			logger.Info("ShutDown signal received")
			break
		case <-complete:
			logger.Info("service stopped")
			break
		}

		server.GracefulStop()
	}()

	wg.Wait()
}

func PreRunE(c *cobra.Command, args []string) error {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(confPath)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	logLvl := viper.GetString("service.log_level")
	logger = createLogger("[CHECKER]", logLvl)
	logger.SetTarget(logging.NewConsoleTarget())
	writeLogs := viper.GetBool("service.write_logs_to_file")
	if writeLogs {
		logPath := viper.GetString("service.log_path")
		logger.ToFile(logPath)
	}

	return nil
}

func PostRun(c *cobra.Command, args []string) {
	logger.Close()
}

func createLogger(prefix string, lvl string) logging.Logger {
	return logging.New(logging.ParseLogLevel(lvl), prefix)
}

func initProxyQueue(proxyQueue proxy.Queue) {
	proxyPath := viper.GetString("service.proxy_path")
	proxies, err := proxy.ParseFromFile(proxyPath)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	if viper.GetBool("service.check_proxy") {
		checkUrl := viper.GetString("service.check_proxy_url")
		proxies = getWorkingProxies(proxies, viper.GetInt("data_provider.timeout_in_ms"), checkUrl)
	}

	logger.Info(fmt.Sprintf("%d proxies", len(proxies)))
	for _, p := range proxies {
		proxyQueue.Enqueue(p)
	}
}

func getWorkingProxies(proxies []*url.URL, timeoutInMs int, uri string) []*url.URL {
	var res []*url.URL
	l := make(chan bool, 10)

	var wg sync.WaitGroup
	for _, p := range proxies {
		l <- true
		wg.Add(1)
		go func(p *url.URL) {
			if checkProxy(uri, timeoutInMs, p) {
				res = append(res)
			}
			_ = <-l
			wg.Done()
		}(p)
	}

	wg.Wait()
	close(l)

	return res
}

func checkProxy(uri string, timeout int, proxy *url.URL) bool {
	c := httpclient.NewCustomHttpClient()
	err := c.Check(uri, timeout, proxy)
	logger.Debug(fmt.Sprintf("proxy %s check err %+v", proxy, err))
	return err != nil
}
