package submitter

import (
	"bytes"
	"context"
	"math/big"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"

	"github.com/taikoxyz/taiko-client/driver/state"
	"github.com/taikoxyz/taiko-client/driver/syncer/beacon"
	"github.com/taikoxyz/taiko-client/driver/syncer/calldata"
	"github.com/taikoxyz/taiko-client/pkg/rpc"
	"github.com/taikoxyz/taiko-client/proposer"
	proofProducer "github.com/taikoxyz/taiko-client/prover/proof_producer"
	"github.com/taikoxyz/taiko-client/prover/server"
	"github.com/taikoxyz/taiko-client/tests"
	"github.com/taikoxyz/taiko-client/tests/helper"
	"github.com/taikoxyz/taiko-mono/packages/protocol/bindings"
)

type ProofSubmitterTestSuite struct {
	tests.ClientTestSuite
	validProofSubmitter *ValidProofSubmitter
	calldataSyncer      *calldata.Syncer
	proposer            *proposer.Proposer
	validProofCh        chan *proofProducer.ProofWithHeader
	invalidProofCh      chan *proofProducer.ProofWithHeader
	rpcClient           *rpc.Client
	proverEndpoints     []*url.URL
	proverServer        *server.ProverServer
}

func (s *ProofSubmitterTestSuite) SetupTest() {
	s.ClientTestSuite.SetupTest()
	s.rpcClient = helper.NewWsRpcClient(&s.ClientTestSuite)
	s.validProofCh = make(chan *proofProducer.ProofWithHeader, 1024)
	s.invalidProofCh = make(chan *proofProducer.ProofWithHeader, 1024)
	var err error
	s.validProofSubmitter, err = NewValidProofSubmitter(
		s.rpcClient,
		&proofProducer.DummyProofProducer{},
		s.validProofCh,
		tests.TaikoL2Address,
		tests.ProverPrivKey,
		&sync.Mutex{},
		false,
		"test",
		1,
		12*time.Second,
		10*time.Second,
		nil,
		2,
		nil,
	)
	s.NoError(err)
	ctx := context.Background()
	// Init calldata syncer
	testState, err := state.New(context.Background(), s.rpcClient)
	s.Nil(err)

	tracker := beacon.NewSyncProgressTracker(s.rpcClient.L2, 30*time.Second)

	s.calldataSyncer, err = calldata.NewSyncer(
		context.Background(),
		s.rpcClient,
		testState,
		tracker,
		s.L1.TaikoL1SignalService,
	)
	s.NoError(err)

	// Init proposer
	proposeInterval := 1024 * time.Hour // No need to periodically propose transactions list in unit tests
	s.proverEndpoints, s.proverServer, err = helper.DefaultFakeProver(&s.ClientTestSuite, s.rpcClient)
	s.NoError(err)
	cfg := &proposer.Config{
		L1Endpoint:                         s.L1.WsEndpoint(),
		L2Endpoint:                         s.L2.WsEndpoint(),
		TaikoL1Address:                     s.L1.TaikoL1Address,
		TaikoL2Address:                     tests.TaikoL2Address,
		TaikoTokenAddress:                  s.L1.TaikoL1TokenAddress,
		L1ProposerPrivKey:                  tests.ProposerPrivKey,
		L2SuggestedFeeRecipient:            tests.ProposerAddress,
		ProposeInterval:                    &proposeInterval,
		MaxProposedTxListsPerEpoch:         1,
		WaitReceiptTimeout:                 10 * time.Second,
		ProverEndpoints:                    s.proverEndpoints,
		BlockProposalFee:                   big.NewInt(1000),
		BlockProposalFeeIterations:         3,
		BlockProposalFeeIncreasePercentage: 2,
	}
	s.proposer, err = proposer.New(ctx, cfg)
	s.NoError(err)
}

func (s *ProofSubmitterTestSuite) TearDownTest() {
	s.proposer.Close(context.Background())
	_ = s.proverServer.Shutdown(context.Background())
	s.rpcClient.Close()
	s.ClientTestSuite.TearDownTest()
}

func (s *ProofSubmitterTestSuite) TestValidProofSubmitterRequestProofDeadlineExceeded() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	s.ErrorContains(
		s.validProofSubmitter.RequestProof(
			ctx, &bindings.TaikoL1ClientBlockProposed{BlockId: common.Big256}), "context deadline exceeded",
	)
}

func (s *ProofSubmitterTestSuite) TestValidProofSubmitterSubmitProofMetadataNotFound() {
	s.Error(
		s.validProofSubmitter.SubmitProof(
			context.Background(), &proofProducer.ProofWithHeader{
				BlockID: common.Big256,
				Meta:    &bindings.TaikoDataBlockMetadata{},
				Header:  &types.Header{},
				ZkProof: bytes.Repeat([]byte{0xff}, 100),
			},
		),
	)
}

func (s *ProofSubmitterTestSuite) TestValidSubmitProofs() {
	events := helper.ProposeAndInsertEmptyBlocks(&s.ClientTestSuite, s.proposer, s.calldataSyncer)

	for _, e := range events {
		s.Nil(s.validProofSubmitter.RequestProof(context.Background(), e))
		proofWithHeader := <-s.validProofCh
		s.Nil(s.validProofSubmitter.SubmitProof(context.Background(), proofWithHeader))
	}
}

func (s *ProofSubmitterTestSuite) TestValidProofSubmitterRequestProofCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.AfterFunc(2*time.Second, func() {
			cancel()
		})
	}()

	s.ErrorContains(
		s.validProofSubmitter.RequestProof(
			ctx, &bindings.TaikoL1ClientBlockProposed{BlockId: common.Big256}), "context canceled",
	)
}

func TestProofSubmitterTestSuite(t *testing.T) {
	suite.Run(t, new(ProofSubmitterTestSuite))
}
