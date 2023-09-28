package accounts

import (
	"reflect"

	ethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
)

type Account struct {
	EthAccount ethaccounts.Account
	FilAddress address.Address
}

func (a Account) IsEth() bool {
	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := a.EthAccount.Address.Bytes()
	return !reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

func (a Account) IsFil() bool {
	return !a.FilAddress.Empty()
}

func (a Account) String() string {
	if a.IsEth() {
		return a.EthAccount.Address.String()
	}
	if a.IsFil() {
		return a.FilAddress.String()
	}
	return ""
}
