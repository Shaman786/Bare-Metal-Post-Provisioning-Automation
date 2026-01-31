package state

import (
	"encoding/json"
	"log"
	"os"
	"syscall"
)

type State struct {
	AssignedIPs []string `json:"assigned_ips"`
}

// Manager handles the state file, including locking and persistence.
type Manager struct {
	file  *os.File
	State State
}

// Load opens the state file, acquires an exclusive lock, and loads the current state.
// It terminates the program on error, matching the opinionated nature of this CLI tool.
func Load(path string) *Manager {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("failed to open state file: %v", err)
	}

	if err := lockFile(f); err != nil {
		log.Fatalf("failed to lock state file: %v", err)
	}

	var state State
	info, _ := f.Stat()
	if info.Size() > 0 {
		if err := json.NewDecoder(f).Decode(&state); err != nil {
			// If we can't read the state, it's safer to fail than separate
			log.Fatalf("failed to decode state: %v", err)
		}
	}

	return &Manager{
		file:  f,
		State: state,
	}
}

// Save writes the current state to the file.
func (m *Manager) Save() {
	m.file.Truncate(0)
	m.file.Seek(0, 0)
	if err := json.NewEncoder(m.file).Encode(m.State); err != nil {
		log.Fatalf("failed to save state: %v", err)
	}
}

// Close releases the lock and closes the file.
func (m *Manager) Close() {
	unlockFile(m.file)
	m.file.Close()
}

func lockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

func (m *Manager) Release(state *State, ip string) {
	filtered := make([]string, 0, len(state.AssignedIPs))

	found := false
	for _, assigned := range state.AssignedIPs {
		if assigned == ip {
			found = true
			continue
		}
		filtered = append(filtered, assigned)
	}

	if !found {
		log.Fatalf("IP %s is not currently assigned", ip)
	}

	state.AssignedIPs = filtered
}
