package main

import (
	"fmt"
	"github.com/abdelrahman146/kunai/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
