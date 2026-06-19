// Package diagnose is the deterministic scaffold of the diagnosis loop described
// in docs/DESIGN_diagnosis_loop.md. It wraps the (read-only) gather step around a
// pluggable verdict step, enforces the convergence guards, accumulates the
// evidence trail, and re-scopes by FOLLOWING runtime/call-graph evidence rather
// than re-searching the symptom (the B4a finding: symptom retrieval has a ceiling
// for infrastructure-layer causes — DESIGN §1a).
//
// WHAT LIVES HERE (deterministic, testable without a model): loop control, the
// guards (iteration cap, scope-must-narrow, evidence-must-grow, no-thrash), the
// evidence trail, and the re-scope mechanism.
//
// WHAT DOES NOT (and is an interface, stubbed in tests): the Verdict step — the
// LLM cite-or-abstain judgement (DESIGN §2). On the chassis this is a real model
// call; here it is a deterministic fake so the scaffold + guards are testable.
//
// NON-NEGOTIABLE BOUNDARIES (DESIGN §4): the loop GATHERS (read-only) and
// PROPOSES (a diagnosis + evidence trail). It NEVER applies a fix and NEVER
// triggers a run to test a hypothesis — it reads what already happened. The
// terminal output is a diagnosis for a human.
package diagnose

import (
	"fmt"
	"sort"
	"strings"
)

// Outcome is the verdict the (LLM) verdict step returns for one iteration.
type Outcome int

const (
	// Unverifiable: the bundle does not settle the hypothesis; NeededEvidence
	// names what would (a table, a log, a symbol to add). A verdict WITHOUT a
	// citation is coerced to Unverifiable by the scaffold (DESIGN §2).
	Unverifiable Outcome = iota
	// Confirmed: the evidence directly supports the hypothesis; Citations prove it.
	Confirmed
	// Refuted: the evidence CONTRADICTS the hypothesis. This is a CORRECT and
	// expected outcome, not a failure (DESIGN §2). RevisedHypothesis + NextScope
	// carry the loop forward.
	Refuted
)

func (o Outcome) String() string {
	switch o {
	case Confirmed:
		return "CONFIRMED"
	case Refuted:
		return "REFUTED"
	default:
		return "UNVERIFIABLE"
	}
}

// Citation is a single piece of grounding evidence, tagged with its tier so a
// verdict resting on weak/stale evidence is visible (DESIGN §2).
type Tier int

const (
	TierStatic  Tier = iota // code_symbols + \d (T1)
	TierState               // dbcontext -rows (T2)
	TierRuntime             // existing logs / error_log / work-item (T3)
)

func (t Tier) String() string {
	switch t {
	case TierState:
		return "state"
	case TierRuntime:
		return "runtime"
	default:
		return "static"
	}
}

type Citation struct {
	Tier  Tier
	Quote string // the actual log line / symbol / row — never paraphrased
	Where string // path:symbol, table, or log source
	Fresh string // freshness note for state/runtime ("" for static)
}

// Verdict is the result of one verdict step.
type Verdict struct {
	Outcome           Outcome
	Citations         []Citation // REQUIRED for Confirmed/Refuted; empty ⇒ coerced Unverifiable
	RevisedHypothesis string     // Refuted: the new hypothesis
	NextScope         []string   // Refuted/Unverifiable: symbols/files to scope next
	NeededEvidence    string     // Unverifiable: what evidence would settle it
	RuntimeSite       string     // if the evidence names a runtime fault site to follow
}

// Scope is the bundle scope for one iteration. The loop narrows this over time.
type Scope struct {
	Symbols      []string // path or path:Symbol, passed to bundle as -scope
	Tables       []string // -schema-tables
	RuntimeSite  string   // -runtime-site (empty if none known)
	RuntimePage  string   // -runtime-page
	Capabilities bool     // -capabilities
}

// size is the scope's specificity measure for the narrowing guard: fewer symbols
// = narrower. (A symbol-qualified entry "f.go:Sym" is more specific than a bare
// file "f.go", but for the guard the count is the simple, robust signal.)
func (s Scope) size() int { return len(s.Symbols) }

// Step is one full iteration's record, for the evidence trail.
type Step struct {
	Iteration  int
	Hypothesis string
	Scope      Scope
	BundlePath string // where the gathered bundle was written (audit)
	Verdict    Verdict
	GuardStop  string // non-empty if a guard halted the loop after this step
}

// Gatherer runs the read-only gather (cmd/bundle) for a scope and returns the
// path to the assembled bundle. The real implementation shells out to
// cmd/bundle; tests use a fake. It MUST be side-effect-free w.r.t. the system
// under diagnosis (read-only).
type Gatherer interface {
	Gather(hypothesis string, s Scope) (bundlePath string, err error)
}

