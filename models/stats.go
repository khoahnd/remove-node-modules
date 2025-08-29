package models

import "sync"

type ScanStats struct {
	totalDirsScanned   int64
	nodeModulesFound   int64
	nodeModulesDeleted int64
	errorCount         int64
	mu                 sync.Mutex
}

func NewScanStats() *ScanStats {
	return &ScanStats{}
}

func (s *ScanStats) IncrementScanned() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalDirsScanned++
}

func (s *ScanStats) IncrementFound() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeModulesFound++
}

func (s *ScanStats) IncrementDeleted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeModulesDeleted++
}

func (s *ScanStats) IncrementErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorCount++
}

func (s *ScanStats) GetCounts() (scanned, found, deleted, errors int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalDirsScanned, s.nodeModulesFound, s.nodeModulesDeleted, s.errorCount
}
