package main

import (
	"net/http"
	"strconv"
	"sync"
)

func (a *application) newServers() error {
	matchMux := http.NewServeMux()
	adminMux := http.NewServeMux()
	//	clusterMux := http.NewServeMux()

	matchMux.HandleFunc("/api/v1/heartbeat", handleGetHeartbeat)
	matchMux.HandleFunc("/api/v1/match", a.handleHttpPostMatch)

	adminMux.HandleFunc("/api/admin/v1/add", a.handleHttpPostAddRule)
	adminMux.HandleFunc("/api/admin/v1/delete-by-key", a.handleHttpDeleteByKey)

	a.apiServer = newServer(":"+strconv.Itoa(a.config.matchServer.port), matchMux)
	a.adminServer = newServer(":"+strconv.Itoa(a.config.adminServer.port), adminMux)
	return nil
}

func (a *application) startServer() (chan struct{}, error) {
	a.chShutdown = make(chan struct{})
	a.serverWg = new(sync.WaitGroup)
	go runServerAsync(a.config.matchServer, a.serverWg, a.apiServer, a.logger)
	go runServerAsync(a.config.adminServer, a.serverWg, a.adminServer, a.logger)
	return a.chShutdown, nil
}
