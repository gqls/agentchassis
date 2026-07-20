// Command reasoningset transforms the fix-loop's persisted reasoning into JSONL
// training/eval records.
//
// It reads the tagged JSON lines produced by extract.sql on stdin and writes one
// record per reasoning STEP, carrying provenance, three families of labels, and
// a guard block pairing the model's RAW verdict against the loop's COERCED one.
//
//	kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
//	  psql -U clients_user -d clients_db -At -f - < cmd/reasoningset/extract.sql \
//	  | go run ./cmd/reasoningset --labels labels.json --out reasoning_v1.jsonl
//
// Deliberately runs OUTSIDE the cluster and touches no DB directly — psql
// extracts, this transforms (the cmd/claimscan idiom). Running it as a pod would
// re-introduce the ephemeral-file problem training_data_export.go:3-8 retired.
//
// Bad rows are FLAGGED, never dropped: a file that silently omits the 19 poisoned
// repropose rows reads as "we have no repropose data". Consumers filter on
// input_complete.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// ── the record we emit ──────────────────────────────────────────────────────

type Record struct {
	TrajectoryID string `json:"trajectory_id"`
	RunID        string `json:"run_id"`
	StepIndex    int    `json:"step_index"`
	Task         string `json:"task"`
	StepName     string `json:"step_name"`

	InputState   string          `json:"input_state"`
	Reasoning    json.RawMessage `json:"reasoning,omitempty"`
	ReasoningRaw string          `json:"reasoning_raw,omitempty"`
	Decision     string          `json:"decision"`

	InputComplete bool   `json:"input_complete"`
	ExcludeReason string `json:"exclude_reason,omitempty"`

	Guard      *Guard     `json:"guard,omitempty"`
	Labels     Labels     `json:"labels"`
	Provenance Provenance `json:"provenance"`
}

// Guard pairs the model's raw verdict against the one the loop actually used.
// pkg/diagnose/step.go:67-101 degrades a verdict that lacks citations or an
// evidence family and PREPENDS a diagnostic sentence to NeededEvidence — a
// natural-language label for why this reasoning was invalid, generated on real
// model output. It is the most distinctive signal in the corpus.
//
// ALIGNMENT IS NOT FREE. llm_call_log has no iteration column, so a verdict call
// can only be matched to its trail entry by order — and that is only sound when
// the counts agree. They frequently do not: runs exist with 5 verdict calls and
// a trail of length 1 (retries and failed calls leave no trail entry). Pairing
// by index regardless produced a raw UNVERIFIABLE against a coerced CONFIRMED,
// which the coercion logic cannot emit — it only ever degrades. So when the
// counts disagree we record Alignment and assert nothing.
type Guard struct {
	RawDecision     string `json:"raw_decision"`
	CoercedDecision string `json:"coerced_decision,omitempty"`
	Tripped         *bool  `json:"tripped,omitempty"`
	Diagnostic      string `json:"diagnostic,omitempty"`
	StopReason      string `json:"stop_reason,omitempty"`
	Alignment       string `json:"alignment,omitempty"`
}

type Labels struct {
	SelfOutcome    string `json:"self_outcome,omitempty"`
	BenchmarkGrade string `json:"benchmark_grade,omitempty"`
	// SubsystemGrade grades a MECHANISM a run exercised (the code-lookup tier,
	// the council) rather than the diagnosis. Kept separate from
	// BenchmarkGrade so a subsystem PASS can never be read as reasoning quality.
	SubsystemGrade string `json:"subsystem_grade,omitempty"`
	Terminal       string `json:"terminal,omitempty"`
}

