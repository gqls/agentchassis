# RUNBOOK — Concept Register (operator tasks)

**What this project is about (read this first).** The concept register
(`docs026_concept_register/`) is a three-stage effort to catalogue every
mechanism/responsibility/behaviour documented anywhere under `docs/`, verify
each one against the real codebase, and eventually turn the category structure
into a roster of expert council agents for the fix-loop. Stage 1 (extract) and
stage 2 (verify) are both complete; stage 3 (council agents) has a grounded
design, not yet implemented. **This document is
the list of tasks that need a human**, with the commands to run. Most of the
work in this project is read-only analysis (grep/find/read against the local
git checkout) — no cluster or DB credentials were needed for stages 1–2.

Conventions used below: tasks are grouped into **standing rituals** (repeat
whenever they apply) and **one-off tasks** (checked off once, tracked in
`RUNNING_NOTES_concept_register.md`). In a Claude Code session you can run any
command yourself by prefixing it with `!` at the prompt.

---

## A. Standing rituals

### A1. Before trusting any register entry for a real decision
The register's `status` column is a **documentary signal**, even after stage 2
verified all 1,627 concepts — verification found a ~7.6% error rate, so most
entries are right but not all, and even corrected entries are a snapshot from
2026-07-14. If a register entry is about to gate a real decision (e.g. a
stage-3 council agent citing it, or a fix-loop diagnosis leaning on it), spend
one grep confirming its `verify-later` pointer still resolves before trusting
it as current.

---

## B. One-off tasks

### B1. Rotate two leaked credentials (found during stage-1 extraction, not yet actioned)
Two live-looking secrets were found in doc files during the extraction sweep.
Neither was touched (extraction is read-only by design) — this needs a human
decision and action:

1. **Thunder API bearer token** in
   `docs/agent_docs/docs024_key_docs_latest/finetuning/working/flywheel_docs/ssh_probe.sh`
   — check whether it's still live in the Thunder Compute console; rotate if so.
2. **AWS console password + account ID** in
   `docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/README_email.md`
   — rotate the password in the AWS console; consider whether the account ID
   disclosure itself needs any follow-up.

Neither file has been modified — decide separately whether to redact the docs
themselves after rotating (a docs-hygiene call, not part of this register).

### B2. ~~Decide on the PUB-001 duplicate retirement~~ — DONE 2026-07-14
Retired to a pointer entry in `register/public-api.md` (kept the ID, kept the
full entry for its distinct P2 source citation, added an explicit
duplicate-of note); index row updated. Kept `public-api.md` as its own
category file rather than folding it into `admin-dashboard-and-api.md` —
single-concept categories are already normal in this register.

### B3. ~~Decide whether to sweep superseded/abandoned before stage 3~~ — DONE 2026-07-14
Swept: `stage2_workflow.js` extended with `superseded` and `abandoned` prompt
modes (hunting, respectively, for claimed replacements that don't actually
exist, and abandoned ideas quietly resurrected). 18 corrections confirmed (9
overturned) across all 102 superseded + 72 abandoned concepts — see
`006_VERIFICATION_stage2.md` batch 3. **Stage 2 is now fully complete: every
concept in the register has been checked at least once.**

### B4. Stage 3 design — RESOLVED 2026-07-14, implementation is your call
The three open design questions (granularity, activation, freshness) now have
concrete recommendations grounded in the live fix-loop mechanism — see
`PLAN_concept_register.md` §Stage 3. Summary: relevance-filtered reviewer
subset per run (not all 107), selected by matching the fix plan's touched
files/tables against each category's `verify-later` footprint, with each
reviewer seeded from its register file but required to live-recheck the
specific pointers relevant to that run (not trust the stale label).

