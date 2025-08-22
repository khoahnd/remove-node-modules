package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
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

// showMessageBox displays a message box on Windows, notification on macOS, and prints to console on Linux
func showMessageBox(title, message string) {
	switch runtime.GOOS {
	case "windows":
		showWindowsMessageBox(title, message)
	case "darwin": // macOS
		showMacNotification(title, message)
	default: // Linux and others
		fmt.Printf("\n🎉 %s\n%s\n", title, message)
	}
}

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

// showMacNotification displays a macOS notification and dialog
func showMacNotification(title, message string) {
	// First, show a notification
	notificationScript := fmt.Sprintf(`
		display notification "%s" with title "%s" sound name "Glass"
	`, message, title)

	cmd := exec.Command("osascript", "-e", notificationScript)
	cmd.Run()

	// Then show a dialog box for better visibility
	dialogScript := fmt.Sprintf(`
		display dialog "%s" with title "%s" buttons {"OK"} default button "OK" with icon note
	`, message, title)

	cmd = exec.Command("osascript", "-e", dialogScript)
	cmd.Run()
}

// showInteractiveMenu displays an interactive menu when no arguments are provided
func showInteractiveMenu() (string, int, bool, bool) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("🧹 Node Modules Cleaner - Interactive Mode")
	fmt.Println("==========================================")
	fmt.Println()

	// Get path
	var path string
	for {
		fmt.Print("📁 Enter the path to scan (or '.' for current directory): ")
		scanner.Scan()
		path = strings.TrimSpace(scanner.Text())

		if path == "" {
			path = "."
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("❌ Path does not exist: %s\n", path)
			fmt.Println("Please try again.")
			continue
		}
		break
	}

	// Get mode
	var dryRun bool
	for {
		fmt.Print("🧪 Mode - (1) Dry Run [Preview only] or (2) Real Delete [1/2]: ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		if choice == "1" || choice == "" {
			dryRun = true
			fmt.Println("✅ Selected: Dry Run (Preview mode)")
			break
		} else if choice == "2" {
			dryRun = false
			fmt.Println("✅ Selected: Real Delete mode")
			break
		} else {
			fmt.Println("❌ Please enter 1 or 2")
		}
	}

	// Get workers
	defaultWorkers := runtime.NumCPU()
	var workers int
	for {
		fmt.Printf("⚙️ Number of workers (1-%d) [default %d]: ", defaultWorkers*2, defaultWorkers)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			workers = defaultWorkers
			break
		}

		var err error
		workers, err = strconv.Atoi(input)
		if err != nil || workers < 1 || workers > defaultWorkers*2 {
			fmt.Printf("❌ Please enter a number between 1 and %d\n", defaultWorkers*2)
			continue
		}
		break
	}

	fmt.Printf("✅ Selected: %d workers\n", workers)
	fmt.Println()

	return path, workers, dryRun, false // false means don't show help
}

type Stats struct {
	totalDirsScanned   int64
	nodeModulesFound   int64
	nodeModulesDeleted int64
	errors             int64
	mu                 sync.Mutex
}

func (s *Stats) addScanned() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalDirsScanned++
}

func (s *Stats) addFound() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeModulesFound++
}

func (s *Stats) addDeleted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeModulesDeleted++
}

func (s *Stats) addError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors++
}

func (s *Stats) getStats() (int64, int64, int64, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalDirsScanned, s.nodeModulesFound, s.nodeModulesDeleted, s.errors
}

func removeNodeModules(rootPath string, maxWorkers int, dryRun bool) error {
	stats := &Stats{}

	log.Printf("🚀 Starting directory scan: %s", rootPath)
	log.Printf("📊 Number of worker threads: %d", maxWorkers)
	if dryRun {
		log.Printf("🧪 Running in DRY RUN mode - no actual deletion")
	}

	startTime := time.Now()

	// Use a simpler approach - process synchronously but with concurrent deletion
	var nodeModulesPaths []string

	// Walk through directories and collect all node_modules paths
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("❌ Error accessing path %s: %v", path, err)
			stats.addError()
			return nil // Continue walking
		}

		if !info.IsDir() {
			return nil
		}

		stats.addScanned()

		// Check if this is a node_modules directory
		if info.Name() == "node_modules" {
			log.Printf("🎯 Found node_modules: %s", path)
			stats.addFound()
			nodeModulesPaths = append(nodeModulesPaths, path)
			return filepath.SkipDir // Don't walk into node_modules
		}

		// Skip system and hidden directories
		if shouldSkipDirectory(info.Name()) {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ Error walking directory tree: %v", err)
		return err
	}

	// Now delete all found node_modules directories concurrently
	if len(nodeModulesPaths) > 0 {
		deleteChan := make(chan string, len(nodeModulesPaths))
		var deleteWg sync.WaitGroup

		// Start deletion workers
		numDeleteWorkers := maxWorkers
		if numDeleteWorkers > len(nodeModulesPaths) {
			numDeleteWorkers = len(nodeModulesPaths)
		}

		for i := 0; i < numDeleteWorkers; i++ {
			deleteWg.Add(1)
			go func(workerID int) {
				defer deleteWg.Done()
				for nodeModulesPath := range deleteChan {
					deleteNodeModules(nodeModulesPath, stats, workerID, dryRun)
				}
			}(i)
		}

		// Send all paths to deletion channel
		for _, path := range nodeModulesPaths {
			deleteChan <- path
		}
		close(deleteChan)

		// Wait for all deletions to complete
		deleteWg.Wait()
	}

	duration := time.Since(startTime)
	totalScanned, totalFound, totalDeleted, totalErrors := stats.getStats()

	log.Printf("✅ Completed!")
	log.Printf("📈 Statistics:")
	log.Printf("   - Total directories scanned: %d", totalScanned)
	log.Printf("   - Total node_modules found: %d", totalFound)
	log.Printf("   - Total node_modules deleted: %d", totalDeleted)
	log.Printf("   - Total errors: %d", totalErrors)
	log.Printf("   - Execution time: %v", duration)

	return nil
}

