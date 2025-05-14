package cmd

import (
	"fmt"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/spf13/cobra"
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
		panic(err)
	}
	RootCmd.AddCommand(killCmd)
}

func kill(cmd *cobra.Command, args []string) {
	p, err := utils.FindProcessByPort(killCmdParams.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	utils.KillProcessTree(p)
	// Summary
	fmt.Println("Done")
}
