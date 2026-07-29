// FILE: platform/orchestration/actions/subworkflow_live_dryrun_test.go
//
// bugs_open/144 — replays the workflow validator over an export of the LIVE agent
// definitions, so a change to what "validated" means can be measured against the
// fleet before it ships rather than discovered by it.
//
// This is the instrument that made the 144 fix safe to ship with hard errors: it
// showed that 0 of the 85 nested steps across 18 live agents failed any of the new
// rules. "A naive recursion could start rejecting workflows that run today" was the
// whole risk, and the only honest answer to it is a measurement.
//
// SKIPPED unless pointed at an export — it needs the cluster, and a test that
// silently needs a database is a test that rots. Get the export with:
//
//	kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
//	  psql -U clients_user -d clients_db -tAc "
//	  SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow'))
//	  FROM agent_definitions
//	  WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active
//	    AND default_config ? 'workflow';" > /tmp/live_workflows.json
//
//	SUBWF_LIVE_EXPORT=/tmp/live_workflows.json go test ./platform/orchestration/actions/ \
//	  -run TestLiveDefinitionsPassSubWorkflowValidation -v
//
// Verify the export parses before trusting a clean result: a truncated kubectl exec
// exits 0 (WRONG_CALLS.md 2026-07-29), and a short file would read here as a small,
// clean fleet. The decode-failure count below is printed for exactly that reason.
//
// It lives in this package because it needs the real action registry — "does this
// action require a topic?" is answered by IsLocalAction, and a fake registry would
// measure a fleet that does not exist.
package actions

import (
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/validation"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type liveAgentWorkflow struct {
	Type     string              `json:"type"`
	Workflow models.WorkflowPlan `json:"workflow"`
}

func TestLiveDefinitionsPassSubWorkflowValidation(t *testing.T) {
	path := os.Getenv("SUBWF_LIVE_EXPORT")
	if path == "" {
		t.Skip("set SUBWF_LIVE_EXPORT to a live agent_definitions export — see the file header")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read the export: %v", err)
	}

	// Decoded per agent rather than in one pass: a single step whose `timeout` is a
	// string rather than a number would otherwise take the whole file down and the
	// run would report nothing rather than reporting less.
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("export is not a JSON array — a truncated kubectl export exits 0: %v", err)
	}

	var (
		decoded        int
		decodeFailed   []string
		subWorkflows   = map[string]bool{}
		nestedSteps    int
		topSteps       int
		failures       []string
		alreadyFailing []string
		warnings       = map[string]int{}
	)

	for _, row := range rows {
		var agent liveAgentWorkflow
		if err := json.Unmarshal(row, &agent); err != nil {
			decodeFailed = append(decodeFailed, err.Error())
			continue
		}
		decoded++

		validation.WalkSteps(agent.Workflow, func(stepPath string, _ models.Step, nested bool) {
			if nested {
				nestedSteps++
				// The sub-workflow is everything up to the final ".<step>".
				for i := len(stepPath) - 1; i >= 0; i-- {
					if stepPath[i] == '.' {
						subWorkflows[agent.Type+"/"+stepPath[:i]] = true
						break
					}
				}
				return
			}
			topSteps++
		})

		// The control. A plan with its sub-workflows stripped out is what the
		// validator saw BEFORE bugs_open/144, so anything failing here fails today
		// too and is not attributable to the recursion. Without this the run reports
		// "3 live definitions rejected" and leaves the reader to decide from the
		// wording of an error message whose fault that is.
		core, _ := observer.New(zapcore.WarnLevel)
		preExisting := validation.NewWorkflowValidator(zap.New(core)).ValidateWorkflow(withoutSubWorkflows(agent.Workflow))

		core, logs := observer.New(zapcore.WarnLevel)
		err := validation.NewWorkflowValidator(zap.New(core)).ValidateWorkflow(agent.Workflow)

		switch {
		case err != nil && preExisting == nil:
			failures = append(failures, agent.Type+": "+err.Error())
		case err != nil:
			alreadyFailing = append(alreadyFailing, agent.Type+": "+err.Error())
		}

		for _, entry := range logs.All() {
			warnings[entry.Message]++
		}
	}

	t.Logf("agents decoded: %d (decode failures: %d)", decoded, len(decodeFailed))
	t.Logf("top-level steps: %d · sub-workflows: %d · nested steps: %d", topSteps, len(subWorkflows), nestedSteps)

	msgs := make([]string, 0, len(warnings))
	for msg := range warnings {
		msgs = append(msgs, msg)
	}
	sort.Strings(msgs)
	for _, msg := range msgs {
		t.Logf("warning ×%d: %s", warnings[msg], msg)
	}

	// A zero nested-step count means the export is wrong or the traversal broke, not
	// that the fleet is clean — without this the test would pass loudest exactly when
	// it had stopped looking.
	if nestedSteps == 0 {
		t.Fatal("no nested steps found in the export: the traversal or the export is broken, " +
			"not the fleet (18 live agents carried sub-workflows on 2026-07-29)")
	}

	// Reported, not failed: these definitions are rejected by the validator as it
	// stands today, with or without sub-workflow recursion. They are a finding about
	// the fleet, not about this change — and burying them in a failure about
	// something else is how a real defect gets read as noise from someone's patch.
	for _, f := range alreadyFailing {
		t.Logf("ALREADY REJECTED BEFORE THIS CHANGE (pre-existing, not caused by sub-workflow validation) — %s", f)
	}

	for _, f := range failures {
		t.Errorf("NEWLY REJECTED BY SUB-WORKFLOW VALIDATION — %s", f)
	}
	for _, f := range decodeFailed {
		t.Logf("undecodable agent row (excluded from the count above): %s", f)
	}
}

// withoutSubWorkflows returns the plan the validator saw before bugs_open/144: the
// same steps with every nested sub-workflow removed. Copies the config map rather
// than mutating it, so the caller's plan is untouched and the two verdicts are
// genuinely two verdicts about the same input.
func withoutSubWorkflows(plan models.WorkflowPlan) models.WorkflowPlan {
	stripped := models.WorkflowPlan{StartStep: plan.StartStep, Steps: make(map[string]models.Step, len(plan.Steps))}
	for name, step := range plan.Steps {
		if len(step.Config) > 0 {
			config := make(map[string]interface{}, len(step.Config))
			for k, v := range step.Config {
				if k == "sub_workflow" || k == "substeps" {
					continue
				}
				config[k] = v
			}
			step.Config = config
		}
		stripped.Steps[name] = step
	}
	return stripped
}
