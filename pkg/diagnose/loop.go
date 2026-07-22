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
	Citations         []Citation    // REQUIRED for Confirmed/Refuted; empty ⇒ coerced Unverifiable
	RevisedHypothesis string        // Refuted: the new hypothesis
	NextScope         []string      // Refuted/Unverifiable: symbols/files to scope next
	NeededEvidence    string        // Unverifiable: what evidence would settle it
	RuntimeSite       string        // if the evidence names a runtime fault site to follow
	DataRequests      []DataRequest // Refuted/Unverifiable: read-only SQL the gather should run next
	//                                 (the DATA analogue of NextScope). Each is linted to read-only
	//                                 at parse (Guard 2) and MUST run under a read-only transaction/role
	//                                 (Guard 3); the prompt instructs SELECT-only (Guard 1).
	CodeRequests []CodeRequest // Refuted/Unverifiable: code-search questions the gather should
	//                            answer from the code_symbols index (the BREADTH analogue of
	//                            NextScope's depth). Kind is validated to a closed set at parse;
	//                            no model-written SQL is ever executed.
	SymptomCheck []SymptomCheck `json:"symptom_check,omitempty"` // Confirmed: REQUIRED coverage of the ORIGINAL symptom (F0.4d)
}

// SymptomCheck is one entry of a CONFIRMED verdict's mapping from the ORIGINAL
// symptom's observations back to the confirmed mechanism — the symptom-closure
// gate (F0.4d). Benchmark run dd1186b9 (2026-07-09) confirmed a mechanism that
// explained half the reported symptom and dismissed the rest ("not a nav
// issue") — well-cited, and still not an answer to what was asked. A confirm
// whose entries are missing or not all Explained is coerced to Unverifiable in
// coerceVerdict, so the loop keeps working the residue instead of stopping on
// a partial. Observation strings are model-authored: the gate enforces that
// coverage was DECLARED, not that the decomposition is complete — that
// judgement stays with the prompt and the human.
type SymptomCheck struct {
	Observation string `json:"observation"`   // one observation from the ORIGINAL symptom
	Explained   bool   `json:"explained"`     // does the confirmed mechanism produce it?
	How         string `json:"how,omitempty"` // one line: how it does — or why it remains unexplained
	// F0.6 (benchmark run 5179a2ea): two refinements the first live coverage
	// exposed. Context marks a COMPARATIVE clause (e.g. "site X works on the
	// same platform") that is background, not an observation of the defect —
	// exempt from the explained/unexplained accounting, so the model is not
	// forced to grade-inflate clauses it structurally cannot verify. Cites are
	// indices into the verdict's citations array: an explained entry must be
	// grounded by at least one in-range index, or it degrades — run 4 marked
	// two clauses explained whose own text said "unverifiable from this bundle".
	Context bool  `json:"context,omitempty"`
	Cites   []int `json:"cites,omitempty"`
}

// DataRequest is one read-only query the verdict asks the next gather to run,
// when the bundle doesn't settle the hypothesis and specific DB evidence would.
// All-strings, so one type serves both the domain Verdict and the wire (unlike
// Outcome/Tier, which are enums needing a wire variant).
type DataRequest struct {
	SQL string `json:"sql"`           // a SINGLE read-only SELECT (or WITH … SELECT)
	Why string `json:"why,omitempty"` // what this query would settle (for the trail)
}

