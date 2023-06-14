package main

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

func handleGetHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(`{"ok":true,"data":{"ping":"pong"}}`))
}

func (a *application) handleHttpPostMatch(w http.ResponseWriter, r *http.Request) {

	type responseModel struct {
		Ok      bool      `json:"ok"`
		Matches *[]string `json:"matches"`
	}

	if r.Method != "POST" {
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Incorrect method (POST only)"],"data":{}}`))
		return
	}

	rule, err := io.ReadAll(r.Body)
	if err != nil {
		a.logger.Warn(err.Error(),
			zap.String("ip", r.RemoteAddr))
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Problem reading request body"],"data":{}}`))
		return
	}

	if len(rule) == 0 {
		a.logger.Warn(err.Error(),
			zap.String("ip", r.RemoteAddr))
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Problem reading request body"],"data":{}}`))
		return
	}

	pq := <-a.pool.Pool
	matches, err := pq.MatchesForEvent([]byte(rule))
	a.pool.Pool <- pq
	if err != nil {
		a.logger.Warn(err.Error(),
			zap.String("ip", r.RemoteAddr))
		w.WriteHeader(500)
		w.Write([]byte(`{"ok":false,"errors":["Problem matching pattern"],"data":{}}`))
		return
	}
	resp := responseModel{
		Ok: true,
	}

	matchList := make([]string, 0)
	for _, v := range matches {
		s := v.(string)
		if len(s) > 0 {
			matchList = append(matchList, s)
		}
	}
	resp.Matches = &matchList
	val, err := json.Marshal(resp)
	if err != nil {
		a.logger.Error(err.Error(),
			zap.String("ip", r.RemoteAddr))
		w.WriteHeader(500)
		w.Write([]byte(`{"ok":false,"errors":["Problem returning results"],"data":{}}`))
		return
	}
	w.WriteHeader(200)
	w.Write(val)
}

func (a *application) handleHttpPostAddRule(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Incorrect method (POST only)"],"data":{}}`))
		return
	}

	qs := r.URL.Query()
	if !qs.Has("key") {
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Missing 'key' in query string'"], "data":{}}`))
		return
	}
	key := qs.Get("key")

	rule, err := io.ReadAll(r.Body)
	if err != nil {
		a.logger.Warn(err.Error(),
			zap.String("ip", r.RemoteAddr))
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Problem reading request body"],"data":{}}`))
		return
	}

	if len(rule) == 0 {
		a.logger.Warn(err.Error(),
			zap.String("ip", r.RemoteAddr))
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Problem reading request body"],"data":{}}`))
		return
	}

	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), (30 * time.Second))
	doneChan := make(chan bool)
	errChan := make(chan error)
	go a.asyncAddRule(key, string(rule), doneChan, errChan)
	select {
	case <-timeoutCtx.Done():
		w.WriteHeader(202)
		w.Write([]byte(`{"ok":true,"errors":["Request timed out, submitted for processing"],"data":{}}`))
		return
	case <-doneChan:
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true,"data":{}`))
		cancelFunc()
		return
	case err = <-errChan:
		a.logger.Error(err.Error())
		w.WriteHeader(500)
		w.Write([]byte(`{"ok":false,"errors":["Problem adding pattern"],"data":{}}`))
		cancelFunc()
		return
	}

}

func (a *application) handleHttpDeleteByKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Incorrect method (POST only)"],"data":{}}`))
		return
	}

	qs := r.URL.Query()
	if !qs.Has("key") {
		w.WriteHeader(400)
		w.Write([]byte(`{"ok":false,"errors":["Missing 'key' in query string'"], "data":{}}`))
		return
	}
	key := qs.Get("key")

	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), (30 * time.Second))

	doneChan := make(chan bool)
	errChan := make(chan error)
	go a.asyncDeleteAllRulesFor(key, doneChan, errChan)
	select {
	case <-timeoutCtx.Done():
		w.WriteHeader(202)
		w.Write([]byte(`{"ok":true,"errors":["Request timed out, submitted for processing"],"data":{}}`))
		return
	case <-doneChan:
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true,"data":{}`))
		cancelFunc()
		return
	case err := <-errChan:
		a.logger.Error(err.Error(),
			zap.String("ip", r.RemoteAddr))
		w.WriteHeader(500)
		w.Write([]byte(`{"ok":false,"errors":["Problem deleting pattern"],"data":{}}`))
		cancelFunc()
		return
	}
}
