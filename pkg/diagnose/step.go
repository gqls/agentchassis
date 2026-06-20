// step.go — the per-iteration decision as a PURE function, shared by Run()
// (standalone, in-process loop) and the chassis diagnose_run action (workflow-
// driven loop, where the verdict is a separate step). Extracting it keeps ONE
// source of truth for the guard + re-scope logic instead of two copies that
// could drift. Run() is refactored to call Step(); the existing tests are the
// proof the behaviour is unchanged.
package diagnose

import (
	"encoding/json"

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
	SeenCitations   map[string]bool
	HypHistory      []string
	PrevScopeSize   int
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

	// updated guard memory to carry into the next iteration
	SeenCitations map[string]bool
	HypHistory    []string
}

// DecideStep applies the verdict to the current state and returns the decision. It
// performs, in order: the no-citation coercion (item-24), then per-outcome
// handling with the convergence guards (DESIGN §3), re-scoping by FOLLOWING the
// call graph (DESIGN §1a). PURE: no IO; the engine's Gather/Verdict IO happens
// outside (in Run, or in the workflow's gather/verdict steps).
func DecideStep(in StepInput) StepDecision {
	cfg := Config{MaxIterations: in.MaxIterations, FollowCallGraph: in.FollowCallGraph}
	verdict := in.Verdict

	// item-24: a Confirmed/Refuted WITHOUT a citation cannot stand → Unverifiable.
	if (verdict.Outcome == Confirmed || verdict.Outcome == Refuted) && len(verdict.Citations) == 0 {
		verdict.Outcome = Unverifiable
		verdict.NeededEvidence = "verdict gave no citation; cannot stand — " + verdict.NeededEvidence
	}

	// working copies of guard memory
	seen := in.SeenCitations
	if seen == nil {
		seen = map[string]bool{}
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
			HypHistory:     hypHistory,
		}

	case Refuted:
		next := nextScope(in.Scope, verdict, in.CallGraph, cfg)
		if stop := guardAfter(verdict, next, in.PrevScopeSize, seen, &hypHistory, in.Hypothesis); stop != "" {
			return StepDecision{
				Decision:       "stop",
				StopReason:     stop,
				TerminalStatus: Unverifiable,
				Conclusion:     bestEffortConclusion(stop, in.Hypothesis, verdict),
				SeenCitations:  seen,
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
				HypHistory:     hypHistory,
			}
		}
		return StepDecision{
			Decision:       "continue",
			NextHypothesis: verdict.RevisedHypothesis,
			NextScope:      next,
			SeenCitations:  seen,
			HypHistory:     hypHistory,
		}

	default: // Unverifiable
		next := nextScope(in.Scope, verdict, in.CallGraph, cfg)
		if stop := guardAfter(verdict, next, in.PrevScopeSize, seen, &hypHistory, in.Hypothesis); stop != "" {
			return StepDecision{
				Decision:       "stop",
				StopReason:     stop,
				TerminalStatus: Unverifiable,
				Conclusion:     bestEffortConclusion(stop, in.Hypothesis, verdict),
				SeenCitations:  seen,
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
				HypHistory:     hypHistory,
			}
		}
		return StepDecision{
			Decision:       "continue",
			NextHypothesis: in.Hypothesis, // unchanged on Unverifiable
			NextScope:      next,
			SeenCitations:  seen,
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
