package walletutils

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
)

type KeyType int

const (
	KeyTypeUnknown KeyType = iota
	KeyTypeEth
	KeyTypeFil
)

func (k KeyType) String() string {
	return [...]string{"Unknown", "Eth", "Fil"}[k]
}

func KeyTypeFromString(s string) KeyType {
	switch strings.ToLower(s) {
	case "eth":
		return KeyTypeEth
	case "fil":
		return KeyTypeFil
	default:
		return KeyTypeUnknown
	}
}

func KeyTypeFromAddr(addr string) KeyType {
	if strings.HasPrefix(addr, "0x") {
		return KeyTypeEth
	}

	_, err := address.NewFromString(addr)
	if err != nil {
		return KeyTypeUnknown
	}

	return KeyTypeFil
}

func DeriveAddrFromPkString(pk string) (common.Address, address.Address, error) {
	pkECDSA, err := crypto.HexToECDSA(pk)
	if err != nil {
		log.Fatal(err)
	}

	return DeriveFEVMAddrsFromPk(pkECDSA)
}

func DeriveEthAddressFromPk(pk *ecdsa.PrivateKey) (common.Address, error) {
	publicKey := pk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("error casting public key to ECDSA")
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}

func DelegatedFromEthAddr(addr common.Address) (address.Address, error) {
	fevmAddr, err := ethtypes.ParseEthAddress(addr.String())
	if err != nil {
		return address.Address{}, err
	}

	return fevmAddr.ToFilecoinAddress()
}

func DeriveFEVMAddrsFromPk(pk *ecdsa.PrivateKey) (common.Address, address.Address, error) {
	evmAddr, err := DeriveEthAddressFromPk(pk)
	if err != nil {
		return common.Address{}, address.Address{}, err
	}

	delegatedAddr, err := DelegatedFromEthAddr(evmAddr)
	if err != nil {
		return common.Address{}, address.Address{}, err
	}

	return evmAddr, delegatedAddr, nil
}

// IsZeroAddress validate if it's a 0 address
func IsZeroAddress(address common.Address) bool {
	if isEmptyStruct(address) {
		return true
	}
	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// isEmptyStruct checks if a variable is an empty instance of a struct
func isEmptyStruct(s interface{}) bool {
	v := reflect.ValueOf(s)

	// Ensure the variable is a struct
	if v.Kind() != reflect.Struct {
		return false
	}

	// Check if all fields have their zero values
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			return false
		}
	}

	return true
}

func TruncateAddr(addr string) string {
	if len(addr) <= 10 {
		return addr
	}

	firstSix := addr[:6]
	lastFour := addr[len(addr)-4:]
	return fmt.Sprintf("%s...%s", firstSix, lastFour)
}
