// test/tools/dashboard/test_dashboard.go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Agent Test Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metrics { display: flex; gap: 20px; margin-bottom: 20px; }
        .metric { 
            background: #f0f0f0; 
            padding: 20px; 
            border-radius: 8px;
            text-align: center;
        }
        .metric h3 { margin: 0 0 10px 0; }
        .metric .value { font-size: 2em; font-weight: bold; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }
        .status-completed { color: green; }
        .status-failed { color: red; }
        .status-running { color: blue; }
        .refresh { margin: 20px 0; }
    </style>
    <script>
        function autoRefresh() {
            if (document.getElementById('autoRefresh').checked) {
                setTimeout(() => location.reload(), 5000);
            }
        }
        window.onload = autoRefresh;
    </script>
</head>
<body>
    <h1>Agent Test Dashboard</h1>
    
    <div class="refresh">
        <label>
            <input type="checkbox" id="autoRefresh" checked> Auto-refresh (5s)
        </label>
    </div>

    <div class="metrics">
        <div class="metric">
            <h3>Total Tests</h3>
            <div class="value">{{.Stats.Total}}</div>
        </div>
        <div class="metric">
            <h3>Completed</h3>
            <div class="value" style="color: green;">{{.Stats.Completed}}</div>
        </div>
        <div class="metric">
            <h3>Failed</h3>
            <div class="value" style="color: red;">{{.Stats.Failed}}</div>
        </div>
        <div class="metric">
            <h3>Active</h3>
            <div class="value" style="color: blue;">{{.Stats.Active}}</div>
        </div>
        <div class="metric">
            <h3>Success Rate</h3>
            <div class="value">{{.Stats.SuccessRate}}%</div>
        </div>
    </div>

    <h2>Recent Test Workflows</h2>
    <table>
        <tr>
            <th>Correlation ID</th>
            <th>Status</th>
            <th>Current Step</th>
            <th>Progress</th>
            <th>Duration</th>
            <th>Started</th>
        </tr>
        {{range .Workflows}}
        <tr>
            <td><a href="/workflow/{{.CorrelationID}}">{{.CorrelationID}}</a></td>
            <td class="status-{{.StatusClass}}">{{.Status}}</td>
            <td>{{.CurrentStep}}</td>
            <td>{{.Progress}}</td>
            <td>{{.Duration}}</td>
            <td>{{.StartedAgo}}</td>
        </tr>
        {{end}}
    </table>

    <h2>Agent Performance</h2>
    <table>
        <tr>
            <th>Agent Type</th>
            <th>Total Calls</th>
            <th>Success Rate</th>
            <th>Avg Response Time</th>
        </tr>
        {{range .Agents}}
        <tr>
            <td>{{.Type}}</td>
            <td>{{.TotalCalls}}</td>
            <td>{{.SuccessRate}}%</td>
            <td>{{.AvgResponseTime}}ms</td>
        </tr>
        {{end}}
    </table>
