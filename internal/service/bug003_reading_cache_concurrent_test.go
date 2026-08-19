package service

import (
	"sync"
	"testing"

	"aquafarm/internal/model"
)

// TestBug003_ReadingCacheConcurrentAccess verifies that concurrent Update and GetLatest
// do not cause data races.
// Bug: Update uses RLock instead of Lock for write operations.
// Run with: go test -race -count=20
func TestBug003_ReadingCacheConcurrentAccess(t *testing.T) {
	cache := NewReadingCache()

	var wg sync.WaitGroup
	// Writer goroutines
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cache.Update(model.SensorReading{
					TankID: 1,
					Type:   model.SensorTemperature,
					Value:  float64(j),
					Unit:   "C",
				})
			}
		}()
	}
	// Reader goroutines
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cache.GetLatest(1, model.SensorTemperature)
			}
		}()
	}
	wg.Wait()
}