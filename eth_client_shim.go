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
	impl EthClientShimMethods
}

func (c *EthClientShim) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return c.impl.PendingNonceAt(ctx, account)
}

func (c *EthClientShim) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	return c.impl.EstimateGas(ctx, call)
}

func (c *EthClientShim) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return c.impl.SendTransaction(ctx, tx)
}

func (c *EthClientShim) GetOuterTxHash(innerHash common.Hash) common.Hash {
	return c.impl.GetOuterTxHash(innerHash)
}