// CodeRequest is one code-SEARCH question the verdict asks the next gather to
// answer from the code_symbols index — the breadth counterpart to NextScope's
// depth. NextScope follows the call graph ("what does this code call?"), which
// only ever reaches code the current scope already touches; a CodeRequest asks
// "does this mechanism exist ELSEWHERE?" — another implementation, a second call
// site, anything referencing X — which is precisely where cross-cutting causes
// hide. Answered by the same diagnose_code_lookup action the council's verify
// tier uses (one convention across both halves of the loop).
//
// Unlike DataRequest, nothing model-written is ever executed: the SQL is fixed
// per kind inside the action and Query arrives only as a bind parameter, so
// there is no lint/containment analogue to Guards 2-3 — the containment is that
// the model writes no SQL at all. Kind IS validated (closed set) at parse.
//
// All-strings, so one type serves both the domain Verdict and the wire.
type CodeRequest struct {
	Kind  string `json:"kind"`          // "symbol" | "content" | "ls" — a CLOSED set
	Query string `json:"query"`         // the pattern to match (a bind parameter, never SQL)
	Why   string `json:"why,omitempty"` // what this would settle (for the trail)
}

// codeRequestKeySep separates kind from query inside a CodeRequestKey. It must
// survive the collected_data jsonb round-trip: the original NUL ("\x00")
// delimiter marshalled to the escape sequence backslash-u0000, the ONE Unicode escape Postgres jsonb
// rejects (22P05), so the route step's first persist after any code request
// killed the whole run (bugs_open/056). Unit separator (0x1F) is jsonb-safe,
// and kind is a closed validated set, so neither half can contain it.
const codeRequestKeySep = "\x1f"

// CodeRequestKey is the identity of a code request for guard/dedup purposes:
// kind + case-folded query. Shared by the spin guard (loop.go), the route's
// cumulative re-forwarding, and the action's dedup, so all three agree on what
// "the same question" means.
func CodeRequestKey(kind, query string) string {
	return strings.ToLower(strings.TrimSpace(kind)) + codeRequestKeySep + strings.ToLower(strings.TrimSpace(query))
}

// SplitCodeRequestKey is CodeRequestKey's single inverse. Every reader must
// come through here: a builder and a reader holding separate copies of the
// separator break cross-iteration dedup silently when one moves — it reads as
// "malformed keys" in the route's counter, not as a crash.
func SplitCodeRequestKey(key string) (kind, query string, ok bool) {
	return strings.Cut(key, codeRequestKeySep)
}

// ValidCodeRequestKind reports whether kind is one the code-lookup action can
// answer. Kept here (not in the action) because the verdict parser drops invalid
// kinds at the wire boundary — the same closed set must be enforced at both ends.
func ValidCodeRequestKind(kind string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "symbol", "content", "ls":
		return true
	}
	return false
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
	// MaxExpandedScope caps nextScope's call-graph enrichment (named entries are
	// ALWAYS all kept). 0 = engine default (defaultMaxExpandedScope); <0 =
	// unlimited (explicit opt-out). Guard-vs-expansion fix, run 17933a83.
	MaxExpandedScope int
}

