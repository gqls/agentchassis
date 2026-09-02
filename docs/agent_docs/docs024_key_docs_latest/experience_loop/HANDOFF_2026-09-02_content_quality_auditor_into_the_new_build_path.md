# HANDOFF — put `content-quality-auditor` into the NEW BUILD path

**Opened 2026-09-02** by the experience_loop lane, on the owner's instruction:
*"yes, content quality auditor should be in the new build path."*

**Start here. Read §1 before planning anything — the premise this task was raised on
(2026-08-31) is now STALE, and the work that remains is a different, smaller job than the
one the CONTRIB describes.**

---

> ## ⛔ GATE RESULT, 2026-09-02 (added by the session that worked this handoff)
>
> **§2's gate was run and it FAILED. Do NOT route this seat yet — the routing is now step 3 of 3.**
>
> The audit output is articulate and its findings are fair, but it names **none** of the three
> things the owner complained about, and it structurally cannot. `load_page_content` hardcodes
> `p.name IN ('index','about','services','contact')` — **3 of 22 pages** on boxingonline,
> **92 of 1,196 fleet-wide (7.7%)** across 36 sites. The four guides, the articles-index
> manifesto and the fighter-comparator form are all outside those four names, so no prompt
> wording could ever have reached them. Three compounding defects in the same query: the
> 1,000-char cap samples an index page at **4.5%**; `rendered_html` is **42.8% CSS** fleet-wide
> and on boxingonline's index `<style>` starts at **character 1**, so 999 of 1,000 chars reaching
> the model were stylesheet; and `string_agg` had **no `ORDER BY`**, so the window drifts across
> runs on a byte-identical page. Plus a prompt/enum mismatch: dimension 5 is AUDIENCE, the enum
> offers `content`, and 10 of 210 stored findings emit an out-of-enum `audience`.
>
> **Full evidence:** `RUNNING_NOTES_experience_loop.md`, entry 2026-09-02.
> **Fix:** migration `docs/agent_docs/sql_for_agents/694_content_quality_auditor_can_see_the_site.sql`
> (+ `_ROLLBACK`), committed `5ff171327`, `Council-Submitted: d52a0e45-5c64-4d32-a1ab-f73532684d37`.
> **NOT YET APPLIED** — apply after the verdict, then observe one real audit before routing.
>
> **§4 is ANSWERED: the owner ruled RECORD ONLY**, asked directly on 2026-09-02. Findings are
> verdict rows for human approval; nothing auto-regenerates, preserving the 2026-08-25 ruling.
> **§4's "name the reader" is discharged and needs no build** — it already exists end to end:
> admin dashboard `recordVerdictsOnly` filter (`frontends/admin-dashboard/src/App.tsx:474`
> → `GET /admin/work-items?filing_mode=record`) and the per-finding Release button
> (`App.tsx:737` → `POST /admin/work-items/:item_id/release` → route
> `internal/core-manager/api/server.go:455` → `HandleReleaseRecordVerdict`).
>
> **Two corrections to this handoff's own figures**, both worth carrying forward:
> - It says **44** auditor runs; today the same query returns **42**. `orchestration_states` is
>   reaped, so every "N runs since" figure here has a shelf life of days.
> - §7's open items are unchanged, but the CONTRIB's parting question — *"the 25 FAILED against
>   49 COMPLETED … is its own question and I have not looked into it"* — **is now answered and
>   is NOT a code defect.** Lifetime the seat shows 7,915 LLM calls with 5,114 failures (65%),
>   but 5,013 of those are `"You have reached your specified API usage limits"` and 99 are
>   `"Your credit balance is too low"`, spanning 2026-04-08 → 2026-08-31. Since 2026-09-01:
>   **74 calls, 0 failures.** It was billing caps, and they have cleared.

---

## 1. The premise changed under us. What is true TODAY, all measured 2026-09-02

The originating CONTRIB (`CONTRIB_2026-08-31_from_the_first_paid_build_four_experience_defects_that_every_check_passed.md`,
defect 4) said: *"We do [have a quality auditor], and it did not run … zero of those runs
touched this site. It is not in the new-build path."* **The first half is no longer true.**

```sql
-- real auditor runs are the ones carrying its own output field, NOT a text match on the
-- agent name (a workflow_plan mentioning the auditor is often the PARENT's spawn record —
-- I made that mistake first and the two queries only agree by luck; use this one)
SELECT count(*) AS auditor_runs,
       count(*) FILTER (WHERE collected_data::text ILIKE '%d2aa5206%') AS on_boxingonline,
       min(created_at)::date AS first_run, max(created_at) AS latest
FROM orchestration_states WHERE collected_data ? 'content_audit';
--  44 | 4 | 2026-09-01 | 2026-09-02 12:35
```

