// test/performance/benchmarks/agent_resolution_benchmark_test.go
package benchmarks

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/gqls/agentchassis/platform/discovery"
	"github.com/gqls/agentchassis/test/unit/helpers"
)

func BenchmarkAgentResolution(b *testing.B) {
	db := helpers.TestDB(&testing.T{})
	defer db.Close()

	// Seed test agents
	seedTestAgents(b, db, 1000)

	discovery := discovery.NewAgentDiscovery(convertToPool(db))

	b.ResetTimer()

	b.Run("ByType", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requirements := discovery.Requirements{
				AgentType: "researcher",
				ClientID:  "demo_client",
			}

			matches, err := discovery.DiscoverAgents(context.Background(), requirements)
			if err != nil {
				b.Fatal(err)
			}
			if len(matches) == 0 {
				b.Fatal("No agents found")
			}
		}
	})

	b.Run("ByCapabilities", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requirements := discovery.Requirements{
				Capabilities: []string{"web_search", "analysis"},
				ClientID:     "demo_client",
			}

			matches, err := discovery.DiscoverAgents(context.Background(), requirements)
			if err != nil {
				b.Fatal(err)
			}
			if len(matches) == 0 {
				b.Fatal("No agents found")
			}
		}
	})

	b.Run("WithFiltering", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requirements := discovery.Requirements{
				AgentType:    "researcher",
				Capabilities: []string{"web_search"},
				MinScore:     0.8,
				MaxResults:   10,
				ClientID:     "demo_client",
			}

			matches, err := discovery.DiscoverAgents(context.Background(), requirements)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkAgentCache(b *testing.B) {
	db := helpers.TestDB(&testing.T{})
	defer db.Close()

	cache := discovery.NewAgentCache(100, 5*time.Minute)
	discovery := discovery.NewAgentDiscoveryWithCache(convertToPool(db), cache)

	requirements := discovery.Requirements{
		AgentType: "researcher",
		ClientID:  "demo_client",
	}

	// Warm up cache
	discovery.DiscoverAgents(context.Background(), requirements)

	b.ResetTimer()

	b.Run("CacheHit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			matches, err := discovery.DiscoverAgents(context.Background(), requirements)
			if err != nil {
				b.Fatal(err)
			}
			if len(matches) == 0 {
				b.Fatal("No agents found")
			}
		}
	})

	b.Run("CacheMiss", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Different requirements each time
			req := discovery.Requirements{
				AgentType: fmt.Sprintf("type-%d", i%100),
				ClientID:  "demo_client",
			}

			discovery.DiscoverAgents(context.Background(), req)
		}
	})
}

func seedTestAgents(b *testing.B, db *sql.DB, count int) {
	b.Helper()

	types := []string{"researcher", "analyzer", "content-creator", "reasoning"}
	capabilities := [][]string{
		{"web_search", "summarization"},
		{"data_analysis", "pattern_recognition"},
		{"writing", "editing"},
		{"review", "critique"},
	}

	for i := 0; i < count; i++ {
		agentType := types[i%len(types)]
		caps := capabilities[i%len(capabilities)]

		_, err := db.Exec(`
            INSERT INTO client_demo_client.agent_instances 
            (id, template_id, owner_user_id, name, config, is_active)
            VALUES ($1, $2, $3, $4, $5, true)
        `,
			uuid.New().String(),
			uuid.New().String(),
			"test_user",
			fmt.Sprintf("test-%s-%d", agentType, i),
			map[string]interface{}{
				"agent_type":        agentType,
				"capabilities":      caps,
				"performance_score": 0.5 + (float64(i%50) / 100),
			},
		)

		if err != nil {
			b.Fatal(err)
		}
	}
}
