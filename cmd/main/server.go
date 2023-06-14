package main

import (
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

func newServer(addr string, mux *http.ServeMux) *http.Server {
	s := &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s
}

func runServerAsync(cfg webServerConfig, wg *sync.WaitGroup, server *http.Server, logger *zap.Logger) {
	wg.Add(1)
	if cfg.useTLS {
		logger.Fatal(server.ListenAndServeTLS(cfg.certFilePath, cfg.keyFilePath).Error())
	} else {
		logger.Fatal(server.ListenAndServe().Error())
	}
}
