package accounts

import (
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/event"
)

// Manager is an overarching account manager that can communicate with various
// backends for signing transactions.
type Manager struct {
	config      *accounts.Config           // Global account manager configurations
	backends    map[reflect.Type][]Backend // Index of backends currently registered
	updaters    []event.Subscription       // Wallet update subscriptions for all backends
	updates     chan WalletEvent           // Subscription sink for backend wallet changes
	newBackends chan newBackendEvent       // Incoming backends to be tracked by the manager
	wallets     []Wallet                   // Cache of all wallets from all registered backends

	feed event.Feed // Wallet feed notifying of arrivals/departures

	quit chan chan error
	term chan struct{} // Channel is closed upon termination of the update loop
	lock sync.RWMutex
}

// NewManager creates a generic account manager to sign transaction via various
// supported backends.
func NewManager(config *accounts.Config, backends ...Backend) *Manager {
	// Retrieve the initial list of wallets from the backends and sort by URL
	var wallets []Wallet
	for _, backend := range backends {
		wallets = merge(wallets, backend.Wallets()...)
	}
	// Subscribe to wallet notifications from all backends
	updates := make(chan WalletEvent, managerSubBufferSize)

	subs := make([]event.Subscription, len(backends))
	for i, backend := range backends {
		subs[i] = backend.Subscribe(updates)
	}
	// Assemble the account manager and return
	am := &Manager{
		config:      config,
		backends:    make(map[reflect.Type][]Backend),
		updaters:    subs,
		updates:     updates,
		newBackends: make(chan newBackendEvent),
		wallets:     wallets,
		quit:        make(chan chan error),
		term:        make(chan struct{}),
	}
	for _, backend := range backends {
		kind := reflect.TypeOf(backend)
		am.backends[kind] = append(am.backends[kind], backend)
	}
	go am.update()

	return am
}
