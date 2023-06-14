package main

import (
	"errors"
	"github.com/highgrav/munchkin/internal/wal"
	"quamina.net/go/quamina"
	"time"
)

var ErrNotImplemented = errors.New("Not implemented!")

// TODO -- when writing to WAL files, may need to buffer changes to an in-memory
//// structure if the server is streaming historical changes to a new cluster
//// node.

// addRule adds a rule to the matcher without logging it
func (a *application) addRule(id quamina.X, rule string) {
	m := <-a.pool.Pool
	err := m.AddPattern(id, rule)
	a.pool.Pool <- m
	if err != nil {
		a.logger.Error(err.Error())
	}
}

// deleteAllRulesFor removes a rule from the matcher without logging it
func (a *application) deleteAllRulesFor(id quamina.X) {
	m := <-a.pool.Pool
	err := m.DeletePatterns(id)
	a.pool.Pool <- m
	if err != nil {
		a.logger.Error(err.Error())
	}
}

func (a *application) deleteMatchingRulesFor(id quamina.X, pattern string) (int, error) {
	return -1, ErrNotImplemented
}

func (a *application) hasKey(id quamina.X) (bool, error) {
	return false, ErrNotImplemented
}

func (a *application) match(ch chan quamina.X, data string) {
	m := <-a.pool.Pool
	results, err := m.MatchesForEvent([]byte(data))
	a.pool.Pool <- m
	if err != nil {
		a.logger.Error(err.Error())
		return
	}
	for _, v := range results {
		ch <- v
	}
	close(ch)
}

// asyncDeleteAllRulesFor is a goroutine for removing keys from the local database. This should not be used
// // for replaying logs since it does not take a timestamp and is intended to run concurrently;
// // log replays should always be handled in a deterministic loop.
// Usage:
//
//	   timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), (30 * time.Second))
//		  doneChan := make(chan bool)
//		  errChan := make(chan error)
//		  go a.asyncDeleteAllRulesFor(key, doneChan, errChan)
func (a *application) asyncDeleteAllRulesFor(key string, doneChan chan bool, errChan chan error) {
	ts := time.Now().UnixNano()
	pq := <-a.pool.Pool
	err := pq.DeletePatterns(key)
	a.pool.Pool <- pq
	if err != nil {
		errChan <- err
		return
	}
	if a.config.writeWalFiles {
		go a.walFileMgr.writeWalFileEntry(ts, key, "-", wal.WAL_DEL, a.logger)
		a.lastUpdatedOn = uint64(ts)
	}

	doneChan <- true
	return
}

// asyncAddRule is a goroutine that adds a rule to the local database. This should not be used
// // for replaying logs since it does not take a timestamp and is intended to run concurrently;
// // log replays should always be handled in a deterministic loop.
// Usage:
//
//	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), (30 * time.Second))
//	doneChan := make(chan bool)
//	errChan := make(chan error)
//	go a.asyncAddRule(key, string(rule), doneChan, errChan)
//
// TODO -- this is where raft logic will go
func (a *application) asyncAddRule(key, rule string, doneChan chan bool, errChan chan error) {
	ts := time.Now().UnixNano()
	pq := <-a.pool.Pool
	err := pq.AddPattern(key, rule)
	a.pool.Pool <- pq
	if err != nil {
		errChan <- err
		return
	}
	if a.config.writeWalFiles {
		go a.walFileMgr.writeWalFileEntry(ts, key, rule, wal.WAL_ADD, a.logger)
		a.lastUpdatedOn = uint64(ts)
	}

	doneChan <- true
	return
}
