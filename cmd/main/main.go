package main

import (
	"context"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
)

var app *application

func main() {
	var err error = nil
	cfg := appConfig{}
	flag.IntVar(&cfg.poolSize, "poolSz", 8, "Number of concurrent workers")

	// WAL import
	flag.StringVar(&cfg.walLoad.fileDirectory, "walLoadDir", "", "Directory to load WAL files from (if any)")
	flag.StringVar(&cfg.walLoad.filePrefix, "walLoadPrefix", "mwal-", "Prefix for WAL files to be loaded (if any)")

	// WAL files
	flag.StringVar(&cfg.walWrite.fileDirectory, "walDir", "", "Directory to save WAL files to")
	flag.StringVar(&cfg.walWrite.filePrefix, "walPrefix", "mwal-", "Prefix for WAL files to be saved")
	flag.IntVar(&cfg.walWrite.maxEntriesPerFile, "walMaxEntries", 10000, "Maximum number of entries to store in any single WAL file")

	// Web server configuration
	flag.IntVar(&cfg.matchServer.port, "matchApiPort", 8080, "Port to run matching API on")
	flag.IntVar(&cfg.adminServer.port, "adminApiPort", 9090, "Port to run admin API on")
	flag.IntVar(&cfg.clusterServer.port, "raftApiPort", 7070, "Port to run cluster API on")

	viper.SetDefault("PoolSize", 10)
	viper.SetDefault("MatchApiPort", 8080)
	err = viper.BindPFlag("PoolSize", flag.Lookup("poolSz"))

	flag.Parse()
	if err != nil {
		log.Fatal("Could not bind command flag 'MatchPoolSize'")
	}

	app = newApplication(cfg)

	app.logger.Info("Starting servers...")
	chShutdown, err := app.startServer()
	if err != nil {
		app.logger.Fatal(err.Error())
	}

	_ = <-chShutdown
	app.logger.Info("Shutting down server...")
	err = app.apiServer.Shutdown(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