type Provenance struct {
	AgentType     string `json:"agent_type,omitempty"`
	Model         string `json:"model,omitempty"`
	ModelResolved string `json:"model_resolved,omitempty"`
	MaxTokens     *int   `json:"max_tokens,omitempty"`
	InputTokens   *int   `json:"input_tokens,omitempty"`
	OutputTokens  *int   `json:"output_tokens,omitempty"`
	Success       *bool  `json:"success,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	PreFixCorpus  bool   `json:"pre_fix_corpus"`
}

// ── what extract.sql gives us ───────────────────────────────────────────────

type inRow struct {
	T            string          `json:"_t"`
	TrajectoryID string          `json:"trajectory_id"`
	RunID        string          `json:"run_id"`
	StepName     string          `json:"step_name"`
	CreatedAt    string          `json:"created_at"`
	InputState   string          `json:"input_state"`
	ReasoningRaw string          `json:"reasoning_raw"`
	Provenance   Provenance      `json:"provenance"`
	Trail        []trailStep     `json:"trail"`
	VerdictRaw   json.RawMessage `json:"verdict_raw"`
	Diagnosis    json.RawMessage `json:"diagnosis"`
	Iteration    int             `json:"iteration"`
	Body         string          `json:"body"`
}

// trailStep mirrors pkg/diagnose.Step, which carries NO json tags — so the
// wire form uses Go field names and integer enums. Do not "tidy" these.
type trailStep struct {
	Iteration int    `json:"Iteration"`
	GuardStop string `json:"GuardStop"`
	Verdict   struct {
		Outcome        int    `json:"Outcome"` // 0=Unverifiable 1=Confirmed 2=Refuted
		NeededEvidence string `json:"NeededEvidence"`
		Citations      []struct {
			Tier int `json:"Tier"` // 0=static 1=state 2=runtime
		} `json:"Citations"`
	} `json:"Verdict"`
}

var outcomeName = map[int]string{0: "UNVERIFIABLE", 1: "CONFIRMED", 2: "REFUTED"}

// blindedMarkers name the fixloop documentation that must never be evaluated
// against, because it contains the benchmark's own answers.
var blindedMarkers = []string{"fixloop_eg_dartsonline", "RUBRIC_", "NOTES_running"}

// theFixLanded is the moment bugs_open/016's render fix hit the live
// fix-proposer row. Every repropose call BEFORE it reasoned against blank
// reviewer objections, so the corpus is bimodal across this instant and the two
// halves must not be pooled. Graded on the RUN start, never the step timestamp:
// a run that straddles the fix carries pre-fix config in its later steps.
var theFixLanded = time.Date(2026, 7, 18, 13, 15, 11, 0, time.UTC)

func main() {
	labelsPath := flag.String("labels", "", "path to LABELS_benchmark.json (hand-curated gold grades)")
	outPath := flag.String("out", "", "output JSONL path (default stdout)")
	flag.Parse()

	labels, err := loadLabels(*labelsPath)
	if err != nil {
		fatal("labels: %v", err)
	}

	steps, trails, bundles, err := readInput(os.Stdin)
	if err != nil {
		fatal("read: %v", err)
	}
	if len(steps) == 0 {
		fatal("no step rows on stdin — did extract.sql run? (psql -At -f -)")
	}

	records := build(steps, trails, bundles, labels)

	out := os.Stdout
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			fatal("create %s: %v", *outPath, err)
		}
		defer f.Close()
		out = f
	}
	w := bufio.NewWriter(out)
	defer w.Flush()
	enc := json.NewEncoder(w)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			fatal("encode: %v", err)
		}
	}
	report(records, len(trails), len(bundles))
}

func readInput(f *os.File) (steps []inRow, trails map[string]inRow, bundles map[string]string, err error) {
	trails = map[string]inRow{}
	bundles = map[string]string{}
	sc := bufio.NewScanner(f)
	// Bundles run to tens of KB and prompts larger still.
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue // psql banners, blank separators between the three queries
		}
		var r inRow
		if e := json.Unmarshal([]byte(line), &r); e != nil {
			// A malformed line is a real signal (a body with an embedded
			// newline would split). Report rather than skip silently.
			fmt.Fprintf(os.Stderr, "WARN: unparseable line skipped (%d bytes): %v\n", len(line), e)
			continue
		}
		switch r.T {
		case "step":
			steps = append(steps, r)
		case "trail":
			trails[r.RunID] = r
		case "bundle":
			bundles[bundleKey(r.TrajectoryID, r.Iteration)] = r.Body
		}
	}
	return steps, trails, bundles, sc.Err()
}

func bundleKey(traj string, iter int) string { return traj + "#" + fmt.Sprint(iter) }

func build(steps []inRow, trails map[string]inRow, bundles map[string]string, labels map[string]Labels) []Record {
	// Verdict calls within a run map onto trail entries in order — but only
	// when the counts agree (see Guard). Count them up front to know that.
	verdictSeen := map[string]int{}
	verdictCalls := map[string]int{}
	for _, s := range steps {
		if classify(s.StepName) == "diagnosis_verdict" {
			verdictCalls[s.RunID]++
		}
	}

	sort.SliceStable(steps, func(i, j int) bool {
		if steps[i].TrajectoryID != steps[j].TrajectoryID {
			return steps[i].TrajectoryID < steps[j].TrajectoryID
		}
		return steps[i].CreatedAt < steps[j].CreatedAt
	})

	var out []Record
	for _, s := range steps {
		task := classify(s.StepName)
		rec := Record{
			TrajectoryID: s.TrajectoryID,
			RunID:        s.RunID,
			Task:         task,
			StepName:     s.StepName,
			InputState:   s.InputState,
			ReasoningRaw: s.ReasoningRaw,
			Provenance:   s.Provenance,
		}

		// reasoning: structured when the model returned JSON, raw otherwise.
		if j := extractJSON(s.ReasoningRaw); j != nil {
			rec.Reasoning = j
			rec.ReasoningRaw = ""
		}
		rec.Decision = decisionOf(task, s.ReasoningRaw, rec.Reasoning)

		// input_complete — the bugs_open/016 filter, plus truncation.
		rec.InputComplete, rec.ExcludeReason = judgeInput(s)

		// provenance: which side of the 016 fix this run STARTED on.
		if t, ok := trails[s.RunID]; ok {
			rec.Provenance.PreFixCorpus = runStartedBeforeFix(t.CreatedAt)
		} else {
			rec.Provenance.PreFixCorpus = runStartedBeforeFix(s.CreatedAt)
		}

		// guard + step_index, verdict steps only.
		if task == "diagnosis_verdict" {
			idx := verdictSeen[s.RunID]
			verdictSeen[s.RunID]++
			rec.StepIndex = idx
			t, haveTrail := trails[s.RunID]
			raw := rec.Decision

			switch {
			case !haveTrail:
				rec.Guard = &Guard{RawDecision: raw, Alignment: "no trail for this run"}
			case len(t.Trail) != verdictCalls[s.RunID]:
				// Counts disagree — index pairing would be a guess. Say so.
				rec.Guard = &Guard{
					RawDecision: raw,
					Alignment: fmt.Sprintf("ambiguous: %d verdict calls vs %d trail entries",
						verdictCalls[s.RunID], len(t.Trail)),
				}
			case idx < len(t.Trail):
				ts := t.Trail[idx]
				if ts.Iteration > 0 {
					rec.StepIndex = ts.Iteration
				}
				coerced := outcomeName[ts.Verdict.Outcome]
				if raw == "" {
					raw = coerced
				}
				tripped := raw != coerced
				rec.Guard = &Guard{
					RawDecision:     raw,
					CoercedDecision: coerced,
					Tripped:         &tripped,
					Diagnostic:      ts.Verdict.NeededEvidence,
					StopReason:      ts.GuardStop,
					Alignment:       "by-index (counts agree)",
				}
				rec.Labels.SelfOutcome = coerced
				// The bundle is the real input_state the verdict saw; the
				// rendered prompt wraps it. Prefer the bundle when we have it.
				if b, ok := bundles[bundleKey(s.TrajectoryID, rec.StepIndex)]; ok && b != "" {
					rec.InputState = b
				}
			}
			if rec.Labels.SelfOutcome == "" {
				rec.Labels.SelfOutcome = rec.Decision
			}
		}

		// benchmark grade / terminal — hand-curated, keyed on an 8-char prefix.
		if l, ok := labels[shortID(s.TrajectoryID)]; ok {
			rec.Labels.BenchmarkGrade = l.BenchmarkGrade
			rec.Labels.SubsystemGrade = l.SubsystemGrade
			rec.Labels.Terminal = l.Terminal
		}

		out = append(out, rec)
	}
	return out
}

func classify(step string) string {
	switch {
	case step == "verdict":
		return "diagnosis_verdict"
	case strings.HasPrefix(step, "review_"):
		return "council_review"
	case step == "propose" || step == "repropose" || step == "reframe":
		return step
	default:
		return step
	}
}

// judgeInput implements the two exclusion rules that matter. Rows are FLAGGED,
// not dropped.
func judgeInput(s inRow) (bool, string) {
	// bugs_open/016: a step whose prompt rendered <no value> reasoned without
	// seeing its inputs. 100% of repropose calls in the historical corpus.
	if strings.Contains(s.InputState, "<no value>") {
		return false, "no_value_injection"
	}
	// CLAUDE.md: output_tokens == max_tokens means the completion was CUT.
	p := s.Provenance
	if p.OutputTokens != nil && p.MaxTokens != nil && *p.MaxTokens > 0 && *p.OutputTokens >= *p.MaxTokens {
		return false, "truncated"
	}
	// The fixloop docs are deliberately excluded from the loop's own input so
	// the benchmark stays honest. They reach the corpus anyway via council-gate
	// submissions that legitimately propose changes TO those scripts. Flag such
	// rows so an eval consumer can drop them; a training-only consumer may keep
	// them. Flag, don't drop — the distinction is the consumer's to make.
	for _, marker := range blindedMarkers {
		if strings.Contains(s.InputState, marker) {
			return false, "blinded_docs"
		}
	}
	// New-regime truncation surfaces as an error, with usage unset.
	if p.Success != nil && !*p.Success {
		if strings.HasPrefix(p.ErrorMessage, "response truncated") {
			return false, "truncated"
		}
		return false, "call_failed"
	}
	return true, ""
}

func runStartedBeforeFix(ts string) bool {
	t, err := parseTS(ts)
	if err != nil {
		return false
	}
	return t.Before(theFixLanded)
}

func parseTS(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999Z07:00", "2006-01-02 15:04:05.999999-07"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable timestamp %q", s)
}

// extractJSON pulls the first balanced JSON object or array out of a completion,
// which often carries prose or a ```json fence around it.
func extractJSON(s string) json.RawMessage {
	if s == "" {
		return nil
	}
	start := strings.IndexAny(s, "{[")
	if start < 0 {
		return nil
	}
	open := s[start]
	close := byte('}')
	if open == '[' {
		close = ']'
	}
	depth, inStr, esc := 0, false, false
	for i := start; i < len(s); i++ {
		c := s[i]
		switch {
		case esc:
			esc = false
		case c == '\\' && inStr:
			esc = true
		case c == '"':
			inStr = !inStr
		case inStr:
		case c == open:
			depth++
		case c == close:
			depth--
			if depth == 0 {
				cand := s[start : i+1]
				if json.Valid([]byte(cand)) {
					return json.RawMessage(cand)
				}
				return nil
			}
		}
	}
	return nil
}

func decisionOf(task, raw string, j json.RawMessage) string {
	if j != nil {
		var m map[string]any
		if json.Unmarshal(j, &m) == nil {
			for _, k := range []string{"outcome", "verdict", "decision"} {
				if v, ok := m[k].(string); ok && v != "" {
					return strings.ToUpper(v)
				}
			}
		}
	}
	if task == "diagnosis_verdict" {
		for _, o := range []string{"CONFIRMED", "REFUTED", "UNVERIFIABLE"} {
			if strings.Contains(raw, o) {
				return o
			}
		}
	}
	return ""
}

func shortID(s string) string {
	if len(s) >= 8 {
		return s[:8]
	}
	return s
}

// loadLabels reads the hand-curated gold grades. Deliberately hand-maintained:
// the benchmark grades live in ~21 lines of a 141KB notes file in four
// incompatible prose shapes, and several uses of "FAILED" mean an API 529 rather
// than a grade — a parser would confidently mislabel them.
func loadLabels(path string) (map[string]Labels, error) {
	if path == "" {
		return map[string]Labels{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Decode leniently: the file carries a "_README" array and per-entry "_note"
	// strings for the humans who maintain it. Underscore-prefixed keys are
	// documentation, not labels.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]Labels, len(raw))
	for k, v := range raw {
		if strings.HasPrefix(k, "_") {
			continue
		}
		var l Labels
		if err := json.Unmarshal(v, &l); err != nil {
			return nil, fmt.Errorf("label %q: %w", k, err)
		}
		out[shortID(k)] = l
	}
	return out, nil
}

// report prints the corpus census to stderr. It names what was EXCLUDED and why
// — a silent cap reads as "we covered everything".
func report(recs []Record, trails, bundles int) {
	byTask := map[string]int{}
	byModel := map[string]int{}
	byExclude := map[string]int{}
	byOutcome := map[string]int{}
	byGrade := map[string]int{}
	byTerminal := map[string]int{}
	guardTripped, guardUnaligned, graded, complete := 0, 0, 0, 0
	trajectories := map[string]bool{}

	for _, r := range recs {
		byTask[r.Task]++
		byModel[r.Provenance.Model]++
		trajectories[r.TrajectoryID] = true
		if r.InputComplete {
			complete++
		} else {
			byExclude[r.ExcludeReason]++
		}
		if r.Guard != nil {
			if r.Guard.Tripped != nil && *r.Guard.Tripped {
				guardTripped++
			}
			if r.Guard.CoercedDecision == "" {
				guardUnaligned++
			}
		}
		if r.Labels.BenchmarkGrade != "" {
			graded++
			byGrade[r.Labels.BenchmarkGrade]++
		}
		if r.Labels.Terminal != "" {
			byTerminal[r.Labels.Terminal]++
		}
		if r.Labels.SelfOutcome != "" {
			byOutcome[r.Labels.SelfOutcome]++
		}
	}

	e := func(f string, a ...any) { fmt.Fprintf(os.Stderr, f, a...) }
	e("\n── reasoningset census ──────────────────────────────\n")
	e("records            %d across %d trajectories\n", len(recs), len(trajectories))
	e("trails / bundles   %d / %d\n", trails, bundles)
	e("input_complete     %d  (excluded %d)\n", complete, len(recs)-complete)
	for _, k := range sortedKeys(byExclude) {
		e("  ! %-22s %d   ← flagged, NOT dropped\n", k, byExclude[k])
	}
	e("guard tripped      %d  (raw verdict != coerced verdict)\n", guardTripped)
	e("guard UNALIGNED    %d  (verdict-call count != trail length; not asserted)\n", guardUnaligned)
	e("benchmark-graded   %d\n", graded)
	e("\nby task\n")
	for _, k := range sortedKeys(byTask) {
		e("  %-22s %d\n", k, byTask[k])
	}
	e("\nby model (do NOT pool generations)\n")
	for _, k := range sortedKeys(byModel) {
		e("  %-22s %d\n", k, byModel[k])
	}
	if len(byOutcome) > 0 {
		e("\nself_outcome\n")
		for _, k := range sortedKeys(byOutcome) {
			e("  %-22s %d\n", k, byOutcome[k])
		}
	}
	if len(byGrade) > 0 {
		e("\nbenchmark_grade (gold, pre-registered rubric)\n")
		for _, k := range sortedKeys(byGrade) {
			e("  %-22s %d\n", k, byGrade[k])
		}
	}
	if len(byTerminal) > 0 {
		e("\nterminal\n")
		for _, k := range sortedKeys(byTerminal) {
			e("  %-22s %d\n", k, byTerminal[k])
		}
	}
	e("─────────────────────────────────────────────────────\n")
}

func sortedKeys(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func fatal(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "reasoningset: "+f+"\n", a...)
	os.Exit(1)
}