</body>
</html>
`

type DashboardData struct {
	Stats     TestStats
	Workflows []WorkflowInfo
	Agents    []AgentStats
}

type TestStats struct {
	Total       int
	Completed   int
	Failed      int
	Active      int
	SuccessRate float64
}

type WorkflowInfo struct {
	CorrelationID string
	Status        string
	StatusClass   string
	CurrentStep   string
	Progress      string
	Duration      string
	StartedAgo    string
}

type AgentStats struct {
	Type            string
	TotalCalls      int
	SuccessRate     float64
	AvgResponseTime int
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := collectDashboardData(db)
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/workflow/", func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.URL.Path[len("/workflow/"):]
		showWorkflowDetails(w, db, correlationID)
	})

	fmt.Println("Test dashboard running on http://localhost:8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func collectDashboardData(db *sql.DB) DashboardData {
	data := DashboardData{}

	// Collect test statistics
	db.QueryRow(`
        SELECT 
            COUNT(*) as total,
            COUNT(*) FILTER (WHERE status = 'COMPLETED') as completed,
            COUNT(*) FILTER (WHERE status = 'FAILED') as failed,
            COUNT(*) FILTER (WHERE status IN ('RUNNING', 'AWAITING_RESPONSES')) as active
        FROM orchestrator_state
        WHERE correlation_id LIKE 'test-%'
          AND created_at > NOW() - INTERVAL '24 hours'
    `).Scan(&data.Stats.Total, &data.Stats.Completed, &data.Stats.Failed, &data.Stats.Active)

	if data.Stats.Total > 0 {
		data.Stats.SuccessRate = float64(data.Stats.Completed) / float64(data.Stats.Total) * 100
	}

	// Collect recent workflows
	rows, _ := db.Query(`
        SELECT 
            correlation_id,
            status,
            current_step,
            execution_metadata,
            created_at,
            updated_at
        FROM orchestrator_state
        WHERE correlation_id LIKE 'test-%'
        ORDER BY created_at DESC
        LIMIT 20
    `)
	defer rows.Close()

	for rows.Next() {
		var w WorkflowInfo
		var metadata json.RawMessage
		var createdAt, updatedAt time.Time

		rows.Scan(&w.CorrelationID, &w.Status, &w.CurrentStep,
			&metadata, &createdAt, &updatedAt)

		// Calculate progress
		var meta map[string]interface{}
		if json.Unmarshal(metadata, &meta) == nil {
			if completed, ok := meta["completed_steps"].(float64); ok {
				if total, ok := meta["total_steps"].(float64); ok {
					w.Progress = fmt.Sprintf("%.0f/%.0f", completed, total)
				}
			}
		}

		// Calculate duration
		if w.Status == "COMPLETED" || w.Status == "FAILED" {
			w.Duration = fmt.Sprintf("%.1fs", updatedAt.Sub(createdAt).Seconds())
		} else {
			w.Duration = fmt.Sprintf("%.1fs", time.Since(createdAt).Seconds())
		}

		// Time ago
		w.StartedAgo = fmt.Sprintf("%.0f min ago", time.Since(createdAt).Minutes())

		// Status class for CSS
		switch w.Status {
		case "COMPLETED":
			w.StatusClass = "completed"
		case "FAILED":
			w.StatusClass = "failed"
		case "RUNNING", "AWAITING_RESPONSES":
			w.StatusClass = "running"
		}

		data.Workflows = append(data.Workflows, w)
	}

	// Collect agent statistics
	agentRows, _ := db.Query(`
        SELECT 
            config->>'agent_type' as type,
            COUNT(*) as total_calls,
            AVG(CASE WHEN status = 'COMPLETED' THEN 100 ELSE 0 END) as success_rate,
            AVG(EXTRACT(EPOCH FROM (updated_at - created_at))) * 1000 as avg_response_time
        FROM orchestrator_state os
        JOIN client_demo_client.agent_instances ai ON ai.id::text = os.collected_data->>'agent_id'
        WHERE os.correlation_id LIKE 'test-%'
        GROUP BY config->>'agent_type'
    `)
	defer agentRows.Close()

	for agentRows.Next() {
		var a AgentStats
		agentRows.Scan(&a.Type, &a.TotalCalls, &a.SuccessRate, &a.AvgResponseTime)
		data.Agents = append(data.Agents, a)
	}

	return data
}

func showWorkflowDetails(w http.ResponseWriter, db *sql.DB, correlationID string) {
	// Show detailed workflow information
	var workflowData json.RawMessage
	err := db.QueryRow(`
        SELECT json_build_object(
            'correlation_id', correlation_id,
            'status', status,
            'workflow_plan', workflow_plan,
            'execution_path', execution_path,
            'collected_data', collected_data,
            'execution_metadata', execution_metadata
        )
        FROM orchestrator_state
        WHERE correlation_id = $1
    `, correlationID).Scan(&workflowData)

	if err != nil {
		http.Error(w, "Workflow not found", 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(workflowData)
}
