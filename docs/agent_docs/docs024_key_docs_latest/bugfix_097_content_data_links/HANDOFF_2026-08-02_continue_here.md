# HANDOFF — bugs_open/097, content_data link resolution — CONTINUE HERE

**Written 2026-08-02 ~18:55 UTC.** Single resume point. Read this top to bottom and
you can continue without re-reading the directory.

---

## State in one paragraph

The headline half of `bugs_open/097` is **built, council-APPROVED, committed and
LIVE on chassis `v1.0.1229`, pod-verified on both replicas.** What is NOT yet done
is the *behavioural* proof that the new pass fires on a real page — an induction
was dispatched at 18:47:48Z and had not been claimed by the dispatcher when this
was written. **The bug is still OPEN and must not be closed until that induction
(or an organic equivalent) is confirmed.**

## The one-line mechanism, so you can hold it without reading the bug file

A page has **three** copies of its links — the deployed string, the stored
`rendered_html`, and the `content_data` every re-render is rebuilt *from*. The
first two had resolvers. The third had none, so dead links regenerated on every
re-render, were silently re-repaired outbound, and were never reported. Every prior
mechanism (`ctaFieldNames`, `DeriveCTAURLFields`) answered "is this field a link"
by **enumerating field names**, so all of them were blind one level down — inside
an array's `items`, which is where 25 of the fleet's component types keep them.

The fix, `datahelpers.RepairContentDataLinks`: **nominate by field NAME at any
depth, judge by VALUE** with the shared `ClassifyLinkScope`. Runs in
`repairSectionsBeforePersist` off the same page index as the markup pass.

## Facts you can rely on (all measured 2026-08-02, do not re-derive)

| fact | value |
|---|---|
| production census, run with the SHIPPING function over all 885 `content_data` rows | **52 findings** — 19 rewrite, 33 phantom; 13 components, 7 domains |
| components with findings | `info-card-grid` 21ph+16rw · `case-studies-grid` 10ph · `tool-cta` 1ph+3rw · `platform-comparison` 1ph |
| components with NO findings | **872 of 885** (the control for the no-exclusion-list design) |
| url-named field values fleet-wide | 1,299 → 655 page / 457 external / 168 asset / 16 empty / 2 anchor / 1 mailto |
| page-scope values not starting with `/` | **0** (the false-positive surface is empty) |
| live agents persisting through `save_page_sections` | **6** (re-verified — see RUNBOOK R8, the obvious query returns 3) |
| resave cadence of the 11 affected pages | **2 to 23 days**; 0 of the 13 affected components locked |

## What is committed (all on `087_towards_multiple_domains`)

| commit | what |
|---|---|
| `d78f70bf1` | **the fix** — `content_data_links.go`, `save_sections_content_data_links.go`, wiring in `save_sections_link_repair.go`, 2 test files |
| `85a10b9cb` | the standing five + the council submission JSON |
| `3e0b22fc5` | concept register **LNK-028** + index count correction |
| `cd90982c2` | `WRONG_CALLS.md` — the two-guards-one-property misstep |
| `c6af6b03a` | `016b` §9 — the enumerates-field-names pattern |
| `b914b271b` | `LANDMINES.md` ×2 (synced to `doc_notes`) |
| `edbb6df9c`, `c0a5418da` | the bug file: what shipped, then the council verdict + a correction |
| `b8c8530bc`, `b71dd94e4` | two header notes (the map is not a copy; the reused predicate) |
| `32e30c6aa` | **council response** — objections checked, `Council-Reviewed:` trailer |
| `da676c8f2` | milestone SUMMARY |
| `42373058f` | the deploy verification |

Council: **`40c0c14d-636c-4d6f-b3a2-9316267d7367` — APPROVED round 1**, 12
reviewers, 4 advisory objections, none high, architecture signal `point_fix`.

## Deploy: PROVEN (this is done, do not redo it)

`v1.0.1229`, pods `agent-chassis-79479769b9-g7fbt` / `-n8nbj`, binary mtime
**2026-08-02 18:28:49 UTC** vs commit `d78f70bf1` at **10:45:30 UTC**.

```
audited content_data internal links before persist  1 / 1     (new)
CONTENT_DATA_LINK_AUDIT                             1 / 1     (new)
RepairContentDataLinks                              4 / 4     (new)
repaired dead internal links before persist         2 / 2     (POSITIVE control)
CONTENT_LINK_REPAIR_DETAIL                          1 / 1     (POSITIVE control)
CONTENT_DATA_LINK_INVENTED                          0 / 0     (negative — INVENTED, see caveat in NOTES)
```

## ⏩ THE ONLY THING LEFT TO CLOSE THIS: confirm the pass FIRES

**Induction in flight.** `site_work_items.id = ab409727-4dd3-48b0-8e7c-1a2e3682702d`
— a `page_rerender` on `gaswholesalers.com/supply-terms-and-eligibility`
(page_id `5bad27de-23c8-48fd-9f68-30d9ffe99a9b`), `reason='section_data_resolved'`,
inserted `triaged` with `handler_agent='page-rerender'` at 18:47:48Z.

**The prediction was written BEFORE dispatch, and it is the bar:**

