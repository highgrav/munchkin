package cluster

import (
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
	"net"
	"os"
)

type SerfEventHandler interface {
	Join(name, addr string) error
	Leave(name, addr string) error
	Fail(name, addr string) error
	Reap(name, addr string) error
	Update() error
}

type ClusterMember struct {
	NodeName string
	Tags     map[string]string
	Peers    []string
	Address  string
	serf     *serf.Serf
	events   chan serf.Event
	logger   *zap.Logger
	handler  SerfEventHandler
}

func (m *ClusterMember) Members() []serf.Member {
	return m.serf.Members()
}

func (m *ClusterMember) Leave() error {
	return m.serf.Leave()
}

func (m *ClusterMember) isLocal(member serf.Member) bool {
	return m.serf.LocalMember().Name == member.Name
}

func (m *ClusterMember) handleMemberJoined(member serf.Member) {
	if err := m.handler.Join(member.Name, member.Tags["rpc-addr"]); err != nil {
		m.logger.Error("serf: join failed", zap.Error(err), zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
	}
}

func (m *ClusterMember) handleMemberLeft(member serf.Member) {
	if err := m.handler.Leave(member.Name, member.Tags["rpc-addr"]); err != nil {
		m.logger.Error("serf: leave failed", zap.Error(err), zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
	}
}

func (m *ClusterMember) handleMemberFailed(member serf.Member) {

}

func (m *ClusterMember) handleMemberReaped(member serf.Member) {

}

func (m *ClusterMember) handleMemberUpdated(member serf.Member) {

}

func (m *ClusterMember) eventHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, mem := range e.(serf.MemberEvent).Members {
				if m.isLocal(mem) {
					continue
				}
				m.handleMemberJoined(mem)
			}
		case serf.EventMemberFailed:
			for _, mem := range e.(serf.MemberEvent).Members {
				if m.isLocal(mem) {
					continue
				}
				m.handleMemberFailed(mem)
			}
		case serf.EventMemberLeave:
			for _, mem := range e.(serf.MemberEvent).Members {
				if m.isLocal(mem) {
					continue
				}
				m.handleMemberLeft(mem)
			}
		case serf.EventMemberReap:
			// Serf reaps members that have timed out and exceeded the
			// recovery duration.
			// TODO
		case serf.EventMemberUpdate:
			// TODO
		case serf.EventQuery:
			// TODO
		case serf.EventUser:
			// TODO
		}
	}
}

func NewClusterMember(nodeName, addr string, tags map[string]string, addrsToJoin []string, logger *zap.Logger) (*ClusterMember, error) {
	m := &ClusterMember{
		NodeName: nodeName,
		Address:  addr,
		Tags:     tags,
		Peers:    addrsToJoin,
		logger:   logger,
	}
	if err := m.setUpSerf(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *ClusterMember) setUpSerf() error {
	addr, err := net.ResolveTCPAddr("tcp", m.Address)
	if err != nil {
		return err
	}

	memcfg := memberlist.DefaultLANConfig()
	memcfg.BindAddr = addr.IP.String()
	memcfg.BindPort = addr.Port
	memcfg.AdvertiseAddr = addr.IP.String()
	memcfg.AdvertisePort = addr.Port
	memcfg.LogOutput = os.Stderr // TODO -- implement an io.Writer that writes to zap

	cfg := serf.DefaultConfig()

	cfg.Init()

	cfg.Tags = m.Tags
	cfg.NodeName = m.NodeName
	cfg.MemberlistConfig = memcfg
	cfg.LogOutput = os.Stderr // TODO -- implement an io.Writer that writes to zap
	m.events = make(chan serf.Event)
	cfg.EventCh = m.events
	m.serf, err = serf.Create(cfg)
	if err != nil {
		return err
	}

	// Set up handler for Serf events
	go m.eventHandler()

	// Join cluster
	if m.Peers != nil {
		_, err = m.serf.Join(m.Peers, true)
		if err != nil {
			return err
		}
	}
	return nil
}
