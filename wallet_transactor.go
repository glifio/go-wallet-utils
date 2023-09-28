package walletutils

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	lotustypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/go-wallet-utils/accounts"
)

// NewWalletTransactor is a utility method to easily create transaction
// options for use with an Ethereum or Filecoin wallet with an optional passphrase.
func NewWalletTransactor(
	ctx context.Context,
	lapi *api.FullNodeStruct,
	client *ethclient.Client,
	wallet accounts.Wallet,
	account *accounts.Account,
	passphrase string,
	proposer address.Address,
	approver address.Address,
	chainID *big.Int,
) (*WrappedEthClient, *bind.TransactOpts, error) {
	if chainID == nil {
		return nil, nil, bind.ErrNoChainID
	}
	if account.IsEth() {
		wrappedClientImpl := WrappedEthClientForEth{client: client}
		wrappedClient := &WrappedEthClient{
			Client: *client,
			impl:   wrappedClientImpl,
		}
		return wrappedClient, &bind.TransactOpts{
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
		from := account.FilAddress
		wrappedClient, auth, err := NewFilecoinLedgerTransactor(context.Background(), lapi, client, from)
		return wrappedClient, auth, err
	}
	return nil, nil, fmt.Errorf("account not matched")
}

type WrappedEthClientForEth struct {
	client *ethclient.Client
}

// PendingNonceAt retrieves the current pending nonce associated with an account.
func (_Client WrappedEthClientForEth) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return _Client.client.PendingNonceAt(ctx, account)
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
// There is no guarantee that this is the true gas limit requirement as other
// transactions may be added or removed by miners, but it should provide a basis
// for setting a reasonable default.
func (_Client WrappedEthClientForEth) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	return _Client.client.EstimateGas(ctx, call)
}

// SendTransaction injects the transaction into the pending pool for execution.
func (_Client WrappedEthClientForEth) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return _Client.client.SendTransaction(ctx, tx)
}

func (_Client WrappedEthClientForEth) SetSignedMessage(txHash common.Hash, signedMsg *lotustypes.SignedMessage) {
}

// FilecoinEthHash takes a transaction hash for a native Ethereum transaction and
// returns the transaction hash for a submitted Ethereum transaction signed with Filecoin keys
func (_Client WrappedEthClientForEth) FilecoinEthHash(txHash common.Hash) common.Hash {
	return txHash
}