func shouldSkipDirectory(dirName string) bool {
	skipDirs := []string{
		".git", ".svn", ".hg", // Version control
		"System Volume Information", "$RECYCLE.BIN", // Windows system
		".DS_Store", ".Trash", // macOS system
		"proc", "sys", "dev", // Linux system
		"Windows", "Program Files", "Program Files (x86)", // Windows system
	}

	for _, skip := range skipDirs {
		if dirName == skip {
			return true
		}
	}

	// Skip hidden directories (starting with dot) except .npm, .yarn
	if len(dirName) > 0 && dirName[0] == '.' {
		allowedHidden := []string{".npm", ".yarn", ".pnpm"}
		for _, allowed := range allowedHidden {
			if dirName == allowed {
				return false
			}
		}
		return true
	}

	return false
}

func deleteNodeModules(nodeModulesPath string, stats *Stats, workerID int, dryRun bool) {
	log.Printf("🗑️  Worker %d - Starting deletion: %s", workerID, nodeModulesPath)

	if dryRun {
		log.Printf("🧪 DRY RUN - Would delete: %s", nodeModulesPath)
		stats.addDeleted()
		return
	}

	// Check size before deletion
	size, err := getDirSize(nodeModulesPath)
	if err != nil {
		log.Printf("⚠️  Worker %d - Cannot calculate size %s: %v", workerID, nodeModulesPath, err)
	} else {
		log.Printf("📦 Worker %d - Size: %s (%.2f MB)", workerID, nodeModulesPath, float64(size)/(1024*1024))
	}

	err = os.RemoveAll(nodeModulesPath)
	if err != nil {
		log.Printf("❌ Worker %d - Error deleting %s: %v", workerID, nodeModulesPath, err)
		stats.addError()
		return
	}

	log.Printf("✅ Worker %d - Successfully deleted: %s", workerID, nodeModulesPath)
	stats.addDeleted()
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func main() {
	var (
		rootPath = flag.String("path", ".", "Root directory path to scan")
		workers  = flag.Int("workers", runtime.NumCPU(), "Number of worker threads")
		dryRun   = flag.Bool("dry-run", false, "Only show what would be deleted, don't actually delete")
		help     = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	// Check if no arguments were provided (interactive mode)
	if len(os.Args) == 1 {
		// Interactive mode
		path, workerCount, isDryRun, showHelp := showInteractiveMenu()
		if showHelp {
			*help = true
		} else {
			*rootPath = path
			*workers = workerCount
			*dryRun = isDryRun
		}
	}

	if *help {
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
		return
	}

	// Check if path exists
	if _, err := os.Stat(*rootPath); os.IsNotExist(err) {
		log.Fatalf("❌ Path does not exist: %s", *rootPath)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(*rootPath)
	if err != nil {
		log.Fatalf("❌ Cannot convert to absolute path: %v", err)
	}

	log.Printf("🔧 Configuration:")
	log.Printf("   - Path: %s", absPath)
	log.Printf("   - Workers: %d", *workers)
	log.Printf("   - CPU cores: %d", runtime.NumCPU())

	if *dryRun {
		fmt.Println("⚠️  DRY RUN MODE - NO ACTUAL DELETION")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
	} else {
		fmt.Printf("⚠️  WARNING: You are about to delete ALL node_modules directories in %s\n", absPath)
		fmt.Print("Type 'yes' to confirm: ")
		var confirmation string
		fmt.Scanln(&confirmation)
		if confirmation != "yes" {
			fmt.Println("❌ Cancelled")
			return
		}
	}

	err = removeNodeModules(absPath, *workers, *dryRun)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	// Show completion message on all platforms
	message := "Node Modules Cleaner has completed successfully!\\nCheck the console for detailed statistics."
	if *dryRun {
		message = "Dry run completed successfully!\\nCheck the console for what would have been deleted."
	}
	showMessageBox("Node Modules Cleaner", message)

	// If running in interactive mode (no command line args), pause before exit
	if len(os.Args) == 1 {
		fmt.Println("\nPress Enter to exit...")
		fmt.Scanln()
	}
}
