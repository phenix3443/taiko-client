package main

import (
	"time"

	"github.com/taikoxyz/taiko-client/testutils"
	"github.com/urfave/cli/v2"
)

var rpcTimeout = 5 * time.Second

type cmdSuit struct {
	testutils.ClientTestSuite
	app  *cli.App
	args map[string]interface{}
}

func (s *cmdSuit) SetupTest() {
	s.ClientTestSuite.SetupTest()
}

func (s *cmdSuit) TearDownTest() {
	s.ClientTestSuite.TearDownTest()
}

func mergeArgs(a, b map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range a {
		m[k] = v
	}
	for k, v := range b {
		m[k] = v
	}
	return m
}
