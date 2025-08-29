//go:build !windows

package services

// showWindowsMessageBox is a no-op on non-Windows platforms
func showWindowsMessageBox(title, message string) {
	// No-op on non-Windows platforms
}
