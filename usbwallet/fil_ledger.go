package usbwallet

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/filecoin-project/go-address"
	"github.com/glifio/go-wallet-utils/accounts"
	ledgerfil "github.com/whyrusleeping/ledger-filecoin-go"
)

const hdHard = 0x80000000

type Hub struct {
	ledgerFil *ledgerfil.LedgerFilecoin
	wallets   []accounts.Wallet
}

// NewLedgerHub creates a new hardware wallet manager for Ledger devices.
func NewLedgerHub() (*Hub, error) {
	ledgerFil, err := ledgerfil.FindLedgerFilecoinApp()
	if err != nil {
		return nil, err
	}

	// FIXME: Hardwire the first ledger path for now
	p := []uint32{hdHard | 44, hdHard | 461, hdHard, 0, 0}
	pubk, err := ledgerFil.GetPublicKeySECP256K1(p)
	if err != nil {
		return nil, err
	}

	addr, err := address.NewSecp256k1Address(pubk)
	if err != nil {
		return nil, err
	}

	account := accounts.Account{FilAddress: addr}

	wallet := FilLedgerWallet{
		accounts: []accounts.Account{account},
	}

	hub := Hub{
		ledgerFil: ledgerFil,
		wallets:   []accounts.Wallet{wallet},
	}

	err = ledgerFil.Close()
	if err != nil {
		return nil, err
	}

	return &hub, nil
}

// Wallets implements accounts.Backend, returning all the currently tracked USB
// devices that appear to be hardware wallets.
func (hub *Hub) Wallets() []accounts.Wallet {
	return hub.wallets
}

type FilLedgerWallet struct {
	accounts []accounts.Account
}

// Accounts returns all the accounts in the wallet
func (fw FilLedgerWallet) Accounts() []accounts.Account {
	return fw.accounts
}

// SignTxWithPassphrase signs the transaction
func (fw FilLedgerWallet) SignTxWithPassphrase(a accounts.Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetPrivateKeyBytes returns the private key bytes if available
func (fw FilLedgerWallet) GetPrivateKeyBytes(account accounts.Account, passphrase string) (privateKey []byte, err error) {
	return []byte{}, fmt.Errorf("not implemented")
}
