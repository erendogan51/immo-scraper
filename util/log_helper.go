package util

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Global log cache with mutex for thread safety
var (
	logCache    = make(map[string]CacheEntry)
	callCounter int64
	cacheMutex  sync.RWMutex
)

type CacheEntry struct {
	count int
	dt    time.Time
}

const (
	resetCounter     = 10000
	maxCacheDuration = 10 * time.Second
)

func CachedLog(logger *zerolog.Event, message string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	callCounter++

	if entry, exists := logCache[message]; exists && time.Since(entry.dt) < maxCacheDuration {
		// Same message within 10 seconds, increment counter
		logCache[message] = CacheEntry{count: entry.count + 1, dt: entry.dt}
	} else {
		// Actually log the message
		logger.Int("cache_count", entry.count).Msg(message)

		// Log the current message for the first time
		logCache[message] = CacheEntry{count: 1, dt: time.Now()}
	}

	// regular cleanup to remove outdated cached fields.
	if callCounter >= resetCounter {
		// remove all entries from the log cache which are older than 10 seconds
		for k, v := range logCache {
			if time.Since(v.dt) < maxCacheDuration {
				delete(logCache, k)
			}
		}
		callCounter = 0
	}
}