**Not implemented against the live workflow** — that's deliberately left to
you. `fixloop_eg_dartsonline/0NN_fix_proposer.sql` belongs to the separately
active [[fixloop-workstream]], and changing a production council's
decision-gating logic is a cross-workstream call, not something to proceed on
without your sign-off. When you're ready, the recommended first step is a
single pilot category-reviewer (pick whichever category the next real
fix-loop incident's blast radius touches) rather than building all
category-reviewers at once.

### B5. ~~Pick a stage-3 pilot seat~~ — candidate B is LIVE, 2026-07-16
Two data-driven candidates were identified (see `PLAN_concept_register.md`
§Stage 3), both matching FIX-036's originally named roster: candidate A —
reuse-agent (`tool-lifecycle.md`) and candidate B — bug-historian (the
"silent content loss during rerender" family). You picked bug-historian.

**Applied to `clients_db` with your explicit sign-off.** The council is now
three reviewers: `review_editquality → review_bug_historian → review_guardian
→ council_decide`. Full record: `PILOT_bug_historian_reviewer.md` §6.
Prior `fix-proposer` row snapshotted first (rollback available via
`agent_definitions_backup`). Committed:
`docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_fix_proposer_v6_bug_historian.sql`.

**Remaining:** watch the next real fix-loop run to confirm the new reviewer's
output parses cleanly in production traffic (verified via direct DB read so
far, not yet exercised end to end).

**2026-07-17: candidate A (reuse-agent) also applied — council is now 4
reviewers.** `review_editquality → review_bug_historian → review_reuse_agent
→ review_guardian → council_decide`. Full record:
`PILOT_reuse_agent_reviewer.md`. Correctly re-grounded to `DEV-001` (not
`tool-lifecycle.md` as first framed — that category's density turned out to
be about a different, already-covered theme). Pre-flight checked for
in-flight council activity first (none found).

**2026-07-17: seat #3 (guidelines) also applied — council is now 5
reviewers.** `review_editquality → review_bug_historian → review_reuse_agent
→ review_guidelines → review_guardian → council_decide`. Full record:
`PILOT_guidelines_agent_reviewer.md`. Distinctive two-lens design: objects on
rule *violations*, but treats a *guideline-gap* (rule itself is wrong) as an
approve+note, not an objection — so a correct fix isn't forced to revise for
exposing a bad rule (grounded in FIX-016/FIX-042, with `MDL-039`/BUG B as the
live example of a backwards rule). Pre-flight confirmed no in-flight council
activity and that live state was v7 before applying.

Per your direction ("do the guidelines member then we can look at the
relevance filtering mechanism"), guidelines was the last always-on seat —
the design of the relevance filter follows (see B7).

### B7. Relevance filter — ENGINE BUILT 2026-07-17; the DEPLOY is your call
Per your "the relevance filter can be next," the Go engine is built, tested,
and committed (`37468ba65`): `select_review_panel_action.go` (a generic,
config-driven footprint matcher) + a `council_decide` change (absent reviewer
= abstention). `go build` + `go test` green; committed **inert** (nothing
calls it until the SQL wires it; the abstention can't trigger until skips
exist — today's behaviour unchanged). Full spec + the ready-to-apply `v10`
SQL wiring (footprint config + per-seat gates, retrofitting the 3 advisory
seats first) is in `DESIGN_relevance_filter.md` §7.

**The remaining step is a DEPLOY, and it's genuinely your call because it's a
different class of change from the SQL-only seat adds:**
- It's a **chassis image**, which is **fleet-wide** — it rebuilds the binary
  every agent runs, not just fix-proposer.
- The Go is **shared with the actively-developed council-gate/feature-builder
  thread**; ideally the SAME `select_review_panel` binary serves both councils,
  so the deploy should be **sequenced with that thread**, not shipped as a
  fix-proposer-only image.
- Deploy sequencing (per `CLAUDE.md`): `make build-<service>` builds from
  committed `HEAD` (my Go is committed, so it's included; no WIP-collision
  worry after the 2026-07-17 build inversion) — but a HEAD build ships ALL
  committed code, incl. other threads' untested work. Bump `IMAGE_TAG`, roll,
  verify against the pod (`strings /app/agent-chassis | grep -c SelectReviewPanelAction`),
  **image before the §7 SQL** (a workflow naming an action the binary lacks
  fails at that step), no dispatch within ~300s of a chassis restart.

Decision needed: (a) I coordinate with the gate thread and drive the deploy;
(b) you/another session folds it into a planned chassis release; (c) hold it.
The 7 specialist seats (#3 adoption … #10 contextkit) build behind the filter
once it's deployed.

### B9. Retire the DEV-001 mis-grounding note once comfortable (housekeeping)
Minor: the reuse-agent's grounding was corrected mid-build (tool-lifecycle →
DEV-001); `PILOT_reuse_agent_reviewer.md` §1 records it honestly. No action
needed — noted only so the correction isn't mistaken for an inconsistency.

### B8. Read the council-gate thread's own open decision — it names this workstream directly
`fixloop_eg_dartsonline/HANDOFF_2026-07-17_council_gate_thread.md` asks, as
one of ITS OWN owner-decisions-to-collect: "Seat roster for the gate: the 3
live seats, or wait for more concept-register stage-3 seats?" That thread is
building the delivery vehicle for everything this workstream produces — its
own words: "no competing design." Worth reading both handoffs together before
deciding how many more seats to build right now vs. let that thread proceed
with what exists.

**2026-07-17 update:** the loop delivered its first real-case CONFIRMED
diagnosis (`MDL-038`, "BUG A" — `GenerateText` never decodes `stop_reason`,
so truncated LLM calls silently look complete), but the fix dispatch to
fix-proposer is "awaits owner go" per fixloop's own notes — so the
bug-historian still has not actually run. Also confirmed a second bug found
the same session (`MDL-039`, "BUG B" — root-vs-step `ai_service` config
shadowing, 17-agent fleet blast radius) does NOT affect `fix-proposer` —
re-read the file directly, no top-level `ai_service` key present, so the
bug-historian's `max_tokens` config is unaffected by this defect.

### B6. Fixloop's case-004 dispatch may be moot — flagging, not deciding
2026-07-16 coordination finding: fixloop chose the image-landing/article-body
trap (`aaa_fails_to_mend/004`) as its first real case. A separate concurrent
session resolved the underlying data loss the same day
(`aaa_fails_to_mend/005`, confirmed via file mtime and an independent
`go test` re-run) — 2 of 004's 3 "open items" are done; only the structural
`missingkey=zero` defect (`STY-049`) remains. Worth checking whether the other
queued real cases (001 replan-clobbers-built-pages, 002 errors-to-fix, 003
spawn-lost-child-response) are now higher-value to dispatch first. This is a
fixloop-workstream decision, surfaced here because this register's sweep is
what caught the discrepancy — not actioned by this workstream.

---

## C. What you should expect the agent to do (so you don't have to)
- All verification, corrections, and register edits — grep/find/read against
  the live repo, never assuming a doc claim without checking.
- Updating `RUNNING_NOTES_concept_register.md` every turn and keeping
  `PLAN_concept_register.md` in sync with decisions and stage status.
- Flagging exactly which of the tasks above (B-numbers) need your input next,
  rather than assuming you've read the whole queue.
- Never touching files outside `docs/agent_docs/docs026_concept_register/` as
  part of this workstream (credential rotation in B1 is the one exception —
  that's your action, on the original doc files, not the agent's).
