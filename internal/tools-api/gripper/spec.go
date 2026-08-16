// Package gripper holds the pure logic of the gripper-dossier intake: the
// spec vocabulary, the chat prompt and its reply parser, the email copy, and
// the report-status poller. Nothing here touches gin or pgx directly, so every
// rule can be unit-tested without a database or a network — the handlers and
// store packages are thin adapters over this one.
//
// WHY THE FIELD NAMES ARE THE CLUSTER'S, NOT THE DESIGN DOC'S. DESIGN §2 named
// the chat fields payload_kg / part_surface / environment / notes and §5.3 said
// the mapping to physics inputs would happen "cluster-side, in score_grippers
// input handling". Checked live 2026-08-16: the report-builder workflow's
// load_request step reads the work-item spec by NAME —
// spec->>'mass_kg', 'travel_mm', 'surface_material', 'ip_min', 'cycle_rate',
// 'mounting', 'part_geometry', 'application' — and the work-item spec is the
// island's spec verbatim plus request_id/submitted_at (report_request_pull_action.go).
// There is no renaming step anywhere in between. A spec emitted in the design's
// vocabulary would arrive and be ignored, and score_grippers would fail on
// "mass_kg is required". So the island speaks the cluster's vocabulary from the
// first turn: what the model records is exactly what the physics reads.
package gripper

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// Field describes one spec key: its JSON name, its value shape and whether the
// spec is complete without it. Order here is the order the assistant is told to
// ask in and the order missing_fields is reported in.
type Field struct {
	Name     string
	Kind     Kind
	Required bool
	// Guidance is the per-field instruction the prompt gives the model.
	Guidance string
}

// Kind is the value shape a field accepts. Anything else the model returns for
// that field is dropped rather than stored, so a downstream reader can trust the
// type without re-validating.
type Kind int

const (
	KindNumber   Kind = iota // JSON number, finite, > 0
	KindInt                  // JSON number that is a whole number >= 0
	KindText                 // non-empty string, trimmed, capped
	KindMaterial             // string from Materials
)

// maxTextLen bounds any free-text spec value. The visitor's message is already
// capped at MaxMessageRunes; this stops the model echoing more than that into a
// single field.
const maxTextLen = 500

// Materials is the closed list of surface materials the assistant may record.
// It mirrors score_grippers' gripperMaterialAliases (the μ-table the physics
// resolves): a value outside it would fail the run at scoring time with
// "matches no material in the index", so the model is offered exactly these.
var Materials = []string{"steel", "aluminium", "plastic", "glass", "cardboard", "rubber"}

// Fields is the whole vocabulary, in ask order.
var Fields = []Field{
	{Name: "mass_kg", Kind: KindNumber, Required: true,
		Guidance: "the mass of ONE workpiece in kilograms, as a number (convert grams or pounds to kg; 250 g is 0.25)"},
	{Name: "part_geometry", Kind: KindText, Required: true,
		Guidance: "the shape and key dimensions of the part in words, in millimetres (e.g. \"cylinder, 60 mm diameter, 120 mm long\")"},
	{Name: "travel_mm", Kind: KindNumber, Required: true,
		Guidance: "the jaw opening the gripper must span to pick the part, in millimetres, as a number — usually the part's width or diameter across the grip; ask for it if the geometry does not make it plain"},
	{Name: "surface_material", Kind: KindMaterial, Required: true,
		Guidance: "the material of the surface being gripped, chosen from exactly: " + strings.Join(Materials, ", ") + " (offer this list; if the visitor's material is not on it, ask which is closest and record that)"},
	{Name: "ip_min", Kind: KindInt, Required: false,
		Guidance: "the minimum IP ingress rating the environment demands, as a whole number (e.g. 54, 65, 67); leave null when the visitor says the environment is dry and indoors or gives no rating"},
	{Name: "cycle_rate", Kind: KindNumber, Required: true,
		Guidance: "picks per minute, as a number"},
	{Name: "mounting", Kind: KindText, Required: true,
		Guidance: "the robot or arm it mounts to and the flange standard if known (e.g. \"UR5e, ISO 9409-1-50-4-M6\")"},
	{Name: "application", Kind: KindText, Required: false,
		Guidance: "anything else about the application worth passing to the engineer: environment (washdown, temperature, dust), orientation, presentation, constraints"},
	{Name: "budget", Kind: KindText, Required: false,
		Guidance: "budget, if the visitor volunteers one; never ask for it"},
}

