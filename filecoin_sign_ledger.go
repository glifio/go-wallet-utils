package walletutils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	ledgerfil "github.com/whyrusleeping/ledger-filecoin-go"
)

const hdHard = 0x80000000

func SignMsgLedger(msg *types.Message) (*types.SignedMessage, error) {
	fl, err := ledgerfil.FindLedgerFilecoinApp()
	if err != nil {
		return nil, err
	}

	p, err := parseHDPath("m/44'/461'/0'/0/0")
	if err != nil {
		return nil, err
	}

	b, err := msg.ToStorageBlock()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Message: %x\n", b.RawData())

	sig, err := fl.SignSECP256K1(p, b.RawData())
	if err != nil {
		return nil, err
	}

	sigBytes := append([]byte{byte(crypto.SigTypeSecp256k1)}, sig.SignatureBytes()...)
	fmt.Printf("Signature: %x\n", sigBytes)

	return &types.SignedMessage{
		Message: *msg,
		Signature: crypto.Signature{
			Type: crypto.SigTypeSecp256k1,
			Data: sig.SignatureBytes(),
		},
	}, nil
}

// from lotus-shed
func parseHDPath(s string) ([]uint32, error) {
	parts := strings.Split(s, "/")
	if parts[0] != "m" {
		return nil, fmt.Errorf("expected HD path to start with 'm'")
	}

	var out []uint32
	for _, p := range parts[1:] {
		var hard bool
		if strings.HasSuffix(p, "'") {
			p = p[:len(p)-1]
			hard = true
		}

		v, err := strconv.ParseUint(p, 10, 32)
		if err != nil {
			return nil, err
		}
		if v >= hdHard {
			return nil, fmt.Errorf("path element %s too large", p)
		}

		if hard {
			v += hdHard
		}
		out = append(out, uint32(v))
	}
	return out, nil
}
