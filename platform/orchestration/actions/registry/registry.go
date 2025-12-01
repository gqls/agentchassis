// FILE: platform/orchestration/actions/registry/registry.go
// Package registry provides the central action registry with metadata support.
// All action packages register themselves with this registry during init().
package registry

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"go.uber.org/zap"
)

// ActionFunc is the signature for all action handlers
type ActionFunc func(context.Context, ActionParams) (interface{}, error)

// ActionParams contains everything an action needs to execute
type ActionParams struct {
	Context          context.Context
	Config           map[string]interface{}
	CollectedData    map[string]interface{}
	ExecutionContext interface{} // *messaging.ExecutionContext
	Logger           *zap.Logger
	StepName         string
	// Add other fields as needed from current ActionParams
}

// ActionDefinition contains an action function plus metadata
type ActionDefinition struct {
	Func        ActionFunc
	Category    string   // control, agent, llm, data, io, scrape, storage, html, image, memory, hitl, planning
	Description string   // Human-readable description
	Status      string   // active, experimental, deprecated
	DomainTags  []string // Optional domain-specific tags
}

// Category constants
const (
	CategoryControl  = "control"  // Workflow control: complete, await, branch
	CategoryAgent    = "agent"    // Agent management: spawn, call, discover
	CategoryLLM      = "llm"      // AI/LLM operations
	CategoryData     = "data"     // Data manipulation: validate, transform, aggregate
	CategoryIO       = "io"       // External I/O: git, http, notifications
	CategoryScrape   = "scrape"   // Web scraping operations
	CategoryStorage  = "storage"  // Storage operations: S3, hosting, assets
	CategoryHTML     = "html"     // HTML generation and manipulation
	CategoryImage    = "image"    // Image generation
	CategoryMemory   = "memory"   // Memory and caching
	CategoryHITL     = "hitl"     // Human-in-the-loop approvals
	CategoryPlanning = "planning" // Planning and evaluation
)

// Status constants
const (
	StatusActive       = "active"
	StatusExperimental = "experimental"
	StatusDeprecated   = "deprecated"
)

var (
	registry = make(map[string]ActionDefinition)
	mu       sync.RWMutex
)

// Register adds an action to the global registry.
// This is typically called from init() functions in action packages.
func Register(name string, def ActionDefinition) {
	mu.Lock()
	defer mu.Unlock()

	if def.Status == "" {
		def.Status = StatusExperimental
	}
	if def.Category == "" {
		def.Category = "uncategorized"
	}

	registry[name] = def
}

// RegisterFunc is a convenience method to register just a function with minimal metadata.
// Prefer Register() with full ActionDefinition for production actions.
func RegisterFunc(name, category, description string, fn ActionFunc) {
	Register(name, ActionDefinition{
		Func:        fn,
		Category:    category,
		Description: description,
		Status:      StatusExperimental,
	})
}

// Get returns the action definition if it exists
func Get(name string) (ActionDefinition, bool) {
	mu.RLock()
	defer mu.RUnlock()
	def, exists := registry[name]
	return def, exists
}

// GetFunc returns just the action function (for backward compatibility)
func GetFunc(name string) (ActionFunc, bool) {
	def, exists := Get(name)
	if !exists {
		return nil, false
	}
	return def.Func, true
}

// Exists checks if an action is registered
func Exists(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, exists := registry[name]
	return exists
}

// List returns all registered action names
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListByCategory returns all actions in a given category
func ListByCategory(category string) []string {
	mu.RLock()
	defer mu.RUnlock()

	var names []string
	for name, def := range registry {
		if def.Category == category {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// ListByStatus returns all actions with a given status
func ListByStatus(status string) []string {
	mu.RLock()
	defer mu.RUnlock()

	var names []string
	for name, def := range registry {
		if def.Status == status {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// GetAll returns the entire registry (for introspection/API endpoints)
func GetAll() map[string]ActionDefinition {
	mu.RLock()
	defer mu.RUnlock()

	// Return a copy to prevent modification
	result := make(map[string]ActionDefinition, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}

// Summary returns a summary of registered actions by category
func Summary() map[string]int {
	mu.RLock()
	defer mu.RUnlock()

	summary := make(map[string]int)
	for _, def := range registry {
		summary[def.Category]++
	}
	return summary
}

// Validate checks if an action can be executed based on its status
// Returns an error if the action is deprecated
func Validate(name string) error {
	def, exists := Get(name)
	if !exists {
		return fmt.Errorf("action not found: %s", name)
	}
	if def.Status == StatusDeprecated {
		return fmt.Errorf("action %s is deprecated", name)
	}
	return nil
}

// WarnIfDeprecated logs a warning if the action is deprecated
func WarnIfDeprecated(name string, logger *zap.Logger) {
	def, exists := Get(name)
	if exists && def.Status == StatusDeprecated {
		logger.Warn("Using deprecated action",
			zap.String("action", name),
			zap.String("description", def.Description),
		)
	}
}
