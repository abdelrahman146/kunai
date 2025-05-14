package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

var freeSomeSpaceCmd = &cobra.Command{
	Use:   "free-some-space",
	Short: "Kill top memory-consuming processes",
	Run:   runFreeSomeSpace,
}

var freeSomeSpaceCmdParams = struct {
	Top int
}{}

func init() {
	freeSomeSpaceCmd.Flags().IntVarP(&freeSomeSpaceCmdParams.Top, "top", "t", 5, "top processes")
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

	fmt.Printf("Top %d memory-consuming processes:\n", top)
	var toKill []*process.Process
	for i := 0; i < top; i++ {
		p := pmems[i].proc
		mem := pmems[i].mem
		pid := p.Pid
		name, _ := p.Name()
		fmt.Printf("%d. PID: %d, Name: %s, Memory: %d bytes\n", i+1, pid, name, mem)
		if isSafeToKill(p) {
			toKill = append(toKill, p)
		} else {
			fmt.Printf("Process PID %d is not safe to kill, skipping.\n", pid)
		}
	}

	// Attempt to kill safe processes
	var killed []int32
	for _, p := range toKill {
		err := p.Kill()
		if err != nil {
			fmt.Printf("Failed to kill PID %d: %v\n", p.Pid, err)
		} else {
			killed = append(killed, p.Pid)
			fmt.Printf("Killed PID %d successfully.\n", p.Pid)
		}
		// Small delay to allow OS to reclaim resources
		time.Sleep(100 * time.Millisecond)
	}

	// Summary
	if len(killed) > 0 {
		fmt.Printf("Successfully killed %d processes: %v\n", len(killed), killed)
	} else {
		fmt.Println("No processes were killed.")
	}
}

// isSafeToKill applies basic heuristics to avoid killing critical or own processes
func isSafeToKill(p *process.Process) bool {
	// Avoid killing this CLI itself
	if p.Pid == int32(os.Getpid()) {
		return false
	}
	// Avoid killing root-owned processes
	username, err := p.Username()
	if err == nil && username == "root" {
		return false
	}
	return true
}
