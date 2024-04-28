package mocks

import (
	"math/big"

	"github.com/0xsequence/bundler/lib/registry"
	"github.com/0xsequence/bundler/lib/registry/source"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type MockRegistrySource struct {
	mock.Mock
}

func (m *MockRegistrySource) ReputationForEndorser(endorser common.Address) (*big.Int, error) {
	args := m.Called(endorser)
	return args.Get(0).(*big.Int), args.Error(1)
}

var _ source.Interface = &MockRegistrySource{}

type MockRegistry struct {
	mock.Mock
}

func (m *MockRegistry) KnownEndorsers() []*registry.KnownEndorser {
	return m.Called().Get(0).([]*registry.KnownEndorser)
}

func (m *MockRegistry) BanEndorser(endorser common.Address, banType registry.BanType) {
	m.Called(endorser, banType)
}

func (m *MockRegistry) IsAcceptedEndorser(endorser common.Address) bool {
	return m.Called(endorser).Bool(0)
}

func (m *MockRegistry) StatusForEndorser(endorser common.Address) registry.EndorserStatus {
	return registry.EndorserStatus(m.Called(endorser).Int(0))
}

func (m *MockRegistry) TrustEndorser(endorser common.Address) {
	m.Called(endorser)
}

var _ registry.Interface = &MockRegistry{}
