# BUG 403 — a hand-authored value inside a `source:"llm"` field has no provenance, so every content_rewrite is licensed to destroy it — and every existing guard is structurally blind to the loss

**Filed 2026-08-25** by the leopardess lane (`docs/leopardessconsulting/`), which has now had the
same page eaten three times. **Status: OPEN.** Owner ruling 2026-08-25 (recorded in
`docs/leopardessconsulting/README_where_we_are.md`, same date): **this lane builds the fix.**

## Why this is filed on first-hand verification, with the diagnosis loop corroborating in flight

Per the 2026-07-31 owner ruling, a structural claim needs the `090` loop or a stated equivalent
substitute. Both, here:

- A `090` run was fired before this file was written — intake correlation
  `2590b5b6-7b77-4dcb-964a-14cced9900ce`, **run correlation `c946b495-115d-4e3e-8186-3819273edb6c`**
  (artifacts are keyed on the run correlation). Its verdict should be read and recorded here before
  the fix is designed. ⚠ The trigger warned local HEAD was 87 commits ahead of
  `origin/087_towards_multiple_domains` at dispatch, so the loop cannot see the newest code; the
  mechanism cited below predates that window except `cta_provenance.go` (2026-08-22), which is
  supporting context, not the defect.
- The substitute itself: the damage is **traced at the row level, not inferred** —
  `page_component_history` names the exact generation that took the content, `site_work_items`
  names the driver from its own row, and the "no guard covers this" negative is grounded in each
  guard's own documented scope (below), each of which was **live when the loss happened**.

## Symptom

A page repaired or authored by hand — values written into `page_components.content_data` — loses
that work wholesale after a `content_rewrite` / `tone_shift` / full page build. The work item
completes green, the page renders and serves correctly, nothing is logged, and the loss is
routinely misdiagnosed as a recurrence of the closed regeneration bugs (`bugs_closed/238` family).
The tell that it is THIS bug: the lost values sat **inside a field the section spec declares
`source:"llm"`** — usually nested in an array (`cards`, `items`) — and the field itself never went
blank; it was regenerated with different content.

## Worked instance `[MEASURED 2026-08-25]`

`leopardessconsulting.co.uk` · site `4851f6fc-71cf-4160-a270-e03d6d3e0732` · page `/services.html`
(`ebc2c413-61e2-465e-b22b-9aab0167abc9`). Hand-restored on 2026-08-14
(`docs/leopardessconsulting/HANDOFF_2026-08-14_services_restore.md` — itself the third restore of
this page). The restore then survived **8 days and at least four rerenders** (generations 08-14
18:25, 08-16 10:47, 08-17 21:57, all byte-identical at 1,794 B for `info-card-grid`, 4,136 B with
six `icon-service-*` references for `teaser-reveal-panel`).

**One generation took all of it — 2026-08-22 11:35:41Z:**

| slot | before (archived 08-22 11:35) | after (same write, archived 08-24 18:36) |
|---|---|---|
| `info-card-grid.cards` | **6** cards | **3** cards, fully rewritten copy |
| `teaser-reveal-panel.items` | **6** items, **6** `icon-service-*.jpg` refs | **5** items, **0** icon refs |

The driver, named from its own row: `site_work_items` key
`offer-analysis_content_rewrite_services_4851f6fc-71cf-4160-a270-e03d6d3e0732`, `item_type
content_rewrite`, created 2026-08-22 11:17:21Z, complete 11:36:20Z. The write path was
`save_page_sections_overwrite` (visible in `page_component_history.source`).

Re-runnable evidence queries:

