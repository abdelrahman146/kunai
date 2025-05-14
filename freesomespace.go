package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

// freeSomeSpaceCmdParams holds CLI flags for free-some-space
var freeSomeSpaceCmdParams = struct {
	Top   int
	Force bool
}{
	Top:   5,
	Force: false,
}

// free-some-space subcommand
var freeSomeSpaceCmd = &cobra.Command{
	Use:   "free-some-space",
	Short: "Kill top memory-consuming processes and their children",
	Run:   runFreeSomeSpace,
}

func init() {
	// register flags
	freeSomeSpaceCmd.Flags().IntVarP(&freeSomeSpaceCmdParams.Top, "top", "t", freeSomeSpaceCmdParams.Top, "number of top memory processes to target")
	freeSomeSpaceCmd.Flags().BoolVarP(&freeSomeSpaceCmdParams.Force, "force", "f", freeSomeSpaceCmdParams.Force, "force delete processes without user confirmation")
	rootCmd.AddCommand(freeSomeSpaceCmd)
}

type processInfo struct {
	proc *process.Process
	mem  uint64
}

func runFreeSomeSpace(cmd *cobra.Command, args []string) {
	// Fetch all processes
	processes, err := process.Processes()
	if err != nil {
		fmt.Printf("Error fetching processes: %v\n", err)
		return
	}
	var infos []processInfo
	for _, p := range processes {
		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		infos = append(infos, processInfo{proc: p, mem: memInfo.RSS})
	}

	// Sort by memory usage descending
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].mem > infos[j].mem
	})

	// Determine how many to handle
	top := freeSomeSpaceCmdParams.Top
	if len(infos) < top {
		top = len(infos)
	}

	// Print the top processes
	fmt.Printf("Top %d memory-consuming processes:\n", top)
	var targets []*process.Process
	for i := 0; i < top; i++ {
		p := infos[i].proc
		mem := infos[i].mem
		pid := p.Pid
		name, _ := p.Name()
		fmt.Printf("%d. PID: %d, Name: %s, Memory: %d bytes\n", i+1, pid, name, mem)

		if isSafeToKill(p) {
			targets = append(targets, p)
		} else {
			fmt.Printf("Skipping PID %d (unsafe to kill)\n", pid)
		}
	}

	if !freeSomeSpaceCmdParams.Force {
		ok, err := WaitForConfirmation("Are you sure you want to kill the processes above and their children?")
		if err != nil {
			fmt.Printf("Failed to read user input: %v\n", err)
			return
		}
		if !ok {
			fmt.Println("Aborting...")
			return
		}
	}

	// Kill each process tree
	var killed []int32
	for _, p := range targets {
		name, _ := p.Name()
		fmt.Printf("Killing process (PID: %d, Name: %s) and its children if any...\n", p.Pid, name)
		killProcessTree(p, &killed)
		fmt.Printf("Killed process (PID: %d, Name: %s)\n", p.Pid, name)
	}

	// Summary
	if len(killed) > 0 {
		fmt.Printf("Successfully killed %d processes\n", len(killed))
	} else {
		fmt.Println("No processes were killed.")
	}
}

// killProcessTree recursively kills a process and its children
func killProcessTree(p *process.Process, killed *[]int32) {
	// Kill children first
	children, err := p.Children()
	if err == nil {
		for _, c := range children {
			killProcessTree(c, killed)
		}
	}

	// Kill this process
	err = p.Kill()
	if err != nil {
		fmt.Printf("--- Failed to kill PID %d: %v\n", p.Pid, err)
	} else {
		*killed = append(*killed, p.Pid)
		fmt.Printf("--- Killed PID %d\n", p.Pid)
	}

	// Small delay to allow OS to reclaim resources
	time.Sleep(100 * time.Millisecond)
}

// isSafeToKill avoids killing critical or the CLI itself
func isSafeToKill(p *process.Process) bool {
	// Never kill this CLI
	if p.Pid == int32(os.Getpid()) {
		return false
	}
	// Avoid root-owned
	username, err := p.Username()
	if err == nil && username == "root" {
		return false
	}
	return true
}
