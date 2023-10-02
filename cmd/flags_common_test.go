package main

import "github.com/taikoxyz/taiko-client/tests"

func commonTestArgs(s *tests.ClientTestSuite) map[string]interface{} {
	return map[string]interface{}{
		// required Common flags
		L1WSEndpointFlag.Name:   s.L1.WsEndpoint(),
		TaikoL1AddressFlag.Name: s.L1.TaikoL1Address,
		TaikoL2AddressFlag.Name: tests.TaikoL2Address,
		// optional Common flags
		VerbosityFlag.Name:          "0",
		LogJsonFlag.Name:            "false",
		MetricsEnabledFlag.Name:     "false",
		MetricsAddrFlag.Name:        "",
		BackOffMaxRetriesFlag.Name:  "10",
		RPCTimeoutFlag.Name:         rpcTimeout.String(),
		WaitReceiptTimeoutFlag.Name: "10s",
	}
}
