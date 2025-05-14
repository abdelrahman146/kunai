package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"time"
)

// KillProcessTree recursively kills a process and its children
func KillProcessTree(p *process.Process) {
	// Kill children first
	children, err := p.Children()
	if err == nil {
		for _, c := range children {
			KillProcessTree(c)
		}
	}

	// Kill this process
	name, _ := p.Name()
	err = p.Kill()
	if err != nil {
		fmt.Printf("Failed to kill (PID: %d, Name: %s): %v\n", p.Pid, name, err)
	} else {
		fmt.Printf("Killed (PID: %d, Name: %s) \n", p.Pid, name)
	}

	// Small delay to allow OS to reclaim resources
	time.Sleep(100 * time.Millisecond)
}

// IsSafeToKill avoids killing critical or the CLI itself
func IsSafeToKill(p *process.Process) bool {
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

// FindProcessByPort scans all processes and returns the first one
// that has a connection with Laddr.Port == port.
func FindProcessByPort(port uint32) (*process.Process, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("could not list processes: %w", err)
	}
	for _, p := range procs {
		conns, err := p.Connections()
		if err != nil {
			// often happens if you lack permissions on some pids
			continue
		}
		for _, c := range conns {
			if c.Laddr.Port == port {
				return p, nil
			}
		}
	}
	return nil, fmt.Errorf("no process found using port %d", port)
}
