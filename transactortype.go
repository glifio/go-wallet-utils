package walletutils

type TransactorType int

const (
	TransactorTypeEthEOA TransactorType = iota
	TransactorTypeFilEOA
	TransactorTypeFilMSig
	TransactorTypeUnknown
)

func TransactorTypeFromString(s string) TransactorType {
	switch s {
	case "eth-secp256k1":
		return TransactorTypeEthEOA
	case "fil-secp256k1":
		return TransactorTypeFilEOA
	case "fil-msig":
		return TransactorTypeFilMSig
	default:
		return TransactorTypeUnknown
	}
}
