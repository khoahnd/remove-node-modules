package services

import (
	"log"
	"time"

	"node-cleaner/models"
)

type StatsService struct {
	stats     *models.ScanStats
	startTime time.Time
}

func NewStatsService(stats *models.ScanStats) *StatsService {
	return &StatsService{
		stats:     stats,
		startTime: time.Now(),
	}
}

func (ss *StatsService) LogFinalStats() {
	duration := time.Since(ss.startTime)
	scanned, found, deleted, errors := ss.stats.GetCounts()

	log.Printf("✅ Completed!")
	log.Printf("📈 Statistics:")
	log.Printf("   - Total directories scanned: %d", scanned)
	log.Printf("   - Total node_modules found: %d", found)
	log.Printf("   - Total node_modules deleted: %d", deleted)
	log.Printf("   - Total errors: %d", errors)
	log.Printf("   - Execution time: %v", duration)
}
