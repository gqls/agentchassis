// step.go — the per-iteration decision as a PURE function, shared by Run()
// (standalone, in-process loop) and the chassis diagnose_run action (workflow-
// driven loop, where the verdict is a separate step). Extracting it keeps ONE
// source of truth for the guard + re-scope logic instead of two copies that
// could drift. Run() is refactored to call Step(); the existing tests are the
// proof the behaviour is unchanged.
package diagnose

import (
	"encoding/json"
	"strings"

	"github.com/gqls/agentchassis/internal/analysis"
)

// StepInput is one iteration's inputs: where we are (iteration, hypothesis,
// scope), the verdict just produced, the call graph for re-scope, and the guard
// memory carried from prior iterations.
type StepInput struct {
	Iteration       int
	MaxIterations   int
	Hypothesis      string
	Scope           Scope
	Verdict         Verdict
	CallGraph       CallGraph
	FollowCallGraph bool
	// MaxExpandedScope threads Config.MaxExpandedScope (see loop.go); 0 = engine default.
	MaxExpandedScope int
	SeenCitations    map[string]bool
	SeenRequests     map[string]bool
	HypHistory       []string
	PrevScopeSize    int
}

// StepDecision is the outcome of one iteration: whether to continue or stop, and
// (if continuing) the next hypothesis/scope, plus the updated guard memory and —
// when stopping — the terminal status + conclusion for emit.
type StepDecision struct {
	Decision       string // "continue" | "stop"
	StopReason     string // "" while continuing; else confirmed|iteration-cap|scope-not-narrowing|evidence-not-growing|hypothesis-thrash
	TerminalStatus Outcome
	Conclusion     string

	NextHypothesis string
	NextScope      Scope
	// NamedScopeSize is the MODEL-NAMED scope size (post-§7D-resolver, PRE
	// call-graph expansion). Advance threads it as the next PrevScopeSize so the
	// narrowing guard compares model intent iteration-over-iteration.
	NamedScopeSize int

	// updated guard memory to carry into the next iteration
	SeenCitations map[string]bool
	SeenRequests  map[string]bool
	HypHistory    []string
}

// coerceVerdict applies the standing verdict coercions in one place so
// DecideStep's decision and Advance's trail record cannot drift apart:
//   - item-24: a Confirmed/Refuted WITHOUT a citation cannot stand → Unverifiable.
//   - tier coverage: a CONFIRMED verdict must carry BOTH evidence families —
//     static (code/schema) and observed (state or runtime) — or it degrades to
//     Unverifiable and the loop continues gathering. Refutation is exempt on
//     purpose: one contradicting log line legitimately breaks a hypothesis
//     (the verdict prompt's rule-3 asymmetry); confirmation must show the
//     mechanism AND its occurrence. Benchmark run 4d43d002 (2026-07-09) showed
//     the doctrine was previously unenforced: the only guard was citations ≥ 1.
func coerceVerdict(v Verdict) Verdict {
	if (v.Outcome == Confirmed || v.Outcome == Refuted) && len(v.Citations) == 0 {
		v.Outcome = Unverifiable
		v.NeededEvidence = "verdict gave no citation; cannot stand — " + v.NeededEvidence
	}
	if v.Outcome == Confirmed && !tierCovered(v.Citations) {
		v.Outcome = Unverifiable
		v.NeededEvidence = "confirmed on one evidence family only; a confirm needs BOTH a static (code/schema) citation showing the mechanism AND a state/runtime citation showing it occurring — " + v.NeededEvidence
	}
	// F0.4d — the symptom-closure gate: a confirm must declare, per observation
	// of the ORIGINAL symptom (rendered at the top of every bundle since F0.4a),
	// that the confirmed mechanism explains it. Missing coverage or an
	// unexplained observation sends the loop back to work the residue instead
	// of stopping on a half-answer (benchmark run dd1186b9: "not a nav issue").
	if v.Outcome == Confirmed {
		if len(v.SymptomCheck) == 0 {
			v.Outcome = Unverifiable
			v.NeededEvidence = "confirm carried no symptom_check; a CONFIRMED verdict must map each observation of the ORIGINAL symptom to the confirmed mechanism, or mark it unexplained — " + v.NeededEvidence
		} else if open := unexplainedObservations(v.SymptomCheck); len(open) > 0 {
			v.Outcome = Unverifiable
			v.NeededEvidence = "confirmed mechanism leaves symptom observations UNEXPLAINED: " +
				strings.Join(open, "; ") +
				" — gather evidence that explains them (or refutes them) before confirming — " + v.NeededEvidence
		}
	}
	return v
}

// unexplainedObservations lists the symptom_check entries the confirmed
// mechanism does not account for.
func unexplainedObservations(cs []SymptomCheck) []string {
	var out []string
	for _, c := range cs {
		if !c.Explained {
			out = append(out, c.Observation)
		}
	}
	return out
}

// tierCovered reports whether the citations span both evidence families:
// static (the mechanism in code) and state/runtime (the mechanism observed).
func tierCovered(cs []Citation) bool {
	var static, observed bool
	for _, c := range cs {
		switch c.Tier {
		case TierStatic:
			static = true
		case TierState, TierRuntime:
			observed = true
		}
	}
	return static && observed
}

