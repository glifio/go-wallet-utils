package walletutils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransactorType int

const (
	TransactorTypeEthEOA TransactorType = iota
	TransactorTypeFilEOA
	TransactorTypeFilMSig
	TransactorTypeUnknown
)

func TransactorTypeFromString(s string) TransactorType {
	switch s {
	case "eth-secp256k1":
		return TransactorTypeEthEOA
	case "fil-secp256k1":
		return TransactorTypeFilEOA
	case "fil-msig":
		return TransactorTypeFilMSig
	default:
		return TransactorTypeUnknown
	}
}

type TransactionShim struct {
	*types.Transaction
	hash common.Hash
}

func (t *TransactionShim) Hash() common.Hash {
	return t.hash
}
