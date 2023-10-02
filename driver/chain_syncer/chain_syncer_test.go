package chainSyncer

import (
	"context"
	"math/big"
	"net/url"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/suite"

	"github.com/taikoxyz/taiko-client/driver/state"
	"github.com/taikoxyz/taiko-client/pkg/rpc"
	"github.com/taikoxyz/taiko-client/proposer"
	"github.com/taikoxyz/taiko-client/prover/server"
	"github.com/taikoxyz/taiko-client/tests"
	"github.com/taikoxyz/taiko-client/tests/helper"
)

type ChainSyncerTestSuite struct {
	tests.ClientTestSuite
	s               *L2ChainSyncer
	snapshotID      string
	p               tests.Proposer
	rpcClient       *rpc.Client
	proverEndpoints []*url.URL
	proverServer    *server.ProverServer
}

func (s *ChainSyncerTestSuite) SetupTest() {
	s.ClientTestSuite.SetupTest()
	s.rpcClient = helper.NewWsRpcClient(&s.ClientTestSuite)
	state, err := state.New(context.Background(), s.rpcClient)
	s.Nil(err)
	ctx := context.Background()
	syncer, err := New(
		ctx,
		s.rpcClient,
		state,
		false,
		1*time.Hour,
		s.L1.TaikoL1SignalService,
	)
	s.Nil(err)
	s.s = syncer

	l1ProposerPrivKey := tests.ProposerPrivKey
	proposeInterval := 1024 * time.Hour // No need to periodically propose transactions list in unit tests

	s.proverEndpoints, s.proverServer, err = helper.DefaultFakeProver(&s.ClientTestSuite, s.rpcClient)
	s.NoError(err)

	cfg := &proposer.Config{
		L1Endpoint:                         s.L1.WsEndpoint(),
		L2Endpoint:                         s.L2.WsEndpoint(),
		TaikoL1Address:                     s.L1.TaikoL1Address,
		TaikoL2Address:                     tests.TaikoL2Address,
		TaikoTokenAddress:                  s.L1.TaikoL1TokenAddress,
		L1ProposerPrivKey:                  l1ProposerPrivKey,
		L2SuggestedFeeRecipient:            tests.L2SuggestedFeeRecipientAddress,
		ProposeInterval:                    &proposeInterval,
		MaxProposedTxListsPerEpoch:         1,
		WaitReceiptTimeout:                 10 * time.Second,
		ProverEndpoints:                    s.proverEndpoints,
		BlockProposalFee:                   big.NewInt(1000),
		BlockProposalFeeIterations:         3,
		BlockProposalFeeIncreasePercentage: 2,
	}
	s.p, err = proposer.New(ctx, cfg)
	s.NoError(err)
}

func (s *ChainSyncerTestSuite) TearDownTest() {
	_ = s.proverServer.Shutdown(context.Background())
	s.rpcClient.Close()
	s.ClientTestSuite.TearDownTest()
}

func (s *ChainSyncerTestSuite) TestGetInnerSyncers() {
	s.NotNil(s.s.BeaconSyncer())
	s.NotNil(s.s.CalldataSyncer())
}

func (s *ChainSyncerTestSuite) TestSync() {
	head, err := s.rpcClient.L1.HeaderByNumber(context.Background(), nil)
	s.Nil(err)
	s.Nil(s.s.Sync(head))
}

func (s *ChainSyncerTestSuite) TestAheadOfProtocolVerifiedHead2() {
	s.TakeSnapshot()
	// propose a couple blocks
	helper.ProposeAndInsertEmptyBlocks(&s.ClientTestSuite, s.p, s.s.calldataSyncer)

	// NOTE: need to prove the proposed blocks to be verified, writing helper function
	// generate bind.TransactOpts to interact with TaikoL1 contract with.
	privKey := tests.ProverPrivKey
	opts, err := bind.NewKeyedTransactorWithChainID(privKey, s.rpcClient.L1ChainID)
	s.Nil(err)

	head, err := s.rpcClient.L1.HeaderByNumber(context.Background(), nil)
	s.Nil(err)

	l2Head, err := s.rpcClient.L2.HeaderByNumber(context.Background(), nil)
	s.Nil(err)

	log.Info("L1HeaderByNumber head", "number", head.Number)
	// (equiv to s.state.GetL2Head().Number)
	log.Info("L2HeaderByNumber head", "number", l2Head.Number)
	log.Info("LatestVerifiedBlock number", "number", s.s.state.GetLatestVerifiedBlock().ID.Uint64())

	config, err := s.s.rpc.TaikoL1.GetConfig(&bind.CallOpts{})
	s.Nil(err)

	// increase evm time to make blocks verifiable.
	var result uint64
	s.Nil(s.rpcClient.L1RawRPC.CallContext(
		context.Background(),
		&result,
		"evm_increaseTime",
		config.ProofRegularCooldown.Uint64()))
	s.NotNil(result)
	log.Info("EVM time increase", "number", result)

	// interact with TaikoL1 contract to allow for verification of L2 blocks
	tx, err := s.s.rpc.TaikoL1.VerifyBlocks(opts, uint64(3))
	s.Nil(err)
	s.NotNil(tx)

	head2, err := s.rpcClient.L1.HeaderByNumber(context.Background(), nil)
	s.Nil(err)

	l2Head2, err := s.rpcClient.L2.HeaderByNumber(context.Background(), nil)
	s.Nil(err)

	log.Info("L1HeaderByNumber head2", "number", head2.Number)
	log.Info("L2HeaderByNumber head", "number", l2Head2.Number)
	log.Info("LatestVerifiedBlock number", "number", s.s.state.GetLatestVerifiedBlock().ID.Uint64())

	s.RevertSnapshot()
}

func TestChainSyncerTestSuite(t *testing.T) {
	suite.Run(t, new(ChainSyncerTestSuite))
}

func (s *ChainSyncerTestSuite) TakeSnapshot() {
	// record snapshot state to revert to before changes
	s.Nil(s.rpcClient.L1RawRPC.CallContext(context.Background(), &s.snapshotID, "evm_snapshot"))
}

func (s *ChainSyncerTestSuite) RevertSnapshot() {
	// revert to the snapshot state so protocol configs are unaffected
	var revertRes bool
	s.Nil(s.rpcClient.L1RawRPC.CallContext(context.Background(), &revertRes, "evm_revert", s.snapshotID))
	s.True(revertRes)
	s.Nil(rpc.SetHead(context.Background(), s.rpcClient.L2RawRPC, common.Big0))
}

func (s *ChainSyncerTestSuite) TestAheadOfProtocolVerifiedHead() {
	s.True(s.s.AheadOfProtocolVerifiedHead())
}
