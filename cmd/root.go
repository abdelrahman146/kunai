package cmd

import (
	"github.com/abdelrahman146/kunai/cmd/codebase"
	"github.com/abdelrahman146/kunai/cmd/ops"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kunai",
	Short: "🗡 Kunai CLI",
	Long:  `Kunai CLI, your development swiss knife! a cli tool that aims to improve your development productivity`,
}

func init() {
	RootCmd.AddCommand(ops.Cmd)
	RootCmd.AddCommand(codebase.Cmd)
}
