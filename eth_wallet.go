package walletutils

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// NewEthWalletTransactor is a utility method to easily create transaction
// options for use with an Ethereum wallet with an optional passphrase.
func NewEthWalletTransactor(wallet accounts.Wallet, account *accounts.Account, passphrase string, chainID *big.Int) (*bind.TransactOpts, error) {
	fmt.Println("Jim NewEthWalletTransactor")
	if chainID == nil {
		return nil, bind.ErrNoChainID
	}
	return &bind.TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, bind.ErrNotAuthorized
			}
			fmt.Printf("Jim sign wallet %+v\n", wallet)
			fmt.Printf("Jim sign account %+v\n", account)
			return wallet.SignTxWithPassphrase(*account, passphrase, tx, chainID)
		},
		Context: context.Background(),
	}, nil
}
