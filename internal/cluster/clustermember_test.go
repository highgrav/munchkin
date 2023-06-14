package cluster

import (
	"fmt"
	"github.com/cmertens/munchkin/internal/net"
	"go.uber.org/zap"
	"log"
	"strconv"
	"testing"
)

var zaplog *zap.Logger
var memberCount int = 5

type testClusterHandler struct {
	chJoin  chan map[string]string
	chLeave chan string
}

func (h *testClusterHandler) Join(name, addr string) error {
	if h.chJoin != nil {
		h.chJoin <- map[string]string{
			"id":   name,
			"addr": addr,
		}
	}
	return nil
}

func (h *testClusterHandler) Leave(name, addr string) error {
	if h.chLeave != nil {
		h.chLeave <- name
	}
	return nil
}

func (h *testClusterHandler) Fail(name, addr string) error {
	return nil
}

func (h *testClusterHandler) Reap(name, addr string) error {
	return nil
}

func (h *testClusterHandler) Update() error {
	return nil
}

func TestClusterMembership(t *testing.T) {
	zaplog = zap.NewExample()
	var mems []*ClusterMember
	for x := 0; x < memberCount; x++ {
		mems, _ = setUpClusterMember(t, nil)
	}
	mems[0].Leave()
}

// setUpClusterMember adds a new ClusterMember to an array of ClusterMembers
func setUpClusterMember(t *testing.T, members []*ClusterMember) ([]*ClusterMember, SerfEventHandler) {
	var clmbr *ClusterMember = nil
	var err error
	// Get a unique ID for this cluster based on total number of cluster members
	// (This works because it's a single-threaded, single-process test scenario)
	id := len(members)

	addrs, _ := net.GetFreePorts(1, 5000, 0)
	if len(addrs) < 1 {
		log.Fatal("No free ports available for cluster!")
	}
	addr := fmt.Sprintf("127.0.0.1:%d", addrs[0])
	fmt.Println("Binding to " + addr)
	tags := map[string]string{"rpc_addr": addr}
	handler := &testClusterHandler{}

	if (len(members)) == 0 {
		handler.chJoin = make(chan map[string]string, memberCount)
		handler.chLeave = make(chan string, memberCount)
		clmbr, err = NewClusterMember(strconv.Itoa(id), addr, tags, []string{}, zaplog)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		mems := []string{
			members[0].Address,
		}
		clmbr, err = NewClusterMember(strconv.Itoa(id), addr, tags, mems, zaplog)
		if err != nil {
			log.Fatal(err)
		}
	}
	members = append(members, clmbr)
	return members, handler
}
