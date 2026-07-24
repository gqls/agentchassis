# BUG 065 — formatGeneratedGo rejects the commit-entry shape its only caller sends

**Filed:** 2026-07-24 · gauntlet_dead_cta / feature-builder B4 first-fire · **CLOSED same day (fixed AND live, behaviourally proven)**
**Severity:** high — the feature-implementer could never ship a single `.go` file; every Go stage died at commit-prep.

## Symptom
Implementer runs refused at `stage_prepare`: `generated <file>.go has a non-string body
(map[string]interface {})` — rounds 2+3 (orchs `2284a8f4`, `70731845`), different file
each time. The stored `implementation` payload had every file body as a jsonb **string**,
so the model's output was valid.

## Root cause
`validateImplementation` (`diagnose_prepare_fix_commit_action.go:316`) wraps every file
body in the GitCommitData entry shape — `{"content": <string>, "encoding": "utf-8"}` —
while `formatGeneratedGo` (same file) type-asserted each body as a bare **string** and
errored otherwise. The two halves were added by different changes (bugs_open/013's
formatter; the GitCommitData wrapping), each unit-tested with its OWN shape, and never
run together until B4's first fire: the unit tests build `files{path: <bare string>}` — a
shape production never sends. A `.sql`-only stage passes because the formatter skips
non-`.go` paths — exactly why round 1's migration stage sailed through and every Go
stage died. Dedup-index/Go-list contract-drift class, one function down.

## Fix (shipped)
`formatGeneratedGo` accepts both shapes (wrapped entry formats `content` in place,
preserves `encoding`; bare string unchanged; fail-loud truncation kept on both paths) +
a regression test running the REAL chain `validateImplementation → formatGeneratedGo`
and a wrapped-truncation loud-failure test. Commit `430ed5c18`; **council gate APPROVED**
corr `6bf3806f` (first run died on an Anthropic endpoint timeout — filed invalid, no
judgement; resubmit approved 2026-07-24 12:02; commit predates verdict so the corr is in
the message, no trailer — forward-only).

## Live proof
Chassis v1.0.1155 rolled; **behaviourally proven in round 7** (orch `863668c1`): the
implementer committed its first-ever Go stage (branch sha `790988cf`) and proceeded to
the build gate — past the exact step that had refused every prior round.

## Trap discovered while shipping it (see bugs_open/066)
"Pod-verified live" on the chassis **deployment** pod was a false green for this fix:
the implementer runs in a **spawned dedicated pod** whose image comes from
`agent_definitions.image_tag`, which pinned `v1.0.1151`. Round 6 failed on the old
binary AFTER the deployment pod-grep passed. Verify the runtime that will EXECUTE the
code path.
