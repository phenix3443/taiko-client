package prover

import (
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"

	"github.com/taikoxyz/taiko-client/driver"
	"github.com/taikoxyz/taiko-client/pkg/jwt"
	"github.com/taikoxyz/taiko-client/pkg/rpc"
	"github.com/taikoxyz/taiko-client/proposer"
	producer "github.com/taikoxyz/taiko-client/prover/proof_producer"
	"github.com/taikoxyz/taiko-client/tests"
	"github.com/taikoxyz/taiko-client/tests/helper"
	"github.com/taikoxyz/taiko-mono/packages/protocol/bindings"
)

type ProverTestSuite struct {
	tests.ClientTestSuite
	p         *Prover
	cancel    context.CancelFunc
	d         *driver.Driver
	rpcClient *rpc.Client
	proposer  *proposer.Proposer
}

func (s *ProverTestSuite) SetupTest() {
	s.ClientTestSuite.SetupTest()

	// Init prover
	l1ProverPrivKey := tests.ProverPrivKey

	proverServerUrl := helper.LocalRandomProverEndpoint()
	port, err := strconv.Atoi(proverServerUrl.Port())
	s.NoError(err)

	ctx, cancel := context.WithCancel(context.Background())

	cfg := &Config{
		L1WsEndpoint:                    s.L1.WsEndpoint(),
		L1HttpEndpoint:                  s.L1.HttpEndpoint(),
		L2WsEndpoint:                    s.L2.WsEndpoint(),
		L2HttpEndpoint:                  s.L2.HttpEndpoint(),
		TaikoL1Address:                  s.L1.TaikoL1Address,
		TaikoL2Address:                  tests.TaikoL2Address,
		L1ProverPrivKey:                 l1ProverPrivKey,
		OracleProverPrivateKey:          l1ProverPrivKey,
		OracleProver:                    false,
		Dummy:                           true,
		MaxConcurrentProvingJobs:        1,
		CheckProofWindowExpiredInterval: 5 * time.Second,
		ProveUnassignedBlocks:           true,
		Capacity:                        1024,
		MinProofFee:                     common.Big1,
		HTTPServerPort:                  uint64(port),
	}
	s.p, err = New(ctx, cfg)
	s.NoError(err)
	jwtSecret, err := jwt.ParseSecretFromFile(tests.JwtSecretFile)
	s.NoError(err)
	s.NotEmpty(jwtSecret)
	s.rpcClient = helper.NewWsRpcClient(&s.ClientTestSuite)
	protocolConfigs, err := s.rpcClient.TaikoL1.GetConfig(nil)
	s.NoError(err)
	s.p.srv, err = helper.NewFakeProver(s.L1.TaikoL1Address, &protocolConfigs, jwtSecret,
		s.rpcClient, l1ProverPrivKey, s.p.capacityManager, proverServerUrl)
	s.NoError(err)
	s.cancel = cancel

	// Init driver

	dCfg := &driver.Config{
		L1Endpoint:       s.L1.WsEndpoint(),
		L2Endpoint:       s.L2.WsEndpoint(),
		L2EngineEndpoint: s.L2.AuthEndpoint(),
		TaikoL1Address:   s.L1.TaikoL1Address,
		TaikoL2Address:   tests.TaikoL2Address,
		JwtSecret:        string(jwtSecret),
	}
	s.d, err = driver.New(ctx, dCfg)
	s.NoError(err)

	// Init proposer
	l1ProposerPrivKey := tests.ProposerPrivKey
	s.Nil(err)

	proposeInterval := 1024 * time.Hour // No need to periodically propose transactions list in unit tests
	pCfg := &proposer.Config{
		L1Endpoint:                         s.L1.WsEndpoint(),
		L2Endpoint:                         s.L2.WsEndpoint(),
		TaikoL1Address:                     s.L1.TaikoL1Address,
		TaikoL2Address:                     tests.TaikoL2Address,
		TaikoTokenAddress:                  s.L1.TaikoL1TokenAddress,
		L1ProposerPrivKey:                  l1ProposerPrivKey,
		L2SuggestedFeeRecipient:            tests.ProposerAddress,
		ProposeInterval:                    &proposeInterval,
		MaxProposedTxListsPerEpoch:         1,
		WaitReceiptTimeout:                 10 * time.Second,
		ProverEndpoints:                    []*url.URL{proverServerUrl},
		BlockProposalFee:                   common.Big256,
		BlockProposalFeeIterations:         3,
		BlockProposalFeeIncreasePercentage: 2,
	}
	s.proposer, err = proposer.New(context.Background(), pCfg)
	s.NoError(err)
}

func (s *ProverTestSuite) TearDownTest() {
	s.proposer.Close(context.Background())
	s.d.Close(context.Background())
	s.p.Close(context.Background())
	s.rpcClient.Close()
	s.ClientTestSuite.TearDownTest()
}

func (s *ProverTestSuite) TestName() {
	s.Equal("prover", s.p.Name())
}

func (s *ProverTestSuite) TestOnBlockProposed() {
	s.p.cfg.OracleProver = true
	// Init prover
	l1ProverPrivKey := tests.ProverPrivKey
	s.p.cfg.OracleProverPrivateKey = l1ProverPrivKey
	// Valid block
	e := helper.ProposeAndInsertValidBlock(&s.ClientTestSuite, s.proposer, s.d.ChainSyncer().CalldataSyncer())
	s.Nil(s.p.onBlockProposed(context.Background(), e, func() {}))
	s.Nil(s.p.validProofSubmitter.SubmitProof(context.Background(), <-s.p.proofGenerationCh))

	// Empty blocks
	for _, e = range helper.ProposeAndInsertEmptyBlocks(
		&s.ClientTestSuite,
		s.proposer,
		s.d.ChainSyncer().CalldataSyncer(),
	) {
		s.Nil(s.p.onBlockProposed(context.Background(), e, func() {}))

		s.Nil(s.p.validProofSubmitter.SubmitProof(context.Background(), <-s.p.proofGenerationCh))
	}
}

func (s *ProverTestSuite) TestOnBlockVerifiedEmptyBlockHash() {
	s.Nil(s.p.onBlockVerified(context.Background(), &bindings.TaikoL1ClientBlockVerified{
		BlockId:   common.Big1,
		BlockHash: common.Hash{},
	},
	))
}

func (s *ProverTestSuite) TestSubmitProofOp() {
	s.NotPanics(func() {
		s.p.submitProofOp(context.Background(), &producer.ProofWithHeader{
			BlockID: common.Big1,
			Meta:    &bindings.TaikoDataBlockMetadata{},
			Header:  &types.Header{},
			ZkProof: []byte{},
		})
	})
	s.NotPanics(func() {
		s.p.submitProofOp(context.Background(), &producer.ProofWithHeader{
			BlockID: common.Big1,
			Meta:    &bindings.TaikoDataBlockMetadata{},
			Header:  &types.Header{},
			ZkProof: []byte{},
		})
	})
}

func (s *ProverTestSuite) TestStartSubscription() {
	s.NotPanics(s.p.initSubscription)
	s.NotPanics(s.p.closeSubscription)
}

func (s *ProverTestSuite) TestCheckChainVerification() {
	s.Nil(s.p.checkChainVerification(0))
	s.p.latestVerifiedL1Height = 1024
	s.Nil(s.p.checkChainVerification(1024))
}

func TestProverTestSuite(t *testing.T) {
	suite.Run(t, new(ProverTestSuite))
}
