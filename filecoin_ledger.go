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
	lotusbig "github.com/filecoin-project/go-state-types/big"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	lotusapi "github.com/filecoin-project/lotus/api"
	lotustypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// NewFilecoinLedgerTransactor is a utility method to easily create transaction
// options for use with the Lotus JSON-RPC API and a Ledger USB Wallet with
// the Filecoin app.
func NewFilecoinLedgerTransactor(ctx context.Context, api *lotusapi.FullNodeStruct, client *ethclient.Client, from address.Address) (*WrappedEthClient, *bind.TransactOpts, error) {
	wrappedClientImpl := WrappedEthClientForFilLedger{
		from:            from,
		api:             api,
		signedMessage:   make(map[common.Hash]*lotustypes.SignedMessage),
		filecoinEthHash: make(map[common.Hash]common.Hash),
	}
	wrappedClient := &WrappedEthClient{
		Client: *client,
		impl:   wrappedClientImpl,
	}

	opts := bind.TransactOpts{
		From: common.Address{}, // unused
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			filecoinToAddr, err := ethtypes.ParseEthAddress(tx.To().String())
			if err != nil {
				return nil, err
			}

			delegatedToAddr, err := filecoinToAddr.ToFilecoinAddress()
			if err != nil {
				return nil, err
			}

			var buffer bytes.Buffer
			if err := cbg.WriteByteArray(&buffer, tx.Data()); err != nil {
				return nil, err
			}
			calldata := buffer.Bytes()

			proposeMsg := &lotustypes.Message{
				From:       from,
				To:         delegatedToAddr,
				GasLimit:   int64(tx.Gas()),
				GasFeeCap:  lotusbig.NewFromGo(tx.GasFeeCap()),
				GasPremium: lotusbig.NewFromGo(tx.GasTipCap()),
				Nonce:      tx.Nonce(),
				Method:     builtintypes.MethodsEVM.InvokeContract,
				Value:      lotusbig.NewFromGo(tx.Value()),
				Params:     calldata,
			}

			signedMsg, err := SignMsgLedger(proposeMsg)
			if err != nil {
				return tx, err
			}

			wrappedClient.SetSignedMessage(tx.Hash(), signedMsg)

			return tx, nil
		},
		Context: context.Background(),
	}
	return wrappedClient, &opts, nil
}

type WrappedEthClientForFilLedger struct {
	from            address.Address
	api             *lotusapi.FullNodeStruct
	signedMessage   map[common.Hash]*lotustypes.SignedMessage
	filecoinEthHash map[common.Hash]common.Hash
}

// PendingNonceAt retrieves the current pending nonce associated with an account.
func (_Client WrappedEthClientForFilLedger) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nonce, err := _Client.api.MpoolGetNonce(ctx, _Client.from)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
// There is no guarantee that this is the true gas limit requirement as other
// transactions may be added or removed by miners, but it should provide a basis
// for setting a reasonable default.
func (_Client WrappedEthClientForFilLedger) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	filecoinToAddr, err := ethtypes.ParseEthAddress(call.To.String())
	if err != nil {
		return 0, err
	}

	delegatedToAddr, err := filecoinToAddr.ToFilecoinAddress()
	if err != nil {
		return 0, err
	}

	var buffer bytes.Buffer
	if err := cbg.WriteByteArray(&buffer, call.Data); err != nil {
		return 0, err
	}
	calldata := buffer.Bytes()

	proposeMsg := &lotustypes.Message{
		From:       _Client.from,
		To:         delegatedToAddr,
		GasFeeCap:  lotusbig.NewFromGo(call.GasFeeCap),
		GasPremium: lotusbig.NewFromGo(call.GasTipCap),
		Method:     builtintypes.MethodsEVM.InvokeContract,
		Value:      lotusbig.NewFromGo(call.Value),
		Params:     calldata,
	}

	msgWithGas, err := _Client.api.GasEstimateMessageGas(ctx, proposeMsg, nil, lotustypes.EmptyTSK)
	if err != nil {
		return 0, err
	}

	return uint64(msgWithGas.GasLimit), nil
}

// SendTransaction injects the transaction into the pending pool for execution.
func (_Client WrappedEthClientForFilLedger) SendTransaction(ctx context.Context, tx *types.Transaction) error {

	signedMessage := _Client.signedMessage[tx.Hash()]
	delete(_Client.signedMessage, tx.Hash())

	cid, err := _Client.api.MpoolPush(ctx, signedMessage)
	if err != nil {
		return err
	}

	txHashFil, err := _Client.api.EthGetTransactionHashByCid(ctx, cid)
	if err != nil {
		return err
	}

	filTxHash := common.Hash{}
	filTxHash.UnmarshalText([]byte(txHashFil.String()))
	_Client.filecoinEthHash[tx.Hash()] = filTxHash

	return nil
}

func (_Client WrappedEthClientForFilLedger) SetSignedMessage(txHash common.Hash, signedMsg *lotustypes.SignedMessage) {
	_Client.signedMessage[txHash] = signedMsg
}

// FilecoinEthHash takes a transaction hash for a native Ethereum transaction and
// returns the transaction hash for a submitted Ethereum transaction signed with Filecoin keys
func (_Client WrappedEthClientForFilLedger) FilecoinEthHash(txHash common.Hash) common.Hash {
	return _Client.filecoinEthHash[txHash]
}
