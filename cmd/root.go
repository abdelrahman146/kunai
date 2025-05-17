package cmd

import (
	"github.com/abdelrahman146/kunai/cmd/ai"
	"github.com/abdelrahman146/kunai/cmd/codebase"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kunai",
	Short: "Kunai CLI tool",
}

func init() {
	RootCmd.AddCommand(freeSomeSpaceCmd)
	RootCmd.AddCommand(killCmd)
	RootCmd.AddCommand(codebase.Cmd)
	RootCmd.AddCommand(ai.Cmd)
}
