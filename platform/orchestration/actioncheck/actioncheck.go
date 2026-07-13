// FILE: platform/actioncheck/actioncheck.go
// Package actioncheck provides a simple interface for checking if actions are local.

package actioncheck

// IsLocalActionFunc is the function signature for checking if an action is local
type IsLocalActionFunc func(action string) bool

// isLocalAction is the registered checker function
var isLocalAction IsLocalActionFunc

// RegisterLocalActionChecker sets the function used to check if an action is local.
// This should be called once during application startup, after the action registry is built.
func RegisterLocalActionChecker(checker IsLocalActionFunc) {
	isLocalAction = checker
}

// IsLocalAction checks if an action is available for local execution.
// Returns false if no checker has been registered.
func IsLocalAction(action string) bool {
	if isLocalAction == nil {
		return false
	}
	return isLocalAction(action)
}
