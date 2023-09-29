package filkeystore

import (
	"fmt"
	"math/big"

	ethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	filcrypto "github.com/filecoin-project/go-crypto"
	"github.com/glifio/go-wallet-utils/accounts"
)

type KeystoreWrapper struct {
	Keystore *keystore.KeyStore
}

// Wallets implements accounts.Backend, returning all the currently tracked USB
// devices that appear to be hardware wallets.
func (ks KeystoreWrapper) Wallets() []accounts.Wallet {
	wallets := make([]accounts.Wallet, 0)
	for _, w := range ks.Keystore.Wallets() {
		wallets = append(wallets, EthWalletWrapper{Wallet: w, ks: ks.Keystore})
	}
	return wallets
}

type EthWalletWrapper struct {
	ethaccounts.Wallet
	ks *keystore.KeyStore
}

// Accounts returns all key files present in the directory.
func (ew EthWalletWrapper) Accounts() []accounts.Account {
	ethAccounts := ew.Wallet.Accounts()
	accts := make([]accounts.Account, 0)
	for _, acct := range ethAccounts {
		keyJSON, err := ew.ks.Export(acct, "", "")
		if err != nil {
			continue
		}
		pk, err := keystore.DecryptKey(keyJSON, "")
		if err != nil {
			continue
		}
		privateKeyBytes := crypto.FromECDSA(pk.PrivateKey)
		publicKey := filcrypto.PublicKey(privateKeyBytes)
		filAddr, err := address.NewSecp256k1Address(publicKey)
		if err != nil {
			continue
		}
		accts = append(accts, accounts.Account{FilAddress: filAddr})
	}
	return accts
}

// SignTxWithPassphrase signs the transaction
func (ew EthWalletWrapper) SignTxWithPassphrase(a accounts.Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetPrivateKeyBytes returns the private key bytes if available
func (ew EthWalletWrapper) GetPrivateKeyBytes(account accounts.Account) (privateKey []byte, err error) {
	ethAccounts := ew.Wallet.Accounts()
	for _, acct := range ethAccounts {
		keyJSON, err := ew.ks.Export(acct, "", "")
		if err != nil {
			continue
		}
		pk, err := keystore.DecryptKey(keyJSON, "")
		if err != nil {
			continue
		}
		privateKeyBytes := crypto.FromECDSA(pk.PrivateKey)
		publicKey := filcrypto.PublicKey(privateKeyBytes)
		filAddr, err := address.NewSecp256k1Address(publicKey)
		if err != nil {
			continue
		}
		if filAddr == account.FilAddress {
			return privateKeyBytes, nil
		}
	}
	return []byte{}, fmt.Errorf("not found")
}
