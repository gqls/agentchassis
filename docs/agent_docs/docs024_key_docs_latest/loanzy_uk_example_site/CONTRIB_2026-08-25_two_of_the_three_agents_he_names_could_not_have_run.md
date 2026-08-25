# CONTRIB into the 2026-08-25 owner review — two of the three agents he names COULD NOT HAVE RUN on homegarden.uk

**From the `vigilant_designer_offer_analysis` lane, 2026-08-25, answering the reframe
`OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md` §6 puts to the owning lanes.**
That section says all three of his proposals map onto agents that already exist, and asks why they did
not help. Your handover added: *"if the honest answer is that none of them runs on this path, that is
worth saying plainly to him."*

**It is close to that, and here is the measurement.**

> ⚠ **NOTHING BELOW CONTRADICTS HIS REVIEW.** The site is as he describes it; every one of your
> served-artefact measurements reproduces. What changes is the CAUSE, and therefore what fixing it
> means. Two of the three verdicts are aimed at agents that were never asked to do the work.

---

## 1. The offer / benefit analyser NEVER RAN on homegarden.uk, and could not have

`[MEASURED 2026-08-25, live clients_db]`

| check | result |
|---|---|
| `site_specs` rows for `homegarden.uk` with `aspect='offer_ordering'` (the enrolment marker) | **0** |
| `site_work_items` for the site with `spec->>'audit_source'='offer-analysis'` | **0** |
| sites carrying `offer_ordering` fleet-wide | **5** — homegarden is not one |
| live sites | **28** |
| `scheduled_tasks` targeting `offer-analyser` | **3, and ALL THREE `enabled=false`**, last triggered 2026-08-14 / 08-15 |
| its only other route, `improvement-loop` | carrier `improvement-sweep` **also `enabled=false`**, last triggered 2026-08-17 |

**So "the offer and benefit analysis agent should make it much more clear what we can offer the
customer" is an instruction to an agent that has never looked at this site and currently has no
enabled path to any site.** Its reach is 5 of 28 by enrolment, and 0 of 28 by schedule.

⚠ **This does NOT make his complaint wrong.** `about.html` really does carry 14 methodology headings
against 3 reader-facing ones, and that is exactly the defect this agent exists to catch. **The finding
is that the catcher was switched off, not that it looked and failed.**

## 2. The visual designer is ACTIVE, STORAGE-GRANTED — AND UNREACHABLE

This is the sharper one, because his verdict is *"it hasn't done its job."*

`[MEASURED 2026-08-25]` `visual-designer` is a live `agent_definitions` row (`is_active=true`), with a
real LLM step (`design`: `action=execute_llm_prompt`, carrying `ai_service` and `prompt_template`).
And **nothing in the estate can dispatch it**:

| route | result |
|---|---|
| `scheduled_tasks` with `target_agent_type='visual-designer'` | **none** |
| any other live agent whose `default_config` names it (`call_agent`, sub-workflow, spawn) | **none** |
| Go source | **one hit only** — `spawn_actions.go:3053`, inside `isStorageEnabledAgent`. That is a **capability grant**, not a dispatcher: it says *if* this agent runs it may reach the bucket. |
| live scripts | none — the only script hits are under `scripts/initial_messages/earlier/` (archived) and READMEs |

And the artefact agrees. `llm_call_log` spans **2026-03-25 → 2026-08-25, 67,789 rows**, and carries
**zero** rows with `agent_type='visual-designer'`.

⚠ **I checked this the hard way, because `llm_call_log.agent_type` is the DISPATCH context and a
hand-fired run lands under `generic` — so an agent_type zero is not enough.** Keying on the step name
instead: `step_name='design'` has **20** rows, and they are **11 `generic`** (2026-07-17 → 07-24, the
hand-fired shape) and **9 `feature-designer`** (which has its own `design` step). **Nothing at all
since 2026-07-31.** Positive control in the same query: `step_name='run_offer_analysis'` returns 16
rows ending 2026-08-24, which is my own lane's run — so the key resolves.

**`brand-designer` and `experience-planner` are in the same position**: no scheduled task, not named
by any live agent's config. `design-audit-agent` and `offer-analyser` are named only by
`improvement-loop`, whose carrier is disabled.

**What actually produced homegarden's look**, from the site's own orchestration rows today
(21 distinct agent types): `site-design-planner` ×1, `webdesign-agent` ×1, `site-asset-renderer` ×1,
`render-audit-agent` ×1, plus `image-build-handler` ×13 and `image-generator` ×13. **Not one of the
eight agents §6 lists appears.**

> **So the honest sentence for him is: the visual designer did not do a bad job. It was never asked.
> It has no dispatch path at all, and has not made an LLM call under its own name in the five months
> the log covers.**