// defaultMaxExpandedScope bounds Neighbourhood enrichment when the caller does
// not set Config.MaxExpandedScope: seed-sized (12) plus modest headroom, so a
// rich graph cannot flood the bundle (B4a: irrelevant context buries signal).
const defaultMaxExpandedScope = 18

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
	seenRequests := map[string]bool{}  // dedup key for in-flight data_requests (guardAfter)
	seenCodeReqs := map[string]bool{}  // dedup key for in-flight code_requests (guardAfter)
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

		// One iteration of the shared decision logic (guards + re-scope + the
		// no-citation coercion). Run() owns the IO above; Step() owns the decision.
		d := DecideStep(StepInput{
			Iteration:        i,
			MaxIterations:    cfg.MaxIterations,
			Hypothesis:       hyp,
			Scope:            scope,
			Verdict:          verdict,
			CallGraph:        cg,
			FollowCallGraph:  cfg.FollowCallGraph,
			SeenCitations:    seenCitations,
			SeenRequests:     seenRequests,
			SeenCodeRequests: seenCodeReqs,
			HypHistory:       hypHistory,
			PrevScopeSize:    prevScopeSize,
		})

		// Record this iteration in the trail coerced the same way DecideStep
		// decided on it — one shared coercion (coerceVerdict), no drift.
		recorded := coerceVerdict(verdict)
		trail = append(trail, Step{
			Iteration: i, Hypothesis: hyp, Scope: scope, BundlePath: bundlePath,
			Verdict: recorded, GuardStop: d.StopReason,
		})

		if d.Decision == "stop" {
			if d.TerminalStatus == Confirmed {
				return Result{Status: Confirmed, Conclusion: d.Conclusion, Trail: trail, StoppedBy: d.StopReason}, nil
			}
			return Result{Status: Unverifiable, Conclusion: d.Conclusion, Trail: trail, StoppedBy: d.StopReason}, nil
		}

		// continue: carry the advanced state into the next iteration
		seenCitations = d.SeenCitations
		seenRequests = d.SeenRequests
		seenCodeReqs = d.SeenCodeRequests
		hypHistory = d.HypHistory
		prevScopeSize = scope.size()
		hyp = d.NextHypothesis
		scope = d.NextScope
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
		limit := cfg.MaxExpandedScope
		if limit == 0 {
			limit = defaultMaxExpandedScope
		}
		// Named entries first — ALWAYS all included (they are the model's ask);
		// the limit caps only the engine's enrichment. limit < 0 = unlimited.
		for _, s := range seeds {
			add(s)
		}
		for _, s := range seeds {
			for _, n := range cg.Neighbourhood(s) {
				if limit > 0 && len(out) >= limit {
					break
				}
				add(n) // follow callers/callees toward the cause
			}
			if limit > 0 && len(out) >= limit {
				break
			}
		}
		next.Symbols = out
	} else {
		next.Symbols = append([]string{}, seeds...)
	}
	sort.Strings(next.Symbols) // stable for the narrowing comparison + tests
	return next
}

// namedScope is the MODEL-NAMED next scope: the verdict's next_scope entries
// (already §7D-resolved by the route action), deduped, runtime fields carried —
// and NO call-graph expansion. The narrowing guard measures THIS; nextScope's
// enrichment is applied only after the guard passes.
func namedScope(prev Scope, v Verdict) Scope {
	s := Scope{
		Tables:       prev.Tables,
		RuntimeSite:  prev.RuntimeSite,
		RuntimePage:  prev.RuntimePage,
		Capabilities: prev.Capabilities,
	}
	if v.RuntimeSite != "" {
		s.RuntimeSite = v.RuntimeSite
	}
	seen := map[string]bool{}
	for _, e := range v.NextScope {
		e = strings.TrimSpace(e)
		if e != "" && !seen[e] {
			seen[e] = true
			s.Symbols = append(s.Symbols, e)
		}
	}
	sort.Strings(s.Symbols)
	return s
}

