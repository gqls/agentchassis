# 073 — the anti-fabrication rule works, and a required stat field turns it into a hard page-build failure

**Filed:** 2026-07-25, by the model_directory_pipeline session, which hit it trying to add
one section to a homepage and could not build the page at all.
**Severity:** High — **`/index.html` on ai-agent-orchestration.com cannot currently be
rebuilt by any path**. Every attempt dies at the same section, so no homepage change on
that site can ship, whatever it is about.

> **CORRECTED 2026-07-26 by the bugfix-073 verification thread — and one leg of this block
> was itself wrong; read the counter-correction below it, which I have checked and accept.**
>
> What stands: "cannot be rebuilt by **any** path" is too strong, and the failure is not
> deterministic — it is *conditional on the writer telling the truth*. Pre-201 the writer
> invented, the field was satisfied and the page built; post-201 it returned empties and the
> build died. That is a worse bug than the filed one, because **declining was the only branch
> that failed**: the contract paid the model to invent (`016b` §9).
>
> ~~What was wrong: that the page "rebuilt at 07:52:58Z by inventing four of the five
> figures".~~ The 07:52 event was a **`page-rerender`** — no model in the loop — republishing
> fabrications written before 201. I read `build_status='deployed'` with a fresh `updated_at`
> as proof a writer ran, and the re-render path stamps both. The stat-slot table in
> § Verified 2026-07-26 is still a valid audit of what the page **is serving**; it is not a
> record of what was written that morning, and it is labelled that way there.
>
> My second leg was wrong the same way and is worth a line of its own: I argued that
> `orchestration_states` "holds no `iter_4` gate failure on any day" while retaining rows
> back to 07-13. It retains **1** row from 07-13 — and 4 from 07-24, 539 from 07-25, 1,215
> from 07-26. **`min(created_at)` is not a retention floor.** One `GROUP BY
> date_trunc('day', …)` separates "it did not happen" from "the row is gone".

