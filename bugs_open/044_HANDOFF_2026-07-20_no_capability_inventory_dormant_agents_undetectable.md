# 044 — nothing detects a capability that exists but nothing routes work to; 57 active agents have never run

**Filed:** 2026-07-20 by the bugfix thread, out of `bugs_closed/002` D.
**Severity:** latent, diagnostic. Nothing errors. No site reports a failure.
**Status:** OPEN — the **detector half is BUILT and committed** (inert until the
next chassis image roll); the `is_active` hygiene half remains an owner decision
(see §The two halves and §UPDATE below).

> **UPDATE 2026-07-21 — detector half BUILT (dormant_agents_inventory workstream).**
> Half 1 (the detector) is implemented, tested, and committed; stays OPEN only
> until a chassis image carrying it is rolled (the bar is fixed AND live).
> - Action `diagnose_dormant_agents` (`platform/orchestration/actions/
>   diagnose_dormant_agents_action.go`) — a deterministic fleet sweep beside
>   triage/silent-check, exactly as this handoff [INFERRED]. Implements the
>   step-fingerprint method verbatim; `owner_agent_type` deliberately unused.
>   Age floor (default 14d) gates emission; the mirrored-agent blind spot is
>   listed but **never flagged** (it cannot enter the never-set by construction).
>   Emits INERT `dormant_agent` items (status='dormant', pipeline='maintenance',
>   unclaimable, anchored to system.internal) for human triage; closes them when
>   the agent is observed to run; one doc_note per sweep (categories
>   `dormant-agents`+`fixloop`). Ships `dry_run=true`, manual trigger.
> - Seed: `docs024.../dormant_agents_inventory/seed_diagnosis_dormant_agents.sql`
>   (image-first — apply only after `grep -ac diagnose_dormant_agents /proc/1/exe`
>   ≥ 1 in the chassis pod). Standing-five docs in that dir.
> - **Live numbers moved since filing:** 156/122/**57** → **155/123/77** (70 past
>   the 14d floor). More agents seeded; the 55-day orchestration window advanced.
> - **Correction to this handoff's validation:** `feature-designer` is listed
>   above as a positive control that "correctly detects as run." Live it reads as
>   **never-run**, and that is correct — its 3 unique steps appear nowhere in
>   `orchestration_states` (not even by text). Its own workflow has never fired;
>   the "PROVEN 2026-07-18" run was the council approving its plan through other
>   machinery (councils log `agent_type='generic'`). The solid positive controls
>   are fix-proposer / page-build-handler / section-editor.
> - **`orchestration_name` rejected** as the mirrored-agent second signal this
>   handoff suggested: live it is `generic-orchestrate-<ts>`, it does not name the
>   agent. The blind spot stays unmeasurable (listed, never flagged).
>
> **UPDATE 2026-07-22 — LIVE on v1.0.1149, and a substrate finding → `bugs_open/060`.**
> The action shipped (pod-grep `diagnose_dormant_agents`=4); seed applied; a real
> sweep ran in ~3s and wrote its dry-run report to `doc_notes` (categories
> `dormant-agents`). **The discoverability half (the inventory report) is LIVE and
> proven.** But running it exposed that the detection SUBSTRATE is too ephemeral for
> reliable *emission*:
> - `orchestration_states` is pruned **hourly at 24h** by the `database-cleanup`
>   task. The ~55-day window this handoff's method was validated against
>   (2026-07-20) was a TRANSIENT — the cleanup wasn't pruning then. Live 07-22:
>   1,737 rows, 9-day span, 94% in the last 36h. On a 24h window "never observed"
>   over-flags: **`fix-proposer` (runs constantly) is now flagged**, its runs pruned.
> - There is **no durable "has ever run" signal**: `usage_count` is 0 for all 162
>   agents (unmaintained), `orchestration_state_audit` is ~2 days. So the LIFETIME
>   question (section-editor's 3 runs) cannot be answered at all today. Filed
>   **`bugs_open/060`** (the real blocker to reliable emission).
> - **Response (committed):** a WINDOW GUARD (no emit while window < age floor;
>   loud banner) + reworded report (recent-activity, not lifetime) + a seed
>   flip-warning. Emission is now correctly gated off until 060 is addressed.
>   `dry_run` stays true (the LIVE 1149 binary has no guard yet — its only
>   protection). **044 stays OPEN**: the report is live, but the emitting detector
>   the two halves describe is blocked on 060.

This is the **producer-side** mirror of `bugs_open/033` (findings reach
`needs_human_review` and no consumer ever actions them). 033 is work with no
worker; this is a worker with no work. Both are invisible because nothing fails.

---

## Why it was filed

`bugs_closed/002` D was signed off twice claiming a repair mechanism did not
exist. It did: `apply_section_edit` / the `section-editor` agent, shipped
2026-02-19, deployed, **3 completed production runs**, declared nonexistent
through a handoff, a re-grounding pass and a sign-off. The council now has a
seat for the assertion side (`review_prior_art`, live and mirrored, always-on).

**The seat closes the claim; it does not close the condition.** A seat only sees
a submission. It cannot see a capability quietly going unused for five months
until someone declares it missing — and it never sees an absence asserted in a
handoff, which is where this one was. The platform has **no inventory of its own
capabilities and no detector for unused ones**, so knowledge of what exists
decays into folklore at the pace of session turnover.

## What the discovery layer covers, and why this falls outside it

All **49** registered discovery checks are **site-scoped** — they take a
`site_id` and walk that site's pages, components, images, CSS, news, tools
(`platform/orchestration/actions/discovery_checks/`; only `check_news_feed.go`
reads `agent_definitions`/`orchestration_states`, and that is for news sources).

Nothing inspects the **platform's own** inventory, and this cannot simply be
added as another site check: *"which agents never run"* is not a per-site
question, so it does not fit `DiscoveryCheckContext`. **[INFERRED]** the right
home is a fleet-level sweep beside the immune system's existing triage /
silent-check rather than a 50th discovery check — an owner should confirm.

## Measurement (2026-07-20, live)

**Method, and why the obvious one is wrong.** `owner_agent_type` **cannot** be
used to tell whether an agent has run: **95,797** of ~101k orchestrations carry
`owner_agent_type='generic'` because that is the dispatch path, not the agent.
Counting that way reports `fix-proposer` and `council-gate` as never-run when
both demonstrably run (councils appear in **92** `workflow_plan`s;
`review_prior_art` in **20**). **My own first pass made exactly this error and
produced 110 — see `WRONG_CALLS.md`.**

What works is a **step fingerprint**: a workflow step key belonging to exactly
one agent, looked for among step keys ever observed in a real
`orchestration_states.workflow_plan`.

```sql
CREATE TEMP TABLE observed_steps AS
SELECT DISTINCT jsonb_object_keys(workflow_plan->'steps') AS step
FROM orchestration_states
WHERE workflow_plan ? 'steps' AND jsonb_typeof(workflow_plan->'steps')='object';

CREATE TEMP TABLE agent_steps AS
SELECT a.type, jsonb_object_keys(a.default_config#>'{workflow,steps}') AS step
FROM agent_definitions a
WHERE a.is_active AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
  AND jsonb_typeof(a.default_config#>'{workflow,steps}')='object';

CREATE TEMP TABLE fingerprints AS          -- step keys unique to ONE agent
SELECT step, min(type) AS type FROM agent_steps
GROUP BY step HAVING count(DISTINCT type)=1;

SELECT f.type FROM fingerprints f GROUP BY f.type          -- never observed
HAVING bool_and(f.step NOT IN (SELECT step FROM observed_steps));
```

| figure | value |
|---|---|
| active, non-snapshot agents with a workflow | **156** |
| …measurable (have ≥1 unique fingerprint step) | **122** |
| …**never observed in any orchestration** | **57** |
| **blind spot:** active agents with NO unique step key, unmeasurable this way | **34** |

**Method validated** against agents known to be live: `fix-proposer`,
`page-build-handler`, `section-editor`, `feature-designer` all correctly detect
as run. **Independent corroboration:** the method flags `feature-implementer`
(seeded 2026-07-17) as never-run, which the feature-builder workstream's own
handoff already records as *"live but NEVER FIRED — the whole remaining gap"*.
The detector found a known-true fact it was not told.

**Known limitation:** `council-gate` is in the 34-agent blind spot — the 099
roster mirror copies its steps verbatim from `fix-proposer`, so it has no unique
key. Any real check needs a second signal (**[INFERRED]** `orchestration_name`,
or a marker step) for mirrored agents.

## The 57 are not one thing — a raw count would mislead

Grouping by seed date makes the categories obvious, and **only the third is the
defect this bug is about**:

- **~34 legacy / superseded** (2025-08 → 2026-02): `multipage-website-builder`,
  `html-assembler`, `chief-strategist`, `portfolio-architect`,
  `content-creator-hero`, `site-planner`, `calculator`, `content_researcher` …
  the previous generation, replaced by the current pipeline. Not dormant
  capability — **retired code still flagged `is_active=true`**, which is its own
  hygiene problem (below).
- **~15 paused workstreams**: `med-*` (7, the quarantined vetcomparison export),
  `ch-*` (4, Companies House), `vet-*` (3). Dormant *by decision*.
- **~8 current-generation capabilities that have never fired** — the sharp set:
  `feature-implementer`, `diagnose-dispatch-loop`, `work-item-archiver`,
  `component-quality-auditor`, `tool-deployer`, `css-patch-agent`,
  `nav-link-fixer`, `color-variable-fixer`, `site-component-linker`.
  **[UNVERIFIED]** I have not opened these to confirm each is genuinely wired-but-
  unrouted rather than superseded; that triage is the first task for whoever
  takes this. `feature-implementer` is confirmed (its workstream says so).
  Several of the names — a nav-link fixer, a colour-variable fixer, a component
  quality auditor — are **exactly the repair shapes threads have recently
  hand-rolled or declared missing**, which is the pattern this bug exists to stop.

`evidence-researcher` (seeded 2026-07-20) is in the 57 but is simply new, not
dormant — **any check needs an age floor or it will flag every fresh seed.**

## The two halves (separable; the second is an owner call)

1. **The detector.** A fleet sweep emitting one work item per
   never-run-in-N-days active agent, with an age floor and the mirrored-agent
   caveat handled. This is the part that would have made `section-editor`
   discoverable before it was declared missing.
2. **`is_active` hygiene.** ~34 rows are retired-but-active. Also found while
   measuring: **5 agent types have MORE THAN ONE active non-snapshot row**
   (`multipage-website-builder`, `chief-strategist`, `content-creator-contact`
   …). `FindByType` does `ORDER BY version DESC LIMIT 1`
   (`platform/discovery/agent_discovery.go:109-125`), so the extras are silently
   shadowed — harmless today, but it means "is it active?" has more than one
   answer. **Deactivating retired agents is a judgement call about intent, not a
   code fix — owner decision.**

## How to verify a fix

Re-run the fingerprint query above; a working detector should produce work items
matching its output minus the age floor. **Do not accept `owner_agent_type` as
the implementation** — that is the trap that produced the wrong 110.

## References

- `bugs_closed/002` D — the case that motivated it; `016b` §9 `asserted-absence /
  dormant-machinery` for the pattern and the four-lookup existence check.
- `WRONG_CALLS.md` 2026-07-20 — three rows from this investigation, including the
  110-vs-57 miscount.
- `bugs_open/033` — the consumer-side mirror (work with no worker).
- Live council seat `review_prior_art` on `fix-proposer` + `council-gate`
  (0 roster drift, verified 2026-07-20) — covers the assertion, not the condition.
