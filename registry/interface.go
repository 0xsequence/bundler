package registry

import "github.com/0xsequence/ethkit/go-ethereum/common"

type EndorserStatus int

const (
	UnknownEndorser EndorserStatus = iota
	AcceptedEndorser
	TrustedEndorser

	TemporaryBanned
	PermanentBanned
)

type BanType int

const (
	TemporaryBan BanType = BanType(TemporaryBanned)
	PermanentBan BanType = BanType(PermanentBanned)
)

type Interface interface {
	TrustEndorser(endorser common.Address)
	IsAcceptedEndorser(endorser common.Address) bool
	StatusForEndorser(endorser common.Address) EndorserStatus
	BanEndorser(endorser common.Address, banType BanType)
}
