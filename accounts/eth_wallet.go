package accounts

import (
	"fmt"
	"math/big"

	ethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthWallet struct {
	ethaccounts.Wallet
}

// Accounts returns all key files present in the directory.
func (ew EthWallet) Accounts() []Account {
	ethAccounts := ew.Wallet.Accounts()
	accounts := make([]Account, 1)
	for _, acct := range ethAccounts {
		accounts = append(accounts, Account{EthAccount: acct})
	}
	return accounts
}

// SignTxWithPassphrase signs the transaction
func (ew EthWallet) SignTxWithPassphrase(a Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return ew.Wallet.SignTxWithPassphrase(a.EthAccount, passphrase, tx, chainID)
}

// GetPrivateKeyBytes returns the private key bytes if available
func (ew EthWallet) GetPrivateKeyBytes(account Account, passphrase string) (privateKey []byte, err error) {
	return []byte{}, fmt.Errorf("not implemented")
}
