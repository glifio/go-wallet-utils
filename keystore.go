package walletutils

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	filcrypto "github.com/filecoin-project/go-crypto"
)

type KeyStorageShim struct {
	*keystore.KeyStore
	*AccountsStorage
}

var keyStore *KeyStorageShim

func KeyStore() *KeyStorageShim {
	return keyStore
}

func NewKeyStore(keydir string, accountsStore *AccountsStorage) {
	keyStore = &KeyStorageShim{
		keystore.NewKeyStore(
			keydir,
			keystore.StandardScryptN,
			keystore.StandardScryptP,
		),
		accountsStore,
	}
}

func (ks *KeyStorageShim) NewFilAccount(passphrase string) (address.Address, accounts.Account, error) {
	account, err := ks.NewAccount(passphrase)
	if err != nil {
		return address.Undef, accounts.Account{}, err
	}

	// we now have to unlock the account to recover the raw private key, and derive the filecoin address from it
	keyJSON, err := ks.Export(account, passphrase, "")
	if err != nil {
		return address.Undef, accounts.Account{}, err
	}

	key, err := keystore.DecryptKey(keyJSON, "")
	if err != nil {
		return address.Undef, accounts.Account{}, err
	}

	filAddr, err := address.NewSecp256k1Address(filcrypto.PublicKey(key.PrivateKey.D.Bytes()))
	if err != nil {
		return address.Undef, accounts.Account{}, err
	}

	return filAddr, account, nil
}

func (ks *KeyStorageShim) NewEthAccount(passphrase string) (accounts.Account, error) {
	return ks.NewAccount(passphrase)
}

func (ks *KeyStorageShim) HasKeyForAddress(addr interface{}, keytype KeyType) bool {
	if keytype != KeyTypeEth && keytype != KeyTypeFil {
		return false
	}

	var ethAddrToLookup common.Address

	if keytype == KeyTypeEth {
		ethAddrToLookup = addr.(common.Address)
	} else {
		// here we have to lookup the associated eth addr for this fil addr, and check if we have a key for it
		filAddr := addr.(address.Address)
		ethAddrAssociate, err := ks.Get(filAddr.String())
		if err != nil {
			return false
		}

		ethAddrToLookup = common.HexToAddress(ethAddrAssociate)
	}

	return ks.HasAddress(ethAddrToLookup)

}
