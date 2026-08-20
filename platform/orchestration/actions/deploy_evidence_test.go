package actions

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// bugs_open/315 / RFC_038. These pin the two things a hand-rolled reader gets
// wrong, both of them MEASURED against the live fleet rather than imagined:
//
//   - the field NAME varies (nine distinct output_field values across the 19
//     live git_commit steps; deploy_result names only three, section-editor
//     uses git_result), so the name must come from config;
//   - the field SHAPE varies (57 of 744 orchestrations over 7 days nest one
//     level deeper, via a called sub-agent), so a literal path index reports
//     "no evidence" on 7.7% of real runs — and, because the caller must fail
//     open, would wave exactly those through in silence.

// directReply is the shape produced when git_commit runs inline in the workflow.
func directReply() map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"response": map[string]interface{}{
			"data": map[string]interface{}{
				"success":    true,
				"repo_url":   "https://github.com/gqls/sites",
				"commit_sha": "abc123def456",
				"files_sha256": map[string]interface{}{
					"tools/css-variables/index.html": "deadbeef",
					"tools/assets/header.js":         "cafebabe",
				},
			},
		},
	}
}

func TestResolveDeployEvidence_DirectShape(t *testing.T) {
	ev, ok := resolveDeployEvidence(map[string]interface{}{"deploy_result": directReply()}, "deploy_result", zap.NewNop())
	if !ok {
		t.Fatal("evidence did not resolve from the ordinary inline shape")
	}
	if ev.CommitSHA != "abc123def456" {
		t.Errorf("CommitSHA = %q, want abc123def456", ev.CommitSHA)
	}
	if ev.FilesSHA256["tools/css-variables/index.html"] != "deadbeef" {
		t.Errorf("fingerprint not resolved: %v", ev.FilesSHA256)
	}
}

// THE REGRESSION PIN. This is the 7.7% shape: the deploy was performed by a
// called sub-agent, so the whole child collected-data comes back under
// `response`, and the reply sits one level deeper. A reader indexing
// `deploy_result.response.data.commit_sha` finds nothing here.
func TestResolveDeployEvidence_NestedSubAgentShape(t *testing.T) {
	nested := map[string]interface{}{
		"deploy_result": map[string]interface{}{
			"response": map[string]interface{}{
				"deploy_result":   directReply(),
				"rendered_page":   map[string]interface{}{"page_id": "irrelevant"},
				"response_status": "complete",
			},
		},
	}
	ev, ok := resolveDeployEvidence(nested, "deploy_result", zap.NewNop())
	if !ok {
		t.Fatal("evidence did not resolve from the NESTED call_agent shape — this is 7.7% of live runs, and failing here means failing open on all of them")
	}
	if ev.CommitSHA != "abc123def456" {
		t.Errorf("CommitSHA = %q, want abc123def456", ev.CommitSHA)
	}
}

// The field name is config-supplied, so a differently-named field must work
// identically. section-editor really does call it git_result.
func TestResolveDeployEvidence_HonoursTheConfiguredFieldName(t *testing.T) {
	collected := map[string]interface{}{"git_result": directReply()}

	if _, ok := resolveDeployEvidence(collected, "git_result", zap.NewNop()); !ok {
		t.Error("evidence did not resolve under the configured field name git_result")
	}
	// ...and must NOT be found under a name nothing wrote, or the guard would be
	// reading some other step's output.
	if _, ok := resolveDeployEvidence(collected, "deploy_result", zap.NewNop()); ok {
		t.Error("evidence resolved under a field name that does not exist — the guard is not reading the field it was told to")
	}
}

func TestResolveDeployEvidence_SkipIsDistinctFromUnreadable(t *testing.T) {
	// GitCommitAction's live skip path: Success:true with a skip reason and no
	// commit at all. It must read as SKIPPED (refuse the stamp), never as
	// unreadable (stamp anyway) — they are opposite responses.
	skipped := map[string]interface{}{"deploy_result": map[string]interface{}{
		"success": true,
		"metadata": map[string]interface{}{
			"status":      "skipped",
			"skip_reason": "no files to commit",
		},
	}}
	ev, ok := resolveDeployEvidence(skipped, "deploy_result", zap.NewNop())
	if !ok {
		t.Fatal("a skip must resolve as evidence — reading it as unreadable would stamp the page, which is the bug")
	}
	if !ev.Skipped || ev.SkipReason != "no files to commit" {
		t.Errorf("skip not reported: %+v", ev)
	}
}

