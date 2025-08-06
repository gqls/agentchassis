// platform/evolution/performance.go
package evolution

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type GroupPerformanceRecord struct {
	GroupID     string
	ExecutionID string
	StartTime   time.Time
	EndTime     time.Time
	TaskType    string

	// Metrics
	SuccessRate   float64
	TotalDuration time.Duration
	StepDurations map[string]time.Duration
	ResourceUsage map[string]int // fuel per agent

	// Quality indicators
	HumanFeedback *HumanFeedback
	OutputQuality map[string]float64

	// Issues
	Bottlenecks []BottleneckInfo
	Failures    []FailureInfo
}

type HumanFeedback struct {
	Rating   float64
	Comments string
	Issues   []string
}

type BottleneckInfo struct {
	AgentID    string
	StepName   string
	Duration   time.Duration
	QueueDepth int
}

type FailureInfo struct {
	AgentID string
	Step    string
	Error   string
	Time    time.Time
}

type Suggestion struct {
	Type   string
	Target string
	Reason string
	Impact string
}

func analyzeGroupPerformance(ctx context.Context, db *pgxpool.Pool, group *AgentGroup) []Suggestion {
	// Get last N executions
	records := getRecentPerformanceRecords(ctx, db, group.ID, 5)

	suggestions := []Suggestion{}

	// Check for consistent bottlenecks
	bottleneckCounts := make(map[string]int)
	for _, record := range records {
		for _, bottleneck := range record.Bottlenecks {
			bottleneckCounts[bottleneck.AgentID]++
		}
	}

	for agentID, count := range bottleneckCounts {
		if count >= 3 { // Consistent problem
			suggestions = append(suggestions, Suggestion{
				Type:   "add_parallel_agent",
				Target: agentID,
				Reason: fmt.Sprintf("Agent bottlenecked in %d of last 5 runs", count),
				Impact: "high",
			})
		}
	}

	// Check for quality issues
	avgQuality := calculateAverageQuality(records)
	if avgQuality < 0.7 {
		suggestions = append(suggestions, Suggestion{
			Type:   "add_quality_checker",
			Reason: "Output quality below threshold",
			Impact: "medium",
		})
	}

	// Check for repeated failures
	failureCounts := make(map[string]int)
	for _, record := range records {
		for _, failure := range record.Failures {
			key := fmt.Sprintf("%s:%s", failure.AgentID, failure.Step)
			failureCounts[key]++
		}
	}

	for failureKey, count := range failureCounts {
		if count >= 2 {
			suggestions = append(suggestions, Suggestion{
				Type:   "add_retry_logic",
				Target: failureKey,
				Reason: fmt.Sprintf("Step failed %d times", count),
				Impact: "medium",
			})
		}
	}

	return suggestions
}

func getRecentPerformanceRecords(ctx context.Context, db *pgxpool.Pool, groupID string, limit int) []GroupPerformanceRecord {
	query := `
        SELECT 
            we.id as execution_id,
            we.started_at,
            we.completed_at,
            we.input_data->>'task_type' as task_type,
            we.output_data
        FROM workflow_executions we
        WHERE we.input_data->>'group_id' = $1
        AND we.status = 'COMPLETED'
        ORDER BY we.completed_at DESC
        LIMIT $2
    `

	rows, err := db.Query(ctx, query, groupID, limit)
	if err != nil {
		return []GroupPerformanceRecord{}
	}
	defer rows.Close()

	var records []GroupPerformanceRecord
	for rows.Next() {
		var record GroupPerformanceRecord
		var outputData json.RawMessage

		err := rows.Scan(
			&record.ExecutionID,
			&record.StartTime,
			&record.EndTime,
			&record.TaskType,
			&outputData,
		)

		if err != nil {
			continue
		}

		record.GroupID = groupID
		record.TotalDuration = record.EndTime.Sub(record.StartTime)

		// Parse output data for metrics
		var output map[string]interface{}
		if err := json.Unmarshal(outputData, &output); err == nil {
			record.SuccessRate = 1.0 // Completed = success for now

			// Extract step durations if available
			if metrics, ok := output["execution_metrics"].(map[string]interface{}); ok {
				record.StepDurations = extractStepDurations(metrics)
				record.ResourceUsage = extractResourceUsage(metrics)
				record.Bottlenecks = identifyBottlenecks(record.StepDurations)
			}

			// Extract quality metrics
			if quality, ok := output["quality_metrics"].(map[string]interface{}); ok {
				record.OutputQuality = extractQualityMetrics(quality)
			}
		}

		records = append(records, record)
	}

	return records
}

func calculateAverageQuality(records []GroupPerformanceRecord) float64 {
	if len(records) == 0 {
		return 1.0
	}

	totalQuality := 0.0
	count := 0

	for _, record := range records {
		// Average quality scores from the record
		for _, quality := range record.OutputQuality {
			totalQuality += quality
			count++
		}

		// Include human feedback if available
		if record.HumanFeedback != nil {
			totalQuality += record.HumanFeedback.Rating
			count++
		}
	}

	if count == 0 {
		return 1.0
	}

	return totalQuality / float64(count)
}

func extractStepDurations(metrics map[string]interface{}) map[string]time.Duration {
	durations := make(map[string]time.Duration)

	if steps, ok := metrics["step_durations"].(map[string]interface{}); ok {
		for step, duration := range steps {
			if d, ok := duration.(float64); ok {
				durations[step] = time.Duration(d) * time.Millisecond
			}
		}
	}

	return durations
}

