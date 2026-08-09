# Register — operator-practice

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

_Concept count retired 2026-08-09 — derived, not stored; run the drift pair in `000_concept_index.md`, or read `concept-register-drift-check`'s daily row (DOC-074)._ (2 consolidated from 4 raw extractions across unit U25 — each of the 2 distinct blocks
appeared byte-identically twice within the cluster input file, treated as duplicate copies of
one extraction, not independent corroboration; + OPP-003, OPP-004 and OPP-005 added later, see
their entries for dates).

### OPP-001 — In-chassis replicability requirement for operator work
- **status:** deployed
- **status-evidence:** REPLICATION_in_chassis.md (2026-07-10) maps every off-platform action to [chassis]/[human]/[gap]; RUNNING_NOTES turn 8: "everything done in this thread must be replicable inside the chassis … Either document the tool-free path or don't use the tool."
- **what:** Standing owner rule: interactive-agent work on sites must be reproducible by the chassis itself (or documented as a human-judgement or platform-gap item). The audit, spec rewrites, artifact verification and imagery deploys all map to normal platform operations; the genuinely human items (choosing a permanent logo, stating personal-history claims, setting engagement terms) are deliberately not automated; named gaps: pinned enforcement, favicon/OG derivation, chart capability.
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md (whole); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-8
- **relations:** hitl; documentation-system; site_specs pinned gap
- **verify-later:** n/a (practice doc); checkpoint_for_review as the in-chassis review surface

### OPP-002 — Operator discipline: verify-by-artifact, dated backups, kcat generic trigger
- **status:** deployed
- **status-evidence:** PLAN standing rule 2: "Verify by artifact, never by report. (This platform has a long history of builds reporting success while building nothing.)"; RUNBOOK landmines 2/17/18; VERDICT §7 "Verified by artifact, never by item status" with md5 predictions.
- **what:** The cross-workstream operating discipline: (1) never trust a complete work item — curl the page, read the DB row, diff the bytes (strongest form: predict output bytes offline and md5-compare); (2) back up before ANY change using dated bak_*/_backup_YYYYMMDD tables and never reuse a name (CREATE TABLE IF NOT EXISTS silently no-ops); (3) trigger any agent by producing to Kafka system.agent.generic.requests via kcat with the standard header set; (4) kubectl exec heredocs need -i or silently run nothing — prefer kubectl cp + psql -f; quote heredoc delimiters.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Landmines; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#7; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/leopardessconsulting/HANDOFF.md#5
- **relations:** silent no-op success class; fleet generalisation doctrine; standing evidence rules (operating-doctrine register, OPD-001)
- **verify-later:** n/a (doctrine); bak_* table inventory in clients_db

### OPP-003 — `check_logged_model_output`: a pre-commit detector for publishing model text
- **status:** deployed (advisory; 8th check in `scripts/pattern-check.py`, run by `.githooks/pre-commit`)
- **what:** Flags a log sink (`log.*`, `logger.*`, `fmt.Print*`) that passes an
  unwrapped payload identifier (`text`, `body`, `completion`, `response`, `raw`,
  …) inside any package that calls `GenerateText`, and points at
  `aiservice.Fingerprint` as the fix. Silent when the value is wrapped in a
  derived-fact helper. Exists because whether to log model output is a *content*
  decision, not a code-review one — it took a full council round to surface once,
  and the reviewer stated it could not be closed by the council alone.
- **sources:** `scripts/pattern-check.py` (`check_logged_model_output`)
- **relations:** LCO-005; `bugs_open/083`
- **note — the check nearly shipped VACUOUS, and this is the transferable part:**
  the first version gated on **the file** containing `GenerateText`. The LLM call
  lives in `defend.go` and the log sink in its sibling `ailog.go`, so it examined
  **zero files and printed a clean result**. A clean result and an unrun check are
  byte-identical output. Caught only by a **positive control** — auditing the
  commit that introduced the defect and *requiring* a finding. Now gated on the
  package and verified three ways: fires on `a37a2037c`, silent on the fix, zero
  findings fleet-wide. **Any new detector needs its failing case demonstrated at
  the moment it is written.**