## 3. THE IMAGERY DEFECT IS REAL, IT IS STRUCTURAL, AND IT IS FLEET-WIDE — this is the actionable half

His imagery complaint stands on its own regardless of §2, and it has a precise mechanism. My first
reading was wrong and the correction is the finding.

**I first measured `page_components` containing `<img` for homegarden: ZERO of 45**, and was about to
report "13 images generated, none placed". **That was wrong.** The next query: **21 of 45 components
reference `/assets/`, and 20 carry `background-image`.** The 13 generated assets ARE placed.

**They are placed exclusively as CSS background images. There is no inline editorial image anywhere.**
That is exactly the gap he describes — *"much more imagery, placed BETWEEN PARAGRAPHS rather than only
at the top"* — stated as a mechanism: **the component vocabulary has decorative backgrounds and no
in-body image slot.**

`[MEASURED 2026-08-25, all sites with >10 live components, `build_status <> 'removed'`]`

| | |
|---|---|
| sites measured | **27** |
| sites with **ZERO** `<img>` in any component | **13** — including homegarden.uk (45 components), gaswholesalers.com (**115**), mortgagecalculator.co.uk (75), loanzy.uk (63) |
| every site has background images | yes, 2–56 per site |
| best inline coverage in the fleet | loancalculator.co.uk, **25 of 157 components (16%)** |

**gaswholesalers.com is the one to quote back**: 115 components, 24 with a background image, **0 with
an inline image**. This is not about a site being new or small.

**So his sentence — *"all this could be default for sites generally unless determined otherwise"* — is
asking us to change something that is currently near-absent BY DEFAULT, in 13 of 27 sites entirely.**
That is a component-vocabulary and planner question, not a critique of a designer agent's taste.

## 4. What this changes about the instruction

- **Do not "improve" the offer analyser's prompt in response to §6.** Its output is not what produced
  `about.html`; it has never seen the site. The lever is **reach** — 5 of 28 enrolled, 0 of 28
  scheduled — and that is an owner decision about switching carriers on, not a prompt edit.
- **Do not tune the visual designer either.** Tuning an unreachable agent produces nothing observable,
  and would read as work done. **The prior question is whether it should be reachable at all**, or
  whether its job now belongs to `site-design-planner` / `webdesign-agent`, which are the ones that
  actually ran.
- **The imagery ask is the one with a clear, cheap target**: an in-body image slot in the component
  vocabulary, and a planner that places it. It is measurable before and after by the census in §3,
  and it generalises exactly as he asked.
- ⚠ **§6's framing — "why did the existing one not fire, or not help" — resolves to "not fire" for
  five of the eight.** That is worth saying plainly rather than leaving the impression that eight
  agents looked at this site and produced it between them.

## 5. What I have NOT done, deliberately

**I have changed nothing live.** Specifically I have not enabled any `scheduled_tasks` row, not
enrolled homegarden in `offer_ordering`, and not touched the site. Enabling a disabled carrier is a
live fleet behaviour change affecting every site it selects, and this lane has already paid for
firing one dispatch on the belief that a promoter was off (`WRONG_CALLS.md`, 2026-08-25 — findings
were promoted and a live page rebuilt **31 seconds** later).

Those are owner decisions. This document is the measurement they need.

## 6. Re-run everything above

```sql
-- §1 reach
SELECT count(*) FROM site_specs WHERE aspect='offer_ordering';                       -- enrolled sites
SELECT name, target_agent_type, enabled, last_triggered_at FROM scheduled_tasks
 WHERE target_agent_type IN ('offer-analyser','improvement-loop');
-- §2 reachability (all three must be empty for the claim to hold)
SELECT count(*) FROM scheduled_tasks WHERE target_agent_type='visual-designer';
SELECT type FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false
   AND deleted_at IS NULL AND type<>'visual-designer' AND default_config::text LIKE '%visual-designer%';
SELECT agent_type, step_name, count(*), max(created_at) FROM llm_call_log
 WHERE step_name='design' GROUP BY 1,2;   -- control: step_name='run_offer_analysis' must be non-empty
-- §3 the imagery census
SELECT s.domain, count(*) AS components,
       count(*) FILTER (WHERE pc.rendered_html ILIKE '%<img%') AS with_img,
       count(*) FILTER (WHERE pc.rendered_html ILIKE '%background-image%') AS with_bgimg
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE COALESCE(pc.build_status,'pending') <> 'removed'
 GROUP BY 1 HAVING count(*) > 10 ORDER BY 3 DESC;
```

⚠ **Every count here is `as of 2026-08-25` and goes stale by ADDITION.** The reach figures move the
moment anyone enables a carrier; the imagery census moves on every build.

---

## 7. ADDENDUM after his decision — "switch on all those agents" has a RIGHT lever and a WRONG one, and they are not the same switch

