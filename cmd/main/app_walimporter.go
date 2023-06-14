package main

import (
	"errors"
	"github.com/highgrav/munchkin/internal/wal"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"quamina.net/go/quamina"
	"sort"
	"strings"
)

func (app *application) importWalFiles(quan *quamina.Quamina, dir, prefix string, logger *zap.Logger) (totalEntries int64, totalErrors int64, err error) {
	s, err := os.Stat(dir)
	if err != nil {
		return -1, -1, err
	}
	if s.IsDir() == false {
		return -1, -1, errors.New(dir + " is not a directory!")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return -1, -1, err
	}

	var files []string = make([]string, 0)
	for _, v := range entries {
		if !v.IsDir() && strings.HasPrefix(v.Name(), prefix) {
			files = append(files, v.Name())
		}
		sort.Strings(files)
	}
	if len(files) == 0 {
		return 0, 0, nil
	}

	for _, fname := range files {
		walf, err := wal.OpenWalFile(filepath.Join(dir, fname))
		if err != nil {
			logger.Error("wal.Open: " + err.Error())
			totalErrors++
			continue
		}
		for walf.HasNext() {
			walEntry, err := walf.Next()
			if err != nil {
				logger.Error("wal.Next: " + err.Error())
				totalErrors++
				continue
			}
			if walEntry.Timestamp >= app.lastUpdatedOn && len(walEntry.Key) > 0 && len(walEntry.Pattern) > 0 && walEntry.Action == wal.WAL_ADD {
				err = quan.AddPattern(string(walEntry.Key), string(walEntry.Pattern))
				if err != nil {
					logger.Error("quan.AddPattern: " + err.Error())
					totalErrors++
					continue
				}
				app.lastUpdatedOn = walEntry.Timestamp
				totalEntries++
				continue
			} else if walEntry.Timestamp >= app.lastUpdatedOn && len(walEntry.Key) > 0 && len(walEntry.Pattern) <= 1 && walEntry.Action == wal.WAL_DEL {
				err = quan.DeletePatterns(string(walEntry.Key))

				if err != nil {
					logger.Error("quan.DeletePatterns: " + err.Error())
					totalErrors++
					continue
				}
				app.lastUpdatedOn = walEntry.Timestamp
				totalEntries++
				continue
			} else {
				// TODO -- Once we can delete matching patterns from a key we'll handle this with a Pattern length > 1 check
			}
		}
		walf.Close()
	}

	return totalEntries, totalErrors, nil
}
