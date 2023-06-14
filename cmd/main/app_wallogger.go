package main

import "fmt"

func (a *application) loadWalFiles() {
	if a.config.walLoad.fileDirectory != "" {
		totalLoaded, totalErrored, err := a.importWalFiles(a.matcher, a.config.walLoad.fileDirectory, a.config.walLoad.filePrefix, a.logger)
		if err != nil {
			a.logger.Fatal(err.Error())
		}
		a.logger.Info(fmt.Sprintf("Loaded %d entries, %d failed\n", totalLoaded, totalErrored))
	} else {
		a.logger.Info("No WAL load parameters specified")
	}
}

func (a *application) newWalLogger() {
	if a.config.walWrite.fileDirectory == "" {
		a.config.writeWalFiles = false
		return
	}
	a.config.writeWalFiles = true
	wfm, err := newWalFileManager(a, a.config.walWrite.fileDirectory, a.config.walWrite.filePrefix, a.config.walWrite.maxEntriesPerFile)
	if err != nil {
		a.logger.Fatal(err.Error())
	}
	a.walFileMgr = wfm
}