// Verdicter is the (LLM) cite-or-abstain step (DESIGN §2). Real = a model call;
// tests = a deterministic fake. It is handed the hypothesis and the bundle path.
type Verdicter interface {
	Assess(hypothesis, bundlePath string) (Verdict, error)
}

// CallGraph resolves the callers/callees of a symbol, so re-scope can FOLLOW the
// call graph from an evidence-named site (the move retrieval cannot do — §1a).
// Backed by the analyser's recorded `calls`; a nil CallGraph disables following
// (re-scope then uses NextScope verbatim).
type CallGraph interface {
	Neighbourhood(symbol string) (related []string)
}

// Config holds the convergence guards (DESIGN §3).
type Config struct {
	MaxIterations int // hard cap; past it, stop + best-effort
	// FollowCallGraph: when a verdict names a RuntimeSite or NextScope symbol,
	// expand the next scope to its call-graph neighbourhood (§1a). Off ⇒ verbatim.
	FollowCallGraph bool
}

func DefaultConfig() Config { return Config{MaxIterations: 5, FollowCallGraph: true} }

// Result is the loop's terminal output: a diagnosis for a HUMAN, plus the full
// evidence trail. Never a fix (DESIGN §4).
type Result struct {
	Status     Outcome // Confirmed = pinned; else best-effort/couldn't-pin
	Conclusion string  // the confirmed cause, or the best-effort summary
	Trail      []Step  // every iteration, for audit
	StoppedBy  string  // "confirmed" | "iteration-cap" | "scope-not-narrowing" | "evidence-not-growing" | "hypothesis-thrash"
}

// Run executes the loop. It is the deterministic core; all reasoning is in the
// Verdicter (stubbed in tests). See DESIGN §1/§1a for the flow.
func Run(seedHypothesis string, seedScope Scope, g Gatherer, v Verdicter, cg CallGraph, cfg Config) (Result, error) {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 5
	}
	var trail []Step
	hyp := seedHypothesis
	scope := seedScope

	// Guard memory across iterations.
	seenCitations := map[string]bool{} // dedup key for evidence-must-grow
	var hypHistory []string            // for no-thrash detection
	prevScopeSize := scope.size() + 1  // ensure first iteration always "narrows"

	for i := 1; i <= cfg.MaxIterations; i++ {
		// Always gather runtime when a runtime-site is known (§1a).
		bundlePath, err := g.Gather(hyp, scope)
		if err != nil {
			return Result{}, fmt.Errorf("iteration %d gather: %w", i, err)
		}

		verdict, err := v.Assess(hyp, bundlePath)
		if err != nil {
			return Result{}, fmt.Errorf("iteration %d verdict: %w", i, err)
		}
		// Coerce: a Confirmed/Refuted WITHOUT a citation is invalid → Unverifiable
		// (DESIGN §2, the item-24 rule as a machine check).
		if (verdict.Outcome == Confirmed || verdict.Outcome == Refuted) && len(verdict.Citations) == 0 {
			verdict.Outcome = Unverifiable
			verdict.NeededEvidence = "verdict gave no citation; cannot stand — " + verdict.NeededEvidence
		}

		step := Step{Iteration: i, Hypothesis: hyp, Scope: scope, BundlePath: bundlePath, Verdict: verdict}

		switch verdict.Outcome {
		case Confirmed:
			step.GuardStop = "confirmed"
			trail = append(trail, step)
			return Result{
				Status:     Confirmed,
				Conclusion: confirmConclusion(hyp, verdict),
				Trail:      trail,
				StoppedBy:  "confirmed",
			}, nil

		case Refuted:
			// The falsification move: adopt the revised hypothesis + re-scope by
			// FOLLOWING evidence (§1a), not re-searching.
			next := nextScope(scope, verdict, cg, cfg)
			// Guards (DESIGN §3) — evaluate BEFORE committing to the next round.
			if stop := guardAfter(verdict, next, prevScopeSize, seenCitations, &hypHistory, hyp); stop != "" {
				step.GuardStop = stop
				trail = append(trail, step)
				return bestEffort(trail, stop), nil
			}
			trail = append(trail, step)
			hyp = verdict.RevisedHypothesis
			prevScopeSize = scope.size()
			scope = next

		case Unverifiable:
			// Widen to get the named missing evidence; same guards apply.
			next := nextScope(scope, verdict, cg, cfg)
			if stop := guardAfter(verdict, next, prevScopeSize, seenCitations, &hypHistory, hyp); stop != "" {
				step.GuardStop = stop
				trail = append(trail, step)
				return bestEffort(trail, stop), nil
			}
			trail = append(trail, step)
			prevScopeSize = scope.size()
			scope = next
			// hypothesis unchanged on Unverifiable (we're gathering, not pivoting)
		}
	}

	return bestEffort(trail, "iteration-cap"), nil
}

