package ai

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "ai",
	Short: `Kunai Assistant`,
	Long:  `Kunai Assistant Utilities`,
}

func init() {
	Cmd.AddCommand(chatCmd)
}
