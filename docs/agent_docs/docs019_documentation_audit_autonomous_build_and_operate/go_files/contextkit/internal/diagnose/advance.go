// advance.go — the CHASSIS-FACING per-iteration API over the tested DecideStep
// (step.go). The standalone Run() owns its own loop; the chassis loop is
// WORKFLOW-driven (gather → verdict step → diagnose_route → back | emit), so the
// chassis needs: (1) a LoopState it can thread through workflow collected_data
// between iterations, (2) a single Advance() call per iteration, and (3) parse
// helpers for a verdict that arrives as an already-unmarshalled map (the
// execute_llm_prompt step's {"result": ...}).
//
// This adds NO new decision logic — Advance is DecideStep + state bookkeeping.
// The guard/re-scope behaviour is exactly step.go's, already covered by
// step_test.go and loop_test.go. advance_test.go below proves Advance threaded
// across calls reproduces Run().
package diagnose

import (
	"encoding/json"
	"fmt"
)

// LoopState is the loop memory the chassis threads between iterations (persisted
// in collected_data by diagnose_route, rehydrated next iteration). It is the
// externalised form of the locals Run() keeps on its stack.
type LoopState struct {
	Iteration     int             `json:"iteration"`
	MaxIterations int             `json:"max_iterations"`
	Hypothesis    string          `json:"hypothesis"`
	Scope         Scope           `json:"scope"`
	SeenCitations map[string]bool `json:"seen_citations"`
	HypHistory    []string        `json:"hyp_history"`
	PrevScopeSize int             `json:"prev_scope_size"`
	Trail         []Step          `json:"trail"`
	Follow        bool            `json:"follow_call_graph"`
}

// AdvanceResult is diagnose_route's view of one iteration's outcome.
type AdvanceResult struct {
	Stop       bool
	Status     Outcome // when Stop
	Conclusion string  // when Stop
	StoppedBy  string  // when Stop
	Hypothesis string  // when continuing: the (possibly revised) next hypothesis
	NextScope  Scope   // when continuing
}

// InitLoopState builds the first-iteration state from the seed hypothesis/scope,
// mirroring Run()'s initialisation (prevScopeSize = size+1 so the first iteration
// always "narrows"; iteration starts at 1 on the first Advance).
func InitLoopState(seedHypothesis string, seed Scope, maxIter int, follow bool) LoopState {
	if maxIter <= 0 {
		maxIter = 5
	}
	return LoopState{
		Iteration:     0,
		MaxIterations: maxIter,
		Hypothesis:    seedHypothesis,
		Scope:         seed,
		SeenCitations: map[string]bool{},
		PrevScopeSize: seed.size() + 1,
		Follow:        follow,
	}
}

// Advance applies one verdict to the state: it runs the SAME DecideStep the
// standalone loop uses, records the iteration in the trail, threads the guard
// memory, and reports continue/stop. The caller (diagnose_route) translates the
// result into a workflow route.
func Advance(st *LoopState, verdict Verdict, cg CallGraph) AdvanceResult {
	st.Iteration++

	d := DecideStep(StepInput{
		Iteration:       st.Iteration,
		MaxIterations:   st.MaxIterations,
		Hypothesis:      st.Hypothesis,
		Scope:           st.Scope,
		Verdict:         verdict,
		CallGraph:       cg,
		FollowCallGraph: st.Follow,
		SeenCitations:   st.SeenCitations,
		HypHistory:      st.HypHistory,
		PrevScopeSize:   st.PrevScopeSize,
	})

	// Record the iteration (coerce the verdict the same way DecideStep does, so
	// the trail shows what was decided on) — identical to Run()'s trail append.
	recorded := verdict
	if (recorded.Outcome == Confirmed || recorded.Outcome == Refuted) && len(recorded.Citations) == 0 {
		recorded.Outcome = Unverifiable
	}
	st.Trail = append(st.Trail, Step{
		Iteration: st.Iteration, Hypothesis: st.Hypothesis, Scope: st.Scope,
		Verdict: recorded, GuardStop: d.StopReason,
	})

	if d.Decision == "stop" {
		return AdvanceResult{
			Stop:       true,
			Status:     d.TerminalStatus,
			Conclusion: d.Conclusion,
			StoppedBy:  d.StopReason,
		}
	}

	// continue: advance the state exactly as Run() does between iterations
	st.SeenCitations = d.SeenCitations
	st.HypHistory = d.HypHistory
	st.PrevScopeSize = st.Scope.size()
	st.Hypothesis = d.NextHypothesis
	st.Scope = d.NextScope

	return AdvanceResult{
		Stop:       false,
		Hypothesis: d.NextHypothesis,
		NextScope:  d.NextScope,
	}
}

// --- state threading (collected_data is JSON-shaped maps) -------------------

// EncodeLoopState / DecodeLoopState round-trip LoopState through the
// map[string]interface{} shape collected_data uses.
func EncodeLoopState(st *LoopState) map[string]interface{} {
	b, _ := json.Marshal(st)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

func DecodeLoopState(v interface{}, out *LoopState) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// EncodeScope / EncodeTrail render sub-objects for diagnose_route's result (so
// the gather step reads scope, and emit reads the trail).
func EncodeScope(s Scope) map[string]interface{} {
	b, _ := json.Marshal(s)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

func EncodeTrail(trail []Step) []interface{} {
	b, _ := json.Marshal(trail)
	var a []interface{}
	_ = json.Unmarshal(b, &a)
	return a
}

// --- verdict parsing from an already-unmarshalled value ---------------------

// ParseVerdictValue accepts the verdict as it arrives in collected_data — an
// already-unmarshalled map (execute_llm_prompt returns {"result": <parsed>}; the
// action passes the .result value here) — and maps it to the domain Verdict via
// the same VerdictWire as ParseVerdict.
func ParseVerdictValue(v interface{}) (Verdict, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return Verdict{}, fmt.Errorf("re-marshal verdict value: %w", err)
	}
	return ParseVerdict(b)
}

// NewCallGraphFromValue builds a call graph from the analyser Output as it sits
// in collected_data (an already-unmarshalled map), re-marshalling to reuse
// NewCallGraphFromJSON.
func NewCallGraphFromValue(v interface{}) (*AnalysisCallGraph, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return NewCallGraphFromJSON(b)
}
