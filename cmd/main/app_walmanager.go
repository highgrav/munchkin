package main

import (
	"fmt"
	"github.com/highgrav/munchkin/internal/wal"
	"go.uber.org/zap"
	"sync"
	"time"
)

type walFileManager struct {
	app               *application
	dir               string
	currFilePath      string
	filePrefix        string
	file              *wal.WalFile
	maxEntries        int
	maxDuration       time.Duration
	mu                sync.Mutex
	totalLogEntries   uint64
	currentLogEntries int64
}

func newWalFileManager(app *application, dir, prefix string, maxEntriesPerFile int) (*walFileManager, error) {
	wfm := &walFileManager{
		app:        app,
		dir:        dir,
		filePrefix: prefix,
		maxEntries: maxEntriesPerFile,
	}
	cfp, err := wal.CreateWalFile(wfm.dir, wfm.filePrefix)
	if err != nil {
		return nil, err
	}
	wfm.currFilePath = cfp
	wf, err := wal.OpenWalFile(wfm.currFilePath)
	if err != nil {
		return nil, err
	}
	wfm.file = wf
	app.logger.Info(fmt.Sprintf("Writing WAL files to %s\n", wfm.dir))
	return wfm, nil
}

func (wfm *walFileManager) rotateWalFile() error {
	wfm.mu.Lock()
	defer wfm.mu.Unlock()
	wfm.file.Close()
	cfp, err := wal.CreateWalFile(wfm.dir, wfm.filePrefix)
	if err != nil {
		return err
	}
	wfm.currFilePath = cfp
	wf, err := wal.OpenWalFile(wfm.currFilePath)
	if err != nil {
		return err
	}
	wfm.file = wf
	return nil
}

func (wfm *walFileManager) writeWalFileEntry(timestamp int64, key, pattern string, action uint16, logger *zap.Logger) {
	if wfm.currentLogEntries >= int64(wfm.maxEntries) {
		err := wfm.rotateWalFile()
		if err != nil {
			logger.Fatal(err.Error())
		}
	}
	err := wfm.file.Write(timestamp, []byte(key), []byte(pattern), action)
	if err != nil {
		logger.Error(err.Error())
	}
	wfm.totalLogEntries++
	wfm.currentLogEntries++
}

func (wfm *walFileManager) closeWalFile() {
	wfm.file.Close()
}
