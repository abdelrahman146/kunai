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
	Top int
}{
	Top: 5,
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
	rootCmd.AddCommand(freeSomeSpaceCmd)
}

func runFreeSomeSpace(cmd *cobra.Command, args []string) {
	// Fetch all processes
	procs, err := process.Processes()
	if err != nil {
		fmt.Printf("Error fetching processes: %v\n", err)
		return
	}

	// Collect memory usage for each process
	type procMem struct {
		proc *process.Process
		mem  uint64
	}
	var pmems []procMem
	for _, p := range procs {
		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		pmems = append(pmems, procMem{proc: p, mem: memInfo.RSS})
	}

	// Sort by memory usage descending
	sort.Slice(pmems, func(i, j int) bool {
		return pmems[i].mem > pmems[j].mem
	})

	// Determine how many to handle
	top := freeSomeSpaceCmdParams.Top
	if len(pmems) < top {
		top = len(pmems)
	}

	// Print the top processes
	fmt.Printf("Top %d memory-consuming processes:\n", top)
	var targets []*process.Process
	for i := 0; i < top; i++ {
		p := pmems[i].proc
		mem := pmems[i].mem
		pid := p.Pid
		name, _ := p.Name()
		fmt.Printf("%d. PID: %d, Name: %s, Memory: %d bytes\n", i+1, pid, name, mem)

		if isSafeToKill(p) {
			targets = append(targets, p)
		} else {
			fmt.Printf("Skipping PID %d (unsafe to kill)\n", pid)
		}
	}

	// Kill each process tree
	var killed []int32
	for _, p := range targets {
		fmt.Printf("Killing process tree rooted at PID %d...\n", p.Pid)
		killProcessTree(p, &killed)
	}

	// Summary
	if len(killed) > 0 {
		fmt.Printf("Successfully killed %d processes: %v\n", len(killed), killed)
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
		fmt.Printf("Failed to kill PID %d: %v\n", p.Pid, err)
	} else {
		*killed = append(*killed, p.Pid)
		fmt.Printf("Killed PID %d\n", p.Pid)
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