### OPP-004 — `check_register_coverage`: the register coverage cadence, on the commit path
- **status:** deployed (advisory; 9th check in `scripts/pattern-check.py`, run by `.githooks/pre-commit`)
- **what:** Fires when a commit **creates a workstream directory the concept register has never heard of**, naming both ways to go quiet — add a register entry, or add the directory to `102_coverage_ratchet.txt`. It closes `bugs_open/106`'s last open scope: the coverage SENSOR (`102_CHECK_register_coverage.py`, sensor + ratchet) already existed and worked, but ran only when a human remembered, which is the same detected-by-coincidence mechanism the bug was filed about, one step earlier. **A commit trigger was chosen over a cron deliberately**: a cron reports drift up to a week late to nobody in particular, whereas this reports it the instant it is created to the one person who can close it in ten seconds. Only NEW directories fire — the 43 already-uncovered workstreams are accepted backlog on the ratchet, and flagging active work on them every commit is how a check becomes wallpaper.
- **why it is registerable:** it is the enforcement half of the register's own coverage guarantee. Any workstream creating a new subsystem directory now learns, at creation, that the estate's capability index does not know about it.
- **measured before inclusion** (this file's own bar): **4 fires / 1,500 commits = 0.27%**, quieter than every existing check (README 0.7%, SUMMARY 2.0%, twin ~2%) — correct for a population of "commits creating a brand-new workstream". **Precision inspected because a very low rate and a dead check look identical:** all 4 genuine (`memory_index`, `bugs_sweep_2026_07`, `bugfix_066_spawn_image_tag`, `gemini_content_provider`), and the last two are exactly the pair 106's own triage records the sensor finding by hand on 2026-07-27 — the same gaps, now caught at creation.
- **verified by induced gap**, as `106` demands: two scratch workstreams staged → both fire; one added to the ratchet → only the other fires; the other given a register entry → it goes quiet too. Negative control: silent and 40 ms on a commit touching no workstream directory.
- **sources:** `scripts/pattern-check.py` (`check_register_coverage`); `docs/agent_docs/docs026_concept_register/102_CHECK_register_coverage.py`; `docs024/bugfix_106_register_coverage_cadence/` (PLAN, RUNBOOK, NOTES)
- **relations:** OPP-003 (sibling commit-path detector); `102_coverage_ratchet.txt`; the `verifier_coverage_test.go` sensor+ratchet shape it is modelled on; `bugs_closed/106`
- **landmine:** it **imports** the sensor rather than reimplementing `is_covered()` — one matching rule, one implementation, because two hand-maintained copies is the `idx_swi_dedup` ↔ `workItemTerminalStatuses` drift class. Change how coverage is decided in the sensor and the hook follows. The import is guarded: if the sensor moves, the check returns silently rather than breaking commits.
- **NOT in scope, deliberately:** it asks only whether a subsystem is *represented*, never whether an entry is *accurate*. **The register can be complete in coverage and stale in content, and nothing detects that** — two live instances hit on 2026-07-28 (a `verify-later` stating an expected answer that had been false for weeks, contributing to `bugs_closed/124` going unnoticed). Recorded, not fixed: a coverage check that starts auditing accuracy becomes slow, noisy and ignored.
- **verify-later:** does `check_register_coverage` still appear in `scripts/pattern-check.py`'s check tuple in `main()`, and does `./scripts/pattern-check.py` stay silent on a commit touching no workstream directory?

### OPP-005 — Session scratchpads: versioned by a snapshot hook, reaped by content class
- **status:** deployed (2026-08-03; `CLAUDE_CODE_TMPDIR` moved to disk in `~/.claude/settings.json`, hook wired in `.claude/settings.json`, both tools in `scripts/`)
- **what:** Three coupled pieces that make the shared session-scratchpad tree survivable. (1) **`CLAUDE_CODE_TMPDIR=/home/ant/.claude-scratch`** — scratchpads move off the 16 GB `/tmp` tmpfs onto the main disk (212 GB free), ending the class where a full tmpfs makes a Bash call return **no output** (ENOSPC on the harness's own stdout capture, which reads like "the command found nothing", not like a disk error). (2) **`scripts/scratch-git-snapshot.py`** — a `PostToolUse` hook on `Write|Edit` that commits each written scratchpad file into a git repo at the scratch root, one file by pathspec, so a scratchpad stops being unrecoverable. (3) **`scripts/scratch-report.py`** — shows every session dir by age, size and content class, and reaps only what is provably reproducible.
- **why it is registerable:** any workstream on this machine writes to a scratchpad, and until now nothing versioned or bounded that tree. It is also the general answer to "how do we clean up shared temp safely", which every lane eventually asks.
- **the measurement the whole design rests on** — the two content classes have opposite risk profiles and the space is almost entirely in the safe one: **33 repo-extraction dirs = 12.13 GB = 99.3%**, versus **370 other entries = 88 MB = 0.7%**. A repo extraction is a `git archive HEAD` unpack (the shared-tree build check CLAUDE.md mandates); it is byte-for-byte reproducible from a sha, so deleting one loses nothing. The 0.7% is irreplaceable and small enough never to need deleting.
- **the separation is structural, not heuristic.** Extractions are created by `tar`/`git archive` in **Bash**; the hook fires on **Write|Edit**. So it is *incapable* of committing an extraction — no size rule is doing the work, and there is no threshold to misfire. The 5 MB cap in the hook is a backstop against a large generated artefact, not the class filter.
- **`.gitignore` is `*` and every add is `-f`.** The repo therefore contains only what was deliberately snapshotted, and `git status` next to 12 GB of untracked bulk returns **empty in 3 ms** instead of listing it. Showing untracked bulk is the report tool's job, not git's.
- **verified in both directions before wiring** (a guard that refuses everything passes a refusal test): positive — a real `Write` in a live session produced a commit, proving the hook fires without a restart; negative — a repo path outside any scratchpad and a 6 MB file were each refused, commit count unchanged; idempotence — re-firing on unchanged content adds no second commit.
- **`--reap` deletes only marker-verified extraction directories**, never a loose file, never a session dir, never anything it cannot positively identify (≥2 of `go.mod`/`platform`/`CLAUDE.md` at the directory top). It re-verifies at the moment of deletion, not just at scan time, and defaults to a dry run at a 2-day threshold.
- **sources:** `scripts/scratch-git-snapshot.py`; `scripts/scratch-report.py`; `~/.claude/settings.json` (`env.CLAUDE_CODE_TMPDIR`); `.claude/settings.json` (`PostToolUse`); `gauntlet_dead_cta/NOTES` 2026-08-03
- **relations:** `scripts/memory-git-snapshot.py` (same argument, same shape — an unversioned shared store is the hazard, and versioning it dissolves the class rather than policing writes); the `bugfix_177` LANDMINE on `/tmp` filling mid-build, which is the same outage seen from inside a build
- **landmine:** **two scratch roots are live at once after the move.** A running session keeps the tmpdir it launched with, so sessions started before 2026-08-03 keep writing to `/tmp` while new ones use the disk. `df -h /tmp` can therefore look healthy while the tree you are actually in is full, or the reverse. Both tools read **both** roots for this reason; a check that inspects only one will be confidently wrong for as long as the old sessions live.
- **NOT in scope, deliberately:** it does not reap hand-written files at all, at any age. They are 0.7% of the volume, so reaping them buys nothing measurable and risks the only irreplaceable thing in the tree.
- **verify-later:** does `git -C /home/ant/.claude-scratch/claude-1000 log` show commits from more than one session id (proving the hook is live fleet-wide, not just in the session that wired it), and does `scripts/scratch-report.py` still report both roots?
