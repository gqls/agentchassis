# NOTES — bugfix_153_build_provenance (append-only, newest at bottom)

## 2026-08-10 — picking this bug up: ownership + validity check

Asked to find the next `bugs_open/` case not already owned by another thread, research it,
plan a robust framework-level fix (via a Fable-model planning pass), check the fleet's own
guidelines and the council, and see it through to a committed, buildable fix.

**Triage across all ~79 open bugs.** Ran `scripts/who-owns.py` against every number in
`bugs_open/`. Nearly all come back "OWNED or recently active" — this fleet is extremely busy,
and the script's heuristic fires on any commit that merely cites a bug number in its subject,
including the filing commit itself, so the raw verdict over-reports ownership. Cross-checked
against the actual `docs024_key_docs_latest/bugfix_*` workstream directory listing (grep for a
directory whose name embeds the bug's number) to separate "someone filed this and moved on"
from "someone is actively iterating on this". That narrowed the field to bugs with either no
matching `bugfix_NNN_*` directory, or a matching directory that has gone quiet (no commits in
14+ days).

**Considered and rejected `bugs_open/040` (kafka-dial timeouts).** Genuinely quiet — its
workstream's last commit was 2026-07-26/27, 14+ days ago. But it's an open-ended,
still-root-cause-unknown infra investigation awaiting more passive measurement days; not a
scoped fix a single session can responsibly plan and close.

**Considered and mostly-rejected `bugs_open/093` (stat-audit has one guarded call site).**
Root cause fully known, candidate 1 (the actual code fix) already shipped live in `v1.0.1172`
back on 07-27. Checked the DB directly: `site_work_items` now has 30 live `claims_unverified`
items (up from the "has never fired" state the file itself recorded as of 08-07), including a
same-day (08-10 08:44Z) revalidation of the exact vonc.com contradiction the file predicted —
so the detection gap is closed in substance, just not yet formally verified/documented/closed
out. What's left is process (council round 6 — confirmed via `diagnosis_artifacts`/`doc_notes`
that no round 6 exists yet) plus an optional, and per the file's *own* analysis NOT
recommended, prevention-path build (candidate 2). Real, unclaimed, small task — but not much
of a "fix" to plan, mostly paperwork. Kept in mind as a fallback, not picked.

**Picked `bugs_open/153`** — "an IMAGE_TAG bump does not imply a rebuild: nothing ties a tag
to the code it was built from". Self-declared `Status: OPEN, unowned` in its own header (the
strongest kind of signal — the filing session said so directly), no `bugfix_153_*` workstream
directory exists, `who-owns.py 153` shows the only touching commits are the filing session's
own (07-29/07-30), nothing since. Root cause is fully diagnosed and the file itself already
ranks fix candidates by robustness — exactly the "prefer a robust, framework-level solution"
brief. Chose it over 093 because there's a real, concrete, still-undone build to plan and do.

**Live-collision check.** `who-owns.py` is blind to uncommitted work in other sessions
(documented fleet-wide practice: grep the live `.jsonl` transcripts). There are ~40+ sessions
active in this repo right now (checked `~/.claude/projects/-home-ant-projects-agentchassis/
*.jsonl` mtimes — dozens modified in the last few minutes). Grepped recent transcripts for the
bug's filename and for `check_unverified_claims_stats`/`scanStoredStatClaims` (093's markers)
and for 153's own filename. All matches traced back to generic `ls bugs_open/` directory
listings (every session's own first-look step, mine included) or to my own session's prior
turns — no other session shows any sign of being inside `153`'s actual content or the
build/dockerfile/makefile files it touches. No `orchestration_states` row and no
`diagnosis_artifacts` row exists yet for anything resembling this fix.

**Re-verified the defect is still live today, not stale.** Both `agent-chassis` pods
(`agent-chassis-8496665bb8-{f6svp,sskxd}`, image `v1.0.1279` — 76 versions past the `v1.0.1202`
the bug was originally filed against) still show `strings /app/agent-chassis | grep -cE
'v1\.0\.1[0-9]{3}'` → 0 and the same for a 40-hex sha → 0. `grep -c ldflags
build/docker/backend/agent-chassis.dockerfile makefile` → 0/0. Nobody has fixed this in the 12
days since it was filed. `[MEASURED]` 2026-08-10.

**Plan drafted via a `claude-fable-5` Plan-type agent** (per the brief's "using fable, prepare
a plan" instruction), given the full research above plus my own direct reads of the makefile
(`ref_build`/`tree_build` defines, `push-backend`, `verify-agent-images`), all 14 backend
Dockerfiles' `go build` lines, `go.mod`'s module path, and `cmd/agent-chassis/main.go`'s
logger idiom. See `PLAN_2026-08-10_build_provenance.md`. Scope: candidate 1 (stamp + verify,
via the makefile's two `define`s so it's generic across all 14 services, not a one-off) plus
candidate 4 (widen `verify-agent-images`); candidates 2 (tag-implies-build) and 3
(push-gating) deliberately deferred — both change the push/deploy contract fleet-wide and
warrant explicit owner sign-off rather than one session's unilateral call.