```sql
-- the generation trail (bytes + array length collapse across 11:35:41)
SELECT created_at, source, length(content_data::text) AS bytes,
       jsonb_array_length(content_data->'cards') AS n_cards
FROM page_component_history
WHERE page_id='ebc2c413-61e2-465e-b22b-9aab0167abc9' AND slot_name='info-card-grid'
  AND created_at > '2026-08-13' ORDER BY created_at;

-- where the icon references lived and when they stopped existing
SELECT slot_name, created_at,
       (length(content_data::text) - length(replace(content_data::text,'icon-service','')))
         / length('icon-service') AS icon_refs
FROM page_component_history
WHERE page_id='ebc2c413-61e2-465e-b22b-9aab0167abc9'
  AND content_data::text LIKE '%icon-service%' ORDER BY created_at DESC;

-- the driver
SELECT item_type, item_key, status, created_at, updated_at FROM site_work_items
WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND updated_at BETWEEN '2026-08-22 10:30' AND '2026-08-22 13:00';
```

All six `icon-service-*.jpg` assets still serve 200 as of 2026-08-25 — the page simply no longer
names them.

### Two further instances, found the same evening (same site, different pages, same query shape)

The 08-11 imagery session wired per-page heroes by merging a `background_image` key into
`hero-about` / `hero-contact` / `hero-services` `content_data` (the plan-row route is
schema-gated off for those components — RUNNING_NOTES ~1710) and verified all of them live.
`page_component_history` shows the wired keys being eaten `[MEASURED 2026-08-25]`:

- `about.hero-about` — `background_image` present in four consecutive generations
  (08-11 16:50 → 08-16 16:04), **absent from the generation written 2026-08-16 16:04** onward.
- `contact.hero-contact` — present 08-11 16:37, **absent from the generation written at that
  timestamp** onward.

Here the eaten key was not even inside an array — it was a top-level key the component's
schema does not declare, which a wholesale regeneration of the slot's `content_data` does not
re-emit. Same mechanism, second shape: **anything a hand writes into a slot the LLM owns —
nested value or extra key — is unrepresentable as "authored" and therefore fair game.**

## Root cause

Three facts compose, none of them individually a defect:

1. **The build/rewrite path REPLACES; only rerender MERGES.** `save_page_sections` regenerates
   every `source:"llm"` field wholesale (this asymmetry is documented and was ruled on in
   `RFC_042`; the REPLACE side is by design — the LLM owns llm fields).
2. **Nothing inside an llm field distinguishes an authored value from a generated one.** A hand
   repair writes into the same JSON the LLM writes into. The one exception proves the shape:
   `platform/orchestration/datahelpers/cta_provenance.go` (`__cta_minted`, LNK-035, shipped
   2026-08-22) marks machine-minted CTAs precisely so a later pass can tell them from authored
   ones — **for CTA fields only.**
3. **Site repair practice writes into llm fields**, because for most sections there is nowhere
   else — the content IS the llm field. Three sanctioned hand-restores of this one page did
   exactly that.

So an authored value inside an llm field is a loan against the next rewrite pass, and rewrite
passes are frequent: this page was hit by `offer-analysis` (08-22) and `cta_label_relevance`
(08-25 19:52Z) within four days.

## Why every existing guard misses it — each scope from the guard's own record

| guard | scope | why blind here |
|---|---|---|
| PBP-039/268 carry (`bugs_closed/238`) | **non-llm**-sourced fields whose source resolves nothing | these fields are `source:"llm"` — explicitly excluded (`llm fields never carried`) |
| `cmd/content-loss-check` (PBP-046) | field-level **non-empty→blank** transitions | the field never blanked; it changed. Array shrink with new content is a legitimate rewrite to it |
| 238's closure census | pairs consecutive generations **by field key** | same reason — `cards` exists on both sides; nested loss invisible |
| `__cta_minted` (LNK-035) | CTA label/url fields | the only provenance of this kind in the tree, and it is CTA-specific |
| "survives N rerenders" verification | rerender path | rerender MERGES — survival there says nothing about the REPLACE path (the 08-14 restore survived four rerenders and died at one rewrite) |