func TestResolveDeployEvidence_OldAdapterIsUNREADABLENotSuccess(t *testing.T) {
	// A git-adapter image predating RFC_038 replies without commit_sha or
	// files_sha256. The chassis and the adapter are separate images, so this is
	// the NORMAL state during a partial roll. It must fail open (ok=false) so
	// the caller logs and stamps — not be mistaken for a successful, verified
	// deploy, and not be mistaken for a skip either.
	old := map[string]interface{}{"deploy_result": map[string]interface{}{
		"success": true,
		"response": map[string]interface{}{"data": map[string]interface{}{
			"success":  true,
			"repo_url": "https://github.com/gqls/sites",
		}},
	}}
	if ev, ok := resolveDeployEvidence(old, "deploy_result", zap.NewNop()); ok {
		t.Errorf("a pre-RFC_038 reply resolved as usable evidence (%+v) — it carries no commit identity and must read as unreadable", ev)
	}
}

func TestResolveDeployEvidence_AbsentOrJunkField(t *testing.T) {
	for name, collected := range map[string]map[string]interface{}{
		"missing field": {},
		"not a map":     {"deploy_result": "a string"},
		"nil tree":      nil,
	} {
		if _, ok := resolveDeployEvidence(collected, "deploy_result", zap.NewNop()); ok {
			t.Errorf("%s: resolved evidence that is not there", name)
		}
	}
	if _, ok := resolveDeployEvidence(map[string]interface{}{"deploy_result": directReply()}, "  ", zap.NewNop()); ok {
		t.Error("a blank field name resolved something — an unset key must mean the guard is OFF")
	}
}

func TestHashForPageFile_KeyIsTheCommittedPath(t *testing.T) {
	files := map[string]string{"tools/css-variables/index.html": "deadbeef"}

	if got := hashForPageFile(files, "/tools/css-variables/index.html"); got != "deadbeef" {
		t.Errorf("page url did not map to its committed file path: got %q", got)
	}
	// A FRAGMENT url must yield no fingerprint. This is not hypothetical: idea.uk
	// carries a live page row at "/tools.html#audience-check" while a DIFFERENT
	// page owns "/tools.html", so stripping the fragment would fingerprint one
	// page with the other's bytes.
	if got := hashForPageFile(map[string]string{"tools.html": "beef"}, "/tools.html#audience-check"); got != "" {
		t.Errorf("a fragment url produced fingerprint %q — it must produce none", got)
	}
	if got := hashForPageFile(files, "/tools/not-committed.html"); got != "" {
		t.Errorf("a page absent from the commit produced fingerprint %q", got)
	}
	if got := hashForPageFile(nil, "/tools/css-variables/index.html"); got != "" {
		t.Errorf("an empty fingerprint map produced %q", got)
	}
}

// THE OBJECTION THIS TEST ANSWERS (council round 2, 377167cd — gating objection
// from prior_art_librarian, seconded by editquality).
//
// The earlier version of deploy_evidence.go justified its safety by borrowing a
// property from datahelpers.ExtractFields: "collect-all / unique-or-nothing
// (RFC_029 §9)", i.e. an ambiguous subtree resolves to nothing rather than to a
// guess. THAT CLAIM WAS FALSE FOR THIS BUILD. findFieldRecursive's own comment
// says the ruling is unique-or-nothing but "PHASE 1 (this build — instrument
// first, refuse second): conflicts still resolve, to the STABLE shallowest-first
// winner, and emit the WARN". Phase 2 has not shipped.
//
// For this caller a guess is the worst outcome available: a fingerprint taken
// from the WRONG git_commit is silently and permanently wrong, and every later
// comparison reports a healthy page as diverged. So the resolver now collects
// candidates itself and REFUSES on conflict, and this pins that.