// FieldNames returns the names in ask order.
func FieldNames() []string {
	out := make([]string, 0, len(Fields))
	for _, f := range Fields {
		out = append(out, f.Name)
	}
	return out
}

// RequiredNames returns the names whose absence keeps a spec incomplete.
func RequiredNames() []string {
	out := make([]string, 0, len(Fields))
	for _, f := range Fields {
		if f.Required {
			out = append(out, f.Name)
		}
	}
	return out
}

func fieldByName(name string) (Field, bool) {
	for _, f := range Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// Spec is a validated, typed spec: only known keys, only well-shaped values.
// It is a plain map so it round-trips through jsonb unchanged, but the ONLY
// way to obtain one from outside input is Normalise, which is what makes the
// map's contents trustworthy.
type Spec map[string]interface{}

// Normalise takes an untrusted map (from the model, from the plain-form
// fallback, from a stored row) and returns the subset that is a well-shaped
// value for a known field. Unknown keys and mis-typed values are dropped, not
// errored: the model is told to send null for anything it does not know, and
// a wrong type is treated as "does not know" rather than as a reason to fail
// the turn — the missing_fields list will simply keep asking.
func Normalise(raw map[string]interface{}) Spec {
	out := Spec{}
	for k, v := range raw {
		f, ok := fieldByName(k)
		if !ok || v == nil {
			continue
		}
		if nv, ok := coerce(f.Kind, v); ok {
			out[k] = nv
		}
	}
	return out
}

func coerce(kind Kind, v interface{}) (interface{}, bool) {
	switch kind {
	case KindNumber:
		n, ok := asFloat(v)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) || n <= 0 {
			return nil, false
		}
		return n, true
	case KindInt:
		n, ok := asFloat(v)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 || n != math.Trunc(n) {
			return nil, false
		}
		return int(n), true
	case KindText:
		s, ok := v.(string)
		s = strings.TrimSpace(s)
		if !ok || s == "" {
			return nil, false
		}
		if len(s) > maxTextLen {
			s = s[:maxTextLen]
		}
		return s, true
	case KindMaterial:
		s, ok := v.(string)
		if !ok {
			return nil, false
		}
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "aluminum" { // the alias score_grippers also accepts
			s = "aluminium"
		}
		for _, m := range Materials {
			if s == m {
				return m, true
			}
		}
		return nil, false
	}
	return nil, false
}

func asFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

// Merge applies a turn's spec on top of the stored one. NON-NULL NEVER
// REGRESSES: a key the model omitted or nulled this turn keeps its stored
// value. The model is asked for the full spec every turn, and it will
// occasionally forget one; forgetting must not un-answer a question the
// visitor already answered. A visitor who genuinely changes their mind gets
// the new value because that arrives non-null.
func Merge(stored, turn Spec) Spec {
	out := Spec{}
	for k, v := range stored {
		out[k] = v
	}
	for k, v := range turn {
		out[k] = v
	}
	return out
}

// Missing returns the required fields the spec does not yet hold, in ask order.
func Missing(s Spec) []string {
	var out []string
	for _, f := range Fields {
		if !f.Required {
			continue
		}
		if _, ok := s[f.Name]; !ok {
			out = append(out, f.Name)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
}

// Complete reports whether every required field is present.
func Complete(s Spec) bool { return len(Missing(s)) == 0 }

// ForCluster returns the spec as the cluster's puller will store it — the
// same map, keys sorted for a stable NDJSON line. Deliberately does NOT add
// request_id or submitted_at: the puller adds those itself, and a duplicate
// key with a different value would be a quiet argument between two writers.
func ForCluster(s Spec) map[string]interface{} {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make(map[string]interface{}, len(s))
	for _, k := range keys {
		out[k] = s[k]
	}
	return out
}

// Summary is a one-line human rendering used in log lines and email subjects.
// Structural, not verbatim: it names which fields are set, never their text.
func Summary(s Spec) string {
	set := make([]string, 0, len(s))
	for _, f := range Fields {
		if _, ok := s[f.Name]; ok {
			set = append(set, f.Name)
		}
	}
	return fmt.Sprintf("%d/%d fields set [%s]", len(set), len(Fields), strings.Join(set, ","))
}
