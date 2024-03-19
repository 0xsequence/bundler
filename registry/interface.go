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

func (e EndorserStatus) String() string {
	switch e {
	case UnknownEndorser:
		return "unknown"
	case AcceptedEndorser:
		return "accepted"
	case TrustedEndorser:
		return "trusted"
	case TemporaryBanned:
		return "temporary-banned"
	case PermanentBanned:
		return "permanent-banned"
	default:
		return "unknown"
	}
}

type BanType int

const (
	TemporaryBan BanType = BanType(TemporaryBanned)
	PermanentBan BanType = BanType(PermanentBanned)
)

type KnownEndorser struct {
	Address common.Address
	Status  EndorserStatus
}

type Interface interface {
	KnownEndorsers() []*KnownEndorser
	TrustEndorser(endorser common.Address)
	IsAcceptedEndorser(endorser common.Address) bool
	StatusForEndorser(endorser common.Address) EndorserStatus
	BanEndorser(endorser common.Address, banType BanType)
}
