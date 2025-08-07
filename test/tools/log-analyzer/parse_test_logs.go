// test/tools/log-analyzer/parse_test_logs.go
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	logFile     = flag.String("file", "", "Log file to analyze")
	correlation = flag.String("correlation", "", "Filter by correlation ID")
	timeRange   = flag.String("time", "", "Time range (e.g., '5m', '1h')")
	errorOnly   = flag.Bool("errors", false, "Show only errors")
	summary     = flag.Bool("summary", false, "Show summary statistics")
)

type LogEntry struct {
	Timestamp     time.Time
	Level         string
	CorrelationID string
	Component     string
	Message       string
	Fields        map[string]interface{}
}

type LogStats struct {
	TotalEntries   int
	ErrorCount     int
	WarningCount   int
	UniqueFlows    int
	ComponentStats map[string]int
	ErrorTypes     map[string]int
}

func main() {
	flag.Parse()

	if *logFile == "" {
		fmt.Println("Please specify a log file with -file")
		os.Exit(1)
	}

	entries, err := parseLogFile(*logFile)
	if err != nil {
		fmt.Printf("Error parsing log file: %v\n", err)
		os.Exit(1)
	}

	// Apply filters
	entries = filterEntries(entries)

	if *summary {
		showSummary(entries)
	} else {
		showEntries(entries)
	}
}

func parseLogFile(filename string) ([]LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)

	// Patterns for different log formats
	jsonPattern := regexp.MustCompile(`^{.*}$`)
	textPattern := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[\.\d]*Z?)\s+(\w+)\s+\[([^\]]+)\]\s+(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()

		if jsonPattern.MatchString(line) {
			// JSON log format
			var entry LogEntry
			if err := parseJSONLog(line, &entry); err == nil {
				entries = append(entries, entry)
			}
		} else if matches := textPattern.FindStringSubmatch(line); matches != nil {
			// Text log format
			entry := LogEntry{
				Timestamp: parseTimestamp(matches[1]),
				Level:     matches[2],
				Component: matches[3],
				Message:   matches[4],
				Fields:    extractFields(matches[4]),
			}

			if cid := extractCorrelationID(matches[4]); cid != "" {
				entry.CorrelationID = cid
			}

			entries = append(entries, entry)
		}
	}

	return entries, scanner.Err()
}

func parseJSONLog(line string, entry *LogEntry) error {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return err
	}

	// Extract standard fields
	if ts, ok := raw["timestamp"].(string); ok {
		entry.Timestamp = parseTimestamp(ts)
	}
	if level, ok := raw["level"].(string); ok {
		entry.Level = level
	}
	if msg, ok := raw["msg"].(string); ok {
		entry.Message = msg
	}
	if cid, ok := raw["correlation_id"].(string); ok {
		entry.CorrelationID = cid
	}
	if comp, ok := raw["component"].(string); ok {
		entry.Component = comp
	}

	// Store other fields
	entry.Fields = make(map[string]interface{})
	for k, v := range raw {
		if k != "timestamp" && k != "level" && k != "msg" && k != "correlation_id" && k != "component" {
			entry.Fields[k] = v
		}
	}

	return nil
}

