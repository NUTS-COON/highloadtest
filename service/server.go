package service

import (
	"context"
	"encoding/json"
	"fmt"
	"highloadtest/checker"
	"highloadtest/logging"
	"highloadtest/search"
	"net/http"
	"net/url"
)

type Config struct {
	ListenAddr string
	UseProxy   bool
}

type SitesResponse struct {
	Host    string
	Threads int
}

type Service struct {
	s              http.Server
	conf           Config
	searchProvider search.Provider
	checker        *checker.HostTester
	l              logging.Logger
}

func NewService(conf Config, l logging.Logger, searchProvider search.Provider, checker *checker.HostTester) *Service {
	return &Service{
		s: http.Server{
			Addr: conf.ListenAddr,
		},
		conf:           conf,
		searchProvider: searchProvider,
		checker:        checker,
		l:              l,
	}
}

func (s *Service) Start() error {
	registerRouts(s)
	s.l.Info(fmt.Sprintf("Server starting on %s", s.conf.ListenAddr))
	return s.s.ListenAndServe()
}

func (s *Service) GracefulStop() {
	s.s.Shutdown(context.Background())
}

func registerRouts(s *Service) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sites", s.sites)

	s.s.Handler = mux
}

func (s *Service) sites(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)
	query := r.URL.Query().Get("search")
	useProxy := r.URL.Query().Get("withoutProxy") != "true" && s.conf.UseProxy
	withoutCache := r.URL.Query().Get("withoutCache") == "true"

	searchRes, err := s.searchProvider.Search(query, false)
	if err != nil {
		s.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	testResult := s.checker.CheckHighLoad(searchRes, useProxy, withoutCache)
	res := []SitesResponse{}
	for _, item := range testResult {
		if !item.NotAvailable && item.MaxThread > 0 {
			res = append(res, SitesResponse{
				Host:    item.Item.Host,
				Threads: item.MaxThread,
			})
		}
	}

	resp, _ := json.Marshal(res)
	w.WriteHeader(200)
	w.Write(resp)
}

func (s *Service) logRequest(r *http.Request) {
	unescaped, _ := url.QueryUnescape(r.URL.RequestURI())
	s.l.Info(fmt.Sprintf("req %s", unescaped))
}
