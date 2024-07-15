package walletutils

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClientShimEthEoa struct {
	EthClientShimMethods
	ethClient *ethclient.Client
}

// NewEthWalletTransactor is a utility method to easily create transaction
// options for use with an Ethereum wallet with an optional passphrase.
func NewEthWalletTransactor(wallet accounts.Wallet, account *accounts.Account, passphrase string, chainID *big.Int, ethClient *ethclient.Client) (*EthClientShim, *bind.TransactOpts, error) {
	if chainID == nil {
		return nil, nil, bind.ErrNoChainID
	}

	shimImpl := &EthClientShimEthEoa{ethClient: ethClient}

	shimmedEthClient := &EthClientShim{
		Client: ethClient,
		shim:   shimImpl,
	}

	opts := &bind.TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, bind.ErrNotAuthorized
			}
			return wallet.SignTxWithPassphrase(*account, passphrase, tx, chainID)
		},
		Context: context.Background(),
	}

	return shimmedEthClient, opts, nil
}

// EthClientShimEthEoa does not actually shim the ethclient, it just provides type consistency for consumers of this package
func (c *EthClientShimEthEoa) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return c.ethClient.PendingNonceAt(ctx, account)
}

func (c *EthClientShimEthEoa) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	return c.ethClient.EstimateGas(ctx, call)
}

func (c *EthClientShimEthEoa) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return c.ethClient.SendTransaction(ctx, tx)
}

func (c *EthClientShimEthEoa) GetOuterTxHash(innerHash common.Hash) common.Hash {
	return innerHash
}
