package walletutils

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
)

type AccountsStorage struct {
	*Storage
}

var accountsStore *AccountsStorage

func AccountsStore() *AccountsStorage {
	return accountsStore
}

func NewAccountsStore(filename string) error {
	accountsDefault := map[string]string{}

	s, err := NewStorage(filename, accountsDefault, true)
	if err != nil {
		return err
	}

	accountsStore = &AccountsStorage{s}

	return nil
}

func (a *AccountsStorage) Exists(key string) bool {
	if _, ok := a.data[key]; !ok {
		return false
	}
	return true
}

func (a *AccountsStorage) GetAddr(key string) (interface{}, KeyType, error) {
	addrStr, ok := a.data[key]
	if !ok {
		return "", KeyTypeUnknown, &ErrKeyNotFound{Key: key}
	}

	// check if the key is an eth address
	if strings.HasPrefix(addrStr, "0x") {
		return common.HexToAddress(addrStr), KeyTypeEth, nil
	}

	// check if the key is a filecoin address
	filAddr, err := address.NewFromString(addrStr)
	if err != nil {
		return "", KeyTypeUnknown, nil
	}

	return filAddr, KeyTypeFil, nil
}

func (a *AccountsStorage) GetFilAddr(key string) (address.Address, error) {
	addr, err := a.Get(key)
	if err != nil || addr == "" {
		return address.Address{}, err
	}
	return address.NewFromString(addr)
}

func (a *AccountsStorage) GetEthAddr(key string) (common.Address, error) {
	addr, err := a.Get(key)
	if err != nil || addr == "" {
		return common.Address{}, err
	}
	return common.HexToAddress(addr), nil
}

func (a *AccountsStorage) SetFilAddr(key string, filAddr address.Address, evmAddr common.Address) error {
	if err := a.Set(key, filAddr.String()); err != nil {
		return err
	}
	// when we set the fil address, we need to also set a mapping from f1 => 0x address to access the correct keystore file
	if err := a.Set(filAddr.String(), evmAddr.Hex()); err != nil {
		return err
	}
	return nil
}

func (a *AccountsStorage) SetEthAddr(key string, evmAddr common.Address) error {
	if err := a.Set(key, evmAddr.Hex()); err != nil {
		return err
	}

	return nil
}
