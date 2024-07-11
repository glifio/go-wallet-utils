package walletutils

import (
	"encoding/hex"
	"errors"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	filcrypto "github.com/filecoin-project/go-crypto"
	"github.com/glifio/glif/v2/util"
)

var (
	ErrKeyAlreadyExists   = errors.New("key already exists")
	ErrUnsupportedKeyType = errors.New("unsupported key type")
	ErrInvalidKeyName     = errors.New("invalid key name")
)

type KeyStorageShim struct {
	ks    *keystore.KeyStore
	cache *Storage
}

var keyStore *KeyStorageShim

func KeyStore() *KeyStorageShim {
	return keyStore
}

func NewKeyStore(cfgDir string) error {
	keydir := cfgDir + "/keystore"
	cachedir := cfgDir + "/accounts.toml"

	accountStore, err := NewStorage(cachedir, map[string]string{}, true)
	if err != nil {
		return err
	}

	keyStore = &KeyStorageShim{
		keystore.NewKeyStore(
			keydir,
			keystore.StandardScryptN,
			keystore.StandardScryptP,
		),
		accountStore,
	}

	return nil
}

// will return true if the account exists in the keystore
func (store *KeyStorageShim) GetAddr(key string) (interface{}, KeyType, error) {
	addrStr, ok := store.cache.data[key]
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
		return "", KeyTypeUnknown, err
	}

	return filAddr, KeyTypeFil, nil
}

func (store *KeyStorageShim) NewAccount(name string, passphrase string, keytype KeyType) (interface{}, error) {
	if keytype == KeyTypeEth {
		return store.NewEthAccount(name, passphrase)
	} else if keytype == KeyTypeFil {
		return store.NewFilAccount(name, passphrase)
	}

	return nil, ErrUnsupportedKeyType
}

func (store *KeyStorageShim) NewReadOnlyAccount(name string, addr string) error {
	if err := ValidateKeyName(name); err != nil {
		return err
	}

	return store.cache.Set(name, addr)
}

func (store *KeyStorageShim) NewFilAccount(name string, passphrase string) (address.Address, error) {
	if err := ValidateKeyName(name); err != nil {
		return address.Undef, err
	}

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

	if err := store.SetFilAddr(name, filAddr, account.Address); err != nil {
		return address.Undef, err
	}

	return filAddr, nil
}

