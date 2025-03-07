package tests

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/taikoxyz/taiko-mono/packages/protocol/bindings/encoding"
)

type CalldataSyncer interface {
	ProcessL1Blocks(ctx context.Context, l1End *types.Header) error
}

type Proposer interface {
	ProposeOp(ctx context.Context) error
	ProposeEmptyBlockOp(ctx context.Context) error
	L2SuggestedFeeRecipient() common.Address
	ProposeTxList(
		ctx context.Context,
		meta *encoding.TaikoL1BlockMetadataInput,
		txListBytes []byte,
		txNum uint,
		nonce *uint64,
	) error
	Close(ctx context.Context)
}
