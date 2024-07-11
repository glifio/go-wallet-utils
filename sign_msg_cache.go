package walletutils

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	lotustypes "github.com/filecoin-project/lotus/chain/types"
)

// maps ethereum tx hashes to filecoin signed messages
type SignedMessageCache struct {
	signedMessages map[common.Hash]*lotustypes.SignedMessage
	mutex          sync.Mutex
}

func (smc *SignedMessageCache) Add(txHash common.Hash, signedMsg *lotustypes.SignedMessage) {
	smc.mutex.Lock()
	defer smc.mutex.Unlock()

	smc.signedMessages[txHash] = signedMsg
}

func (smc *SignedMessageCache) Get(txHash common.Hash) *lotustypes.SignedMessage {
	smc.mutex.Lock()
	defer smc.mutex.Unlock()

	return smc.signedMessages[txHash]
}

func (smc *SignedMessageCache) Delete(txHash common.Hash) {
	smc.mutex.Lock()
	defer smc.mutex.Unlock()

	delete(smc.signedMessages, txHash)
}

func NewSignedMsgCache() *SignedMessageCache {
	return &SignedMessageCache{
		signedMessages: make(map[common.Hash]*lotustypes.SignedMessage),
	}
}
