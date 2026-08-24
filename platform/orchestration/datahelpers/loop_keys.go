package datahelpers

import "fmt"

// LoopItemKey is the CollectedData key under which loop expansion stores the
// i-th item of a named loop.
//
// WHY THIS IS A FUNCTION AND NOT A FORMAT STRING YOU RETYPE. The spelling
// "<loop_name>_item_<index>" is a CONTRACT between four places that never call
// each other: the expander writes it (loop_expansion_handler.go), setLoopVariable
// reads it back on every iteration to bind the loop variable, LoopCompleteAction
// reads it to name each iteration's result, and — as of bugs_closed/283's
// RFC_032 step 3 — component_instance_occurrence.go counts the items before the
// one being rendered in order to derive a section's per-instance element-id
// occurrence.
//
// Before this function existed the format was spelled out three separate times.
// A typo in any one of them produces a MISS, not an error: the reader gets
// "key not present", which every one of those call sites treats as a legitimate
// empty case (no item this iteration / nothing to count). So the drift would
// have presented as a quietly wrong occurrence or a silently unnamed iteration
// result, never as a failure — which is the class this estate keeps filing bugs
// about. One spelling makes the contract a compile-time fact.
//
// The loop-expansion CONFIG keys that travel with it (loop_item_index,
// loop_name, loop_var_name, loop_iteration) are declared in this package's
// frameworkStepConfigKeys — see IsFrameworkStepConfigKey. Together those two
// declarations are the whole published surface of loop expansion that an action
// may read; anything else about a loop is its internals.
func LoopItemKey(loopName string, idx int) string {
	return fmt.Sprintf("%s_item_%d", loopName, idx)
}