// DecideStep applies the verdict to the current state and returns the decision. It
// performs, in order: the verdict coercions (coerceVerdict), then per-outcome
// handling with the convergence guards (DESIGN §3), re-scoping by FOLLOWING the
// call graph (DESIGN §1a). PURE: no IO; the engine's Gather/Verdict IO happens
// outside (in Run, or in the workflow's gather/verdict steps).
func DecideStep(in StepInput) StepDecision {
	cfg := Config{MaxIterations: in.MaxIterations, FollowCallGraph: in.FollowCallGraph, MaxExpandedScope: in.MaxExpandedScope}
	verdict := coerceVerdict(in.Verdict)

	// working copies of guard memory
	seen := in.SeenCitations
	if seen == nil {
		seen = map[string]bool{}
	}
	seenReq := in.SeenRequests
	if seenReq == nil {
		seenReq = map[string]bool{}
	}
	hypHistory := append([]string{}, in.HypHistory...)

	switch verdict.Outcome {
	case Confirmed:
		return StepDecision{
			Decision:       "stop",
			StopReason:     "confirmed",
			TerminalStatus: Confirmed,
			Conclusion:     confirmConclusion(in.Hypothesis, verdict),
			NextHypothesis: in.Hypothesis,
			NextScope:      in.Scope,
			SeenCitations:  seen,
			SeenRequests:   seenReq,
			HypHistory:     hypHistory,
		}

	case Refuted:
		// Guard on the MODEL-NAMED scope (guard-vs-expansion fix, run 17933a83):
		// expansion is the engine's enrichment and must not count against "is the
		// model narrowing". nextScope runs only after the guard passes.
		named := namedScope(in.Scope, verdict)
		if stop := guardAfter(verdict, named, in.PrevScopeSize, seen, seenReq, &hypHistory, in.Hypothesis); stop != "" {
			return StepDecision{
				Decision:       "stop",
				StopReason:     stop,
				TerminalStatus: Unverifiable,
				Conclusion:     bestEffortConclusion(stop, in.Hypothesis, verdict),
				SeenCitations:  seen,
				SeenRequests:   seenReq,
				HypHistory:     hypHistory,
			}
		}
		// check the iteration cap AFTER a clean iteration
		if in.Iteration >= in.MaxIterations {
			return StepDecision{
				Decision:       "stop",
				StopReason:     "iteration-cap",
				TerminalStatus: Unverifiable,
				Conclusion:     bestEffortConclusion("iteration-cap", verdict.RevisedHypothesis, verdict),
				SeenCitations:  seen,
				SeenRequests:   seenReq,
				HypHistory:     hypHistory,
			}
		}
		next := nextScope(in.Scope, verdict, in.CallGraph, cfg)
		return StepDecision{
			Decision:       "continue",
			NamedScopeSize: named.size(),
			NextHypothesis: verdict.RevisedHypothesis,
			NextScope:      next,
			SeenCitations:  seen,
			SeenRequests:   seenReq,
			HypHistory:     hypHistory,
		}

	default: // Unverifiable
		// Guard on the MODEL-NAMED scope (guard-vs-expansion fix, run 17933a83):
		// expansion is the engine's enrichment and must not count against "is the
		// model narrowing". nextScope runs only after the guard passes.
		named := namedScope(in.Scope, verdict)
		if stop := guardAfter(verdict, named, in.PrevScopeSize, seen, seenReq, &hypHistory, in.Hypothesis); stop != "" {
			return StepDecision{
				Decision:       "stop",
				StopReason:     stop,
				TerminalStatus: Unverifiable,
				Conclusion:     bestEffortConclusion(stop, in.Hypothesis, verdict),
				SeenCitations:  seen,
				SeenRequests:   seenReq,
				HypHistory:     hypHistory,
			}
		}
		if in.Iteration >= in.MaxIterations {
			return StepDecision{
				Decision:       "stop",
				StopReason:     "iteration-cap",
				TerminalStatus: Unverifiable,
				Conclusion:     bestEffortConclusion("iteration-cap", in.Hypothesis, verdict),
				SeenCitations:  seen,
				SeenRequests:   seenReq,
				HypHistory:     hypHistory,
			}
		}
		next := nextScope(in.Scope, verdict, in.CallGraph, cfg)
		return StepDecision{
			Decision:       "continue",
			NamedScopeSize: named.size(),
			NextHypothesis: in.Hypothesis, // unchanged on Unverifiable
			NextScope:      next,
			SeenCitations:  seen,
			SeenRequests:   seenReq,
			HypHistory:     hypHistory,
		}
	}
}

// NewCallGraphFromJSON builds a call graph from analyser Output JSON bytes (the
// chassis action has the Output as a re-marshalled map; this is the entry point
// it uses). Mirrors NewCallGraphFromFile minus the file read.
func NewCallGraphFromJSON(raw []byte) (*AnalysisCallGraph, error) {
	var out analysis.Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return NewCallGraph(out), nil
}
