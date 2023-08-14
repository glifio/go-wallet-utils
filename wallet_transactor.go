package walletutils

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/glifio/go-wallet-utils/accounts"
)

// NewWalletTransactor is a utility method to easily create transaction
// options for use with an Ethereum or Filecoin wallet with an optional passphrase.
func NewWalletTransactor(wallet accounts.Wallet, account *accounts.Account, passphrase string, chainID *big.Int) (*bind.TransactOpts, error) {
	if chainID == nil {
		return nil, bind.ErrNoChainID
	}
	if account.IsEth() {
		return &bind.TransactOpts{
			From: account.EthAccount.Address,
			Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
				if address != account.EthAccount.Address {
					return nil, bind.ErrNotAuthorized
				}
				return wallet.SignTxWithPassphrase(*account, passphrase, tx, chainID)
			},
			Context: context.Background(),
		}, nil
	}
	if account.IsFil() {

	}
}
