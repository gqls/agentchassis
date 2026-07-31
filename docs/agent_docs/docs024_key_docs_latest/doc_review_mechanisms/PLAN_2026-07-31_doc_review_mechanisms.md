# PLAN — doc review mechanisms (LANDMINES.md verifier + bugs_open staleness sweep)

## Progress log

**2026-07-31 — Part A, first pass built and test-driven.**

- `landmine-verifier` agent definition applied
  (`docs/agent_docs/sql_for_agents/276_landmine_verifier.sql`): load_entry
  (query_database) -> derive_checks (execute_llm_prompt) -> run_checks
  (diagnose_code_lookup, reused from the diagnose loop) -> verify
  (execute_llm_prompt) -> persist_verdict (append_doc_note). No new Go code
  for the agent itself.
- End-to-end test run (`LANDMINES.md#deployimageasset-resolves-its-source-
  image-by-purpose-not-by-the-assetid-you-pas`) got through all four steps
  and failed only at the last one, `persist_verdict`.
- Root cause was NOT this agent's design — it was a real, pre-existing
  platform bug: `validDocSubjectTypes`
  (`platform/orchestration/actions/doc_subjects_common.go`) never got
  `'landmine'` added when migration 270 widened `doc_notes_subject_type_check`
  two days ago. Exactly the split-contract shape `bugs_closed/064` exists to
  prevent, and it slipped past that bug's own regression test because the
  test only ever watched `doc_plans_subject_type_check` (migration 270
  correctly never touches `doc_plans`). Fixed both the list and the test
  (commit `7290433f2`) — the test now checks the union of both tables'
  constraints, closing the blind spot rather than just the immediate case.
  `go build`/`vet`/`test` clean on the package.
- **Submitted to the council gate**: `SUBMISSION_CORR=cb405b7b-c32a-463c-9687-ee3b52acc2fb`,
  queued. **Process note against myself**: committed `7290433f2` before
  submitting, so that commit carries no `Council-Reviewed`/`Council-Submitted`
  trailer — forward-only forbids amending it now. Recording the correlation
  here for traceability since the automatic 098 join won't find it on the
  commit itself. Read the verdict before trusting the fix is architecturally
  sound, even though it's already committed and tested locally.
- **Blocked on an image roll, not yet requested**: the fix is committed to
  the working tree but Go changes are inert until the chassis image is
  rebuilt and deployed — the running pods still carry the old binary, so
  re-running the verifier right now would hit the same failure again. This
  is a fleet-wide action (a new agent-chassis image affects every agent, not
  just this one), so it's paused for an explicit go-ahead rather than done
  unilaterally.
