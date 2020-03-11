package commands

import "log"

// ReportNetworkUsage lets an external system know about some network usage
// For the time being we're just logging stuff locally to show that it's happening
func (app *AppImplementation) ReportNetworkUsage(runID string, source string, in uint64, out uint64) error {
	log.Printf("Network Usage: %v source: %v, in: %v, out: %v", runID, source, in, out)
	return nil
}
