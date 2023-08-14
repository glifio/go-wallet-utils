package walletutils

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	lotustypes "github.com/filecoin-project/lotus/chain/types"
)

type WrappedEthClientMethods interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	SetSignedMessage(txHash common.Hash, signedMsg *lotustypes.SignedMessage)
	FilecoinEthHash(txHash common.Hash) common.Hash
}

type WrappedEthClient struct {
	ethclient.Client
	impl WrappedEthClientMethods
}

func (c *WrappedEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return c.impl.PendingNonceAt(ctx, account)
}

func (c *WrappedEthClient) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	return c.impl.EstimateGas(ctx, call)
}

func (c *WrappedEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return c.impl.SendTransaction(ctx, tx)
}

func (c *WrappedEthClient) SetSignedMessage(txHash common.Hash, signedMsg *lotustypes.SignedMessage) {
	c.impl.SetSignedMessage(txHash, signedMsg)
}

func (c *WrappedEthClient) FilecoinEthHash(txHash common.Hash) common.Hash {
	return c.impl.FilecoinEthHash(txHash)
}
