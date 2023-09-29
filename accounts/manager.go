package accounts

import (
	"fmt"

	ethaccounts "github.com/ethereum/go-ethereum/accounts"
)

// managerSubBufferSize determines how many incoming wallet events
// the manager will buffer in its channel.
const managerSubBufferSize = 50

// Manager is an overarching account manager that can communicate with various
// backends for signing transactions.
type Manager struct {
	*ethaccounts.Manager
	filBackends []Backend
}

// NewManager creates a generic account manager to sign transaction via various
// supported backends.
func NewManager(config *ethaccounts.Config, backends []ethaccounts.Backend, filbackends []Backend) *Manager {
	ethManager := ethaccounts.NewManager(config, backends...)
	manager := Manager{
		Manager:     ethManager,
		filBackends: filbackends,
	}
	return &manager
}

// Find attempts to locate the wallet corresponding to a specific account. Since
// accounts can be dynamically added to and removed from wallets, this method has
// a linear runtime in the number of wallets.
func (am *Manager) Find(account Account) (Wallet, error) {
	if account.IsEth() {
		ethWallet, err := am.Manager.Find(account.EthAccount)
		if err != nil {
			return nil, err
		}
		wrappedEthWallet := EthWallet{ethWallet}
		return wrappedEthWallet, nil
	} else {
		fmt.Printf("Jim backends: %d\n", len(am.filBackends))
		for i, backend := range am.filBackends {
			fmt.Printf("Jim backend: %d: %+v\n", i, backend)
			for _, wallet := range backend.Wallets() {
				fmt.Printf("Jim wallet: %+v\n", wallet)
				for _, walletAccount := range wallet.Accounts() {
					fmt.Printf("Jim account: %+v\n", walletAccount)
					if walletAccount.FilAddress == account.FilAddress {
						return wallet, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("account not found in backends")
}
