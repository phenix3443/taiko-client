package main

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/taikoxyz/taiko-client/proposer"
	"github.com/taikoxyz/taiko-client/testutils"
	"github.com/urfave/cli/v2"
)

var (
	proverEndpoints  = "http://localhost:9876,http://localhost:1234"
	blockProposalFee = "10000000000"
	proposeInterval  = "10s"
)

type proposerCmdSuite struct {
	cmdSuit
}

func (s *proposerCmdSuite) SetupTest() {
	s.cmdSuit.SetupTest()
	proposerConf = &proposer.Config{}
	s.app = cli.NewApp()
	s.app.Flags = proposerFlags
	s.app.Action = func(c *cli.Context) error {
		parseMultiUsedFlags()
		return proposerConf.Validate()
	}
	s.args = s.testArgs()
}

func (s *proposerCmdSuite) TearDownTest() {
	s.cmdSuit.TearDownTest()
}

func TestProposerCmdSuit(t *testing.T) {
	suite.Run(t, new(proposerCmdSuite))
}

func (s *proposerCmdSuite) TestFlags() {
	s.app.After = func(ctx *cli.Context) error {
		s.Equal(s.L1.WsEndpoint(), proposerConf.L1Endpoint)
		s.Equal(s.L2.HttpEndpoint(), proposerConf.L2Endpoint)
		s.Equal(s.L1.TaikoL1Address, proposerConf.TaikoL1Address)
		s.Equal(testutils.TaikoL2Address, proposerConf.TaikoL2Address)
		s.Equal(s.L1.TaikoL1TokenAddress, proposerConf.TaikoTokenAddress)
		s.Equal(float64(10), proposerConf.ProposeInterval.Seconds())
		s.Equal(1, len(proposerConf.LocalAddresses))
		s.Equal(uint64(5), proposerConf.ProposeBlockTxReplacementMultiplier)
		s.Equal(rpcTimeout, *proposerConf.RPCTimeout)
		s.Equal(10*time.Second, proposerConf.WaitReceiptTimeout)
		for i, e := range strings.Split(proverEndpoints, ",") {
			s.Equal(proposerConf.ProverEndpoints[i].String(), e)
		}
		fee, _ := new(big.Int).SetString(blockProposalFee, 10)
		s.Equal(fee, proposerConf.BlockProposalFee)
		s.Equal(uint64(10), proposerConf.BlockProposalFeeIncreasePercentage)
		s.Equal(uint64(100), proposerConf.BlockProposalFeeIterations)
		return nil
	}
	s.NoError(s.app.Run(flagsFromArgs(s.T(), s.args)))
	s.app.After = nil
}

func (s *proposerCmdSuite) TestPrivKeyErr() {
	s.args[L1ProposerPrivKeyFlag.Name] = "0x"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid L1 proposer private key")
}

func (s *proposerCmdSuite) TestL2ReceiptErr() {
	s.args[L2SuggestedFeeRecipientFlag.Name] = "notAnAddress"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid L2 suggested fee recipient address")
}

func (s *proposerCmdSuite) TestTxPoolLocalsErr() {
	s.args[TxPoolLocalsFlag.Name] = "notAnAddress"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid account in --txpool.locals")
}

func (s *proposerCmdSuite) TestBlockTxReplacementMultiplier() {
	s.args[ProposeBlockTxReplacementMultiplierFlag.Name] = "0"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid --proposeBlockTxReplacementMultiplier value")
}

func (s *proposerCmdSuite) testArgs() map[string]interface{} {
	a := map[string]interface{}{
		// proposer flags
		L2HTTPEndpointFlag.Name:                      s.L2.HttpEndpoint(),
		L1ProposerPrivKeyFlag.Name:                   testutils.ProposerPrivateKey,
		L2SuggestedFeeRecipientFlag.Name:             testutils.L2SuggestedFeeRecipientAddress.Hex(),
		ProposeIntervalFlag.Name:                     proposeInterval,
		TxPoolLocalsFlag.Name:                        testutils.ProposerAddress.Hex(),
		TxPoolLocalsOnlyFlag.Name:                    false,
		ProposeEmptyBlocksIntervalFlag.Name:          proposeInterval,
		MaxProposedTxListsPerEpochFlag.Name:          1,
		ProposeBlockTxGasLimitFlag.Name:              "100000",
		ProposeBlockTxReplacementMultiplierFlag.Name: "5",
		ProposeBlockTxGasTipCapFlag.Name:             "100000",
		ProverEndpointsFlag.Name:                     proverEndpoints,
		BlockProposalFeeFlag.Name:                    blockProposalFee,
		BlockProposalFeeIncreasePercentageFlag.Name:  "10",
		BlockProposalFeeIterationsFlag.Name:          "100",
		TaikoTokenAddressFlag.Name:                   s.L1.TaikoL1TokenAddress,
	}
	return mergeArgs(commonTestArgs(&s.ClientTestSuite), a)
}
