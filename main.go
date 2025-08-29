package main

import (
	"flag"
	"fmt"
	"log"
	"node-cleaner/models"
	"node-cleaner/services"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	config := parseCommandLineArgs()

	if config.ShowHelp {
		showHelpAndExit()
		return
	}

	if err := validateAndExecute(config); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
}

func parseCommandLineArgs() *models.Config {
	var (
		rootPath = flag.String("path", ".", "Root directory path to scan")
		workers  = flag.Int("workers", runtime.NumCPU(), "Number of worker threads")
		dryRun   = flag.Bool("dry-run", false, "Only show what would be deleted, don't actually delete")
		help     = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	// Interactive mode if no args provided
	if len(os.Args) == 1 {
		uiService := services.NewUserInterfaceService()
		return uiService.ShowInteractiveMenu()
	}

	config := models.NewConfig(*rootPath, *workers, *dryRun)
	config.ShowHelp = *help
	return config
}

func validateAndExecute(config *models.Config) error {
	// Validate path
	if _, err := os.Stat(config.RootPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", config.RootPath)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(config.RootPath)
	if err != nil {
		return fmt.Errorf("cannot convert to absolute path: %v", err)
	}
	config.RootPath = absPath

	// Log configuration
	logConfiguration(config)

	// Get user confirmation for real deletions
	if !config.DryRun {
		uiService := services.NewUserInterfaceService()
		if !uiService.PromptForConfirmation(absPath) {
			fmt.Println("❌ Cancelled")
			return nil
		}
	} else {
		fmt.Println("⚠️  DRY RUN MODE - NO ACTUAL DELETION")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
	}

	// Execute the cleaning process
	stats := models.NewScanStats()
	directoryService := services.NewDirectoryService(config, stats)
	statsService := services.NewStatsService(stats)
	notificationService := services.NewNotificationService()

	if err := directoryService.FindAndProcessNodeModules(); err != nil {
		return err
	}

	statsService.LogFinalStats()
	notificationService.ShowCompletionMessage(config.DryRun)

	// Wait for exit in interactive mode
	if len(os.Args) == 1 {
		uiService := services.NewUserInterfaceService()
		uiService.WaitForExit()
	}

	return nil
}

func logConfiguration(config *models.Config) {
	log.Printf("🔧 Configuration:")
	log.Printf("   - Path: %s", config.RootPath)
	log.Printf("   - Workers: %d", config.Workers)
	log.Printf("   - CPU cores: %d", runtime.NumCPU())
}

func showHelpAndExit() {
	fmt.Println("🧹 Node Modules Cleaner")
	fmt.Println("")
	fmt.Println("Tool to find and delete all node_modules directories in a path.")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  Interactive mode: Just run the executable")
	fmt.Println("  Command line mode:")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  node-cleaner.exe")
	fmt.Println("  node-cleaner.exe -path C:\\Users\\Projects")
	fmt.Println("  node-cleaner.exe -path C:\\Users\\Projects -workers 8")
	fmt.Println("  node-cleaner.exe -path C:\\Users\\Projects -dry-run")
	fmt.Println("")
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}
