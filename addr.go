package walletutils

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/glifio/go-pools/util"
)

func GetGenericAddr(addr string) (interface{}, KeyType, error) {
	// check if the key is an eth address
	if strings.HasPrefix(addr, "0x") {
		return common.HexToAddress(addr), KeyTypeEth, nil
	}

	// check if the key is a filecoin address
	filAddr, err := address.NewFromString(addr)
	if err != nil {
		return "", KeyTypeUnknown, err
	}

	return filAddr, KeyTypeFil, nil
}

func GetFilAddr(ctx context.Context, addrStr string, lapi api.FullNode) (address.Address, error) {
	addr, keyType, err := GetGenericAddr(addrStr)
	if err != nil {
		return address.Undef, err
	}

	if keyType == KeyTypeFil {
		// make sure the fil key type is not an ID address of an EVM actor type
		if err := util.CheckIDNotEVMActorType(ctx, addr.(address.Address), lapi); err != nil {
			return address.Undef, err
		}
		return addr.(address.Address), nil
	} else if keyType == KeyTypeEth {
		// convert to filecoin address
		ethAddr, err := ethtypes.ParseEthAddress(addr.(common.Address).String())
		if err != nil {
			return address.Undef, err
		}

		return ethAddr.ToFilecoinAddress()
	}

	return address.Undef, ErrUnsupportedKeyType
}

func GetEthAddr(ctx context.Context, addrStr string, lapi api.FullNode) (common.Address, error) {
	addr, keyType, err := GetGenericAddr(addrStr)
	if err != nil {
		return common.Address{}, err
	}

	if keyType == KeyTypeFil {
		filAddr := addr.(address.Address)
		// make sure the fil key type is not an ID address of an EVM actor type
		if err := util.CheckIDNotEVMActorType(ctx, filAddr, lapi); err != nil {
			return common.Address{}, err
		}

		// if the address is f1, f2, f3, go fetch the ID addr and return the masked hex id
		if filAddr.Protocol() != address.ID && filAddr.Protocol() != address.Delegated {
			// convert to filecoin f0 address
			filAddr, err = lapi.StateLookupID(ctx, filAddr, types.EmptyTSK)
			if err != nil {
				return common.Address{}, err
			}
		}

		ethAddr, err := ethtypes.EthAddressFromFilecoinAddress(filAddr)
		if err != nil {
			return common.Address{}, err
		}
		return common.HexToAddress(ethAddr.String()), nil
	} else if keyType == KeyTypeEth {
		return addr.(common.Address), nil
	}

	return common.Address{}, ErrUnsupportedKeyType
}
