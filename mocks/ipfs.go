package mocks

import (
	"github.com/0xsequence/bundler/ipfs"
	"github.com/stretchr/testify/mock"
)

type MockIpfs struct {
	mock.Mock
}

func (m *MockIpfs) Report(data []byte) (string, error) {
	args := m.Called(data)
	return args.String(0), args.Error(1)
}

var _ ipfs.Interface = &MockIpfs{}
