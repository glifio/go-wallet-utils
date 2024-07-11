package walletutils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	filcrypto "github.com/filecoin-project/go-crypto"
)

var (
	ErrKeyAlreadyExists   = errors.New("key already exists")
	ErrUnsupportedKeyType = errors.New("unsupported key type")
	ErrInvalidKeyName     = errors.New("invalid key name")
)

type KeyStorageShim struct {
	ks *keystore.KeyStore
	as *AccountsStorage
}

var keyStore *KeyStorageShim

func KeyStore() *KeyStorageShim {
	return keyStore
}

func NewKeyStore(cfgDir string) error {
	keydir := cfgDir + "/keystore"
	accountsStore, err := NewAccountsStore(cfgDir + "/accounts.toml")
	fmt.Println(keydir, cfgDir+"/accounts.json")
	if err != nil {
		return err
	}
	keyStore = &KeyStorageShim{
		keystore.NewKeyStore(
			keydir,
			keystore.StandardScryptN,
			keystore.StandardScryptP,
		),
		accountsStore,
	}

	return nil
}

func (store *KeyStorageShim) NewAccount(name string, passphrase string, keytype KeyType) (interface{}, error) {
	re := regexp.MustCompile(`^[tf][0-9]`)
	if strings.HasPrefix(name, "0x") || re.MatchString(name) {
		return nil, ErrKeyAlreadyExists
	}

	if keytype == KeyTypeEth {
		return store.NewEthAccount(name, passphrase)
	} else if keytype == KeyTypeFil {
		return store.NewFilAccount(name, passphrase)
	}

	return nil, ErrUnsupportedKeyType
}

func (store *KeyStorageShim) NewReadOnlyAccount(name string, addr string) error {
	re := regexp.MustCompile(`^[tf][0-9]`)
	if strings.HasPrefix(name, "0x") || re.MatchString(name) {
		return ErrInvalidKeyName
	}

	return store.as.Set(name, addr)
}

func (store *KeyStorageShim) NewFilAccount(name string, passphrase string) (address.Address, error) {
	account, err := store.ks.NewAccount(passphrase)
	if err != nil {
		return address.Undef, err
	}

	// we now have to unlock the account to recover the raw private key, and derive the filecoin address from it
	keyJSON, err := store.ks.Export(account, passphrase, "")
	if err != nil {
		return address.Undef, err
	}

	key, err := keystore.DecryptKey(keyJSON, "")
	if err != nil {
		return address.Undef, err
	}

	filAddr, err := address.NewSecp256k1Address(filcrypto.PublicKey(key.PrivateKey.D.Bytes()))
	if err != nil {
		return address.Undef, err
	}

	if err := store.as.SetFilAddr(name, filAddr, account.Address); err != nil {
		return address.Undef, err
	}

	return filAddr, nil
}

func (store *KeyStorageShim) NewEthAccount(name string, passphrase string) (accounts.Account, error) {
	account, err := store.ks.NewAccount(passphrase)
	if err != nil {
		return accounts.Account{}, err
	}

	if err := store.as.SetEthAddr(name, account.Address); err != nil {
		return accounts.Account{}, err
	}

	return account, nil
}

func (store *KeyStorageShim) HasKeyForAddress(addr interface{}, keytype KeyType) bool {
	if keytype != KeyTypeEth && keytype != KeyTypeFil {
		return false
	}

	var ethAddrToLookup common.Address

	if keytype == KeyTypeEth {
		ethAddrToLookup = addr.(common.Address)
	} else {
		// here we have to lookup the associated eth addr for this fil addr, and check if we have a key for it
		filAddr := addr.(address.Address)
		ethAddrAssociate, err := store.as.Get(filAddr.String())
		if err != nil {
			return false
		}

		ethAddrToLookup = common.HexToAddress(ethAddrAssociate)
	}

	return store.ks.HasAddress(ethAddrToLookup)
}

type AccountListEntry struct {
	Name    string
	KeyType KeyType
	Addr    interface{}
}

func (store *KeyStorageShim) List(includeReadOnly bool) ([]AccountListEntry, error) {
	allNames := store.as.AccountNames()
	var entries []AccountListEntry

	for _, name := range allNames {
		// if the account name is a fil address, it's a mapping to an eth addr for private key lookups, don't include it to avoid confusion
		re := regexp.MustCompile(`^[tf][0-9]`)
		if re.MatchString(name) {
			continue
		}

		addr, keytype, err := store.as.GetAddr(name)
		if err != nil {
			return nil, err
		}

		if store.HasKeyForAddress(addr, keytype) || includeReadOnly {
			entries = append(entries, AccountListEntry{
				Name:    name,
				KeyType: keytype,
				Addr:    addr,
			})
		}
	}

	return entries, nil
}