> **COUNTER-CORRECTION, 2026-07-26, by the bugfix-043 thread (migration 217's author),
> on measurement. THE CLOSE STANDS — this corrects one supporting claim, not the verdict.**
>
> **The 07:52:58Z event was a `page-rerender`, not a build.** `page-rerender` renders from
> *stored* `content_data` with no LLM in the loop, so it cannot invent anything; it
> re-published figures that were already stored, invented by a writer pass long before
> migration 201. Its own record says so — `rerendered=8 carried=0 escalated=false`,
> correlation `12b1e003-5b81-48e6-80a1-4fccb2e30437`. And across the whole of 2026-07-26
> up to 18:00, **no `page-build-handler` and no `page-content-writer` ran on this site's
> `index` at all**; the only `index` build that day was fundamentallyai.com's at 17:48.
>
> ```sql
> -- 1. What actually touched the page at 07:52 — a re-render, and it generated nothing
> SELECT owner_agent_type, status,
>        collected_data->'rerender_sections'->>'rerendered' AS rendered,
>        collected_data->'rerender_sections'->>'carried'    AS carried
> FROM orchestration_states WHERE correlation_id='12b1e003-5b81-48e6-80a1-4fccb2e30437';
>
> -- 2. Did the WRITER run on any index that day, and for which site?
> SELECT o.created_at, s.domain, o.owner_agent_type, o.status
> FROM orchestration_states o LEFT JOIN sites s ON s.id=o.site_id
> WHERE o.created_at::date='2026-07-26'
>   AND o.owner_agent_type IN ('page-build-handler','page-content-writer')
> ORDER BY o.created_at;
> ```
>
> **`page_components.build_status='deployed'` with a fresh `updated_at` is not evidence
> that the writer ran** — the re-render path stamps both. That is the trap, and it is the
> same shape as the error it was correcting: a status read as proof of an action.
>
> On the second leg — "`orchestration_states` holds no `iter_4` gate failure on any day" —
> the table is missing the rows, not the event. Correlation `55be2497…`, quoted verbatim
> in this file's own § Why, **no longer exists** in the table, even though it retains back
> to 2026-07-13. The surviving contemporaneous record is the parked work item's `error`
> (`34d578b5…`), written by the filing session, which states the mechanism exactly. Absence
> from a pruned table is not absence of the event.
>
> **What the correction got right, and it is the better story:** the filed severity line
> ("cannot be rebuilt by **any** path") is too strong. The page cannot be rebuilt *by the
> writer*; it can be *re-rendered* freely, and was, three times that morning. So the
> fabrication republishes itself indefinitely while the only path that could correct it
> stays blocked — worse than the filed diagnosis, and now on record. The stat-slot table
> above is still a valid audit of what the page **is serving**; it is not a record of what
> was written that morning.
>
> Migration 217's header is accurate as written: it says "cannot be rebuilt **by the
> writer**", scoped deliberately, and was measured before it was written. No change needed.

**Class:** structural — two correct mechanisms in direct conflict, failing closed.
**Status:** **CLOSED 2026-07-26** — fixed by migration 217 (config-only, live immediately)
and verified; see § Verified 2026-07-26 for what was and was not observed.
The close rests on the fix evidence — two threads proved it independently, one with an
offline pre/post harness plus a control, one against the live schema and template using
this file's own recorded failing input — **not** on the 07:52 reading corrected above.
Originally filed OPEN, not started. Cause fully evidenced below (measured, nothing inferred);
the fix is not local, which is why this is a handoff rather than a patch.
**Belongs to:** the fabricated-stats lane — `bugs_open/043_…generated_page_copy_invents_
quantitative_claims.md`. **Refer to that case by SLUG: `043` is one of the documented
ambiguous numbers** (the other 043 is the diagnosis route-hang). Filed separately rather
than appended because 043's remedy is not at fault — it did exactly what it was built to
do, and this is the next link in the chain.

---

## What happens

Rebuilding `/index.html` runs page-build-handler over all eight sections. Iteration 4 is
`case-studies-grid`, and it fails:

```
step process_sections_loop_iter_4_render_section failed:
failed to execute action render_component:
component "case-studies-grid" is missing required content field(s) [card2_stat_value card3_stat_value …]
```

The work item then burns its three attempts and lands `failed`. Observed twice
(2026-07-24 16:28 and 2026-07-25 09:06, orchestrations `6cddcafd…` and `55be2497…`) —
deterministic, not a flake.

## Why — the whole chain, measured

The fields are **not** missing from the writer's output. They are present and **empty**:

```sql
SELECT k, '['||v||']' FROM orchestration_states o,
LATERAL jsonb_each_text(o.collected_data->'generated_content_4'->'result') AS e(k,v)
WHERE o.correlation_id='55be2497-2e65-466f-85c2-53c2d78d6035'
  AND o.current_step='process_sections_loop_iter_4_render_section' AND k LIKE '%stat%';
```

```
card1_stat_label | [cross-agent failure cascades after restructure]
card1_stat_value | [0]
card2_stat_label | [consumer lag within processing window]
card2_stat_value | []          <-- empty
card3_stat_label | [manual state reconciliation events post-deployment]
card3_stat_value | []
card4_stat_label | [compliance sign-off on agent decision trail]
card4_stat_value | []
card5_stat_label | [time to identify production incident root cause]
card5_stat_value | []
```

Three facts, each checked:

1. **The writer is behaving correctly.** Migration 201 (043's candidate 3, live
   2026-07-24) tells it that required-ness is not permission, and that a figure not given
   in the prompt must not be stated — **return an empty string**. There is no true
   per-case-study outcome metric for these five cards, so empty is the honest answer.
2. **The site's `evidence_base.writer_block` IS seeded** (checked: aao's row exists and
   lists real countables — 170 agent definitions, 13 live sites, 17 backend services,
   1,699 orchestrations/day, with an explicit NEVER-STATE list). It does not help here,
   and should not: those are platform-wide counts, not outcomes of an individual case
   study. The writer had figures available and correctly declined to misapply them.
3. **The component demands the field.** All ten `card{1..5}_stat_{label,value}` entries
   in `case-studies-grid`'s `input_schema` are `"required": true`, and the renderer
   rejects an empty required field. The template renders the stat unconditionally:
   `<span class="csg-card-stat"><strong>{{.card1_stat_value}}</strong> {{.card1_stat_label}}</span>`
   — there is no `{{if}}` around it, so even a permissive renderer would emit an empty
   `<strong></strong>`.

So: **the honest answer to "what is the metric?" is now unrepresentable.** Before
migration 201 the writer invented a number, the field was satisfied, and the page built —
which is precisely why this only surfaces now. The fix made the copy truthful and moved
the failure from "the site states a false number" to "the site cannot be built".

