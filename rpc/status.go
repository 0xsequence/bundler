package rpc

import (
	"context"
	"net/http"
	"time"

	"github.com/0xsequence/bundler"
	"github.com/0xsequence/bundler/proto"
)

func (s *RPC) Version(ctx context.Context) (*proto.Version, error) {
	return &proto.Version{
		WebrpcVersion: proto.WebRPCVersion(),
		SchemaVersion: proto.WebRPCSchemaVersion(),
		SchemaHash:    proto.WebRPCSchemaHash(),
		NodeVersion:   bundler.GITCOMMIT,
	}, nil
}

func (s *RPC) Status(ctx context.Context) (*proto.Status, error) {
	hostID := s.Host.HostID().String()

	addrs := s.Host.Addrs()
	hostAddrs := make([]string, len(addrs))
	for i := range addrs {
		hostAddrs[i] = addrs[i].String()
	}

	priorityPeers := s.Host.PriorityPeers()
	statusPeers := make([]string, len(priorityPeers))
	for i := range priorityPeers {
		statusPeers[i] = priorityPeers[i].String()
	}

	status := &proto.Status{
		HealthOK:   true,
		StartTime:  s.startTime,
		Uptime:     uint64(time.Now().UTC().Sub(s.startTime).Seconds()),
		Ver:        bundler.VERSION,
		Branch:     bundler.GITBRANCH,
		CommitHash: bundler.GITCOMMIT,

		HostID:        hostID,
		HostAddrs:     hostAddrs,
		PriorityPeers: statusPeers,
	}
	return status, nil
}

func (s *RPC) Peers(ctx context.Context) ([]string, []string, error) {
	peers := s.Host.Peers()
	statusPeers := make([]string, len(peers))
	for i := range peers {
		statusPeers[i] = peers[i].String()
	}

	priorityPeers := s.Host.PriorityPeers()
	statusPriorityPeers := make([]string, len(priorityPeers))
	for i := range priorityPeers {
		statusPriorityPeers[i] = priorityPeers[i].String()
	}

	return statusPeers, statusPriorityPeers, nil
}

func (s *RPC) statusPage(w http.ResponseWriter, r *http.Request) {
	status, err := s.Status(r.Context())
	if err != nil {
		s.renderJSON(w, r, err.Error(), 500)
		return
	}
	s.renderJSON(w, r, status, 200)
}

func (s *RPC) peersPage(w http.ResponseWriter, r *http.Request) {
	peers, priorityPeers, err := s.Peers(r.Context())
	if err != nil {
		s.renderJSON(w, r, err.Error(), 500)
		return
	}

	result := struct {
		Peers         []string `json:"peers"`
		PriorityPeers []string `json:"priorityPeers"`
	}{peers, priorityPeers}

	s.renderJSON(w, r, result, 200)
}