func (store *KeyStorageShim) NewEthAccount(name string, passphrase string) (accounts.Account, error) {
	if err := ValidateKeyName(name); err != nil {
		return accounts.Account{}, err
	}

	account, err := store.ks.NewAccount(passphrase)
	if err != nil {
		return accounts.Account{}, err
	}

	if err := store.SetEthAddr(name, account.Address); err != nil {
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
		ethAddrAssociate, err := store.cache.Get(filAddr.String())
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
	allNames := store.cache.AccountNames()
	var entries []AccountListEntry

	for _, name := range allNames {
		// if the account name is a fil address, it's a mapping to an eth addr for private key lookups, don't include it to avoid confusion
		re := regexp.MustCompile(`^[tf][0-9]`)
		if re.MatchString(name) {
			continue
		}

		addr, keytype, err := store.GetAddr(name)
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

func (store *KeyStorageShim) GetFilAddr(key string) (address.Address, error) {
	addr, err := store.cache.Get(key)
	if err != nil || addr == "" {
		return address.Address{}, err
	}
	return address.NewFromString(addr)
}

func (store *KeyStorageShim) GetEthAddr(key string) (common.Address, error) {
	addr, err := store.cache.Get(key)
	if err != nil || addr == "" {
		return common.Address{}, err
	}
	return common.HexToAddress(addr), nil
}

func (store *KeyStorageShim) SetFilAddr(key string, filAddr address.Address, evmAddr common.Address) error {
	if err := store.cache.Set(key, filAddr.String()); err != nil {
		return err
	}
	// when we set the fil address, we need to also set a mapping from f1 => 0x address to access the correct keystore file
	if err := store.cache.Set(filAddr.String(), evmAddr.Hex()); err != nil {
		return err
	}
	return nil
}

func (store *KeyStorageShim) SetEthAddr(key string, evmAddr common.Address) error {
	if err := store.cache.Set(key, evmAddr.Hex()); err != nil {
		return err
	}

	return nil
}

func (store *KeyStorageShim) Delete(name string, passphrase string) error {
	addr, keytype, err := store.GetAddr(name)
	if err != nil {
		return err
	}

	var ethAddrToRemove string
	// if the keytype is fil key, we have to delete the key's associated eth address derived from the same private key
	if keytype == KeyTypeFil {
		ethAddrToRemove, err = store.cache.Get(addr.(address.Address).String())
		if err != nil {
			return err
		}
	} else if keytype == KeyTypeEth {
		ethAddrToRemove = addr.(common.Address).Hex()
	}

	account, err := store.ks.Find(accounts.Account{Address: common.HexToAddress(ethAddrToRemove)})
	if err != nil {
		return err
	}
	// delete the account from keystore
	if err := store.ks.Delete(account, passphrase); err != nil {
		return err
	}

	// delete the account from the cache
	if err := store.cache.Delete(name); err != nil {
		return err
	}

	// if the keytype was a filecoin key, we need to delete the mapping from fil address to eth address
	if keytype == KeyTypeFil {
		if err := store.cache.Delete(addr.(address.Address).String()); err != nil {
			return err
		}
	}

	return nil
}

func (store *KeyStorageShim) Export(name string, passphrase string) ([]byte, error) {
	addr, keytype, err := store.GetAddr(name)
	if err != nil {
		return nil, err
	}

	var ethAddrToExport string
	// if the keytype is fil key, we have to export the key's associated eth address derived from the same private key
	if keytype == KeyTypeFil {
		ethAddrToExport, err = store.cache.Get(addr.(address.Address).String())
		if err != nil {
			return nil, err
		}
	} else if keytype == KeyTypeEth {
		ethAddrToExport = addr.(common.Address).Hex()
	}

	account, err := store.ks.Find(accounts.Account{Address: common.HexToAddress(ethAddrToExport)})
	if err != nil {
		return nil, err
	}

	return store.ks.Export(account, passphrase, passphrase)
}

// imports key type with passphrase into the key store and caches the address
func (store *KeyStorageShim) Import(name string, keyBytesStr string, passphrase string, keytype KeyType, encryptedJSON bool) error {
	if err := ValidateKeyName(name); err != nil {
		return err
	}

	// first check to see if we have an existing key with the same name, we won't overwrite it
	var e *ErrKeyNotFound
	_, _, err := store.GetAddr(name)
	if !errors.As(err, &e) {
		return ErrKeyAlreadyExists
	}

	var account accounts.Account
	if encryptedJSON {
		keyBytes, err := hex.DecodeString(keyBytesStr)
		if err != nil {
			return err
		}

		account, err = util.KeyStore().Import(keyBytes, passphrase, passphrase)
		if err != nil {
			return err
		}
	} else {
		pkECDSA, err := crypto.HexToECDSA(keyBytesStr)
		if err != nil {
			return err
		}

		account, err = store.ks.ImportECDSA(pkECDSA, passphrase)
		if err != nil {
			return err
		}
	}

	// if the keytype is a fil key, we need to cache the mapping from the fil address to the eth address
	if keytype == KeyTypeFil {
		keyJSON, err := store.ks.Export(account, passphrase, passphrase)
		if err != nil {
			return err
		}

		key, err := keystore.DecryptKey(keyJSON, passphrase)
		if err != nil {
			return err
		}

		filAddr, err := address.NewSecp256k1Address(filcrypto.PublicKey(key.PrivateKey.D.Bytes()))
		if err != nil {
			return err
		}
		store.cache.Set(filAddr.String(), account.Address.Hex())
		store.cache.Set(name, filAddr.String())
	} else if keytype == KeyTypeEth {
		store.cache.Set(name, account.Address.Hex())
	} else {
		return ErrUnsupportedKeyType
	}

	return nil
}

func (store *KeyStorageShim) Rename(oldName string, newName string) error {
	if err := ValidateKeyName(newName); err != nil {
		return err
	}

	addrStr, ok := store.cache.data[oldName]
	if !ok {
		return &ErrKeyNotFound{Key: oldName}
	}

	if err := store.cache.Set(newName, addrStr); err != nil {
		return err
	}

	if err := store.cache.Delete(oldName); err != nil {
		return err
	}

	return nil
}

func (store *KeyStorageShim) ChangePassphrase(name string, oldPassphrase string, newPassphrase string) error {
	addr, keytype, err := store.GetAddr(name)
	if err != nil {
		return err
	}

	var ethAddrToChange string
	// if the keytype is fil key, we have to change the passphrase of the key's associated eth address derived from the same private key
	if keytype == KeyTypeFil {
		ethAddrToChange, err = store.cache.Get(addr.(address.Address).String())
		if err != nil {
			return err
		}
	} else if keytype == KeyTypeEth {
		ethAddrToChange = addr.(common.Address).Hex()
	}

	account, err := store.ks.Find(accounts.Account{Address: common.HexToAddress(ethAddrToChange)})
	if err != nil {
		return err
	}

	return store.ks.Update(account, oldPassphrase, newPassphrase)

}

func ValidateKeyName(name string) error {
	re := regexp.MustCompile(`^[tf][0-9]`)
	if strings.HasPrefix(name, "0x") || re.MatchString(name) {
		return ErrInvalidKeyName
	}

	return nil
}