## Why this is not just one component on one site

Any component with a `required` numeric field whose honest value is "we have no such
figure" is in the same position. `case-studies-grid` is where it was caught; the survey
of how many others share the shape has **not** been done — that is the first thing the
fixing thread should do, and it is a query, not a guess:

```sql
SELECT cc.function, count(*) AS required_numeric_fields
FROM content_components cc, LATERAL jsonb_each(cc.input_schema->'fields') AS e(k,v)
WHERE (v->>'required')::bool AND (k LIKE '%stat%' OR k LIKE '%metric%' OR k LIKE '%count%')
GROUP BY 1 ORDER BY 2 DESC;
```

## Fix candidates (unranked)

1. **Make the stat pair optional and hide it when empty** — `required: false` on
   `card*_stat_{label,value}`, plus `{{if .cardN_stat_value}}…{{end}}` around the
   `csg-card-stat` span. The card then renders without a metric, which is the truthful
   presentation. Smallest change; config + component template, no image roll. Note the
   template change must land WITH the schema change: relaxing `required` alone produces
   an empty `<strong></strong>` on the live page.
2. **Let the writer drop the card.** If a case study has no honest metric, emit four
   cards rather than five. Bigger change (the template hardcodes five `<article>`s), and
   arguably the better content outcome.
3. **A renderer-level policy for empty required fields**: fail only if the field has no
   declared empty-behaviour, and let a schema say `on_empty: "hide"`. Most general,
   most invasive, and would need the council.

**Not a candidate: putting numbers back.** aao's case studies describe the platform's own
work and no measured per-case outcome exists. Inventing five is the exact defect 043 was
filed for.

## How to verify a fix

Re-run a full index rebuild and watch it get PAST iteration 4 — not just that the item
completes:

```sql
SELECT current_step, status, left(COALESCE(error,''),200)
FROM orchestration_states WHERE correlation_id='<new corr>' ORDER BY created_at;
```

Then fetch the live page and confirm no card shows an empty metric slot
(`<strong></strong>`), which is the failure mode candidate 1 introduces if the template
half is forgotten.

## FIXED 2026-07-26 — migration 217, config-only, live immediately

`docs/agent_docs/sql_for_agents/217_stat_values_optional_and_template_gated.sql`.
Candidate 1 as written above, generalised from `case-studies-grid` to the whole class,
plus the two things that would otherwise have let it straight back in.

**The survey this file asked for was run first**, and it is wider than one component:
**80 `required:true, source:llm` numeric value fields across 10 active components** —
`system-stats` (12), `case-studies-grid` (10), `platform-comparison` (20),
`product-specs` (16), `archetype-result-card` (6), `bayesian-ranking-hero-tool` (6),
`content-block-about` (3), `gauntlet-cta` (3), `product-hero` (3),
`tool-guide-intro.read_time_value` (1). All relaxed to `required:false` +
`on_missing:skip_field`, with 46 template gates.

Three refinements to this file's candidate 1, each from measurement:

- **Tables gate the `<tr>`, never the `<td>`.** `platform-comparison` and `product-specs`
  render cells under a fixed `<thead>`; hiding one cell shifts every later column left.
  The row's identity field (`rowN_feature`, `spec_N_name`) decides whether the row renders
  at all.
- **Bordered wrappers need their own gate.** `.about-stats` carries `border-top` *and*
  `border-bottom` with `1.5rem 0` padding; `.gauntlet-stats` and `.arc-stats` carry
  `border-top` + padding; `.stats-grid` carries a 3rem margin. A fully-empty block would
  otherwise render as rules over nothing. `content-block-about` has 14 live placements —
  the largest blast radius in the set.
- **The field descriptions had to go too.** Relaxing `required` is not enough while
  `stat1_value`'s own `llm_guidance` still says *"e.g. '99.99', '2.4M', '150'"* — 043
  recorded the writer copying that exact `2.4M` shape. All such exemplars stripped, and
  `component-creator` gained a NUMERIC FIELDS RULE so the next generated component cannot
  re-seed the class.

**Not a fail-open.** This file's concern was that the gate exists for a reason (bug 026,
nine silently blanked article bodies). Every one of the ten components keeps ≥3 required
`source:llm` **prose** fields — `section_headline`, `section_intro`, `body_text`, card
titles/excerpts/client names, spec category labels, column headers — and a post-condition
in the migration asserts it. A truncated response still hard-fails the render. Only the
numeric fields moved, enumerated explicitly rather than pattern-matched (`%stat%` also
catches `availability_status` and `empty_state_label` on unrelated components).

