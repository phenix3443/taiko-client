package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/taikoxyz/taiko-client/prover"
	"github.com/taikoxyz/taiko-client/testutils"
	"github.com/urfave/cli/v2"
)

var minProofFee = "1024"

type proverCmdSuite struct {
	cmdSuit
}

func (s *proverCmdSuite) SetupTest() {
	s.cmdSuit.SetupTest()
	proverConf = &prover.Config{}
	s.app = cli.NewApp()
	s.app.Flags = proverFlags
	s.app.Action = func(c *cli.Context) error {
		parseMultiUsedFlags()
		return proverConf.Validate()
	}
	s.args = s.testArgs()
}

func (s *proverCmdSuite) TearDownTest() {
	s.cmdSuit.TearDownTest()
}

func TestProverCmdSuit(t *testing.T) {
	suite.Run(t, new(proverCmdSuite))
}

func (s *proverCmdSuite) TestOracleProver() {
	s.app.After = func(ctx *cli.Context) error {
		s.Equal(s.L1.WsEndpoint(), proverConf.L1WsEndpoint)
		s.Equal(s.L1.HttpEndpoint(), proverConf.L1HttpEndpoint)
		s.Equal(s.L2.WsEndpoint(), proverConf.L2WsEndpoint)
		s.Equal(s.L2.HttpEndpoint(), proverConf.L2HttpEndpoint)
		s.Equal(s.L1.TaikoL1Address, proverConf.TaikoL1Address)
		s.Equal(testutils.TaikoL2Address, proverConf.TaikoL2Address)
		s.Equal(30*time.Minute, *proverConf.RandomDummyProofDelayLowerBound)
		s.Equal(time.Hour, *proverConf.RandomDummyProofDelayUpperBound)
		s.True(proverConf.Dummy)
		s.True(proverConf.OracleProver)
		s.Equal("", proverConf.Graffiti)
		s.Equal(30*time.Second, proverConf.CheckProofWindowExpiredInterval)
		s.True(proverConf.ProveUnassignedBlocks)
		s.Equal(rpcTimeout, *proverConf.RPCTimeout)
		s.Equal(uint64(8), proverConf.Capacity)
		s.Equal(minProofFee, proverConf.MinProofFee.String())
		s.Equal(uint64(3), proverConf.ProveBlockTxReplacementMultiplier)
		s.Equal(uint64(256), proverConf.ProveBlockMaxTxGasTipCap.Uint64())
		s.Equal(15*time.Second, proverConf.TempCapacityExpiresAt)
		return nil
	}

	s.NoError(s.app.Run(flagsFromArgs(s.T(), s.args)))
	s.app.After = nil
}

func (s *proverCmdSuite) TestOracleProverError() {
	s.args[OracleProverFlag.Name] = true
	delete(s.args, OracleProverPrivateKeyFlag.Name)
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "oracleProver flag set without oracleProverPrivateKey set")
}

func (s *proverCmdSuite) TestProverKeyError() {
	s.args[L1ProverPrivKeyFlag.Name] = "0x"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid L1 prover private key")
}

func (s *proverCmdSuite) TestOracleProverKeyError() {
	s.args[OracleProverFlag.Name] = true
	s.args[OracleProverPrivateKeyFlag.Name] = ""
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid oracle private key")
}

func (s *proverCmdSuite) TestNewConfigFromCliContext_RandomDelayError() {
	s.args[RandomDummyProofDelayFlag.Name] = "130m"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid random dummy proof delay value")
}

func (s *proverCmdSuite) TestNewConfigFromCliContext_RandomDelayErrorLower() {
	s.args[RandomDummyProofDelayFlag.Name] = "30x-1h"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid random dummy proof delay value")
}

func (s *proverCmdSuite) TestRandomDelayErrorUpper() {
	s.args[RandomDummyProofDelayFlag.Name] = "30m-1x"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid random dummy proof delay value")
}

func (s *proverCmdSuite) TestRandomDelayErrorOrder() {
	s.args[RandomDummyProofDelayFlag.Name] = "1h-30m"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid random dummy proof delay value (lower > upper)")
}

func (s *proverCmdSuite) testArgs() map[string]interface{} {
	a := map[string]interface{}{
		L1HTTPEndpointFlag.Name:                    s.L1.HttpEndpoint(),
		L2WSEndpointFlag.Name:                      s.L2.WsEndpoint(),
		L2HTTPEndpointFlag.Name:                    s.L2.HttpEndpoint(),
		ZkEvmRpcdEndpointFlag.Name:                 os.Getenv("ZK_EVM_RPCD_ENDPOINT"),
		ZkEvmRpcdParamsPathFlag.Name:               os.Getenv("ZK_EVM_RPCD_PARAMS_PATH"),
		L1ProverPrivKeyFlag.Name:                   testutils.ProverPrivateKey,
		MinProofFeeFlag.Name:                       minProofFee,
		StartingBlockIDFlag.Name:                   0,
		MaxConcurrentProvingJobsFlag.Name:          1,
		DummyFlag.Name:                             true,
		RandomDummyProofDelayFlag.Name:             "30m-1h",
		OracleProverFlag.Name:                      true,
		OracleProverPrivateKeyFlag.Name:            testutils.ProverPrivateKey,
		OracleProofSubmissionDelayFlag.Name:        "10s",
		ProofSubmissionMaxRetryFlag.Name:           3,
		ProveBlockTxReplacementMultiplierFlag.Name: 3,
		ProveBlockMaxTxGasTipCapFlag.Name:          256,
		GraffitiFlag.Name:                          "",
		CheckProofWindowExpiredIntervalFlag.Name:   "30s",
		ProveUnassignedBlocksFlag.Name:             true,
		ProverCapacityFlag.Name:                    8,
		TempCapacityExpiresAtFlag.Name:             "15s",
		TaikoTokenAddressFlag.Name:                 s.L1.TaikoL1TokenAddress.Hex(),
	}
	return mergeArgs(commonTestArgs(&s.ClientTestSuite), a)
}
