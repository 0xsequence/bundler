package mocks

import (
	"context"

	"github.com/0xsequence/bundler/lib/pricefeed"
	"github.com/stretchr/testify/mock"
)

type MockFeed struct {
	mock.Mock
}

func (m *MockFeed) Snapshot() (*pricefeed.Snapshot, error) {
	args := m.Called()
	err := args.Error(1)
	if err != nil {
		return nil, err
	}

	return args.Get(0).(*pricefeed.Snapshot), nil
}

func (m *MockFeed) Ready() bool {
	return m.Called().Bool(0)
}

func (m *MockFeed) Name() string {
	return m.Called().String(0)
}

func (m *MockFeed) Start(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

var _ pricefeed.Feed = &MockFeed{}