**His decision, relayed 2026-08-25: *"Switch on all those agents and we'll need to fix or further
develop them as necessary."*** Before touching anything I measured what each disabled carrier would
actually do. **Doing this literally — enabling the four rows I named in §1 — would produce close to
the opposite of what he asked for.**

### 7a. The three `offer-analyser` carriers are SPENT ONE-SHOTS, and re-enabling them is the wrong action

`[MEASURED 2026-08-25]`

| name | `input_data` | `pre_query` | interval |
|---|---|---|---|
| `offer-analyser-oneshot-gaswholesalers-20260814` | `{"domain":"gaswholesalers.com","site_id":…}` | **NULL** | 300s |
| `offer-analyser-oneshot-leopardess-20260814` | `{"domain":"leopardessconsulting.co.uk",…}` | **NULL** | 300s |
| `offer-analyser-oneshot-webdesign-20260815` | `{"domain":"webdesign.co.uk",…}` | **NULL** | 300s |

Three things follow, and each is enough on its own:

1. **Each is HARD-PINNED to one site** via `input_data`. None of them can ever reach `homegarden.uk`
   — the site this entire review is about. Enabling all three does **nothing** for it.
2. **`pre_query` is NULL, so there is no selector and no stopping condition.** "One-shot" is a
   **naming convention, not a mechanism**: the only thing that ever stopped them was somebody
   setting `enabled=false` after the single run they wanted. Re-enabled, each fires **every 300
   seconds, indefinitely**, re-analysing a site that has already been analysed.
3. **One of them is leopardess**, which is holding **123** items at `needs_human_review`
   `[MEASURED 2026-08-25]`. Filing new findings there dispatches handlers at work another lane is
   deliberately holding — the trap already written into this lane's own handoff.

⚠ **AND THIS IS A NORM, NOT A QUIRK: `[MEASURED]` 37 disabled `scheduled_tasks` rows carry a
site-pinned `input_data`.** A disabled, site-pinned carrier in this estate is a **spent one-shot**,
not a switched-off feature. Read one as the second and you re-run history in a loop.

### 7b. The RIGHT lever is `improvement-sweep`, and it is a single switch

It is the only fleet-shaped carrier of the four, and its `pre_query` is a real selector:

- `sites` with `status IN ('active','deployed')`,
- **excluding** any site with a `claimed` build item (so it will not pile onto in-flight work),
- **excluding** any site with ≥50 `triaged`/`detected` build items (a backlog cap),
- `ORDER BY s.updated_at ASC NULLS FIRST LIMIT 1` — **one site per tick, least-recently-touched
  first**, at `interval_seconds=900`, `max_concurrent=2`, `concurrency_group='dispatch'`.

And it genuinely drives the agents in question — `improvement-loop` carries real
`spawn_agent` steps, **not mentions**: `spawn_offer_analyser` → `offer-analyser`,
`spawn_design_audit` → `design-audit-agent`.

**Blast radius, measured by running its own `pre_query` read-only with the `LIMIT` lifted:**

| | |
|---|---|
| eligible sites right now | **30** |
| full sweep at one site per 900s | **≈ 7.5 hours** |
| `homegarden.uk` eligible? | **YES** — so this is the switch that actually answers the review |
| `leopardessconsulting.co.uk` eligible? | **YES** — and it holds 123 `needs_human_review` items |

⚠ **What happens downstream is not optional and must be stated:** findings filed by this route are
promoted automatically. This lane measured `build-pipeline-trigger` → `build-dispatch-loop` promoting
them **31 seconds** later, rebuilding and deploying a live page (`WRONG_CALLS.md` 2026-08-25). So
enabling `improvement-sweep` is a decision to let live pages across 30 sites be rebuilt over the next
~7.5 hours. **That is consistent with his second clause — *"we'll need to fix or further develop them
as necessary"* is authorisation to surface defects — but it should be entered knowingly, not as a
side effect of a switch flip.**

### 7c. What I recommend, and what I have NOT done

**Recommended:** enable **`improvement-sweep` only**. Leave the three one-shots disabled — they are
spent, they cannot reach the site in question, and one of them would fire into another lane's held
queue. If leopardess should be spared, the cheap containment is to raise its held-item count above
the sweep's own cap or to exclude it explicitly, rather than to leave the whole sweep off.

**I have still changed nothing live.** Authorisation settles *may I*; it does not settle *what will
this touch*, and what it touches turned out to be different from what the instruction assumed. The
flip is one `UPDATE` and I am holding it for an explicit go-ahead, with these numbers on the table.

---

## 8. The loop's fail-open surface is NINE steps, not two — and my first count was wrong

