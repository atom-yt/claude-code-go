package memory

import (
	"os"
	"time"
)

// LockState represents the current state of the consolidation lock.
type LockState struct {
	LastConsolidation time.Time // Time of last consolidation (from .consolidate-lock mtime)
	IsRunning         bool      // Whether consolidation is currently running
}

// GetLockState reads the current lock state.
func GetLockState() (LockState, error) {
	state := LockState{}

	lockPath, err := LockFilePath()
	if err != nil {
		return state, err
	}

	// Check .consolidate-lock.running
	runningPath, err := RunningLockFilePath()
	if err != nil {
		return state, err
	}
	if _, err := os.Stat(runningPath); err == nil {
		state.IsRunning = true
	}

	// Get mtime of .consolidate-lock
	info, err := os.Stat(lockPath)
	if err == nil {
		state.LastConsolidation = info.ModTime()
	} else if !os.IsNotExist(err) {
		return state, err
	}

	return state, nil
}

// AcquireLock attempts to acquire the consolidation lock.
// Returns true if successful, false if lock is held by another process.
func AcquireLock() (bool, error) {
	runningPath, err := RunningLockFilePath()
	if err != nil {
		return false, err
	}

	// Try to create .consolidate-lock.running with exclusive access
	// This will fail if the file already exists
	f, err := os.OpenFile(runningPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return false, nil // Lock is held by another process
		}
		return false, err
	}
	f.Close()
	return true, nil
}

// ReleaseLock releases the consolidation lock.
func ReleaseLock() error {
	runningPath, err := RunningLockFilePath()
	if err != nil {
		return err
	}

	// Remove .consolidate-lock.running
	if err := os.Remove(runningPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Update .consolidate-lock mtime to current time
	lockPath, err := LockFilePath()
	if err != nil {
		return err
	}

	// Touch the lock file to update mtime
	now := time.Now()
	if err := os.Chtimes(lockPath, now, now); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