- exactly **2 rewrites** — `cards[4].link_url` and `cards[5].link_url`, both
  `/contact` → `/contact.html`
- exactly **4 phantoms** — `cards[0..3]` = `/eligibility`, `/pricing`, `/delivery`,
  `/products`, and their values **UNCHANGED** in `content_data`
- one `agent_error_log` row, `error_code='CONTENT_DATA_LINK_AUDIT'`,
  `action='save_page_sections'`, component named as `info-card-grid`

**Pre-state hashes** (`md5(content_data::text)`, so you can prove which slots moved):

```
hero               2f773e4e0e362a0e34495beca4fcbc53
generic-text-block 1dae897b2f0b579b60183fb9a75d0685
info-card-grid     45215dac0aac739a68b181d881bb4345   <- ONLY this one should change
generic-text-block 5aeee5f74fb7239638610f3e1cec772e
faq                5f6b87f7f821ea3f7517058c2aad8cf9
call-to-action     fcee6d170924f0e0c1924bebc76bd4c6
```

### The three checks, in order

```sql
-- 1. did the run happen, and did it end healthy?
SELECT status, error, updated_at FROM site_work_items
WHERE id='ab409727-4dd3-48b0-8e7c-1a2e3682702d';

-- 2. did the audit fire?  (NOTE: the timestamp column is occurred_at, NOT created_at)
SELECT occurred_at, domain, action, context->>'rewritten', context->>'phantom',
       jsonb_pretty(context->'findings')
FROM agent_error_log WHERE error_code='CONTENT_DATA_LINK_AUDIT'
ORDER BY occurred_at DESC LIMIT 3;

-- 3. did the SOURCE actually change, in the right direction?
SELECT pc.slot_name, md5(pc.content_data::text) AS now_hash, e.idx-1 AS card,
       e.card->>'link_url' AS link_url
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id,
LATERAL jsonb_array_elements(pc.content_data->'cards') WITH ORDINALITY e(card, idx)
WHERE s.domain='gaswholesalers.com' AND p.name='supply-terms-and-eligibility' ORDER BY 3;
```

### If it stayed `triaged` and nothing ran

That is a dispatch problem, **not this bug** — `bugs_open/029` (hung spawns) and
`bugs_open/154`'s starvation notes. Do not conclude the fix is broken from a work
item that was never claimed. Re-check `handler_agent='page-rerender'` is set (an
item without it goes `blocked` with *"No handler_agent set"*), and note **no
orchestration dispatch lands within ~300s of a chassis pod restart**. Alternative
targets with known findings, all unlocked: `finetuning.uk/index` (5 phantoms,
`case-studies-grid`), `robot-hands.com/gripper-catalog` (1 rewrite),
`leopardessconsulting.co.uk/llm-cost-calculator` (1 rewrite).
**Do NOT use idea.uk** — one live session had 976 mentions of it in 90 minutes.

## Then: is 097 closable?

**My reading: the HEADLINE is dischargeable once the induction confirms, but the
FILE is not automatically closable**, and this is a judgement the next session
should make deliberately rather than inherit. 097 still carries four items that are
not mine and not fixed:

1. the single-component `content_data` writers of `bugs_open/136` (separate bug, has an owner);
2. a fleet dispatcher for `check_phantom_internal_links` (**owner call**, `083`/`033`);
3. the deploy-path completeness sub-question from the 07-28 round (never claimed);
4. whether the phantom arm should escalate rather than record — the same open
   question `link_repair.go` carries for its unlink arm; they should move together.

If you close it, move it with **both paths named on the commit** and verify at HEAD:
```
git add bugs_closed/097_*.md && git commit bugs_open/097_*.md bugs_closed/097_*.md -m "..."
git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 097   # must be exactly ONE line
```
(`git mv` + a pathspec commit silently ships a COPY — that is a live landmine.)

## Landmines this lane wrote (also in LANDMINES.md, synced to doc_notes)

1. **An offline link census MUST use `linkablePageStatusPredicate`**
   (`status NOT IN ('deleted','archived')`). robot-hands has `/learning-center.html`
   (active) **and** `/learning-center/index.html` (**archived**), and
   `NormalizePagePath` maps the second to `/learning-center` — include archived
   pages and a **correct** rewrite reads as a false positive.
2. **`CONTENT_LINK_REPAIR_DETAIL` no longer means "the link machinery saw this
   page".** A page whose only defect is in `content_data` produces no row under
   that code at all. Three codes, three questions.
3. **`agent_error_log`'s timestamp is `occurred_at`, not `created_at`.**
4. **"six agents persist through `save_page_sections`" cannot be measured with
   `WHERE s.value->>'action'='save_page_sections'`** — that returns 3. Three more
   reach it via a `loop` step. RUNBOOK R8 has both queries and the reconciliation.

## The two missteps, so you do not repeat them

- I wrote **two** mechanisms guaranteeing one ordering property, so deleting either
  left every test green and neither was load-bearing. Deleted the spare. After
  adding a guard, **delete it and watch a named test fail.**
- Two mutation runs printed `FAIL` for a **compile error**, which looks exactly
  like proof. `[build failed]` is not a red test.