> **CORRECTED 2026-08-25, same session, hours after filing — "no guard covers this" was
> OVERSTATED, and reading the save action found what the filing missed.** `page_components`
> has `locked_by` / `lock_type` columns, and `save_page_sections` **already excludes locked
> rows from its DELETE and re-matches them into the new section set**
> (`loadActiveLockedRows` / `matchLockedRow`, `save_page_sections_action.go:1218`, predicate
> shared with the list side via `datahelpers.AgentWritableSQLFor` — bugs 058/285 lineage).
> `lock_type='permanent'` is LIVE on **51 rows across 7 lanes** as of 2026-08-25 — other
> lanes protect hand-authored rows with exactly this. **So the missing thing is narrower
> than filed:** (a) ROW-level protection exists and none of this site's three restores used
> it — a **discoverability** defect (no leopardess handoff, runbook or landmine names it —
> the register entry to check for is the lock helpers'); (b) **field-level** protection does
> not exist — a row lock stops ALL automated improvement of that slot (e.g. a locked
> `call-to-action` row would also freeze legitimate CTA-relevance passes), which is exactly
> the case the provenance candidate below still owes; (c) the detection blindness stands
> unchanged. What caught it: reading the whole action rather than the diff site —
> `editing-one-file-is-not-knowing-the-package`.

## Fix candidates, ordered by what closes the door

0. **Use the existing row lock — available TODAY, no code.** For a wholly hand-authored or
   hand-restored slot, set `lock_type='permanent'`, `locked_by='<lane>'` on the row.
   `save_page_sections` keeps it out of the DELETE and re-seats it. This is the correct
   immediate protection for the `/services.html` restore and the hero-only slots. Its cost
   is coarseness: nothing automated may improve the locked row again until a human unlocks.
1. **Generalise provenance (the field-level fix — makes the bad state unrepresentable).** An authored
   marker inside `content_data` (the inverse of `__cta_minted`: mark what a HUMAN wrote, e.g. an
   `__authored` key listing authored paths or values), honoured at plan/save time — the rewrite
   carries authored values into its output, or declines to regenerate them. Ships as **opt-in with
   the unsafe side default-OFF** (absent marker = today's behaviour), per the 2026-08-02 owner
   ruling on new authority at shared seams; consumers of `save_page_sections` must be named and
   told. Council-scope; register the seam in the same commit.
2. **Re-source the authored content** (available today, per-site labour, does not close the
   fleet door): move authored values to a non-llm source (`site_specs` aspect), where the
   PBP-039 carry and the aspect-authority mechanism already protect them. This is also the
   correct *site repair* for `/services.html` regardless of 1.
3. **Detection only** (weakest): teach `content-loss-check` an authored-loss census. Requires
   provenance anyway to avoid firing on every legitimate rewrite — so it is 1's little sibling,
   not an alternative.

## How to verify a fix

Behavioural, at the artefact: author a marked value inside an llm field on a test page, fire a
real `content_rewrite` through the real handler, and read the stored row + served page — the
authored value must survive while unmarked siblings regenerate. Then the mutation: remove the
marker and watch the same pass destroy it (a guard proven only by its green path is not proven —
see `mutate-the-code-to-prove-the-guard`).

## Related

- `bugs_closed/238` — the REPLACE-vs-MERGE asymmetry, closed for the non-llm field class; its
  closure explicitly says a recurrence is a NEW case. This is that case, one day after closure,
  in the class its fix deliberately excluded.
- `bugs_open/248` (CTA recompute clobbers authored contact links) — same disease, CTA organ;
  its fix line (`__cta_minted`) is the provenance pattern candidate 1 generalises.
- `RFC_042` — the asymmetry ruling (option c: watcher, not guard). Candidate 1 does not reopen
  it: the watcher answer was given for the *non-llm* loss classes measured at zero; the authored
  class was not in that population.
- 016b §9 entry added 2026-08-25 (same session); LANDMINES entry `hand-edit to a source:"llm"
  field` (same session).