// guardAfter applies the convergence guards (DESIGN §3). Returns a non-empty
// stop-reason if the loop must halt, else "". `next` is the MODEL-NAMED scope
// (see namedScope) — the narrowing check measures model intent, never the
// engine's own call-graph enrichment (guard-vs-expansion fix, run 17933a83:
// on the first real 515-file graph, unbounded expansion of six named symbols
// tripped scope-not-narrowing at iteration 1; the stale 69-file graph had
// masked the difference because empty neighbourhoods made expansion a no-op).
func guardAfter(v Verdict, next Scope, prevSize int, seen, seenReq, seenCodeReq map[string]bool, hypHistory *[]string, currentHyp string) string {
	// Scope must NARROW (or hold) — a widening scope is not converging.
	// Exception: Unverifiable MAY widen once to fetch named evidence, but the
	// guard still trips if it balloons beyond the previous size + a small slack.
	if next.size() > prevSize+2 {
		return "scope-not-narrowing"
	}

	// Evidence must GROW — but a verdict that issues a NEW (previously-unseen)
	// read-only data_request is making progress, not spinning: its answer arrives in
	// the NEXT gather, so the stale-citation test must not stop the loop one iteration
	// before it. Issued requests are tracked like citations, so a RE-issue of the same
	// query does NOT count as progress (a true re-issue spin still trips; the iteration
	// cap bounds the worst case). (DESIGN §3.)
	newEvidence := false
	for _, c := range v.Citations {
		key := c.Where + "|" + c.Quote
		if !seen[key] {
			seen[key] = true
			newEvidence = true
		}
	}
	newRequest := false
	if seenReq != nil {
		for _, dr := range v.DataRequests {
			key := strings.TrimSpace(dr.SQL)
			if key == "" {
				continue
			}
			if !seenReq[key] {
				seenReq[key] = true
				newRequest = true
			}
		}
	}
	// A NEW code_request is progress on exactly the same reasoning as a new
	// data_request: the answer arrives in the next gather, so stopping now would
	// halt the loop one iteration before the evidence it just asked for. Tracked
	// in their OWN map, not seenReq — the route re-forwards seenReq's keys as SQL
	// (withPriorRequests), so a code question living there would be re-issued as a
	// query and silently dropped by the read-only lint. Same key as the action's
	// dedup, so "the same question" means one thing across guard, route and action.
	if seenCodeReq != nil {
		for _, cr := range v.CodeRequests {
			if strings.TrimSpace(cr.Query) == "" {
				continue
			}
			key := CodeRequestKey(cr.Kind, cr.Query)
			if !seenCodeReq[key] {
				seenCodeReq[key] = true
				newRequest = true
			}
		}
	}
	if len(v.Citations) > 0 && !newEvidence && !newRequest {
		return "evidence-not-growing"
	}

	// No hypothesis THRASH — oscillating between two hypotheses without new
	// discriminating evidence (DESIGN §3). Detect a repeat of a hypothesis seen
	// two steps ago.
	if v.Outcome == Refuted && v.RevisedHypothesis != "" {
		h := normaliseHyp(v.RevisedHypothesis)
		hist := *hypHistory
		for j := len(hist) - 1; j >= 0 && j >= len(hist)-2; j-- {
			if hist[j] == h && !newEvidence && !newRequest {
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
	// F0.4d: show the human the coverage the gate enforced — which observation
	// of the original symptom this mechanism produces, and how.
	if len(v.SymptomCheck) > 0 {
		b.WriteString("\nSymptom coverage:\n")
		for _, s := range v.SymptomCheck {
			mark := "explained"
			switch {
			case s.Context:
				mark = "context"
			case !s.Explained:
				mark = "UNEXPLAINED"
			}
			fmt.Fprintf(&b, "  [%s] %s — %s\n", mark, s.Observation, s.How)
		}
	}
	return b.String()
}

func bestEffort(trail []Step, stoppedBy string) Result {
	var lastHyp string
	var lastVerdict Verdict
	if n := len(trail); n > 0 {
		lastHyp = trail[n-1].Hypothesis
		lastVerdict = trail[n-1].Verdict
	}
	return Result{
		Status:     Unverifiable,
		Conclusion: bestEffortConclusion(stoppedBy, lastHyp, lastVerdict),
		Trail:      trail,
		StoppedBy:  stoppedBy,
	}
}

// bestEffortConclusion renders the not-confirmed summary. Shared by bestEffort
// (standalone Run) and Step (the chassis per-iteration decision) so the wording
// is identical in both paths.
func bestEffortConclusion(stoppedBy, lastHypothesis string, lastVerdict Verdict) string {
	var b strings.Builder
	fmt.Fprintf(&b, "NOT CONFIRMED (stopped: %s).\n", stoppedBy)
	if lastHypothesis != "" {
		fmt.Fprintf(&b, "  last hypothesis: %s\n", lastHypothesis)
		fmt.Fprintf(&b, "  last verdict: %s\n", lastVerdict.Outcome)
		if lastVerdict.NeededEvidence != "" {
			fmt.Fprintf(&b, "  still needed: %s\n", lastVerdict.NeededEvidence)
		}
	}
	fmt.Fprintf(&b, "Hand to a human with the full trail; do NOT auto-conclude.")
	return b.String()
}
