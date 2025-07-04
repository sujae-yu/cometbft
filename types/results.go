package types

import (
	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/cometbft/cometbft/v2/crypto/merkle"
)

// ABCIResults wraps the deliver tx results to return a proof.
type ABCIResults []*abci.ExecTxResult

// NewResults strips non-deterministic fields from ExecTxResult responses
// and returns ABCIResults.
func NewResults(responses []*abci.ExecTxResult) ABCIResults {
	res := make(ABCIResults, len(responses))
	for i, d := range responses {
		res[i] = abci.DeterministicExecTxResult(d)
	}
	return res
}

// Hash returns a merkle hash of all results.
func (a ABCIResults) Hash() []byte {
	return merkle.HashFromByteSlices(a.toByteSlices())
}

// ProveResult returns a merkle proof of one result from the set.
func (a ABCIResults) ProveResult(i int) merkle.Proof {
	_, proofs := merkle.ProofsFromByteSlices(a.toByteSlices())
	return *proofs[i]
}

func (a ABCIResults) toByteSlices() [][]byte {
	l := len(a)
	bzs := make([][]byte, l)
	for i := 0; i < l; i++ {
		bz, err := a[i].Marshal()
		if err != nil {
			panic(err)
		}
		bzs[i] = bz
	}
	return bzs
}