func TestResolveDeployEvidence_AmbiguousSubtreeREFUSES(t *testing.T) {
	// Two git_commit results under ONE named field — the shape that arises when
	// a called sub-agent performed more than one commit.
	ambiguous := map[string]interface{}{
		"deploy_result": map[string]interface{}{
			"response": map[string]interface{}{
				"page_commit": map[string]interface{}{"data": map[string]interface{}{
					"commit_sha":   "aaaaaaaaaaaa",
					"files_sha256": map[string]interface{}{"tools/x/index.html": "hash-A"},
				}},
				"asset_commit": map[string]interface{}{"data": map[string]interface{}{
					"commit_sha":   "bbbbbbbbbbbb",
					"files_sha256": map[string]interface{}{"tools/x/index.html": "hash-B"},
				}},
			},
		},
	}

	ev, ok := resolveDeployEvidence(ambiguous, "deploy_result", zap.NewNop())
	if ok {
		t.Fatalf("ambiguous subtree RESOLVED to %+v — it must refuse. A fingerprint from the wrong commit is silently and permanently wrong, and 'no fingerprint' is the recoverable direction", ev)
	}
}

func TestResolveDeployEvidence_AgreeingDuplicatesStillResolve(t *testing.T) {
	// The other half of unique-or-nothing: the SAME value appearing twice (the
	// envelope legitimately echoes a reply at two depths) is not a conflict, and
	// refusing it would make the guard useless on ordinary nested runs.
	agreeing := map[string]interface{}{
		"deploy_result": map[string]interface{}{
			"deploy_result": directReply(),
			"response":      map[string]interface{}{"deploy_result": directReply()},
		},
	}
	ev, ok := resolveDeployEvidence(agreeing, "deploy_result", zap.NewNop())
	if !ok {
		t.Fatal("duplicate but AGREEING candidates were refused — only a genuine disagreement may refuse")
	}
	if ev.CommitSHA != "abc123def456" {
		t.Errorf("CommitSHA = %q, want abc123def456", ev.CommitSHA)
	}
}

func TestCollectUniqueValue_ConflictBeatsFound(t *testing.T) {
	// A caller that read `found` and ignored `conflict` would get the shallowest
	// value — exactly the guess this design exists to avoid. Pin the precedence.
	tree := map[string]interface{}{
		"a": map[string]interface{}{"commit_sha": "one"},
		"b": map[string]interface{}{"commit_sha": "two"},
	}
	_, found, conflict := collectUniqueValue(tree, "commit_sha", 0)
	if !conflict {
		t.Fatal("disagreeing values did not report a conflict")
	}
	if !found {
		t.Error("found should still be true — the caller distinguishes 'absent' from 'ambiguous'")
	}
}

// Council round 3, editquality advisory: the stamp statement is built with a
// conditionally-included clause AND a conditionally-appended arg. Those are two
// facts that must agree and nothing forces them to — a literal "$3" beside an
// append is a runtime error waiting on whichever branch the tests miss.
//
// This calls the REAL construction (buildPageDeployStampQuery), which is why it
// was extracted: the first version of this test mirrored the code, and a mirror
// passes happily while production is broken.
func TestStampStatementPlaceholdersMatchArgs(t *testing.T) {
	re := regexp.MustCompile(`\$(\d+)`)
	for _, guardRan := range []bool{false, true} {
		q, args := buildPageDeployStampQuery([]interface{}{"page-id", "deployed"}, guardRan, "abc123")

		highest := 0
		for _, m := range re.FindAllStringSubmatch(q, -1) {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				t.Fatalf("unparsable placeholder %q", m[1])
			}
			if n > highest {
				highest = n
			}
		}
		if highest != len(args) {
			t.Errorf("guardRan=%v: statement references up to $%d but %d args are supplied — psql would reject this at runtime on the deploy path",
				guardRan, highest, len(args))
		}
		if guardRan && !strings.Contains(q, "content_hash") {
			t.Error("guardRan=true but the statement does not write content_hash")
		}
		if !guardRan && strings.Contains(q, "content_hash") {
			t.Error("guardRan=false but the statement touches content_hash — an unarmed path must leave the column entirely alone")
		}
	}
}
