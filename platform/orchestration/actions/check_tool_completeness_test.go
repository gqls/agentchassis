package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// First test coverage for this action (council 70cf0da5 round 2, editquality
// advisory: the bugs_open/303 fix touched it with no coverage stated). The
// balance core is content.StructuralTagCounts, tested in platform/content;
// what is pinned HERE is the action's own behaviour: markup-context counts
// feed the issues list, the check is advisory (flags, never fails), and the
// marker is stripped from the passed-through HTML.

func checkToolCompletenessParams(html string) ActionParams {
	return ActionParams{
		Context:          context.Background(),
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"tool_recreation": map[string]interface{}{"result": html},
		},
		StepConfig: models.Step{Config: map[string]interface{}{}},
	}
}

func TestCheckToolCompleteness(t *testing.T) {
	pad := strings.Repeat("<!-- pad -->", 50) // past the 500-char substance floor

	base := `<section><div id="out"></div></section>` +
		`<script>var protect=/<script[^>]*>/g; /* mentions <style> too */ run();</script>` +
		pad + toolCompleteMarker

	t.Run("mentions in JS are not unbalanced tags (bugs_open/303)", func(t *testing.T) {
		res, err := CheckToolCompletenessAction(context.Background(), checkToolCompletenessParams(base))
		if err != nil {
			t.Fatalf("action errored: %v", err)
		}
		m := res.(map[string]interface{})
		if !m["complete"].(bool) {
			t.Fatalf("complete=false, issues=%v — the substring counter's false positive is back", m["issues"])
		}
		if strings.Contains(m["clean_html"].(string), toolCompleteMarker) {
			t.Fatalf("completion marker not stripped from clean_html")
		}
	})

	t.Run("cut mid-script is flagged but ADVISORY — no error", func(t *testing.T) {
		cut := `<section><div></div></section><script>var t=['a','b` + pad + toolCompleteMarker
		res, err := CheckToolCompletenessAction(context.Background(), checkToolCompletenessParams(cut))
		if err != nil {
			t.Fatalf("advisory check must not fail the step: %v", err)
		}
		m := res.(map[string]interface{})
		if m["complete"].(bool) {
			t.Fatalf("a cut script read as complete")
		}
		issues := m["issues"].([]string)
		joined := strings.Join(issues, " | ")
		if !strings.Contains(joined, "script") {
			t.Fatalf("issues %v do not name the unbalanced script", issues)
		}
	})

	t.Run("case-folded: <SCRIPT> cut is caught (was case-sensitive pre-303)", func(t *testing.T) {
		cut := `<SECTION><DIV></DIV></SECTION><SCRIPT>var t=1;` + pad + toolCompleteMarker
		res, err := CheckToolCompletenessAction(context.Background(), checkToolCompletenessParams(cut))
		if err != nil {
			t.Fatalf("action errored: %v", err)
		}
		if res.(map[string]interface{})["complete"].(bool) {
			t.Fatalf("uppercase cut script slipped through — the pre-303 case-sensitivity is back")
		}
	})
}
