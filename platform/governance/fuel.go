// FILE: platform/governance/fuel.go
package governance

import (
	"fmt"
	"strconv"
)

const (
	// FuelHeader is the standard Kafka message header key for the fuel budget
	FuelHeader = "fuel_budget"
)

// CostTable defines the "price" in fuel units for various agent actions
var CostTable = map[string]int{
	"default_step":                   1,
	"fan_out":                        5,
	"ai_text_generate_claude_haiku":  10,
	"ai_text_generate_claude_sonnet": 25,
	"ai_text_generate_claude_opus":   50,
	"ai_image_generate_sdxl":         40,
	"web_search":                     5,
	"database_query":                 1,
	"memory_store":                   2,
	"memory_search":                  2,
	"pause_for_human_input":          0, // No cost for waiting
}

// FuelManager provides methods for checking and managing task fuel
type FuelManager struct{}

// NewFuelManager creates a new fuel manager
func NewFuelManager() *FuelManager {
	return &FuelManager{}
}

// GetCost returns the fuel cost for a given action
func (fm *FuelManager) GetCost(action string) int {
	if cost, ok := CostTable[action]; ok {
		return cost
	}
	// Return a default cost if the specific action isn't priced
	return CostTable["default_step"]
}

// HasEnoughFuel checks if the current budget is sufficient for an action
func (fm *FuelManager) HasEnoughFuel(currentFuel int, action string) bool {
	cost := fm.GetCost(action)
	return currentFuel >= cost
}

// DeductFuel subtracts the cost of an action from the current budget
func (fm *FuelManager) DeductFuel(currentFuel int, action string) int {
	cost := fm.GetCost(action)
	return currentFuel - cost
}

// GetFuelFromHeader safely parses the fuel value from Kafka message headers
func GetFuelFromHeader(headers map[string]string) (int, error) {
	fuelStr, ok := headers[FuelHeader]
	if !ok {
		return 0, fmt.Errorf("'%s' header not found", FuelHeader)
	}
	fuel, err := strconv.Atoi(fuelStr)
	if err != nil {
		return 0, fmt.Errorf("invalid fuel value in header: %w", err)
	}
	return fuel, nil
}

// SetFuelHeader sets the fuel budget in the headers map
func SetFuelHeader(headers map[string]string, fuel int) {
	headers[FuelHeader] = strconv.Itoa(fuel)
}