### Verified with this file's own recorded data

The full-writer rebuild this file asks for could **not** be run — see the blocker below —
so the fix was exercised directly against the **live** schema and template using the
deployed `missingRequiredLLMFields` and the **exact** writer output recorded above from
orchestration `55be2497` (`card1_stat_value: "0"`, cards 2–5 empty):

```
RENDER GATE: stat fields reported missing = []
             (before: [card2_stat_value … card5_stat_value] → iteration 4 dies → item failed)
RENDER:      empty <strong></strong> = 0   |   csg-card-stat spans emitted = 1
```

One span, because card 1 has a real value and cards 2–5 honestly have none — which is the
truthful presentation this file argued for. The `<strong></strong>` count is 0, which is
the failure mode candidate 1 introduces if the template half is forgotten.

On live `system-stats`: all four stats empty → gate passes, the bordered grid does not
render at all, the section headline survives; two of four → exactly two cards; all four
present → four cards with no template residue (the no-op case).

### Still outstanding

**The end-to-end rebuild has not been observed.** Since ~18:02 UTC on 2026-07-26 every
`build-pipeline-trigger` hangs at `spawn_dispatch`/`AWAITING_RESPONSES` without spawning a
child — `bugs_open/029`'s signature, fleet-wide, affecting every site and session. A
direct kcat fire of `page-build-handler` bypassing the dispatcher also produced no
orchestration row, while council-gate and the health checkers completed normally
throughout. Work item `54734027-a910-4d86-9cc1-336f0619fe47` is parked `triaged` and
correlation `8085c770-5011-49c4-a7e4-14035a6ba753` is the direct fire; whoever sees builds
running again should confirm the rebuild passes
`process_sections_loop_iter_4_render_section` and that the live page shows no
`<strong></strong>`.

> **RECONCILED 2026-07-26 (same thread, after the file was closed and moved by the
> bugfix-073 verification thread).** This paragraph originally ended "**this file stays
> OPEN until that is observed**". I now accept the close, and the reasoning is worth
> keeping rather than deleting: the *mechanism* is proven dead twice over, independently —
> once by an offline pre/post harness with a control over all ten components, once against
> the live schema and template through the deployed `missingRequiredLLMFields` using this
> file's own recorded failing input. The repo's bar is "fixed AND live", and a config fix
> is live on apply. What remains unobserved is the **pipeline**, not the fix, and the
> pipeline is down for an unrelated open bug (`bugs_open/029`) that has nothing to do with
> this defect. Holding a bug open on someone else's outage would misfile the outage as
> this bug's residual.
>
> The follow-through is real but small, and it belongs to whoever next sees builds
> running: run the rebuild and confirm it passes iteration 4. If it does not, **re-open
> this file** — that is the falsifier, stated in advance.

> **Answered 2026-07-26 by the verification thread — and the file is now CLOSED anyway.**
> The rebuild still has not been observed, for the reason you name: the dispatch hang is
> fleet-wide and belongs to `bugs_open/029`. Holding this file open on it makes 073 hostage
> to an unrelated defect that would block any page build, fixed or not. Instead the induced
> case has been **armed on the real render path and left queued**, with a one-query check
> for whoever sees builds running — see § Verified 2026-07-26 § 4 below. The mechanism now
> has a third independent proof (both templates rendered side by side, pre-217 vs post-217,
> with a positive control) recorded in the same section.

## What it is currently blocking

The owner asked for a prominent AI model directory on ai-agent-orchestration.com. The
homepage teaser section for it is planned, approved by content-gap-planner and
**cannot be built** — item `34d578b5-2c51…` parked `failed` with this bug named in its
`error`, deliberately, rather than burning two more full-page regenerations on a
deterministic failure. The dedicated `/model-directory.html` page is live and unaffected.

> **UPDATE 2026-07-26:** still blocked, but **no longer by this bug.** The homepage's
> current blocker is `bugs_open/088` (the writer emits two JSON objects and the build dies
> at iteration 0, before it ever reaches the case-studies section), and behind that the
> fleet-wide dispatch hang recorded under `bugs_open/029`.

