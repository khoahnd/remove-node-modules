//go:build !windows

package main

// showWindowsMessageBox is a no-op on non-Windows platforms
func showWindowsMessageBox(title, message string) {
	// No-op on non-Windows platforms
}
