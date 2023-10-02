package submitter

import (
	"context"
	"math/big"

	proofProducer "github.com/taikoxyz/taiko-client/prover/proof_producer"
	"github.com/taikoxyz/taiko-mono/packages/protocol/bindings"
)

type ProofSubmitter interface {
	RequestProof(ctx context.Context, event *bindings.TaikoL1ClientBlockProposed) error
	SubmitProof(ctx context.Context, proofWithHeader *proofProducer.ProofWithHeader) error
	CancelProof(ctx context.Context, blockID *big.Int) error
}
