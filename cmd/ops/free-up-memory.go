package ops

import (
	"fmt"
	"github.com/abdelrahman146/kunai/internal/system"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	"sort"
)

// freeUpMemoryParams holds CLI flags for free-some-space
var freeUpMemoryParams struct {
	Top   int
	Force bool
}

// free-some-space subcommand
var freeUpMemory = &cobra.Command{
	Use:   "free-up-memory",
	Short: "Kill top memory-consuming processes and their children",
	Run:   runFreeSomeSpace,
}

func init() {
	// register flags
	freeUpMemory.Flags().IntVarP(&freeUpMemoryParams.Top, "top", "t", 5, "number of top memory processes to target")
	freeUpMemory.Flags().BoolVarP(&freeUpMemoryParams.Force, "force", "f", false, "force delete processes without user confirmation")
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
	top := freeUpMemoryParams.Top
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
		memMB := fmt.Sprintf("%.2f", float64(mem)/1024/1024/1024)
		fmt.Printf("%d. PID: %d, Name: %s, Memory: %s GB\n", i+1, pid, name, memMB)

		if system.IsSafeToKill(p) {
			targets = append(targets, p)
		} else {
			fmt.Printf("Skipping PID %d (unsafe to kill)\n", pid)
		}
	}

	if !freeUpMemoryParams.Force {
		ok, err := utils.RequestConfirmation("Are you sure you want to kill the processes above and their children?")
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
	for _, p := range targets {
		name, _ := p.Name()
		fmt.Printf("Killing process (PID: %d, Name: %s) and its children if any...\n", p.Pid, name)
		system.KillProcessTree(p)
		fmt.Printf("Killed process (PID: %d, Name: %s)\n", p.Pid, name)
	}

	// Summary
	fmt.Println("Done")
}