func filterEntries(entries []LogEntry) []LogEntry {
	var filtered []LogEntry

	for _, entry := range entries {
		// Filter by correlation ID
		if *correlation != "" && !strings.Contains(entry.CorrelationID, *correlation) {
			continue
		}

		// Filter by error level
		if *errorOnly && entry.Level != "ERROR" && entry.Level != "FATAL" {
			continue
		}

		// Filter by time range
		if *timeRange != "" {
			duration, err := time.ParseDuration(*timeRange)
			if err == nil {
				cutoff := time.Now().Add(-duration)
				if entry.Timestamp.Before(cutoff) {
					continue
				}
			}
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

func showEntries(entries []LogEntry) {
	for _, entry := range entries {
		color := getColorForLevel(entry.Level)

		fmt.Printf("%s[%s]%s %s [%s] ",
			color,
			entry.Level,
			"\033[0m",
			entry.Timestamp.Format("15:04:05.000"),
			entry.Component,
		)

		if entry.CorrelationID != "" {
			fmt.Printf("<%s> ", entry.CorrelationID[:8])
		}

		fmt.Printf("%s", entry.Message)

		if len(entry.Fields) > 0 {
			fmt.Printf(" %v", entry.Fields)
		}

		fmt.Println()
	}
}

func showSummary(entries []LogEntry) {
	stats := calculateStats(entries)

	fmt.Println("=== Log Analysis Summary ===")
	fmt.Printf("Total entries: %d\n", stats.TotalEnt

}

// test/tools/log-analyzer/parse_test_logs.go (continued)
func showSummary(entries []LogEntry) {
	stats := calculateStats(entries)

	fmt.Println("=== Log Analysis Summary ===")
	fmt.Printf("Total entries: %d\n", stats.TotalEntries)
	fmt.Printf("Errors: %d (%.1f%%)\n", stats.ErrorCount,
		float64(stats.ErrorCount)/float64(stats.TotalEntries)*100)
	fmt.Printf("Warnings: %d (%.1f%%)\n", stats.WarningCount,
		float64(stats.WarningCount)/float64(stats.TotalEntries)*100)
	fmt.Printf("Unique workflows: %d\n", stats.UniqueFlows)

	fmt.Println("\nEntries by Component:")
	for comp, count := range stats.ComponentStats {
		fmt.Printf("  %s: %d\n", comp, count)
	}

	if stats.ErrorCount > 0 {
		fmt.Println("\nTop Error Types:")
		errorList := sortMapByValue(stats.ErrorTypes)
		for i, kv := range errorList {
			if i >= 10 {
				break
			}
			fmt.Printf("  %s: %d\n", kv.Key, kv.Value)
		}
	}

	// Timeline analysis
	if len(entries) > 0 {
		fmt.Println("\nTimeline:")
		showTimeline(entries)
	}
}

func calculateStats(entries []LogEntry) LogStats {
	stats := LogStats{
		TotalEntries:   len(entries),
		ComponentStats: make(map[string]int),
		ErrorTypes:     make(map[string]int),
	}

	uniqueFlows := make(map[string]bool)

	for _, entry := range entries {
		// Count by level
		switch entry.Level {
		case "ERROR", "FATAL":
			stats.ErrorCount++
			errorType := extractErrorType(entry.Message)
			stats.ErrorTypes[errorType]++
		case "WARN", "WARNING":
			stats.WarningCount++
		}

		// Count by component
		if entry.Component != "" {
			stats.ComponentStats[entry.Component]++
		}

		// Track unique workflows
		if entry.CorrelationID != "" {
			uniqueFlows[entry.CorrelationID] = true
		}
	}

	stats.UniqueFlows = len(uniqueFlows)
	return stats
}

func showTimeline(entries []LogEntry) {
	if len(entries) == 0 {
		return
	}

	// Group by time buckets (1 minute buckets)
	buckets := make(map[string]int)
	var minTime, maxTime time.Time

	for i, entry := range entries {
		if i == 0 {
			minTime = entry.Timestamp
			maxTime = entry.Timestamp
		}

		if entry.Timestamp.Before(minTime) {
			minTime = entry.Timestamp
		}
		if entry.Timestamp.After(maxTime) {
			maxTime = entry.Timestamp
		}

		bucket := entry.Timestamp.Truncate(time.Minute).Format("15:04")
		buckets[bucket]++
	}

	// Show timeline
	fmt.Printf("  Time range: %s - %s (%.1f minutes)\n",
		minTime.Format("15:04:05"),
		maxTime.Format("15:04:05"),
		maxTime.Sub(minTime).Minutes())

	// Sort buckets
	var sortedBuckets []string
	for bucket := range buckets {
		sortedBuckets = append(sortedBuckets, bucket)
	}
	sort.Strings(sortedBuckets)

	// Display histogram
	fmt.Println("  Activity by minute:")
	maxCount := 0
	for _, count := range buckets {
		if count > maxCount {
			maxCount = count
		}
	}

	for _, bucket := range sortedBuckets {
		count := buckets[bucket]
		barLength := int(float64(count) / float64(maxCount) * 40)
		bar := strings.Repeat("█", barLength)
		fmt.Printf("  %s |%-40s| %d\n", bucket, bar, count)
	}
}

func extractCorrelationID(message string) string {
	patterns := []string{
		`correlation_id=([a-zA-Z0-9-]+)`,
		`correlationId:([a-zA-Z0-9-]+)`,
		`<([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(message); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func extractErrorType(message string) string {
	// Common error patterns
	if strings.Contains(message, "timeout") {
		return "Timeout"
	}
	if strings.Contains(message, "connection refused") {
		return "Connection Refused"
	}
	if strings.Contains(message, "not found") {
		return "Not Found"
	}
	if strings.Contains(message, "validation") {
		return "Validation Error"
	}
	if strings.Contains(message, "kafka") {
		return "Kafka Error"
	}
	if strings.Contains(message, "database") || strings.Contains(message, "sql") {
		return "Database Error"
	}

	// Extract first few words as error type
	words := strings.Fields(message)
	if len(words) > 3 {
		return strings.Join(words[:3], " ")
	}
	return message
}

func getColorForLevel(level string) string {
	switch level {
	case "ERROR", "FATAL":
		return "\033[31m" // Red
	case "WARN", "WARNING":
		return "\033[33m" // Yellow
	case "INFO":
		return "\033[32m" // Green
	case "DEBUG":
		return "\033[36m" // Cyan
	default:
		return "\033[0m"  // Default
	}
}

func parseTimestamp(ts string) time.Time {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t
		}
	}

	return time.Now()
}

func extractFields(message string) map[string]interface{} {
	fields := make(map[string]interface{})

	// Extract key=value pairs
	re := regexp.MustCompile(`(\w+)=([^\s]+)`)
	matches := re.FindAllStringSubmatch(message, -1)

	for _, match := range matches {
		if len(match) > 2 {
			fields[match[1]] = match[2]
		}
	}

	return fields
}

type KeyValue struct {
	Key   string
	Value int
}

func sortMapByValue(m map[string]int) []KeyValue {
	var kvs []KeyValue
	for k, v := range m {
		kvs = append(kvs, KeyValue{k, v})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value > kvs[j].Value
	})

	return kvs
}