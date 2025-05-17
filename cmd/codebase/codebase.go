package codebase

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "codebase",
	Short: `Code base utilities`,
	Long:  `Code base utilities, such as search, ai assistant, create PR, generate commits...`,
}

const (
	alias = "codebase"
)

func init() {
	Cmd.AddCommand(indexCmd)
	Cmd.AddCommand(searchCmd)
	Cmd.AddCommand(statsCmd)
	Cmd.AddCommand(chatCmd)
}