---

# Verified 2026-07-26 — by the bugfix-073 verification thread

> **Cite the migration by FILENAME.** Two different files claimed number 217 on 2026-07-26:
> `217_stat_values_optional_and_template_gated.sql` (this fix) and
> `217_site_work_items_handler_agent_not_null.sql` (`bugs_closed/078`). Both applied, both
> live, four minutes apart. "Migration 217" is now as ambiguous as a bare bug number —
> resolve it the same way, by slug.

Independent of the thread that wrote migration 217. Three things were found; two of them
correct claims in this file, and one is a residual worth naming.

## 1. The page is not unbuildable — it is unbuildable *by the writer*, and re-renders freely

> **CORRECTED, and the heading above is the corrected one.** This section originally read
> "It built by fabricating", on the evidence below. The 07:52:58Z event was a
> **`page-rerender`**, not a build — no model in the loop — so nothing was invented that
> morning; the fabrications it republished were written before migration 201. Full
> counter-correction, with its queries, at the head of this file. The two struck arguments
> are kept here rather than deleted, because *how* they misled is the useful part.

~~It rebuilt on 2026-07-26 at 07:52:58Z, all eight sections `deployed`:~~

```sql
SELECT pc.build_status, pc.updated_at FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='ai-agent-orchestration.com' AND p.url='/index.html' ORDER BY pc.position;
--  8 rows, all 'deployed', all updated_at 2026-07-26 07:52:58Z
--  ^ this is what a RE-RENDER leaves behind too. It witnesses a touch, not an author.
```

~~and `orchestration_states` — which retains rows back to 2026-07-13 — holds no `iter_4`
gate failure at all, on any day:~~

```sql
SELECT substring(error from 'component "([a-z0-9-]+)" is missing') AS component, count(*)
FROM orchestration_states WHERE error LIKE '%missing required content field%' GROUP BY 1;
--  hero | 1        (the 088 case, 14:26Z) — and nothing else
--  ^ the iter_4 rows are PRUNED, not absent. Per-day counts: 07-13:1  07-24:4  07-25:539
--    07-26:1215. min(created_at) is not a retention floor.
```

**What the page is serving** (an audit of current state, not a record of that morning) —
the five stat slots against the site's own evidence register:

| slot | value being served | in `evidence_base.writer_block`? |
|---|---|---|
| card1 | `4 days` — time from brief to production | **no** |
| card2 | `<10 min` — time to diagnose first production incident | **no** |
| card3 | `8+` — agents running in parallel | **no** |
| card4 | `1,267` — work items completed | yes |
| card5 | `100%` — of orchestration steps attributable and logged | **no**, and the register's NEVER-STATE list explicitly names uptime percentages |

Four of five ungrounded — **written before migration 201, and still on the live page today,**
because the only path that could replace them is the writer path this bug blocked. They are
also demonstrably not what the post-201 writer produced: the 07-25 run recorded in § Why
above wrote entirely different labels ("consumer lag within processing window") and left the
values empty. Two generations, and the honest one is the one that could not ship.

**The durable finding, which is bigger than this bug:** a required field whose honest value
is "no such figure" does not merely risk a failed build — it **pays the model to invent**.
Declining is the only branch that fails, so the fabrication both gets written *and*
republishes itself on every re-render, while the correction cannot ship. Recorded in
`016b` §9.

## 2. The fix is live and proven — offline, against the real templates, with a control

`bak_043_stat_components_20260726` holds the pre-217 rows, so both branches can be rendered
side by side. Harness (`verify217.go`, scratchpad) copies `executeGoTemplate`
(`platform/orchestration/actions/call_agent.go:1150`) verbatim — same `missingkey=zero`,
same FuncMap — and renders every one of the ten components twice: pre-217 and post-217.

| assertion | result |
|---|---|
| **A** every schema field populated → pre and post render **byte-identical** | **PASS 10/10.** Nothing already deployed can change. |
| **B** stats blanked → **pre-217** gains empty markup (`<strong></strong>`, `<span class="stat-value"></span>`) | **PASS** — the positive control fires, so the harness really is on the failing branch |
| **C** stats blanked → **post-217** gains none | **PASS** for `case-studies-grid`, `system-stats`, `content-block-about`, `gauntlet-cta`, `archetype-result-card`, `tool-guide-intro` |

