package msigwallet

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/filecoin-project/go-address"
	"github.com/glifio/go-wallet-utils/accounts"
)

type Hub struct {
	wallets []accounts.Wallet
}

// NewMsigLedgerHub creates a new hardware wallet manager for Ledger devices.
func NewMsigLedgerHub() *Hub {
	hub := Hub{
		wallets: []accounts.Wallet{},
	}

	return &hub
}

// Wallets implements accounts.Backend, returning all the currently tracked USB
// devices that appear to be hardware wallets.
func (hub *Hub) Wallets() []accounts.Wallet {
	return hub.wallets
}

// AddMsig adds a wallet and account representing a Filecoin msig address, which has
// a proposer (address with a related private key stored in the go-ethereum keystore)
// and an approver (Filecoin address via Ledger hardware wallet)
func (hub *Hub) AddMsig(msigAddr address.Address, proposer address.Address, approver address.Address) {
	account := accounts.Account{FilAddress: msigAddr}

	wallet := FilMsigLedgerWallet{
		accounts: []accounts.Account{account},
		proposer: proposer,
		approver: approver,
	}

	hub.wallets = append(hub.wallets, wallet)
}

// SetManager updates all the wallets with a pointer to a manager for proposer/approver account lookups
func (hub *Hub) SetManager(manager *accounts.Manager) {
	fmt.Println("Jim SetManager", manager)
	newWallets := make([]accounts.Wallet, 0)
	for _, wallet := range hub.wallets {
		w := wallet.(FilMsigLedgerWallet)
		w.manager = manager
		newWallets = append(newWallets, w)
		fmt.Printf("Jim SetManager wallet: %+v\n", w)
	}
	hub.wallets = newWallets
}

type FilMsigLedgerWallet struct {
	accounts []accounts.Account
	proposer address.Address
	approver address.Address
	manager  *accounts.Manager
}

// Accounts returns all the accounts in the wallet
func (fw FilMsigLedgerWallet) Accounts() []accounts.Account {
	return fw.accounts
}

// SignTxWithPassphrase signs the transaction
func (fw FilMsigLedgerWallet) SignTxWithPassphrase(a accounts.Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetPrivateKeyBytes returns the private key bytes if available
func (fw FilMsigLedgerWallet) GetPrivateKeyBytes(account accounts.Account) (privateKey []byte, err error) {
	return []byte{}, fmt.Errorf("not implemented")
}

func (fw FilMsigLedgerWallet) GetProposerPrivateKey() (privateKey []byte, err error) {
	acct := accounts.Account{FilAddress: fw.proposer}
	fmt.Printf("Jim GetProposerPrivateKey: %+v\n", acct)
	fmt.Printf("Jim GetProposerPrivateKey fw: %+v\n", fw)

	if fw.manager == nil {
		return []byte{}, fmt.Errorf("manager not set")
	}
	wallet, err := fw.manager.Find(acct)
	if err != nil {
		return []byte{}, err
	}

	fmt.Println("Jim GetProposerPrivateKey", wallet)

	return wallet.GetPrivateKeyBytes(acct)
}