So: **the auditor runs, fleet-wide, and has run on the paid site four times.** It began on
**2026-09-01** — the day AFTER the CONTRIB was written. The improvement sweep
(`scheduled_tasks.improvement-sweep`, target `improvement-loop`) is `enabled=t`, 900s, last
fired 2026-09-02 12:45. It has covered 12+ sites in two days.

**The gap is real but it is TIMING and EFFECT, not existence:**

| | measured |
|---|---|
| **exists / runs** | ✅ 44 runs since 2026-09-01, 4 on boxingonline |
| **timing** | ❌ post-hoc only, one site per 15-min tick over 54 sites. boxingonline was built 2026-08-31 and first audited **2026-09-01 21:32** — roughly 30 hours later, and after the owner had already found the defects by hand |
| **effect** | ❌ `filing_mode='record'` is LIVE on it, so a finding is a verdict row neither promoter dispatches. On boxingonline **1 row of 574** carries the record-mode stamp `spec ? 'would_have_routed_at'`, and that one came from `offer-analysis`, not content-quality |

**That is why the owner's instruction is still right.** A checker that runs a day after
delivery, and whose findings nothing acts on, cannot stop a paid site shipping with padding.

---

## 2. THE FIRST MOVE, and do not skip it

**Read what a run actually produces before wiring anything in.** Routing a seat that reports
nothing just moves the silence earlier.

```sql
SELECT left(collected_data->>'content_audit', 3000)
FROM orchestration_states
WHERE collected_data ? 'content_audit' AND collected_data::text ILIKE '%d2aa5206%'
ORDER BY created_at DESC LIMIT 1;
```

Judge it against what the owner actually complained about on that site: guide pages padded
with "how we build the calendar" prose, an articles index writing a manifesto instead of
listing articles, tools that ask the reader for data we should have supplied. **If the audit
names those, wiring it into the build path is the whole job. If it does not, the job is
bigger and the prompt is the work** — say so plainly rather than shipping a route to a seat
that will stay quiet.

---

## 3. Where it goes, and the two candidate shapes

The new-build path is `site-adoption-agent` (dispatched by
`scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh`) →
**`site-work-orchestrator`**, which is the agent to edit. Its live tail, read 2026-09-02:

```
fix_items_loop -> apply_site_design -> update_site_status -> complete
```

`apply_site_design` is the last step that changes the site. **Insert between
`apply_site_design` and `update_site_status`** — the site is fully built and designed, and
nothing has been marked delivered yet.

`site-work-orchestrator` spawns, as of 2026-09-02: site-planner, content-reviewer ×2,
page-content-writer ×2, webdesign-agent ×2, image-generator ×3. **`content-quality-auditor`
is not among them** (verified with `jsonb_path_query_array(default_config,'$.**.agent_type')`,
which descends nested steps — a `for` over `workflow.steps` would miss loop substeps).

