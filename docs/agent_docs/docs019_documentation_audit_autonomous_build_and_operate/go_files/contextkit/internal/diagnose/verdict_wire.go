// verdict_wire.go — the JSON wire format the (LLM) Verdicter emits, and the
// parser that translates it into the domain Verdict the scaffold consumes.
//
// WHY a wire format separate from the domain Verdict: a model should emit
// human-legible STRINGS ("REFUTED", "runtime") and snake_case keys, not the
// domain type's int enums — strings are far less error-prone for a model to
// produce and for a human to script. This file is the ONE seam between the
// verdict prompt's specified output (docs/PROMPT_diagnosis_verdict.md) and the
// scaffold. A verdict-script in this format is therefore a FAITHFUL stand-in for
// the model: the same bytes the model would return drive the scripted loop.
//
// If the prompt's output schema and this struct ever drift, the loop breaks at
// this join — so this file is tested against example model outputs.
package diagnose

import (
	"encoding/json"
	"fmt"
	"strings"
)

// VerdictWire is the JSON the model returns for one iteration. Field names and
// string values MUST match docs/PROMPT_diagnosis_verdict.md's output schema.
type VerdictWire struct {
	Outcome           string         `json:"outcome"`                 // "CONFIRMED" | "REFUTED" | "UNVERIFIABLE"
	Citations         []CitationWire `json:"citations"`               // verbatim quotes from the bundle
	RevisedHypothesis string         `json:"revised_hypothesis"`      // REFUTED only
	NextScope         []string       `json:"next_scope"`              // REFUTED/UNVERIFIABLE
	NeededEvidence    string         `json:"needed_evidence"`         // UNVERIFIABLE only
	RuntimeSite       string         `json:"runtime_site"`            // optional: a runtime fault site to follow
	DataRequests      []DataRequest  `json:"data_requests,omitempty"` // optional: read-only SQL to gather next
}

type CitationWire struct {
	Tier  string `json:"tier"`            // "static" | "state" | "runtime"
	Quote string `json:"quote"`           // VERBATIM from the bundle — never paraphrased
	Where string `json:"where"`           // path:symbol, table, or log source
	Fresh string `json:"fresh,omitempty"` // freshness for state/runtime
}

func parseOutcome(s string) (Outcome, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "CONFIRMED":
		return Confirmed, nil
	case "REFUTED":
		return Refuted, nil
	case "UNVERIFIABLE", "":
		return Unverifiable, nil
	default:
		return Unverifiable, fmt.Errorf("unknown outcome %q (want CONFIRMED|REFUTED|UNVERIFIABLE)", s)
	}
}

func parseTier(s string) Tier {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "state":
		return TierState
	case "runtime":
		return TierRuntime
	default:
		return TierStatic
	}
}

// toVerdict translates one wire verdict into the domain type. An unknown outcome
// string is treated as UNVERIFIABLE (fail safe — never accidentally CONFIRM on a
// malformed verdict), with the error surfaced in NeededEvidence so it's visible
// in the trail rather than silently swallowed.
func (w VerdictWire) toVerdict() Verdict {
	oc, err := parseOutcome(w.Outcome)
	v := Verdict{
		Outcome:           oc,
		RevisedHypothesis: w.RevisedHypothesis,
		NextScope:         append([]string{}, w.NextScope...),
		NeededEvidence:    w.NeededEvidence,
		RuntimeSite:       w.RuntimeSite,
	}
	for _, c := range w.Citations {
		v.Citations = append(v.Citations, Citation{
			Tier:  parseTier(c.Tier),
			Quote: c.Quote,
			Where: c.Where,
			Fresh: c.Fresh,
		})
	}
	// Guard 2 at the parse boundary: keep only read-only data requests; drop any
	// that fail the lint, surfacing the reason in NeededEvidence so it shows in the
	// trail rather than silently vanishing. This is defence in depth — the gather
	// MUST still run survivors under a read-only transaction/role (Guard 3).
	for _, dr := range w.DataRequests {
		if lintErr := IsReadOnlySQL(dr.SQL); lintErr != nil {
			v.NeededEvidence = strings.TrimSpace(v.NeededEvidence +
				" — dropped non-read-only data_request (" + lintErr.Error() + ")")
			continue
		}
		v.DataRequests = append(v.DataRequests, dr)
	}
	if err != nil {
		v.Outcome = Unverifiable
		v.NeededEvidence = strings.TrimSpace(err.Error() + " — " + v.NeededEvidence)
	}
	return v
}

// ParseVerdict parses a single model verdict (one iteration's JSON object).
func ParseVerdict(raw []byte) (Verdict, error) {
	var w VerdictWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return Verdict{}, fmt.Errorf("verdict JSON: %w", err)
	}
	return w.toVerdict(), nil
}

// ParseVerdicts parses a JSON ARRAY of verdicts (a verdict-script: one per
// iteration). Used by cmd/diagnose's scripted mode — the same format the model
// emits, so the script is a faithful model stand-in.
func ParseVerdicts(raw []byte) ([]Verdict, error) {
	var ws []VerdictWire
	if err := json.Unmarshal(raw, &ws); err != nil {
		return nil, fmt.Errorf("verdict-script JSON: %w", err)
	}
	out := make([]Verdict, 0, len(ws))
	for _, w := range ws {
		out = append(out, w.toVerdict())
	}
	return out, nil
}
