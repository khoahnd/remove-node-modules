package services

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"node-cleaner/models"
)

type DirectoryService struct {
	config *models.Config
	stats  *models.ScanStats
}

func NewDirectoryService(config *models.Config, stats *models.ScanStats) *DirectoryService {
	return &DirectoryService{
		config: config,
		stats:  stats,
	}
}

func (ds *DirectoryService) FindAndProcessNodeModules() error {
	log.Printf("🚀 Starting directory scan: %s", ds.config.RootPath)
	log.Printf("📊 Workers: %d", ds.config.Workers)

	if ds.config.DryRun {
		log.Printf("🧪 DRY RUN mode - no actual deletion")
	}

	nodeModulesPaths, err := ds.findNodeModulesDirectories()
	if err != nil {
		return err
	}

	if len(nodeModulesPaths) > 0 {
		ds.processNodeModulesDirectories(nodeModulesPaths)
	}

	return nil
}

func (ds *DirectoryService) findNodeModulesDirectories() ([]string, error) {
	var nodeModulesPaths []string

	err := filepath.Walk(ds.config.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("❌ Error accessing path %s: %v", path, err)
			ds.stats.IncrementErrors()
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		ds.stats.IncrementScanned()

		if ds.isNodeModulesDirectory(info.Name()) {
			log.Printf("🎯 Found node_modules: %s", path)
			ds.stats.IncrementFound()
			nodeModulesPaths = append(nodeModulesPaths, path)
			return filepath.SkipDir
		}

		if ds.shouldSkipDirectory(info.Name()) {
			return filepath.SkipDir
		}

		return nil
	})

	return nodeModulesPaths, err
}

func (ds *DirectoryService) processNodeModulesDirectories(paths []string) {
	pathChan := make(chan string, len(paths))
	var wg sync.WaitGroup

	workerCount := ds.calculateWorkerCount(len(paths))

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go ds.deletionWorker(i, pathChan, &wg)
	}

	// Send paths to workers
	for _, path := range paths {
		pathChan <- path
	}
	close(pathChan)

	wg.Wait()
}

func (ds *DirectoryService) deletionWorker(workerID int, pathChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range pathChan {
		ds.deleteNodeModulesDirectory(path, workerID)
	}
}

func (ds *DirectoryService) deleteNodeModulesDirectory(path string, workerID int) {
	log.Printf("🗑️  Worker %d - Processing: %s", workerID, path)

	if ds.config.DryRun {
		log.Printf("🧪 DRY RUN - Would delete: %s", path)
		ds.stats.IncrementDeleted()
		return
	}

	ds.logDirectorySize(path, workerID)

	if err := os.RemoveAll(path); err != nil {
		log.Printf("❌ Worker %d - Error deleting %s: %v", workerID, path, err)
		ds.stats.IncrementErrors()
		return
	}

	log.Printf("✅ Worker %d - Successfully deleted: %s", workerID, path)
	ds.stats.IncrementDeleted()
}

func (ds *DirectoryService) logDirectorySize(path string, workerID int) {
	if size, err := ds.calculateDirectorySize(path); err == nil {
		sizeMB := float64(size) / (1024 * 1024)
		log.Printf("📦 Worker %d - Size: %.2f MB for %s", workerID, sizeMB, path)
	}
}

func (ds *DirectoryService) calculateDirectorySize(path string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

func (ds *DirectoryService) isNodeModulesDirectory(dirName string) bool {
	return dirName == "node_modules"
}

func (ds *DirectoryService) shouldSkipDirectory(dirName string) bool {
	systemDirs := []string{
		".git", ".svn", ".hg",
		"System Volume Information", "$RECYCLE.BIN",
		".DS_Store", ".Trash",
		"proc", "sys", "dev",
		"Windows", "Program Files", "Program Files (x86)",
	}

	for _, skipDir := range systemDirs {
		if dirName == skipDir {
			return true
		}
	}

	return ds.isHiddenDirectoryToSkip(dirName)
}

func (ds *DirectoryService) isHiddenDirectoryToSkip(dirName string) bool {
	if len(dirName) == 0 || dirName[0] != '.' {
		return false
	}

	allowedHiddenDirs := []string{".npm", ".yarn", ".pnpm"}
	for _, allowed := range allowedHiddenDirs {
		if dirName == allowed {
			return false
		}
	}

	return true
}

func (ds *DirectoryService) calculateWorkerCount(pathCount int) int {
	if ds.config.Workers > pathCount {
		return pathCount
	}
	return ds.config.Workers
}
