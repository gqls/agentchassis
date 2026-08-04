// FILE: platform/orchestration/datahelpers/declared_bool.go
//
// GetBoolFieldLoud — read a declared bool from a step config and SAY SO when the
// declaration is malformed, instead of silently substituting the default.
//
// WHY THIS EXISTS (bugs_open/193, and bugs_open/173 before it)
//
// The shape it replaces, written five times over in this codebase:
//
//	continueOnError := false
//	if coe, ok := config["continue_on_error"].(bool); ok {
//	    continueOnError = coe
//	}
//
// A step declaring `continue_on_error: "true"` — a STRING, the likely
// JSON-authoring mistake — takes the `ok == false` branch and silently keeps the
// default. Nothing errors, nothing logs, and the config-key audit accepts the key
// because the key IS spelled correctly. The author has written the key, believes
// it is in force, and it is not. The first thing a reader infers from "no
// warning" is "no problem".
//
// bugs_open/173 fixed exactly this at the SUBSTEP level and left the LOOP-level
// twin three lines away untouched — the council's `bug_historian` seat objected
// that this is the documented recurring shape *"one call site of a shared
// judgement gets the rigorous fix; the sibling stays heuristic"*, and filed
// bugs_open/193. This helper is that ticket's preferred remedy: **one
// implementation with two callers, rather than two implementations plus a test
// that they agree** (016b §9).
//
// WHAT IT IS NOT. It is not a retirement of the bool-parsing class. This estate
// carries at least five silent bool readers (`GetBoolField` above, plus private
// clones in several actions) and converging them all on one measurement would be
// a rule about the sample rather than the system. This is the LOUD reader, for
// author-facing keys where a mistype changes behaviour; sites converge onto it
// when their own measurement justifies it, and bugs_open/193 records which ones
// were deliberately left alone and why.
//
// THE FALLBACK IS A PARAMETER, and that is what lets one function serve both
// callers: the loop-level reads default to `false` (the historical default),
// while the substep-level read falls back to the LOOP's resolved value, which is
// inheritance rather than a default. Folding that difference into the helper
// would have meant two helpers again.
package datahelpers

import (
	"fmt"

	"go.uber.org/zap"
)

// GetBoolFieldLoud reads m[key] as a bool.
//
//	key absent            → fallback, silently. Saying nothing is not a mistake.
//	key present, bool     → that value.
//	key present, non-bool → fallback, plus a Warn naming the key, the declared
//	                        type, the declared value and the fallback applied.
//
// `fields` are appended to the warning so a caller can name its own context (the
// substep, the site, whatever identifies the declaration for a reader).
//
// PRESENCE IS NOT TRUTH: the type assertion is tested ON ITS OWN and never folded
// into `ok && value`. That fold is a real bug, not a style preference — it reads a
// declared `false` as no declaration at all, so an explicit opt-out silently
// becomes an inherited opt-in. `loop_error_handler.go` may write `ok && cont`
// safely only because the value it reads was stamped by this platform as a Go
// bool, never by an author.
//
// A nil logger is tolerated (the value is still returned correctly and the warning
// is dropped) so that a caller in a context without one cannot be tempted to skip
// the helper and hand-roll the silent read again.
func GetBoolFieldLoud(m map[string]interface{}, key string, fallback bool, logger *zap.Logger, fields ...zap.Field) bool {
	declared, present := m[key]
	if !present {
		return fallback
	}

	value, isBool := declared.(bool)
	if isBool {
		return value
	}

	if logger != nil {
		warnFields := append([]zap.Field{
			zap.String("config_key", key),
			zap.String("declared_type", fmt.Sprintf("%T", declared)),
			zap.Any("declared_value", declared),
			zap.Bool("fallback_applied", fallback),
		}, fields...)
		logger.Warn("Step config declares a bool key with a non-bool value; the declaration is being IGNORED and the fallback applied", warnFields...)
	}
	return fallback
}
