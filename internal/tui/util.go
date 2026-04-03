package tui

import "os"

// getCWD returns the current working directory, or empty string on error.
func getCWD() (string, error) {
	return os.Getwd()
}
