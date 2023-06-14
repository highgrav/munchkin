package main

import (
	"context"
	api "github.com/highgrav/munchkin/api/v1"
	"github.com/highgrav/munchkin/internal/wal"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type walConfig struct {
	walDir    string
	walPrefix string
}

type walFileList []string

func (fl *walFileList) getFile(timestamp uint64) {

}

type walServer struct {
	api.UnimplementedWalServer
	app        *application
	Config     *walConfig
	mu         sync.Mutex
	dirName    string
	filePrefix string
	server     *grpc.Server
}

var _ api.WalServer = (*walServer)(nil)

func newWalServer(config *walConfig) (*walServer, error) {
	svr := grpc.NewServer()
	ws := &walServer{
		Config: config,
		server: svr,
	}
	api.RegisterWalServer(svr, ws)
	return ws, nil
}

// PublishEntryStreamFromTime takes a timestamp and returns all WalEntries with a timestamp equal to or greater than
// the requested timestamp.
// TODO -- there's a race condition with file rotation and possible file writes.
// // I need to think through how to implement handling buffering writes in a clean CSP manner.
func (ws *walServer) PublishEntryStreamFromTime(req *api.PublishEntryRequest, server api.Wal_PublishEntryStreamServer) error {
	var doneWithFiles bool = false
	startFiles, err := wal.FindFilesOnOrAfter(ws.dirName, ws.filePrefix, req.GetTimestamp())
	if err != nil {
		return err
	}
	if len(startFiles) < 1 {
		return nil
	}
	for _, f := range startFiles {
		if doneWithFiles {
			break
		}
		file, err := wal.OpenWalFile(f)
		if err != nil {
			return err
		}
		for !doneWithFiles && file.HasNext() {
			evt, err := file.Next()
			if evt.Timestamp > req.GetTimestamp() {
				file.Close()
				doneWithFiles = true
			}
			if err != nil {
				return err
			}
			res := &api.PublishEntryResponse{
				FileName: file.FileName,
				Entry: &api.WalEntry{
					Timestamp: evt.Timestamp,
					Key:       evt.Key,
					Pattern:   evt.Pattern,
					Action:    uint32(evt.Action),
				},
			}
			server.Send(res)
		}
		file.Close()
	}
	if len(ws.app.cluster.replMemLog) > 0 {
		for _, evt := range ws.app.cluster.replMemLog {
			res := &api.PublishEntryResponse{
				FileName: "",
				Entry: &api.WalEntry{
					Timestamp: evt.Timestamp,
					Key:       evt.Key,
					Pattern:   evt.Pattern,
					Action:    uint32(evt.Action),
				},
			}
			server.Send(res)
		}
	}
	return nil
}

// LogEntry takes a WalEntry and processes it locally, updating the matcher and writing it to the
// local logfiles.
func (ws *walServer) LogEntry(ctx context.Context, req *api.LogEntryRequest) (*api.LogEntryResponse, error) {
	// Don't apply if you've applied later entries already
	if req.Entry.GetTimestamp() < app.lastUpdatedOn {
		return &api.LogEntryResponse{
			Timestamp: 0,
		}, nil
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	// TODO -- using the object pool may introduce a potential race condition
	ts := time.Now().UnixNano()
	if req.GetEntry().GetAction() == uint32(wal.WAL_ADD) {
		ws.app.addRule(req.GetEntry().GetKey(), string(req.GetEntry().GetPattern()))
	} else if req.GetEntry().GetAction() == uint32(wal.WAL_DEL) {
		ws.app.deleteAllRulesFor(req.GetEntry().GetKey())
	}

	go ws.asyncLogEntryToFile(req)

	return &api.LogEntryResponse{
		Timestamp: uint64(ts),
	}, nil
}

func (ws *walServer) LogEntryStream(server api.Wal_LogEntryStreamServer) error {
	for {
		streamReq, err := server.Recv()
		if err != nil {
			return err
		}
		res, err := ws.LogEntry(server.Context(), streamReq)
		if err != nil {
			return err
		}
		if err = server.Send(res); err != nil {
			return err
		}
	}
	return nil
}

func (ws *walServer) asyncLogEntryToFile(evt *api.LogEntryRequest) {
	// TODO
}
