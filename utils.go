package walletutils

import (
	filtypes "github.com/filecoin-project/lotus/chain/types"
	"golang.org/x/xerrors"
)

func propagateErr(returnTrace filtypes.ReturnTrace) error {
	// chop off the first two bytes to get the FEVM error selector
	// TODO: return FEVM call error for our smart contracts
	// matched, err := matchSelector(ToByte4Selector(returnTrace.Return[1:]))
	// if matched {
	// 	return xerrors.Errorf("Simulation reverted with error, tx not submitted: %s", err.Error())
	// }

	return xerrors.Errorf("subcall error code: %s - return (could not decode) %x", returnTrace.ExitCode, returnTrace.Return[1:])
}

func recursiveSubcallErrorThrown(trace *filtypes.ExecutionTrace) error {

	// Recursively check the exit code of each subcall
	for _, subcall := range trace.Subcalls {
		err := recursiveSubcallErrorThrown(&subcall)
		if err != nil {
			return err
		}
	}

	// Check the exit code of the current trace last to make sure we propagate the error
	if trace.MsgRct.ExitCode != 0 {
		return propagateErr(trace.MsgRct)
	}

	// If we've made it this far, none of the exit codes were non-zero
	return nil
}
