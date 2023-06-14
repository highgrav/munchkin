package main

import (
	"github.com/highgrav/munchkin/internal/wal"
)

type ClusterState struct {
	replMemLog []wal.WalEntry
}
