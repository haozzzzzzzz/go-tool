package cmd

import (
	"github.com/haozzzzzzzz/go-tool/code/cmd/precompile"
)

var RootCmd = rootCmd

func init() {
	RootCmd.AddCommand(precompile.CommandPrecompile())
}
