package services

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"node-cleaner/models"
)

type UserInterfaceService struct {
	scanner *bufio.Scanner
}

func NewUserInterfaceService() *UserInterfaceService {
	return &UserInterfaceService{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (ui *UserInterfaceService) ShowInteractiveMenu() *models.Config {
	fmt.Println("🧹 Node Modules Cleaner - Interactive Mode")
	fmt.Println("==========================================")
	fmt.Println()

	path := ui.promptForPath()
	dryRun := ui.promptForMode()
	workers := ui.promptForWorkerCount()

	return models.NewConfig(path, workers, dryRun)
}

func (ui *UserInterfaceService) promptForPath() string {
	for {
		fmt.Print("📁 Enter the path to scan (or '.' for current directory): ")
		ui.scanner.Scan()
		path := strings.TrimSpace(ui.scanner.Text())

		if path == "" {
			path = "."
		}

		if ui.isValidPath(path) {
			return path
		}

		fmt.Printf("❌ Path does not exist: %s\nPlease try again.\n", path)
	}
}

func (ui *UserInterfaceService) promptForMode() bool {
	for {
		fmt.Print("🧪 Mode - (1) Dry Run [Preview only] or (2) Real Delete [1/2]: ")
		ui.scanner.Scan()
		choice := strings.TrimSpace(ui.scanner.Text())

		switch choice {
		case "1", "":
			fmt.Println("✅ Selected: Dry Run (Preview mode)")
			return true
		case "2":
			fmt.Println("✅ Selected: Real Delete mode")
			return false
		default:
			fmt.Println("❌ Please enter 1 or 2")
		}
	}
}

func (ui *UserInterfaceService) promptForWorkerCount() int {
	defaultWorkers := runtime.NumCPU()
	maxWorkers := defaultWorkers * 2

	for {
		fmt.Printf("⚙️ Number of workers (1-%d) [default %d]: ", maxWorkers, defaultWorkers)
		ui.scanner.Scan()
		input := strings.TrimSpace(ui.scanner.Text())

		if input == "" {
			fmt.Printf("✅ Selected: %d workers\n", defaultWorkers)
			return defaultWorkers
		}

		workers, err := strconv.Atoi(input)
		if err != nil || workers < 1 || workers > maxWorkers {
			fmt.Printf("❌ Please enter a number between 1 and %d\n", maxWorkers)
			continue
		}

		fmt.Printf("✅ Selected: %d workers\n", workers)
		return workers
	}
}

func (ui *UserInterfaceService) isValidPath(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (ui *UserInterfaceService) PromptForConfirmation(absPath string) bool {
	fmt.Printf("⚠️  WARNING: You are about to delete ALL node_modules directories in %s\n", absPath)
	fmt.Print("Type 'yes' to confirm: ")

	var confirmation string
	fmt.Scanln(&confirmation)
	return confirmation == "yes"
}

func (ui *UserInterfaceService) WaitForExit() {
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}
