package main

import (
	"github.com/highgrav/munchkin/internal/util"
	"go.uber.org/zap"
	"log"
	"net/http"
	"quamina.net/go/quamina"
	"sync"
)

type application struct {
	config        *appConfig
	matcher       *quamina.Quamina
	lastUpdatedOn uint64
	pool          *util.ObjectPool[quamina.Quamina]
	apiServer     *http.Server
	adminServer   *http.Server
	walServer     *walServer
	chShutdown    chan struct{}
	serverWg      *sync.WaitGroup
	logger        *zap.Logger
	walFileMgr    *walFileManager
	cluster       *ClusterState
}

func newApplication(cfg appConfig) *application {
	a := &application{}
	a.config = &cfg

	err := a.newLogger()
	if err != nil {
		log.Fatal(err.Error())
	}
	a.logger.Info("Logger created")
	a.logger.Info("Creating matcher...")
	err = a.newMatcher()
	if err != nil {
		a.logger.Fatal(err.Error())
	}

	a.logger.Info("Creating worker pools...")
	err = a.newPool()
	if err != nil {
		a.logger.Fatal(err.Error())
	}

	a.logger.Info("Checking for importable WAL logs...")
	a.loadWalFiles()

	a.logger.Info("Starting WAL logger...")
	a.newWalLogger()

	a.logger.Info("Creating servers...")
	err = a.newServers()
	if err != nil {
		a.logger.Fatal(err.Error())
	}

	err = a.newWalServer()
	if err != nil {
		a.logger.Fatal(err.Error())
	}

	return a
}