// nextScope builds the next iteration's scope. It FOLLOWS the call graph from an
// evidence-named site/symbol when enabled (§1a) — the move retrieval can't do —
// otherwise uses the verdict's NextScope verbatim. Runtime-site from the verdict
// is carried so the next bundle re-gathers runtime at the named site.
func nextScope(prev Scope, v Verdict, cg CallGraph, cfg Config) Scope {
	next := Scope{
		Tables:       prev.Tables,
		RuntimeSite:  prev.RuntimeSite,
		RuntimePage:  prev.RuntimePage,
		Capabilities: prev.Capabilities,
	}
	if v.RuntimeSite != "" {
		next.RuntimeSite = v.RuntimeSite // steer runtime gather to the named site
	}

	seeds := v.NextScope
	if cfg.FollowCallGraph && cg != nil {
		expanded := map[string]bool{}
		var out []string
		add := func(s string) {
			if s != "" && !expanded[s] {
				expanded[s] = true
				out = append(out, s)
			}
		}
		for _, s := range seeds {
			add(s)
			for _, n := range cg.Neighbourhood(s) {
				add(n) // follow callers/callees toward the cause
			}
		}
		next.Symbols = out
	} else {
		next.Symbols = append([]string{}, seeds...)
	}
	sort.Strings(next.Symbols) // stable for the narrowing comparison + tests
	return next
}

// guardAfter applies the convergence guards (DESIGN §3). Returns a non-empty
// stop-reason if the loop must halt, else "".
func guardAfter(v Verdict, next Scope, prevSize int, seen map[string]bool, hypHistory *[]string, currentHyp string) string {
	// Scope must NARROW (or hold) — a widening scope is not converging.
	// Exception: Unverifiable MAY widen once to fetch named evidence, but the
	// guard still trips if it balloons beyond the previous size + a small slack.
	if next.size() > prevSize+2 {
		return "scope-not-narrowing"
	}

	// Evidence must GROW — if this verdict adds NO new citation vs everything
	// seen, the loop is spinning (DESIGN §3).
	newEvidence := false
	for _, c := range v.Citations {
		key := c.Where + "|" + c.Quote
		if !seen[key] {
			seen[key] = true
			newEvidence = true
		}
	}
	if len(v.Citations) > 0 && !newEvidence {
		return "evidence-not-growing"
	}

	// No hypothesis THRASH — oscillating between two hypotheses without new
	// discriminating evidence (DESIGN §3). Detect a repeat of a hypothesis seen
	// two steps ago.
	if v.Outcome == Refuted && v.RevisedHypothesis != "" {
		h := normaliseHyp(v.RevisedHypothesis)
		hist := *hypHistory
		for j := len(hist) - 1; j >= 0 && j >= len(hist)-2; j-- {
			if hist[j] == h && !newEvidence {
				return "hypothesis-thrash"
			}
		}
		*hypHistory = append(hist, h)
	} else {
		*hypHistory = append(*hypHistory, normaliseHyp(currentHyp))
	}

	return ""
}

func normaliseHyp(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func confirmConclusion(hyp string, v Verdict) string {
	var b strings.Builder
	fmt.Fprintf(&b, "CONFIRMED: %s\n\nGrounded by:\n", hyp)
	for _, c := range v.Citations {
		fmt.Fprintf(&b, "  [%s] %s — %q\n", c.Tier, c.Where, c.Quote)
	}
	return b.String()
}

func bestEffort(trail []Step, stoppedBy string) Result {
	var b strings.Builder
	fmt.Fprintf(&b, "NOT CONFIRMED (stopped: %s). Best-effort state after %d iteration(s):\n",
		stoppedBy, len(trail))
	if n := len(trail); n > 0 {
		last := trail[n-1]
		fmt.Fprintf(&b, "  last hypothesis: %s\n", last.Hypothesis)
		fmt.Fprintf(&b, "  last verdict: %s\n", last.Verdict.Outcome)
		if last.Verdict.NeededEvidence != "" {
			fmt.Fprintf(&b, "  still needed: %s\n", last.Verdict.NeededEvidence)
		}
	}
	fmt.Fprintf(&b, "Hand to a human with the full trail; do NOT auto-conclude.")
	return Result{Status: Unverifiable, Conclusion: b.String(), Trail: trail, StoppedBy: stoppedBy}
}
