package accounts

import (
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
func (eq EthWallet) SignTxWithPassphrase(a Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return eq.Wallet.SignTxWithPassphrase(a.EthAccount, passphrase, tx, chainID)
}