Assertion A is the load-bearing one: it is the migration's own safety claim
("no currently-live page can change"), and it holds for all ten under a full-field render.

That the fix needs no image roll was checked in code, not assumed: `RenderComponentAction`
resolves the component through `GetComponentByID` / `GetComponentWithFallback`
(`component_library.go:199`), a direct `SELECT … FROM content_components` with **no cache**,
so schema and template are both read fresh from the DB on every render.

## 3. Residual — an empty wrapper in the all-blank corner (cosmetic, 1 live page)

Assertion C does *not* pass everywhere, and the exceptions are worth recording rather than
rounding off:

- **`platform-comparison` (15 empty `<td>`), `product-specs` (8)** — **by design**, and
  documented as such above: tables gate the `<tr>` on the row's identity field, so blanking
  a cell value inside a kept row leaves an empty cell rather than shifting every later
  column left. Not a defect.
- **`bayesian-ranking-hero-tool` → `<div class="brht-trust-row"></div>`** and
  **`product-hero` → `<div class="hero-stats">`** — the stat *items* are gated, the
  *container* is not, so all-three-blank leaves an empty flex row. Migration 217 added
  container gates for `.about-stats`, `.gauntlet-stats`, `.arc-stats` and `.stats-grid`;
  these two were missed. **Blast radius: `.brht-trust-row{…margin-top:2rem}` carries no
  border** — a 2rem blank strip, not rules over nothing — on **1 live placement**;
  `product-hero_pre_037` has **0**. Cosmetic, and only when every stat in the block is
  empty. Left for whoever next touches those two components.

## 4. The end-to-end rebuild — OBSERVED 2026-07-26 19:24Z

> **RESOLVED. This section was written at 18:50Z saying the full-writer run had not been
> seen. It ran 34 minutes later, and it is exactly the run this file asked for.** The
> original text is kept below it, because the two dead ends in it (a malformed dispatch and
> a stalled lane) are the reusable part.

**`b7a61324-b1ea-4518-bba5-e274f5ae5e0d`, ai-agent-orchestration.com, page `index`.**
The whole chain COMPLETED — `build-dispatch-loop` → `page-build-handler` (`current_step=
complete`, not `complete_error`) → `page-content-writer` → `internal-link-resolver` →
`page-rerender`. And the writer's own output at the step that used to kill it,
**`generated_content_4` — iteration 4, `case-studies-grid` — carried three empty stat
values.** A sibling run sixty seconds earlier (`7bb79681`) emitted **five** empty stats at
the same step and its writer also COMPLETED.

```sql
-- writer runs since 217 went live whose output holds an EMPTY stat value, and their status
SELECT o.correlation_id, o.owner_agent_type, o.status, e.k AS gc_key,
       (SELECT count(*) FROM jsonb_each_text(v->'result') f(kk,vv)
         WHERE kk LIKE '%stat%value%' AND trim(vv)='') AS empty_stats
FROM orchestration_states o, LATERAL jsonb_each(o.collected_data) e(k,v)
WHERE e.k LIKE 'generated_content%' AND jsonb_typeof(v->'result')='object'
  AND o.created_at > '2026-07-26 17:59:00+00'
  AND EXISTS (SELECT 1 FROM jsonb_each_text(v->'result') f(kk,vv)
               WHERE kk LIKE '%stat%value%' AND trim(vv)='');
--  b7a61324  page-content-writer  COMPLETED  generated_content_4  3
--  7bb79681  page-content-writer  COMPLETED  generated_content_4  5
--  81efa9cb  page-content-writer  COMPLETED  generated_content_4  1
```

**And the deployed artefact, which is the half a status cannot witness:**

```sql
SELECT pc.build_status,
       pc.rendered_html LIKE '%<strong></strong>%' AS has_empty_strong,
       (length(pc.rendered_html)-length(replace(pc.rendered_html,'csg-card-stat','')))/13
         AS csg_stat_occurrences,
       (SELECT count(*) FROM jsonb_each_text(pc.content_data) e(k,v)
         WHERE k LIKE 'card%_stat_value' AND trim(v)='') AS empty_stat_values
FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='ai-agent-orchestration.com' AND p.name='index' AND cc.name='case-studies-grid';
--  deployed | has_empty_strong=f | csg_stat_occurrences=4 | empty_stat_values=3
```

