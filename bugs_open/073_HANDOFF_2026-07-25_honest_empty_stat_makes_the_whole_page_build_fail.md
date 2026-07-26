# 073 — the anti-fabrication rule works, and a required stat field turns it into a hard page-build failure

**Filed:** 2026-07-25, by the model_directory_pipeline session, which hit it trying to add
one section to a homepage and could not build the page at all.
**Severity:** High — **`/index.html` on ai-agent-orchestration.com cannot currently be
rebuilt by any path**. Every attempt dies at the same section, so no homepage change on
that site can ship, whatever it is about.
**Class:** structural — two correct mechanisms in direct conflict, failing closed.
**Status:** OPEN, not started. Cause fully evidenced below (measured, nothing inferred);
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
`<strong></strong>`. **This file stays OPEN until that is observed** — the mechanism is
proven dead, the pipeline run is not.

## What it is currently blocking

The owner asked for a prominent AI model directory on ai-agent-orchestration.com. The
homepage teaser section for it is planned, approved by content-gap-planner and
**cannot be built** — item `34d578b5-2c51…` parked `failed` with this bug named in its
`error`, deliberately, rather than burning two more full-page regenerations on a
deterministic failure. The dedicated `/model-directory.html` page is live and unaffected.
