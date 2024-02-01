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
	// peerID := s.Node.PeerID().String()

	// addrs := s.Node.Addrs()
	// statusAddrs := make([]string, len(addrs))
	// for i := range addrs {
	// 	statusAddrs[i] = addrs[i].String()
	// }

	status := &proto.Status{
		HealthOK:   true,
		StartTime:  s.startTime,
		Uptime:     uint64(time.Now().UTC().Sub(s.startTime).Seconds()),
		Ver:        bundler.VERSION,
		Branch:     bundler.GITBRANCH,
		CommitHash: bundler.GITCOMMIT,

		// PeerID:       peerID,
		// Multiaddrs:   statusAddrs,
		// JoinedTopics: []string{}, // TODO..
		// Peers:        statusPeers,
	}
	return status, nil
}

func (s *RPC) Peers(ctx context.Context) ([]string, error) {
	return nil, nil
	// peers := s.Node.Peers()
	// statusPeers := make([]string, len(peers))
	// for i := range peers {
	// 	statusPeers[i] = peers[i].String()
	// }
	// return statusPeers, nil
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
	peers, err := s.Peers(r.Context())
	if err != nil {
		s.renderJSON(w, r, err.Error(), 500)
		return
	}
	s.renderJSON(w, r, peers, 200)
}