- Not yet done: wiring an automatic trigger (currently a manual kafka
  message, see the SQL file's own header); the in-file-vs-doc_notes verdict
  question from the original plan is now settled in favour of doc_notes
  (built that way already, for the concurrency reasons discussed there).

**2026-07-31, later — end-to-end CONFIRMED live.** Fresh chassis image
deployed (`v1.0.1215`, both pods started ~15:00 UTC — verified via
`kubectl get pods`, not assumed from the roll alone, per this repo's own
"a roll is not evidence your fix shipped" rule). Re-ran the exact same
test (correlation `11ef1686-a28b-4665-8865-0f7e239e640f`) against the same
entry: **COMPLETED**, all 5 steps, no error. Verdict written to `doc_notes`
(id `0f1cc9b3-b722-40ba-a5eb-4b6172156786`):

> **last verified (landmine-verifier): STILL_VALID.** All three code
> components of the bug footprint — `resolveStorageURIFromAsset` reading
> `purpose+"_uri"` from site `content_data`, `StoreAssetAction` doing
> last-write-wins into that same key, and the `updateContentDataField`
> helper — are confirmed present and unchanged at commit d98010e8. Checked
> against 087_towards_multiple_domains.

Correct on all counts: the `deploy_image_asset` bug (`bugs_open/155`) is
genuinely still unfixed in the platform, and the verdict cites the real code,
not a rephrasing of the entry text. **Part A is done for a first pass** —
built, tested against a real entry, verified live on the deployed image, and
the platform bug it surfaced along the way is fixed, tested, and
council-submitted.

Remaining, for a later increment, not blocking: automatic dispatch on
new/changed entries (currently manual per-entry); the shell/kubectl-shaped
`the check` limitation (LLM-judgement fallback only); RFC_005 §3.3 (the
weekly staleness sweep) is unstarted.

**2026-07-31, later still — automatic dispatch wired. Part A complete.**

Extended `scripts/landmines-sync.py`:
- `existing_bodies()` — fetches the current `{source: body}` map via
  `jsonb_object_agg`, NOT line-split `-t -A` text like the existing
  `existing_sources()`. Bodies are genuine multi-line prose (confirmed live:
  every entry's footprint line alone forces a newline), so the naive
  approach would have silently truncated every multi-line body to its first
  line — caught before shipping, not after.
- `changed` detection — a body-hash-equivalent diff against `existing_bodies()`
  for slugs already present, alongside the existing `new`/`gone` lists (which
  only ever saw whether a slug existed, not whether its content had drifted
  under an unchanged title).
- `NEEDS_VERIFICATION:<source>` lines printed after `--apply`, one per new or
  changed entry, machine-parseable by the dispatcher below.
- A slug-collision warning — **found live, not by design**: the very first
  test run flagged a genuine duplicate (`snapshot_agent` entry, identical
  title at two line numbers), which `slugify()`'s title-only hashing
  silently collapses to one dict key, permanently dropping whichever entry
  loses. Wrote the check before realising it wasn't hypothetical; another
  session fixed the actual duplicate concurrently mid-session (`bcecb65e3`)
  while this check was being written. The detector stays as a permanent
  safety net for the next occurrence, not a one-off cleanup.

New scripts: `scripts/trigger-landmine-verifier.sh` (single-entry kafka
trigger, promoted from the scratch version used for earlier manual tests)
and `scripts/landmines-verify-dispatch.sh` (runs `landmines-sync.py --apply`,
parses `NEEDS_VERIFICATION:`, dispatches one trigger per entry, 2s apart to
avoid bursting `kubectl run` pods into the `kafka` namespace).

**Live proof, unplanned and better than a synthetic test could have been:**
mid-session, another thread appended a genuinely new entry to
`LANDMINES.md` (`hasvisiblearea-reports-0-for-any-axis...`). Running
`landmines-verify-dispatch.sh` caught it automatically via `new` detection,
no manual identification, and dispatched a verifier
(correlation `2ca82ffe-bd92-472b-958f-3391f7cbafa6`). Verdict:

> **last verified (landmine-verifier): NEEDS_HUMAN_REVIEW.** Code index is
> stale (d98010e8, 2026-07-28) relative to both the bug-fix commit
> (71680ad513, 2026-07-30) and current HEAD; the file exists but no symbol
> or content match for `VisibleArea` or `has_visible_area` was found, so it
> is impossible to mechanically confirm whether the old trap code or the fix
> is present.

Exactly the calibrated behaviour the design asked for: it did NOT guess past
what the lookup evidence showed, and said so plainly. Side-finding, out of
scope here: the code-lookup index itself has a real freshness gap (three
days stale relative to a landed fix) — worth someone's attention separately,
not something this workstream should chase.

**Part A (RFC_005 §3.2) is now fully done**: built, tested against a real
entry, verified live on the deployed image, wired to dispatch automatically,
and proven twice on genuinely live (not staged) data — once returning
STILL_VALID with real citations, once correctly declining to guess and
returning NEEDS_HUMAN_REVIEW. Only RFC_005 §3.3 (the weekly staleness sweep)
remains unstarted.

---

**Started 2026-07-31.** Implements the owner ruling on
`architecture_review/RFC_005_targeted_review_for_docs_that_feed_the_fleet.md`
(2026-07-31): §3.1 (diagnosis-loop discipline for `bugs_open/`) is already
adopted as a practice norm, recorded in `CLAUDE.md` directly — nothing to
build there. This workstream is the other two rulings:

- **§3.2 — a dedicated single-pass verifier agent for `LANDMINES.md` entries**
  (not a mechanical sync-script grep — the owner's explicit choice, because a
  grep only confirms a footprint exists, not that the entry's claim is still
  true).
- **§3.3 — a weekly staleness sweep over `bugs_open/`** (flags, never
  auto-closes; the fixed-AND-live bar for closing a bug stays a human/thread
  judgment call).

Why this is its own workstream and not inline in the RFC: both are new,
standing, reusable mechanisms — CLAUDE.md's own rule ("when you build a new
reusable mechanism, register it") and the standing-five convention both apply,
and neither should be improvised live in a conversation that was really about
whether to build them at all.

---

## Part A — LANDMINES.md verifier agent (§3.2)

### What it checks, per entry

Each `LANDMINES.md` entry has four load-bearing fields (see the file's own
"Entry format" section): `footprint`, `fires when`, `the tell`, `the check`.
The verifier's job is narrower than a full diagnosis — it is not asking
"is this a real bug", it is asking **"is this entry still an accurate
description of the system"**:

1. Does the **footprint** still resolve (file/table/symbol exists at all)?
2. Does **the check**, run for real, still produce what **the tell** claims?
3. Is the entry internally consistent (does the footprint actually relate to
   the fires-when clause, or has drift made them describe different things)?

### Reuse before build

`the check` field is free text — sometimes a SQL query, sometimes a
`grep`/`kubectl exec` command, sometimes a code citation. Two existing
capabilities cover most of this without new Go code:

- **Footprint resolution** (file/table/symbol exists): the diagnose loop
  already does exactly this — its "static" evidence-trail citations
  (confirmed live in `bugs_open/155`'s verification run) come from
  `diagnose_code_lookup_action.go` / `code_symbols_actions.go`. Reuse this
  action directly rather than writing a second code-search path.
- **SQL-shaped checks**: `query_database`, already generic and widely used
  (confirmed in the `$ctx.` param-namespace work).
- **Shell/kubectl-shaped checks** (a real fraction of entries — `strings
  /app/agent-chassis`, `grep -ac`, etc.): **no existing action runs an
  arbitrary shell command from inside a workflow**, by design (agents don't
  get a shell). Open question for implementation, not decided here: either
  (a) these entries fall back to LLM judgment on internal consistency only
  (weaker, but honest about the limitation), or (b) a narrow, allow-listed
  "run this specific class of read-only pod check" action gets built — which
  would itself be new platform code, council-gate scope, and its own
  submission. Do not build (b) speculatively; only if (a) proves insufficient
  in practice.

### Design sketch

A single-step agent (`landmine-verifier`, matching the `asset-deployer` /
`section-editor` shape — an `agent_definitions` row, not new Go code for the
agent itself):

1. Input: one entry's four fields, resolved from the `doc_notes` row
   `landmines-sync.py` already writes (`categories ? 'landmine'`,
   `source LIKE 'LANDMINES.md#%'`).
2. Run the footprint/check-resolution actions above for whatever's
   mechanically checkable.
3. One `execute_llm_prompt` call: given the entry + the fresh check results,
   judge whether the entry still holds. Single pass — no iteration, no
   hypothesis refinement (that's what makes it "dedicated" rather than a
   reuse of the full diagnose-orchestrator loop, which is built for
   multi-round root-causing, not a one-shot doc fact-check).
4. Output: a verdict + citations, written back as... **open question**:
   append to the entry itself (a `**last verified:** YYYY-MM-DD, <verdict>`
   line, mechanical and visible in-file), or a separate `doc_notes` row
   (queryable, doesn't touch the append-only ledger's own text). Leaning
   toward the former — matches this repo's preference for visible,
   in-place, dated corrections over a parallel ledger — but this is a real
   decision for whoever builds it, not settled here.

### Trigger

Wire as a step appended after `landmines-sync.py --apply` (new/changed
entries only — diff against the previous sync state, which the script
already tracks via its hash-based dedup), rather than a separate periodic
sweep. New entries get checked once, at the moment they're most likely to
be wrong (freshly written, unreviewed) and most likely to matter (about to
be read by a council seat for the first time).

### Blast radius / scope note

This is additive and inert until wired: it does not change what
`landmines-sync.py` writes, what `doc_notes` contains today, or what any
council seat's `schema_hint` carries. Per the 2026-07-29 owner ruling's
narrowing (additive-and-inert vs. additive-and-guarantee-changing), this
should not need its own architecture RFC beyond this one — but if the agent
definition's SQL touches `platform/`-scoped Go (only true if the shell-check
allow-list gets built, per the open question above), that piece specifically
goes through the normal council gate before shipping.

---

## Part B — weekly `bugs_open/` staleness sweep (§3.3)

### Design sketch, cheap-first-pass only (per RFC_005 §3.3)

A CronJob-triggered agent, same family as `build-pipeline-trigger`
(`docs/agent_docs/sql_for_agents/052_build_pipeline_trigger.sql`):

1. Enumerate `bugs_open/*.md` files whose status line does not read CLOSED
   (the existing 016b §10 index already tracks this by number; the sweep can
   read the files directly rather than trusting the index, since the index
   is itself hand-maintained and can lag).
2. **Cheap pass only, this phase**: for each file, extract cited
   `path:line`/`path:Symbol` references and confirm the path still exists and
   the symbol is still greppable nearby (reuse the same code-lookup action as
   Part A — one reusable capability, two consumers). This catches refactors,
   renames, and deletions mechanically, with no LLM cost.
3. **Explicitly deferred, not built this phase**: re-running each bug's own
   "How to verify a fix" query/command and diffing the result. Heterogeneous
   per bug, meaningfully more engineering, and the RFC's own §3.3 language
   ("deeper pass, sampled, not every run") treats this as a later increment,
   not the initial build.
4. Output: a flagged worklist (which files, which citations no longer
   resolve) — written where a human/thread will actually see it. Candidate:
   a `doc_notes` row per run (queryable, matches how the diagnose loop
   already persists notes) plus a short append to a new
   `docs/agent_docs/docs024_key_docs_latest/bugs_open_staleness/` log, rather
   than editing any `bugs_open/*.md` file directly — this sweep should never
   auto-write into a bug file it didn't author.

### Cadence

Weekly, per the owner's ruling — a K8s CronJob, matching
`build-pipeline-trigger`'s pattern (that one runs on a 30-minute heartbeat;
this is intentionally much less frequent, since `bugs_open/` entries don't
churn at anything like that rate).

### What this sweep must get right from day one (found live, while drafting RFC_005)

Any check that reads "the current code" needs an explicit, resolvable ref —
**never a bare `HEAD`**, and refuse rather than silently fall back to a stale
`main`. This session's own branch was 547 commits ahead of `origin` at one
point mid-conversation, and the function `bugs_open/155` describes was not
yet an ancestor of `origin` either — running a diagnosis against that stale
ref would have "returned a confident wrong answer" (090's own words, from a
2026-07-19 incident it already learned this lesson from). The sweep should
copy `090`'s ref-resolution discipline (current branch, refuse if unresolvable
on the remote) rather than re-learn it.

---

## Open questions for whoever builds this (not decided in this PLAN)

1. Part A: shell/kubectl-shaped checks — accept the weaker LLM-only fallback,
   or build the narrow allow-listed action? (recommend: ship without it,
   revisit if it proves insufficient)
2. Part A: verdict written in-file vs. a separate `doc_notes` row?
   (leaning in-file; not settled)
3. Part B: where exactly does the flagged worklist live, and who is expected
   to act on it? (a worklist nobody reads is the exact "unadopted ledger"
   failure this whole thread has been trying to avoid)
4. Both: concept-register entries owed on delivery, per CLAUDE.md's own rule
   — not optional, just not yet done because nothing has shipped.

## Next steps

Neither mechanism is built yet. This PLAN scopes both; building either is a
separate piece of work (new `agent_definitions` rows at minimum, a CronJob
manifest for Part B, possibly a council-gate submission if the shell-check
allow-list in Part A gets built). Continue here rather than opening a third
document once work starts.
