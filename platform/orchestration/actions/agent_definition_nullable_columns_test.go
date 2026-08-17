// FILE: platform/orchestration/actions/agent_definition_nullable_columns_test.go
//
// bugs_closed/290 — an agent seeded WITHOUT a description is unspawnable, and the
// failure looks like a spawn handshake flake.
//
// agent_definitions.description is nullable with no default, and five readers
// (three agent-definition loaders in this package, the messaging processor's
// resolver, the admin list) scanned it into a plain Go string. brief-fidelity-
// auditor's seed omitted the column, so the FIRST spawn after it was wired into
// the improvement loop (mig 419, 2026-08-16) failed with
//   sql: Scan error on column index 3, name "description": converting NULL to string
// — before the auditor ran, and behind an error_step that would have carried the
// sweep on as COMPLETED every time. The other three NULL rows fleet-wide were
// inactive scratch probes, which is why nothing had ever hit this: the column is
// almost always set, so a seed that forgets it is the whole population.
//
// The fix is the idiom three other loaders in this package already use for
// nullable text (fork_theme_composition.go, load_component_library_actions.go,
// render_js_snippets_for_site_action.go): COALESCE at the query. This test makes
// the class unrepresentable in SOURCE — any SELECT from agent_definitions that
// names description bare fails here — so the sixth reader cannot re-open it.
//
// The scan is deliberately over the whole module (platform/, internal/, pkg/,
// cmd/), because two of the five readers were outside this package; a scan
// scoped to `actions` would have passed while the processor still broke.
// Comment lines are dropped before matching (comment text is not a query).

package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// bareDescriptionSelect matches a SELECT list that names `description` as a
// bare column (leading comma/space, trailing comma/space) inside a query that
// reads agent_definitions. Two-stage: find candidate SELECT statements whose
// FROM names agent_definitions, then check the column list.
var (
	agentDefSelect  = regexp.MustCompile(`(?is)SELECT\s+(.*?)\s+FROM\s+agent_definitions\b`)
	bareDescription = regexp.MustCompile(`(?i)(^|[\s,])description([\s,]|$)`)
)

func TestAgentDefinitionDescriptionIsNeverScannedBare(t *testing.T) {
	root := moduleRoot(t)
	var hits []string
	for _, top := range []string{"platform", "internal", "pkg", "cmd"} {
		_ = filepath.WalkDir(filepath.Join(root, top), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			src, rerr := os.ReadFile(path)
			if rerr != nil {
				return nil
			}
			body := stripLineComments(string(src))
			for _, m := range agentDefSelect.FindAllStringSubmatchIndex(body, -1) {
				cols := body[m[2]:m[3]]
				// An INSERT ... SELECT that COPIES description (discovery_actions.go's
				// variant clone: `description || ' [Variant: ...]'`) is not a scan into
				// a Go string; NULL || text is NULL and lands back in the same
				// nullable column. Only a SELECT that is not preceded by INSERT INTO
				// on the same statement is a reader.
				pre := body[max(0, m[0]-200):m[0]]
				if strings.Contains(strings.ToUpper(pre[strings.LastIndex(pre, ";")+1:]), "INSERT INTO") {
					continue
				}
				if bareDescription.MatchString(cols) && !strings.Contains(strings.ToLower(cols), "coalesce(description") {
					rel, _ := filepath.Rel(root, path)
					hits = append(hits, rel+": "+strings.Join(strings.Fields(cols), " "))
				}
			}
			return nil
		})
	}
	for _, h := range hits {
		t.Errorf("agent_definitions.description scanned BARE — a NULL (a seed that forgot the column) "+
			"fails the Scan and the agent cannot be spawned/resolved (bugs_closed/290). "+
			"Use COALESCE(description, '') as the five existing readers do:\n\t%s", h)
	}
}

// TestBareDescriptionPatternStillBites guards the guard: the regexes must still
// match the exact pre-fix query shape and must not match the fixed one.
func TestBareDescriptionPatternStillBites(t *testing.T) {
	bad := `SELECT id, type, display_name, description, category FROM agent_definitions WHERE type = $1`
	good := `SELECT id, type, display_name, COALESCE(description, ''), category FROM agent_definitions WHERE type = $1`
	other := `SELECT name, description FROM sites WHERE id = $1`

	m := agentDefSelect.FindStringSubmatch(bad)
	if m == nil || !bareDescription.MatchString(m[1]) {
		t.Fatalf("pattern no longer matches the pre-fix shape — the guard is disarmed")
	}
	m = agentDefSelect.FindStringSubmatch(good)
	if m == nil || (bareDescription.MatchString(m[1]) && !strings.Contains(strings.ToLower(m[1]), "coalesce(description")) {
		t.Fatalf("pattern false-positives on the fixed shape")
	}
	if agentDefSelect.MatchString(other) {
		t.Fatalf("pattern matches a query on a different table")
	}
}

func stripLineComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("go.mod not found above test dir")
	return ""
}