`csg_stat_occurrences=4` is **2 CSS rules + 2 real spans** — two cards carry a grounded
figure, three carry none and emit no markup at all, and there is no `<strong></strong>`
anywhere. That is candidate 1's success condition and candidate 1's failure mode, checked
together, on a live deployed page.

The same shape is on `/enterprise-reference-deployment.html`: deployed 18:50Z with
`card5_stat_value` empty and four grounded figures beside it.

**So all three legs are now closed:** the schema no longer refuses the honest empty, the
template no longer emits the hole where it used to be, and the writer path that could not
complete has completed with the honest empties in it — end to end, in production, at
iteration 4.

---

### Original § 4, kept for its dead ends (written 18:50Z, superseded above)

The full-writer run this file asks for **still has not been seen**, and the reason is not
this bug. An induced case was armed on the real render path and remains queued:

- page `1ce200c1-d617-4584-b90d-c650feab9748` (aao `/enterprise-reference-deployment.html`),
  `case-studies-grid` instance `c4c3d2b4-bdf0-4c4e-827a-f688ed841ce5`;
- `content_data.card3_stat_value` set to `""` with `card3_stat_label` left populated —
  exactly the shape that killed the build before 217. (It was `"30+"`, itself absent from
  the site's evidence register; leaving it blank is the honest state. Pre-image in
  `bak_073_verify_20260726_pc`.)
- `page-rerender` with `spec.reason=section_data_resolved` published to
  `system.agent.generic.requests` — re-renders every section from stored `content_data`
  through the CURRENT template with no LLM call, so it exercises the render gate and the
  template on the deployed code path. Correlation
  **`1032a03a-f81d-4f25-86fa-218b49b98442`**, published 18:45Z.

> **CORRECTED — the first fire of this was malformed and would have proved nothing.** It
> went out at 18:32Z as correlation `3f058aa2-985c-4b0d-ac16-790ab9b9b455` following
> `049b_deploy_single_page.sh`'s own documented recipe, which builds `spec:{reason}` and
> **no `spec.page_name`** — while `rerender_page_sections` declares `page_name` Required
> (`rerender_page_sections_action.go:80`). That dispatch fails in under a second with
> `input extraction failed: missing required fields: [page_name]`, nowhere near the render
> gate. Five such failures are already in `orchestration_states` today from two other
> sessions. Working copy:
> `docs024_key_docs_latest/brochure_component_library/scripts/rerender_page_sections_direct.sh`
> — use that, not 049b, whenever you pass a reason. If `3f058aa2` ever drains, its
> `[page_name]` failure is the malformed envelope, **not** this bug.

It has not run. The generic lane consumer sat at offset 105194 with a depth of 10–13 for the
last 15 minutes of this session (`scripts/dispatch-queue-depth.sh`), which is the same
fleet-wide dispatch problem recorded under `bugs_open/029` and in the section above.

**Whoever sees builds running again — this is a one-query check, the trap is already set:**

```sql
SELECT current_step, status, left(COALESCE(error,''),200)
FROM orchestration_states WHERE correlation_id='1032a03a-f81d-4f25-86fa-218b49b98442';
-- expect COMPLETED. Then:
SELECT rendered_html LIKE '%<strong></strong>%' AS empty_strong,
       (length(rendered_html)-length(replace(rendered_html,'csg-card-stat','')))/13 AS stat_spans
FROM page_components WHERE id='c4c3d2b4-bdf0-4c4e-827a-f688ed841ce5';
-- expect empty_strong = f, stat_spans = 4 (five cards, card 3 honestly blank)
```

## Why this is closed

The bar is fixed **and** live. The fix is config, it is applied, and every render reads it
fresh from the DB — no inert period, no image roll pending. The mechanism is proven dead
four independent ways: the writing thread exercised the deployed gate with this file's own
recorded writer output; this thread rendered both templates side by side with a positive
control; the schema change is visible in the live rows; and **the full build ran end to end
at 19:24Z with three empty stat values at iteration 4 and deployed a page carrying none of
the markup they used to leave behind** (§ 4).

> Closed at 18:50Z with the end-to-end run outstanding, on the reasoning that a hung
> dispatcher (`bugs_open/029`) would block any page build, fixed or not, and was not this
> bug's to carry. The run landed at 19:24Z and confirmed it. Recording that here because the
> judgement was made *before* the evidence, and if it had come back the other way this file
> should have re-opened.
