package walletutils

import (
	"bytes"
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/filecoin-project/go-address"
	filbig "github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/builtin"
	msig14 "github.com/filecoin-project/go-state-types/builtin/v14/multisig"
	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors"
	lotustypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type EthClientShimFilMsigProposer struct {
	EthClientShimMethods
	from           address.Address
	api            *lapi.FullNodeStruct
	msig           address.Address
	signedMsgCache *SignedMessageCache
}

func NewFilMsigProposerWalletTransactor(
	ctx context.Context,
	api *lapi.FullNodeStruct,
	client *ethclient.Client,
	from address.Address,
	fromPrivateKey []byte,
	msig address.Address,
) (*EthClientShim, *bind.TransactOpts, error) {
	shimImpl := &EthClientShimFilMsigProposer{
		from:           from,
		api:            api,
		msig:           msig,
		signedMsgCache: NewSignedMsgCache(),
	}

	shimmedEthClient := &EthClientShim{
		Client: client,
		shim:   shimImpl,
	}

	opts := &bind.TransactOpts{
		From: common.Address{}, // unused
		Signer: func(_ common.Address, tx *types.Transaction) (*types.Transaction, error) {
			// get the address of the FEVM smart contract we want to send a msig proposal to
			filecoinToAddr, err := ethtypes.ParseEthAddress(tx.To().String())
			if err != nil {
				return nil, err
			}

			// convert the 0x -> f4 for interacting with the lotus rpc
			delegatedToAddr, err := filecoinToAddr.ToFilecoinAddress()
			if err != nil {
				return nil, err
			}

			var buffer bytes.Buffer
			if err := cbg.WriteByteArray(&buffer, tx.Data()); err != nil {
				return nil, err
			}
			calldata := buffer.Bytes()

			var signedMsg *lotustypes.SignedMessage
			// msig
			enc, actErr := actors.SerializeParams(&msig14.ProposeParams{
				To:     delegatedToAddr,
				Value:  filbig.NewFromGo(tx.Value()),
				Method: builtin.MethodsEVM.InvokeContract,
				Params: calldata,
			})

			if actErr != nil {
				return nil, actErr
			}

			proposeMsg := &lotustypes.Message{
				To:         msig,
				From:       from,
				Value:      filbig.Zero(),
				Method:     builtin.MethodsMultisig.Propose,
				Params:     enc,
				Nonce:      tx.Nonce(),
				GasLimit:   int64(tx.Gas()),
				GasFeeCap:  filbig.NewFromGo(tx.GasFeeCap()),
				GasPremium: filbig.NewFromGo(tx.GasTipCap()),
			}

			signedMsg, err = SignMsg(fromPrivateKey, proposeMsg)
			if err != nil {
				return tx, err
			}

			shimImpl.signedMsgCache.Add(tx.Hash(), signedMsg)

			return tx, nil
		},
		Context: ctx,
	}
	return shimmedEthClient, opts, nil
}

func (c *EthClientShimFilMsigProposer) PendingNonceAt(ctx context.Context, _ common.Address) (uint64, error) {
	nonce, err := c.api.MpoolGetNonce(ctx, c.from)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (c *EthClientShimFilMsigProposer) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	// call.To is the fevm smart contract we wish to transact with
	filecoinToAddr, err := ethtypes.ParseEthAddress(call.To.String())
	if err != nil {
		return 0, err
	}
	// convert the 0x -> f4 for interacting with the lotus rpc
	delegatedToAddr, err := filecoinToAddr.ToFilecoinAddress()
	if err != nil {
		return 0, err
	}

	var buffer bytes.Buffer
	if err := cbg.WriteByteArray(&buffer, call.Data); err != nil {
		return 0, err
	}
	calldata := buffer.Bytes()

	var proposeMsg *lotustypes.Message

	// serialize the inner msig proposal params
	enc, actErr := actors.SerializeParams(&msig14.ProposeParams{
		To:     delegatedToAddr,
		Value:  filbig.NewFromGo(call.Value),
		Method: builtin.MethodsEVM.InvokeContract,
		Params: calldata,
	})

	if actErr != nil {
		return 0, actErr
	}

	proposeMsg = &lotustypes.Message{
		To:     c.msig,
		From:   c.from,
		Value:  filbig.Zero(),
		Method: builtin.MethodsMultisig.Propose,
		Params: enc,
	}

	msgWithGas, err := c.api.GasEstimateMessageGas(ctx, proposeMsg, nil, lotustypes.EmptyTSK)
	if err != nil {
		return 0, err
	}

	state, err := c.api.StateCall(ctx, msgWithGas, lotustypes.EmptyTSK)
	if err != nil {
		return 0, err
	}

	if err := recursiveSubcallErrorThrown(&state.ExecutionTrace); err != nil {
		return 0, err
	}

	return uint64(msgWithGas.GasLimit), nil
}

func (c *EthClientShimFilMsigProposer) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	signedMessage := c.signedMsgCache.Get(tx.Hash())
	c.signedMsgCache.Delete(tx.Hash())

	// regular filecoin tx
	cid, err := c.api.MpoolPush(ctx, signedMessage)
	if err != nil {
		return err
	}

	filtx, err := c.api.EthGetTransactionHashByCid(ctx, cid)
	if err != nil {
		return err
	}

	filTxHash := common.Hash{}
	filTxHash.UnmarshalText([]byte(filtx.String()))

	c.signedMsgCache.MapHash(tx.Hash(), filTxHash)

	return nil
}

func (c *EthClientShimFilMsigProposer) GetOuterTxHash(innerHash common.Hash) common.Hash {
	return c.signedMsgCache.GetOuterHash(innerHash)
}
