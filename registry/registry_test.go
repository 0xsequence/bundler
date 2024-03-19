package registry_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/0xsequence/bundler/config"
	"github.com/0xsequence/bundler/mocks"
	"github.com/0xsequence/bundler/registry"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
)

func TestTrustEndorser(t *testing.T) {
	mockSource := new(mocks.MockRegistrySource)
	logger := httplog.NewLogger("")
	r, err := registry.NewRegistry(&config.RegistryConfig{}, logger, nil, nil)
	assert.NoError(t, err)

	endorser := common.HexToAddress("0x08FFc248A190E700421C0aFB4135768406dCebfF")

	r.AddSource(mockSource, 1)
	r.TrustEndorser(endorser)

	assert.True(t, r.IsAcceptedEndorser(endorser))

	r, err = registry.NewRegistry(&config.RegistryConfig{
		Trusted: []string{endorser.String()},
	}, logger, nil, nil)

	assert.NoError(t, err)
	assert.True(t, r.IsAcceptedEndorser(endorser))
}

func TestUnknownEndorser(t *testing.T) {
	mockSource := new(mocks.MockRegistrySource)
	logger := httplog.NewLogger("")
	r, err := registry.NewRegistry(&config.RegistryConfig{
		AllowUnusable: true,
		MinReputation: 10,
	}, logger, nil, nil)
	assert.NoError(t, err)

	endorser := common.HexToAddress("0x08FFc248A190E700421C0aFB4135768406dCebfF")
	r.AddSource(mockSource, 1)

	mockSource.On("ReputationForEndorser", endorser).Return(big.NewInt(0), nil).Once()

	assert.False(t, r.IsAcceptedEndorser(endorser))
}

func TestAcceptEndorser(t *testing.T) {
	mockSource := new(mocks.MockRegistrySource)
	logger := httplog.NewLogger("")
	r, err := registry.NewRegistry(&config.RegistryConfig{
		AllowUnusable: true,
		MinReputation: 10,
	}, logger, nil, nil)
	assert.NoError(t, err)

	endorser := common.HexToAddress("0x08FFc248A190E700421C0aFB4135768406dCebfF")
	r.AddSource(mockSource, 1)

	mockSource.On("ReputationForEndorser", endorser).Return(big.NewInt(11), nil).Once()

	assert.True(t, r.IsAcceptedEndorser(endorser))
}

func TestTempBanEndorser(t *testing.T) {
	mockSource := new(mocks.MockRegistrySource)
	logger := httplog.NewLogger("")
	r, err := registry.NewRegistry(&config.RegistryConfig{
		AllowUnusable:  true,
		MinReputation:  10,
		TempBanSeconds: 1,
	}, logger, nil, nil)
	assert.NoError(t, err)

	endorser := common.HexToAddress("0x08FFc248A190E700421C0aFB4135768406dCebfF")
	r.AddSource(mockSource, 1)

	mockSource.On("ReputationForEndorser", endorser).Return(big.NewInt(11), nil).Twice()
	assert.True(t, r.IsAcceptedEndorser(endorser))

	r.BanEndorser(endorser, registry.TemporaryBan)

	assert.False(t, r.IsAcceptedEndorser(endorser))

	// Wait a second
	time.Sleep(1*time.Second + 1*time.Millisecond)

	assert.True(t, r.IsAcceptedEndorser(endorser))
}

func TestPermanentBanEndorser(t *testing.T) {
	mockSource := new(mocks.MockRegistrySource)
	logger := httplog.NewLogger("")
	r, err := registry.NewRegistry(&config.RegistryConfig{
		AllowUnusable:  true,
		MinReputation:  10,
		TempBanSeconds: 1,
	}, logger, nil, nil)
	assert.NoError(t, err)

	endorser := common.HexToAddress("0x08FFc248A190E700421C0aFB4135768406dCebfF")
	r.AddSource(mockSource, 1)

	mockSource.On("ReputationForEndorser", endorser).Return(big.NewInt(11), nil).Twice()
	assert.True(t, r.IsAcceptedEndorser(endorser))

	r.BanEndorser(endorser, registry.PermanentBan)

	assert.False(t, r.IsAcceptedEndorser(endorser))

	// Wait a second
	time.Sleep(1*time.Second + 1*time.Millisecond)

	assert.False(t, r.IsAcceptedEndorser(endorser))
}

func TestCacheAcceptEndorser(t *testing.T) {
	mockSource := new(mocks.MockRegistrySource)
	logger := httplog.NewLogger("")
	r, err := registry.NewRegistry(&config.RegistryConfig{
		AllowUnusable: true,
		MinReputation: 10,
	}, logger, nil, nil)
	assert.NoError(t, err)

	endorser := common.HexToAddress("0x08FFc248A190E700421C0aFB4135768406dCebfF")
	r.AddSource(mockSource, 1)

	mockSource.On("ReputationForEndorser", endorser).Return(big.NewInt(11), nil).Once()

	assert.True(t, r.IsAcceptedEndorser(endorser))
	assert.True(t, r.IsAcceptedEndorser(endorser))
	mockSource.AssertNumberOfCalls(t, "ReputationForEndorser", 1)
}