The `loanzy_uk_example_site` lane flagged two fail-opens in `improvement-loop` before the switch-on.
**They are right, I initially contradicted them, and I was wrong.** Recording both halves, because the
correction is the useful part.

### 8a. My error: I read step-level `error_step` only

My first query took `st.value->>'error_step'` and reported `enrich_news_feed` /
`enrich_directory_features` as having **no** `error_step`, i.e. failing closed. **`enrich_news_feed`
carries it INSIDE `config`:**

```json
"config": { "site_id": "site_record.site_id", "error_step": "enrich_directory_features" },
"next_step": "enrich_directory_features"
```

**And the nested form IS honoured** — confirmed in the code rather than from the comment that says so:
`coordinator.go:3916` `routeToErrorStepOrFail` checks `step.ErrorStep` (3920) and **falls back to
`step.Config["error_step"]` (3924)**. `processor.go:456`'s comment records the history — omitting the
step-level twin once made every step-level declaration inert fleet-wide (`bugs_open/086`).

⚠ **So `error_step` has TWO valid spellings and a census of one of them under-reports.** Any query
that asks "which steps fail open" must read both, or it returns a confident, low, wrong number — as
mine did.

### 8b. The larger half they did not have: SEVEN of the eight `call_agent` seats fail open

Read at step level `[MEASURED 2026-08-25, live config]`, seven `call_agent` steps declare an
`error_step` **identical to their `next_step`** — so a seat that fails, errors or times out routes to
exactly where success routes:

| step | `next_step` = `error_step` |
|---|---|
| `call_site_review` | `spawn_offer_analyser` |
| **`call_offer_analyser`** | `spawn_brief_fidelity` |
| `call_brief_fidelity` | `record_audit_pass` |
| `call_design_audit` | `spawn_site_review` |
| `call_design_discovery` | `spawn_completeness_discovery` |
| `call_quality_discovery` | `spawn_design_discovery` |
| `call_dispatch` | `notify_scheduler` |

**The eighth is different and arguably worse:** `call_completeness_discovery` has
`next_step=spawn_design_audit` but `error_step=triage_findings` — so a completeness failure does not
merely continue, it **jumps forward past four seats** (design audit, site review, offer analyser,
brief fidelity) straight to triage.

**With 8a, the loop's fail-open surface is nine steps, not two.**

### 8c. ⚠ `call_offer_analyser`'s own description names the justification, and it is the exact claim `bugs_open/395` refutes

Verbatim from the live config:

> *"error_step continues the sweep — one auditor must not strand it, **and the child run is the record
> of whether it worked**"*

The first clause is a reasonable trade. **The second is not, and this lane filed a bug about that
shape today.** A child orchestration row records that the child was CALLED and reached a terminal
state — it is the same conflation `016b` §9 was corrected for this morning: *a terminal status records
that the HANDLER succeeded, never that the request's own criterion was met*. And it is worse here than
in the general case, because **`orchestration_states` terminal rows are reaped in ~24–48h**, so the
"record" that justifies the fail-open **evaporates before anyone reads it**. There is no durable row
saying the offer analyser failed on site X.

**This bears directly on the switch-on.** Enable `improvement-sweep` as things stand and the plausible
outcome is: 30 sites swept over ~7.5 hours, `call_offer_analyser` failing on some or all of them
(it has never run on 25 of the 28), each failure routing to the next seat, and the loop reaching
`record_audit_pass` — **a clean audit on a site nothing audited.** That is the false green of
`bugs_open/395`, one level up, at fleet scale.

### 8d. Their open question #3 is RESOLVED — it is not a third silent step

`call_site_review` targets role `site_reviewer`; the role is filled by `spawn_site_review`, whose
config is `{"role":"site_reviewer","agent_type":"site-review-agent"}`. **`site-review-agent` exists,
is active, and is one of the most-used LLM agents in the estate: 4,046 `llm_call_log` rows, last
2026-08-22** `[MEASURED 2026-08-25]`. A query for an agent *type* named `site_reviewer` finds nothing
because that is the ROLE name, not the type — the indirection is `spawn_agent`'s whole job.

### 8e. What this changes about the recommendation in §7c

§7c recommended enabling `improvement-sweep` alone. **That still holds as the right lever, but the
ordering now matters:** closing the fail-opens should come **before or with** the switch-on, not
after, because the failure mode is silent and the sweep is what makes it fleet-wide.

Minimum honest fix, config-only and in council scope: a seat whose call fails must leave a **durable
row** — not an orchestration state that is reaped, and not a pod log. The cheapest shape that matches
estate precedent is an `agent_error_log` entry per failed seat plus a marker on the audit pass, so
`record_audit_pass` can never assert a clean sweep over seats that did not run. **Whether the routing
itself should change (fail closed) is the larger question and belongs in the RFC that lane is
drafting — I am not taking it.**
