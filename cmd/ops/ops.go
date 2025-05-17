package ops

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "ops",
	Short: `System Operations Utilities`,
	Long:  `System Operations Utilities, such as kill process, free memory space...`,
}

func init() {
	Cmd.AddCommand(freeUpMemory)
	Cmd.AddCommand(killCmd)
}
