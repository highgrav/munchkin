package cluster

import (
	"google.golang.org/grpc"
)

type ConsensusMember struct {
	DialOptions []grpc.DialOption
}
