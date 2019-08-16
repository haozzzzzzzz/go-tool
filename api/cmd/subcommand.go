package cmd

import (
	"github.com/haozzzzzzzz/go-tool/api/cmd/compile"
	_init "github.com/haozzzzzzzz/go-tool/api/cmd/init"
)

// export root cmd
var RootCmd = rootCmd

func init() {
	rootCmd.AddCommand(_init.CommandApiInit())
	rootCmd.AddCommand(_init.CommandApiAddService())
	rootCmd.AddCommand(compile.CommandApiCompile())
	rootCmd.AddCommand(compile.GenerateSwaggerSpecification())
	rootCmd.AddCommand(compile.GenerateCommentSwagger())
}