**It is not redundant with the `content-reviewer` already in that loop, and this is the
argument to make if challenged.** `content-reviewer` is per-PAGE ("reviews page content for
quality, accuracy, and brand alignment", inside `build_items_loop`).
`content-quality-auditor` is a GROUP auditor: "loads the site brief, page content samples,
and target audience … assesses tone alignment, content gaps, CTA effectiveness, and
differentiation". Boxingonline is the case for the distinction — every page passed
individually and the SITE was wrong.

**Prefer `call_agent` over `spawn_agent`.** Two reasons, both evidenced:
- the spawn→call handshake fails about half the time fleet-wide (2 COMPLETED / 2 FAILED
  all-history when last measured — grep `spawn-call-handshake-races` in the auto-memory);
- `apply_site_design` in this very workflow is already a `call_agent` and is the shape to
  copy verbatim (`action: call_agent`, `config.agent_type`, `config.input_mapping`,
  `timeout_seconds: 300`, `output_field`).

The auditor needs no storage client — its steps are `ensure_site_record → load_brief →
load_page_content → check_empty_pages → run_content_llm_audit → set_audit_source →
write_findings`, i.e. DB queries plus one LLM call. So the inline-chassis storage landmine
(`params.StorageClient` is nil on the chassis) does **not** bite here. Say so in the
submission; a reviewer will ask.

---

## 4. The decision the owner may need to make: record, or dispatch?

`filing_mode='record'` is live on `content-quality-auditor`, `visual-design-auditor`,
`brief-fidelity-auditor` and `reader-experience-auditor`. It exists because of the owner's
**2026-08-25 ruling**: *"switch off the evolutionary aspect of the improvement loop … it is
causing too many bad / unexpected renders."* Those seats' findings used to route to handlers
that REGENERATE pages — lifetime, from design-audit alone: 976 `content_rewrite`, 399
`needs_content_page`, 964 `needs_content_planning`, 26 `tone_shift`.

So there are two shapes, and they are a real choice, not a detail:

- **(a) Route it in RECORD mode (recommended).** The build-time audit produces verdict rows
  and nothing auto-regenerates. It satisfies the owner's 2026-09-02 cut-line for this site
  ("build and fix everything before approval") by putting the findings in front of whoever
  approves, and it cannot reintroduce the churn the 08-25 ruling was written to stop.
  **But then something must READ them, or this repeats defect 4 one layer along** — a
  finding nobody dispatches is the same silence as a check that never ran. Name the reader.
- **(b) Route it and let findings dispatch.** Before delivery, regeneration is arguably
  wanted. It is also exactly the mechanism switched off in August, now running inside the
  build with no cooldown between it and the writer that just produced the page. If you
  propose this, propose a bound with it.

**This is the owner's call and it should be put to him in one short question**, not decided
in a migration.

---

## 5. Ownership — talk to the loanzy lane BEFORE you edit

`architecture_review/RFC_056_the_site_acceptance_council_is_the_improvement_loop_and_its_seats_are_the_benchmark.md`
(DRAFT, raised 2026-08-25 by `loanzy_uk_example_site`) **owns the audit-seat question**, and
`filing_mode` is its seam. Read it first. Two things follow:

- **RFC_056 never mentions the new-build path** (grep: zero hits for `new build`,
  `site-work-orchestrator`, `at build time`). Its whole frame is the owner's 08-25 placement
  ruling, *"the checkers will check after the fact (improvement loop)"*. **Today's
  instruction changes that placement**, so it is not a detail to slip into a migration — it
  is a consumer of their seam changing, and the 2026-07-29 owner ruling says a shared
  mechanism's other consumers must be **told, not merely measured**. Tell them.
- **`_HOLD` migrations 619, 620, 623 and 624 are NOT applied** and 623 is already stale
  against live config: it flips `call_completeness_discovery.next_step` from
  `spawn_design_audit`, and the live value is **`spawn_acceptance_discovery`**. Do not apply
  a HOLD file without re-reading the live row first — the seed is not the system.

---

## 6. What NOT to do

- **Do not treat the CONTRIB's defect 4 as current.** It says the auditor never runs; it has
  run 44 times since 2026-09-01. Quoting it will make a false claim in a submission.
- **Do not count auditor runs by text-matching the agent name in `workflow_plan`** — that
  matches the parent's spawn record. Match on `collected_data ? 'content_audit'`.
- **Do not hand-patch `council-gate`** if seats change: `099_SYNC_gate_roster.py --apply` is
  suspended (see CLAUDE.md), so mirror with a surgical migration anchored on a verbatim line.
- **Do not skip the council gate.** A migration under `docs/agent_docs/sql_for_agents/` is in
  scope (widened 2026-08-19). `DRY_RUN=1` on the 097 trigger tests admission for free, and
  `Council-Submitted: <corr>` lets you commit before the verdict lands.

---

## 7. State of this lane's own work, for context

Two detectors built this week and both LIVE, verified at the pod and at a `doc_notes` receipt:

- **`listing-class-promise-check`** (`25 7 * * *` UTC) — a listing must show the content class
  its heading promises. Register **SQ-004**.
- **`experience-promise-check`** (`40 7 * * *` UTC) — rule A two header entries with one label
  pointing at different pages; rule B a `tool` page serving no control, no data and no fetch;
  rule C an index listing none of its own directory. Register **SQ-005**.

Open from the CONTRIB after this task: ask 2's listing half (an index whose set is empty and
which says nothing true about it — rule B covers the tool-page form only), and the
planner-side half of ask 3 (refusing to SELECT a tool at planning time when we would have no
data to put in it). Both are noted in `RUNNING_NOTES_experience_loop.md`.

**Refused deliberately, do not re-attempt as a regex:** "does this page contain the thing its
title asserts?" — see SQ-005 and the detector docstring for the disproof.
