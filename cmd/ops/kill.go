package ops

import (
	"fmt"
	"github.com/abdelrahman146/kunai/internal/system"
	"github.com/spf13/cobra"
	"log"
)

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill a process in specific port",
	Run:   kill,
}

var killCmdParams struct {
	Port uint32
}

func init() {
	killCmd.Flags().Uint32VarP(&killCmdParams.Port, "port", "p", 0, "the port of the process to kill")
	if err := killCmd.MarkFlagRequired("port"); err != nil {
		log.Fatalln("port is required")
	}
}

func kill(cmd *cobra.Command, args []string) {
	p, err := system.FindProcessByPort(killCmdParams.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	system.KillProcessTree(p)
	// Summary
	fmt.Println("Done")
}
