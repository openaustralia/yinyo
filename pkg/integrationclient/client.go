package integrationclient

import (
	"log"
	"time"
)

// ReportNetworkUsage lets an external system know about some network usage
// For the time being we're just logging stuff locally to show that it's happening
// TODO: Will need the reporting URL as well
func ReportNetworkUsage(runID string, source string, in uint64, out uint64) error {
	log.Printf("Network Usage: %v source: %v, in: %v, out: %v", runID, source, in, out)
	return nil
}

// ReportMemoryUsage lets an external system know about some memory usage
// For the time being we're just logging stuff locally to show that it's happening
// TODO: Will need the reporting URL as well
func ReportMemoryUsage(runID string, memory uint64, duration time.Duration) error {
	log.Printf("Memory Usage: %v memory: %v, duration: %v", runID, memory, duration)
	return nil
}
