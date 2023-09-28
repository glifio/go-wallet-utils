package walletutils

import (
	"github.com/filecoin-project/go-crypto"
	crypto2 "github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/minio/blake2b-simd"
)

func SignMsg(pk []byte, msg *types.Message) (*types.SignedMessage, error) {
	b2sum := blake2b.Sum256(msg.Cid().Bytes())
	sig, err := crypto.Sign(pk, b2sum[:])
	if err != nil {
		return nil, err
	}

	return &types.SignedMessage{
		Message: *msg,
		Signature: crypto2.Signature{
			Type: crypto2.SigTypeSecp256k1,
			Data: sig,
		},
	}, nil
}