func extractResourceUsage(metrics map[string]interface{}) map[string]int {
	usage := make(map[string]int)

	if resources, ok := metrics["resource_usage"].(map[string]interface{}); ok {
		for agent, fuel := range resources {
			if f, ok := fuel.(float64); ok {
				usage[agent] = int(f)
			}
		}
	}

	return usage
}

func extractQualityMetrics(quality map[string]interface{}) map[string]float64 {
	metrics := make(map[string]float64)

	for key, value := range quality {
		if v, ok := value.(float64); ok {
			metrics[key] = v
		}
	}

	return metrics
}

func identifyBottlenecks(stepDurations map[string]time.Duration) []BottleneckInfo {
	if len(stepDurations) == 0 {
		return nil
	}

	// Calculate average duration
	var totalDuration time.Duration
	for _, duration := range stepDurations {
		totalDuration += duration
	}
	avgDuration := totalDuration / time.Duration(len(stepDurations))

	// Identify steps that take significantly longer than average
	var bottlenecks []BottleneckInfo
	threshold := avgDuration * 2 // 2x average is considered a bottleneck

	for step, duration := range stepDurations {
		if duration > threshold {
			bottlenecks = append(bottlenecks, BottleneckInfo{
				StepName: step,
				Duration: duration,
				// AgentID would be extracted from step name or additional metadata
			})
		}
	}

	return bottlenecks
}

// RecordGroupPerformance records performance metrics for a group execution
func RecordGroupPerformance(ctx context.Context, db *pgxpool.Pool, record GroupPerformanceRecord) error {
	// Calculate metrics
	successRate := 1.0
	if record.SuccessRate > 0 {
		successRate = record.SuccessRate
	}

	// Update group performance metrics
	_, err := db.Exec(ctx, `
        UPDATE agent_groups 
        SET 
            performance_metrics = jsonb_set(
                COALESCE(performance_metrics, '{}'::jsonb),
                '{last_execution}',
                $1::jsonb
            ),
            usage_count = usage_count + 1,
            last_used_at = NOW()
        WHERE id = $2
    `, map[string]interface{}{
		"execution_id":   record.ExecutionID,
		"success_rate":   successRate,
		"duration_ms":    record.TotalDuration.Milliseconds(),
		"resource_usage": record.ResourceUsage,
		"timestamp":      time.Now(),
	}, record.GroupID)

	return err
}

// AnalyzeEvolutionImpact compares performance before and after evolution
func AnalyzeEvolutionImpact(ctx context.Context, db *pgxpool.Pool, oldGroupID, newGroupID string) (map[string]interface{}, error) {
	oldRecords := getRecentPerformanceRecords(ctx, db, oldGroupID, 10)
	newRecords := getRecentPerformanceRecords(ctx, db, newGroupID, 10)

	if len(oldRecords) == 0 || len(newRecords) == 0 {
		return nil, fmt.Errorf("insufficient data for comparison")
	}

	// Calculate averages
	oldAvgDuration := calculateAverageDuration(oldRecords)
	newAvgDuration := calculateAverageDuration(newRecords)

	oldAvgQuality := calculateAverageQuality(oldRecords)
	newAvgQuality := calculateAverageQuality(newRecords)

	oldAvgFuel := calculateAverageFuel(oldRecords)
	newAvgFuel := calculateAverageFuel(newRecords)

	// Calculate improvements as percentages
	// Convert durations to float64 (milliseconds) for calculation
	oldDurationMs := float64(oldAvgDuration.Milliseconds())
	newDurationMs := float64(newAvgDuration.Milliseconds())

	var durationImprovement float64
	if oldDurationMs > 0 {
		durationImprovement = ((oldDurationMs - newDurationMs) / oldDurationMs) * 100
	}

	var qualityImprovement float64
	if oldAvgQuality > 0 {
		qualityImprovement = ((newAvgQuality - oldAvgQuality) / oldAvgQuality) * 100
	}

	var fuelEfficiency float64
	if oldAvgFuel > 0 {
		fuelEfficiency = ((oldAvgFuel - newAvgFuel) / oldAvgFuel) * 100
	}

	return map[string]interface{}{
		"duration_improvement_pct": durationImprovement,
		"quality_improvement_pct":  qualityImprovement,
		"fuel_efficiency_pct":      fuelEfficiency,
		"old_avg_duration_ms":      oldDurationMs,
		"new_avg_duration_ms":      newDurationMs,
		"old_avg_quality":          oldAvgQuality,
		"new_avg_quality":          newAvgQuality,
		"old_avg_fuel":             oldAvgFuel,
		"new_avg_fuel":             newAvgFuel,
		"recommendation":           generateRecommendation(durationImprovement, qualityImprovement, fuelEfficiency),
	}, nil
}

func calculateAverageDuration(records []GroupPerformanceRecord) time.Duration {
	var total time.Duration
	for _, record := range records {
		total += record.TotalDuration
	}
	return total / time.Duration(len(records))
}

func calculateAverageFuel(records []GroupPerformanceRecord) float64 {
	totalFuel := 0
	count := 0

	for _, record := range records {
		for _, fuel := range record.ResourceUsage {
			totalFuel += fuel
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return float64(totalFuel) / float64(count)
}

func generateRecommendation(durationImprovement, qualityImprovement, fuelEfficiency float64) string {
	score := (durationImprovement + qualityImprovement + fuelEfficiency) / 3

	if score > 20 {
		return "Highly successful evolution - recommend keeping"
	} else if score > 0 {
		return "Moderate improvement - monitor performance"
	} else {
		return "Evolution did not improve performance - consider reverting"
	}
}
