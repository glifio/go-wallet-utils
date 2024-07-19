package walletutils

import (
	"context"
	"testing"

	"github.com/filecoin-project/go-address"
	actorstypes "github.com/filecoin-project/go-state-types/actors"
	"github.com/filecoin-project/go-state-types/manifest"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
)

var idStr = "f01234"
var maskedIDStr = "0xFF000000000000000000000000000000000004d2"
var evmActorID = "f05678"
var ethAccountID = "f09876"

type MockFullNodeAPI struct {
	api.FullNode
}

func (m *MockFullNodeAPI) StateLookupID(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	return address.NewFromString(idStr)
}

func (m *MockFullNodeAPI) StateGetActor(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error) {
	code := cid.Cid{}
	if actor.String() == evmActorID {
		code, _ = actors.GetActorCodeID(actorstypes.Version(actors.LatestVersion), manifest.EvmKey)
	} else if actor.String() == ethAccountID {
		code, _ = actors.GetActorCodeID(actorstypes.Version(actors.LatestVersion), manifest.EthAccountKey)
	}

	return &types.Actor{
		Code:    code,
		Head:    cid.Cid{},
		Nonce:   0,
		Balance: types.NewInt(0),
	}, nil
}

func TestGetFilAddr(t *testing.T) {
	// Create a mock FullNodeAPI
	mockAPI := &MockFullNodeAPI{}

	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name: "Valid Ethereum Address",

			input:       "0x3972E844729522d367BFA1D64368346D7ccEEa59",
			expected:    "f410fhfzoqrdssurngz57uhleg2bunv6m52sznazommi",
			expectError: false,
		},
		{
			name:        "Valid Filecoin ID Address",
			input:       idStr,
			expected:    idStr,
			expectError: false,
		},
		{
			name:        "Valid Filecoin Account f1 Address",
			input:       "f1ys5qqiciehcml3sp764ymbbytfn3qoar5fo3iwy",
			expected:    "f1ys5qqiciehcml3sp764ymbbytfn3qoar5fo3iwy",
			expectError: false,
		},
		{
			name:        "Valid Filecoin Account f3 Address",
			input:       "f3vpyybzycb3wvhwkxcrodn3rqv66sd5hfho4lfq6p6igmrlgyb22v3ekdghp6km47ioki3gfo4zb4ezirhfaq",
			expected:    "f3vpyybzycb3wvhwkxcrodn3rqv66sd5hfho4lfq6p6igmrlgyb22v3ekdghp6km47ioki3gfo4zb4ezirhfaq",
			expectError: false,
		},
		{
			name:        "Valid Filecoin Account f4 Address",
			input:       "f410fmdqxonrwz5peuit5tlbe6ih6zibu5ys223xctfi",
			expected:    "f410fmdqxonrwz5peuit5tlbe6ih6zibu5ys223xctfi",
			expectError: false,
		},
		{
			name:        "Filecoin masked ID Addr",
			input:       maskedIDStr,
			expected:    idStr,
			expectError: false,
		},
		{
			name:        "Invalid Address",
			input:       "invalid_address",
			expectError: true,
		},
		{
			name:        "EVM Actor",
			input:       evmActorID,
			expectError: true,
		},
		{
			name:        "Eth Account",
			input:       ethAccountID,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetFilAddr(context.Background(), tc.input, mockAPI)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result.String())
			}
		})
	}
}

func TestGetEthAddr(t *testing.T) {
	// Create a mock FullNodeAPI
	mockAPI := &MockFullNodeAPI{}

	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Valid Ethereum Address",
			input:       "0x3972E844729522d367BFA1D64368346D7ccEEa59",
			expected:    "0x3972E844729522d367BFA1D64368346D7ccEEa59",
			expectError: false,
		},
		{
			name:        "Valid Filecoin ID Address",
			input:       idStr,
			expected:    maskedIDStr,
			expectError: false,
		},
		{
			name:        "Valid Filecoin Account f1 Address",
			input:       "f1ys5qqiciehcml3sp764ymbbytfn3qoar5fo3iwy",
			expected:    maskedIDStr,
			expectError: false,
		},
		{
			name:        "Valid Filecoin Account f3 Address",
			input:       "f3vpyybzycb3wvhwkxcrodn3rqv66sd5hfho4lfq6p6igmrlgyb22v3ekdghp6km47ioki3gfo4zb4ezirhfaq",
			expected:    maskedIDStr,
			expectError: false,
		},
		{
			name:        "Valid Filecoin Account f4 Address",
			input:       "f410fhwnyp6tw6n7be5ebmi2izbwvffgenhcnj2gywgy",
			expected:    "0x3d9B87FA76f37e12748162348C86D5294c469c4D",
			expectError: false,
		},
		{
			name:        "Filecoin masked ID Addr",
			input:       maskedIDStr,
			expected:    maskedIDStr,
			expectError: false,
		},
		{
			name:        "Invalid Address",
			input:       "invalid_address",
			expectError: true,
		},
		{
			name:        "EVM Actor",
			input:       evmActorID,
			expectError: true,
		},
		{
			name:        "Eth Account",
			input:       ethAccountID,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetEthAddr(context.Background(), tc.input, mockAPI)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result.String())
			}
		})
	}
}
