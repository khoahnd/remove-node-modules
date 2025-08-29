//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// Windows API constants for MessageBox
const (
	MB_OK       = 0x00000000
	MB_ICONINFO = 0x00000040
)

// Windows API functions
var (
	user32      = syscall.NewLazyDLL("user32.dll")
	messageBoxW = user32.NewProc("MessageBoxW")
)

// showWindowsMessageBox displays a Windows message box
func showWindowsMessageBox(title, message string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	messageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(MB_OK|MB_ICONINFO),
	)
}
