package mocks

import (
	"context"

	"github.com/0xsequence/bundler/p2p"
	"github.com/stretchr/testify/mock"
)

type MockP2p struct {
	mock.Mock
}

// Address implements p2p.Interface.
func (m *MockP2p) Address() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// Broadcast implements p2p.Interface.
func (m *MockP2p) Broadcast(ctx context.Context, topic p2p.PubsubTopic, payload interface{}) error {
	args := m.Called(ctx, topic, payload)
	return args.Error(0)
}

// BroadcastData implements p2p.Interface.
func (m *MockP2p) BroadcastData(ctx context.Context, topic p2p.PubsubTopic, payload []byte) error {
	args := m.Called(ctx, topic, payload)
	return args.Error(0)
}

// HandleTopic implements p2p.Interface.
func (m *MockP2p) HandleTopic(ctx context.Context, topic p2p.PubsubTopic, handler p2p.MsgHandler) error {
	args := m.Called(ctx, topic, handler)
	return args.Error(0)
}

// Sign implements p2p.Interface.
func (m *MockP2p) Sign(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

var _ p2p.Interface = &MockP2p{}
