package cmd

import (
	"github.com/haozzzzzzzz/go-tool/api/cmd/compile"
	"github.com/haozzzzzzzz/go-tool/api/cmd/summary"
)

// export root cmd
var RootCmd = rootCmd

func init() {
	//rootCmd.AddCommand(_init.CommandApiInit())
	//rootCmd.AddCommand(_init.CommandApiAddService())
	rootCmd.AddCommand(compile.CommandApiCompile())
	rootCmd.AddCommand(compile.GenerateSwaggerSpecification())
	rootCmd.AddCommand(summary.CommandVersion())
	//rootCmd.AddCommand(compile.GenerateCommentSwagger())
}
