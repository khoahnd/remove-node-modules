package services

import (
	"fmt"
	"os/exec"
	"runtime"
)

type NotificationService interface {
	ShowCompletionMessage(isDryRun bool)
}

type PlatformNotificationService struct{}

func NewNotificationService() NotificationService {
	return &PlatformNotificationService{}
}

func (n *PlatformNotificationService) ShowCompletionMessage(isDryRun bool) {
	title := "Node Modules Cleaner"
	message := n.buildMessage(isDryRun)

	switch runtime.GOOS {
	case "windows":
		showWindowsMessageBox(title, message)
	case "darwin":
		n.showMacNotification(title, message)
	default:
		fmt.Printf("\n🎉 %s\n%s\n", title, message)
	}
}

func (n *PlatformNotificationService) buildMessage(isDryRun bool) string {
	if isDryRun {
		return "Dry run completed successfully!\\nCheck the console for what would have been deleted."
	}
	return "Node Modules Cleaner has completed successfully!\\nCheck the console for detailed statistics."
}

func (n *PlatformNotificationService) showMacNotification(title, message string) {
	// Show notification
	notificationScript := fmt.Sprintf(`display notification "%s" with title "%s" sound name "Glass"`, message, title)
	exec.Command("osascript", "-e", notificationScript).Run()

	// Show dialog
	dialogScript := fmt.Sprintf(`display dialog "%s" with title "%s" buttons {"OK"} default button "OK" with icon note`, message, title)
	exec.Command("osascript", "-e", dialogScript).Run()
}
