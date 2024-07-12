package walletutils

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClientShimMethods interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	GetOuterTxHash(tx common.Hash) common.Hash
}

type EthClientShim struct {
	*ethclient.Client
	shim EthClientShimMethods
}

func (c *EthClientShim) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return c.shim.PendingNonceAt(ctx, account)
}

func (c *EthClientShim) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	return c.shim.EstimateGas(ctx, call)
}

func (c *EthClientShim) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return c.shim.SendTransaction(ctx, tx)
}

func (c *EthClientShim) GetOuterTxHash(innerHash common.Hash) common.Hash {
	return c.shim.GetOuterTxHash(innerHash)
}
