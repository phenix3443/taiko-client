package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v2"

	"github.com/taikoxyz/taiko-client/driver"
	"github.com/taikoxyz/taiko-client/tests"
)

type driverCmdSuite struct {
	cmdSuit
}

func (s *driverCmdSuite) SetupTest() {
	s.cmdSuit.SetupTest()
	driverConf = &driver.Config{}
	s.app = cli.NewApp()
	s.app.Flags = driverFlags
	s.app.Action = func(ctx *cli.Context) error {
		parseMultiUsedFlags()
		return driverConf.Validate(context.Background())
	}
	s.args = s.testArgs()
}

func (s *driverCmdSuite) TearDownTest() {
	s.cmdSuit.TearDownTest()
}

func TestDriverCmdSuit(t *testing.T) {
	suite.Run(t, new(driverCmdSuite))
}

func (s *driverCmdSuite) testArgs() map[string]interface{} {
	a := map[string]interface{}{
		L2WSEndpointFlag.Name:          s.L2.WsEndpoint(),
		L2AuthEndpointFlag.Name:        s.L2.AuthEndpoint(),
		JWTSecretFlag.Name:             tests.JwtSecretFile,
		P2PSyncVerifiedBlocksFlag.Name: true,
		P2PSyncTimeoutFlag.Name:        "120s",
		CheckPointSyncUrlFlag.Name:     "http://localhost:8545",
	}
	return mergeArgs(commonTestArgs(&s.ClientTestSuite), a)
}

func (s *driverCmdSuite) TestParseFlags() {
	s.app.After = func(ctx *cli.Context) error {
		s.Equal(s.L1.WsEndpoint(), driverConf.L1Endpoint)
		s.Equal(s.L2.WsEndpoint(), driverConf.L2Endpoint)
		s.Equal(s.L2.AuthEndpoint(), driverConf.L2EngineEndpoint)
		s.Equal(s.L1.TaikoL1Address, driverConf.TaikoL1Address)
		s.Equal(tests.TaikoL2Address, driverConf.TaikoL2Address)
		s.Equal(120*time.Second, driverConf.P2PSyncTimeout)
		s.Equal(rpcTimeout, *driverConf.RPCTimeout)
		s.NotEmpty(driverConf.JwtSecret)
		s.True(driverConf.P2PSyncVerifiedBlocks)
		s.Equal("http://localhost:8545", driverConf.L2CheckPoint)
		return nil
	}
	s.NoError(s.app.Run(flagsFromArgs(s.T(), s.args)))
	s.app.After = nil
}

func (s *driverCmdSuite) TestJWTError() {
	s.args[JWTSecretFlag.Name] = "wrongsecretfile.txt"
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "invalid JWT secret file")
}

func (s *driverCmdSuite) TestEmptyL2CheckPoint() {
	delete(s.args, CheckPointSyncUrlFlag.Name)
	s.ErrorContains(s.app.Run(flagsFromArgs(s.T(), s.args)), "empty L2 check point URL")
}

func flagsFromArgs(t *testing.T, args map[string]interface{}) []string {
	flags := []string{t.Name()}
	for k, v := range args {
		flags = append(flags, fmt.Sprintf("--%s=%v", k, v))
	}
	return flags
}
