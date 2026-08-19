// FILE: cmd/config-key-audit/conditionalshape.go
//
// bugs_open/313. Which live conditional dots into a field whose producer
// cannot carry that path?
//
// `internal-linker.check_candidates` tested `candidate_pages.count > 0` while
// `load_candidate_pages` declared `output_format: "array"` — a bare slice with
// no `count` key. The path resolves through no strategy
// (conditional_branch_action.go: strategy 5 requires a map), the numeric arm
// returns false, and every run since the agent's creation took else_step:
// the agent completed 57 link jobs over four months and never planned a link,
// with nothing erroring and nothing retrying. The mismatch is fully visible in
// config — producer's declared shape on one step, consumer's path two lines
// below — which is what makes it checkable offline, before a run pays for it.
//
// The rule: for each string-form condition on a conditional step, take every
// dotted left-side path `F.k…`. If every producer of `output_field: F` in the
// same agent is a `query_database` step whose EFFECTIVE output_format is
// "array" (explicit, or absent — database_actions.go defaults to array), and
// the first path segment after F is not a numeric index (a dot-path may index
// an array — WFA-012), the path can never resolve against what its producer
// emits: finding. Fields with no known producer, non-query_database producers,
// and at least one object-format producer are all skipped — this mode asserts
// "can never resolve", not "looks odd".
//
// Same validation.WalkSteps traversal as every other mode here, top-level and
// nested, for the bugs_open/144 reason: two hand-written descents go blind in
// different directions and then agree.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/validation"
)

// arrayProducerFinding is one condition path that cannot resolve against its
// producer's declared output shape.
type arrayProducerFinding struct {
	Agent          string `json:"agent"`
	StepPath       string `json:"step_path"`       // the conditional step
	Condition      string `json:"condition"`       // the full condition text
	ConditionPath  string `json:"condition_path"`  // the dotted path that cannot resolve
	ProducerPath   string `json:"producer_path"`   // the step writing the field
	ProducerAction string `json:"producer_action"` // always query_database today
	OutputFormat   string `json:"output_format"`   // "array", or "" when defaulted to array
}

type producerInfo struct {
	path   string
	action string
	// format is the RAW declared output_format; "" means absent, which
	// database_actions.go treats as array — kept raw so a defaulted array is
	// distinguishable from a declared one in the finding.
	format string
}

// conditionLeftPaths extracts the left-hand field paths a string condition
// resolves, mirroring evaluateStringCondition's own parse exactly: split on
// " OR " then " AND ", and within each clause take the text before the FIRST
// operator in the runtime's probe order (!=, ==, >=, <=, >, <); a clause with
// no operator is a truthy check on the whole path. Right-hand sides are never
// resolved by the runtime (they are literals), so they are not returned.
func conditionLeftPaths(cond string) []string {
	operators := []string{" != ", " == ", " >= ", " <= ", " > ", " < "}
	var paths []string
	for _, orPart := range strings.Split(cond, " OR ") {
		for _, clause := range strings.Split(orPart, " AND ") {
			clause = strings.TrimSpace(clause)
			left := clause
			for _, op := range operators {
				if idx := strings.Index(clause, op); idx >= 0 {
					left = clause[:idx]
					break
				}
			}
			left = strings.Trim(strings.TrimSpace(left), "()")
			left = strings.TrimSpace(left)
			if left != "" {
				paths = append(paths, left)
			}
		}
	}
	return paths
}

// findArrayProducerConditions is the pure check (see findUnregisteredActions
// for why the I/O is split off).
//
// Producers are collected across the WHOLE agent, nested steps included,
// because collected_data is shared: a conditional inside a loop legitimately
// reads a field a top-level step wrote. A field written by two steps is judged
// on the full producer set — a finding requires that NO producer could satisfy
// the path, which is the only version of "can never be true" that survives a
// second producer being added.
func findArrayProducerConditions(agents []liveAgent) (findings []arrayProducerFinding, conditionals int) {
	findings = []arrayProducerFinding{}
	for _, agent := range agents {
		producers := make(map[string][]producerInfo)
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			if step.OutputField == "" {
				return
			}
			raw, _ := step.Config["output_format"].(string)
			producers[step.OutputField] = append(producers[step.OutputField],
				producerInfo{path: path, action: step.Action, format: raw})
		})

		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			if step.Action != "conditional" && step.Action != "conditional_branch" {
				return
			}
			cond, ok := step.Config["condition"].(string)
			if !ok {
				return // object-form conditions have their own vocabulary; out of scope
			}
			conditionals++
			for _, fieldPath := range conditionLeftPaths(cond) {
				parts := strings.Split(fieldPath, ".")
				if len(parts) < 2 {
					continue // a bare field is satisfiable by any producer shape
				}
				if _, err := strconv.Atoi(parts[1]); err == nil {
					continue // numeric index into the array — resolvable (WFA-012)
				}
				known := producers[parts[0]]
				if len(known) == 0 {
					continue // no known producer: input_data etc. — not this mode's claim
				}
				unsatisfiable := true
				var culprit producerInfo
				for _, p := range known {
					isArray := p.format == "" || p.format == "array"
					if p.action != "query_database" || !isArray {
						unsatisfiable = false
						break
					}
					culprit = p
				}
				if !unsatisfiable {
					continue
				}
				findings = append(findings, arrayProducerFinding{
					Agent:          agent.Type,
					StepPath:       path,
					Condition:      cond,
					ConditionPath:  fieldPath,
					ProducerPath:   culprit.path,
					ProducerAction: culprit.action,
					OutputFormat:   culprit.format,
				})
			}
		})
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Agent != findings[j].Agent {
			return findings[i].Agent < findings[j].Agent
		}
		return findings[i].StepPath < findings[j].StepPath
	})
	return findings, conditionals
}

// emitArrayProducerConditions reads the same stdin shape as --live-pairs and
// the other export-driven modes.
//
// Refuses to print over an empty or broken export, and ALSO when the export
// contained no conditional steps at all: the live fleet carries ~145, so a
// zero there means the decode went wrong, and a clean report over it would be
// a blind pass. Zero findings against a non-zero conditional count is a real,
// meaningful pass.
func emitArrayProducerConditions() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --array-producer-conditions: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--array-producer-conditions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --array-producer-conditions: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --array-producer-conditions: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	findings, conditionals := findArrayProducerConditions(agents)
	if conditionals == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --array-producer-conditions: %d agents decoded but 0 conditional steps found — "+
				"refusing to print a clean report the export could not have failed.\n", len(agents))
		os.Exit(2)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --array-producer-conditions: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr,
		"config-key-audit --array-producer-conditions: %d agents decoded (%d undecodable), %d conditional step(s) checked, %d finding(s)\n",
		len(agents), failed, conditionals, len(findings))
	if len(findings) > 0 {
		os.Exit(1)
	}
}
