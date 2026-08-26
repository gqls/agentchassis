# 396 — a design run erases every appended CSS repair, and the work items stay `complete`

> ⚠ **396 IS A DUPLICATE NUMBER — resolve by SLUG, and `git log` the FILE PATH, not the number.**
> The other 396 is `396_HANDOFF_2026-08-25_work_items_parked_at_deferred_with_a_named_handler_are_undispatchable_unrefilable_and_carry_no_provenance.md`
> (filed 13:49 BST, three hours before this one; unrelated mechanism). This file's own filing
> session checked the number and got "free" from a compound `ls bugs_open/N_* bugs_closed/N_*`,
> which exits non-zero when EITHER glob fails to match — the twin was in `bugs_open/` only, so the
> check inverted. The trap is in `WRONG_CALLS.md` 2026-08-26.

> ## STATUS 2026-08-25 — OPEN, LIVE, MEASURED ONCE, NOT DESIGNED
>
> **The one-line version:** `webdesign-agent`'s `persist_css_to_theme` writes the freshly-rendered
> stylesheet into `css_themes.css_content` **byte-for-byte**, which deletes every rule
> `css-patch-agent` has appended to that row since the last design run. The repairs vanish from the
> served site; their work items stay `complete`.
>
> **This is designed behaviour whose consequence was never drawn.** Migration 543 added the step
> deliberately and its own header says *"whichever agent ran last owns the file entirely"*. What
> nobody wrote down is that this makes **every appended repair expire at the site's next design
> run**.
>
> ⚠ **It also makes a live concept-register sentence FALSE** — see §3. That matters more than the
> bug: council seats and sessions read the register as ground truth.
>
> Split out of `bugs_open/390` (where it was found) rather than folded in: 390 is about a repair
> that never applies, this is about a repair that applies and is then deleted. Different cause,
> different fix, same symptom from the outside.

---

## 1. The mechanism

`css-patch-agent` repairs a contrast finding by **appending** a rule to `css_themes.css_content`
(migration 318) and git-committing the whole row over `assets/css/styles.css`.

`webdesign-agent` renders a site's CSS from the palette / layout / typography FK rows and, since
migration `543_webdesign_agent_persists_rendered_css_to_theme_row.sql`, **writes that render into
the same column byte-for-byte**. `render_css_from_spec` composes from the spec; it has never heard
of the patch agent, so the appended rules are simply not in what it produces.

543's four write guards (`octet_length >= 4096`, `origin <> 'seed'`, exactly one linking site,
`IS DISTINCT FROM`) all protect against *other* harms. **None of them notices that the column it is
about to overwrite contains repairs.**

## 2. The measured case [MEASURED 2026-08-25]

`agritec.uk`, theme `fa96f5ab-9d4c-40a6-b96b-f7baa8b63d22` (`origin='adopted'`, one linking site,
18,691 bytes — so all four of 543's guards pass and the write proceeds):

| when | what |
|---|---|
| 2026-08-24 20:54:02 → 20:55:14 | five `contrast_failure` items complete; each appends a rule and git-commits it (e.g. item `495317f1`, `a.bl-read-link { color: #E8EAF0; }`, commit `a327926e`, message *"CSS fix: contrast (theme v4)"*) |
| 2026-08-25 12:09:57 | `css_themes.updated_at` — the row is rewritten |
| 2026-08-25 12:10:29 | a `needs_design` item completes, handler `webdesign-agent`, `spec.reason = 'palette_changed'` (*"Owner instruction 2026-08-25: lighter palette, dark text"*) |

State now: the theme row contains **0** occurrences of `css-patch-agent`, and the served
`https://agritec.uk/assets/css/styles.css` (HTTP 200, 18,879 bytes; control
`/guides/does-not-exist-390.html` → 404) contains **no** `bl-read-link` rule at all. **All five
work items are still `complete`.**

## 3. ⚠ THE REGISTER SAYS THIS IS FIXED, AND IT IS NOT

`docs/agent_docs/docs026_concept_register/register/styling-render-pipeline.md:25` (STY entry for
the webdesign-agent flow), verbatim:

> *"(1) a per-site CSS repair used to expire at this flow's next run, roughly weekly, with no
> signal — **that is now fixed at source**"*

It is not fixed. 543 fixed the *opposite* direction — the patch agent clobbering a stale or empty
theme row with a fragment (`bugs_closed/198`). The direction in this file, the design run
overwriting appended repairs, is **exactly what 543 institutionalised**, because the step it adds
writes the render byte-for-byte over whatever the column held.

A dated correction has been added beneath that line rather than editing it away.

## 4. What is NOT established, stated because I nearly claimed it

I first counted three sites as victims — `dartsonline.com` (16 completed repairs, 0 markers left),
`lendzy.co.uk` (12, 0) and agritec (5, 0) — and labelled the first two `[INFERRED]`.

**That inference is not supported and I withdraw it.** Both have **zero** `needs_design` items of
any status, so there is no design run to blame:

```sql
SELECT s.domain,
       count(*) FILTER (WHERE w.item_type='contrast_failure' AND w.status='complete') AS repairs,
       count(*) FILTER (WHERE w.item_type='needs_design') AS design_items,
       (SELECT (length(ct.css_content)-length(replace(ct.css_content,'css-patch-agent','')))/15
          FROM style_collections sc JOIN css_themes ct ON ct.id=sc.css_theme_id
         WHERE sc.id = s.style_collection_id) AS markers_left
FROM sites s JOIN site_work_items w ON w.site_id = s.id GROUP BY s.id, s.domain, s.style_collection_id;
```

So their zero is **UNEXPLAINED, not refuted** — `needs_design` may not be webdesign-agent's only
trigger, or their theme row may have been re-linked or recreated. Whoever takes this bug should
settle that first: it is the difference between "one owner-instructed palette flip cost five
repairs" and "this quietly eats every repair on the fleet".

## 5. Blast radius — deliberately not asserted

**Unmeasured.** The honest sizing question is *"how many completed repairs were appended before
their site's most recent theme rewrite"*, and it cannot be answered from the current row: the
overwrite leaves no trace of what it removed, and `css_themes` keeps no history (`version` is
bumped by the patch agent's own append, and 543's write does not bump it). Candidate instruments:
the `agent_snapshots` rows each migration takes, the git history of `assets/css/styles.css` in the
`sites` repo (which DOES have every version), or the `updated_at` ordering above applied per site.
**The git history is the one that can actually answer it** — it holds each deployed stylesheet.

## 6. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make the render carry the repairs.** The durable repair surface is the one the renderer
   READS — the palette / spec rows — not the artefact it overwrites. A contrast repair expressed as
   a token or snippet change survives every subsequent render by construction. Largest change,
   collides with the palette lane, and is the same conclusion `bugs_open/390` §2 and `bugs_open/296`
   §10.5 reach from their own directions.
2. **Re-apply the appended tail after rendering.** Keep the patch block in a column of its own (or
   delimited by the existing `/* css-patch-agent … */` markers) and have `persist_css_to_theme`
   re-append it. Cheap and mechanical; keeps dead rules alive too, which is its own cost.
3. **Do not silently discard: re-open what the overwrite invalidated.** When the write removes
   patch markers, re-file or re-open the `contrast_failure` items whose repairs were in them, so the
   ledger stops asserting a repair that no longer exists. Does not fix anything, but it stops the
   lie and it is the smallest honest step.
4. **Refuse the write when it would drop repairs, and park.** Safest-looking, worst in practice: it
   would block legitimate owner-instructed palette changes behind a queue nobody drains
   (`bugs_open/033`).

## 6a. ⚠ CROSS-REFERENCE OWED TO MIGRATION 616 (council recommendation, 2026-08-25)

`bugs_open/390`'s commit 1 is **migration `616_css_patch_agent_prompt_stops_instructing_the_losing_move.sql`**,
applied 2026-08-25. It makes css-patch-agent's appended rule actually WIN the cascade.

The council's `bug_historian` seat, reviewing 616 (corr `ef5f9a0d`), approved it but noted — medium
severity — that **a fix which makes the appended rule win today has no value once
`persist_css_to_theme` next runs on that site**, and recommended that this filing name 616
explicitly *"so the two are not resolved independently and left disagreeing"*. That is recorded
here as instructed.

**The practical consequence for whoever fixes this bug:** 616 increases the value of fixing 396,
it does not reduce it. Before 616 an appended repair was usually inert anyway, so erasing it cost
little. After 616 the appended rule is expected to work — so from now on, erasure destroys repairs
that were actually holding. **Do not close 396 by arguing that the repairs were worthless.**

## 7. Provenance

- Found 2026-08-25 by the `bugfix_390_cascade_attribution` lane while measuring why contrast
  repairs do not stick; **not** a 390 arm.
- Lane working record: `docs/agent_docs/docs024_key_docs_latest/bugfix_390_cascade_attribution/`
  (`NOTES_cascade_attribution.md`, entry 2026-08-25 (d)).
- Related: `bugs_closed/198` (the opposite direction, which 543 fixed), `bugs_open/390` (a repair
  that never applies), `bugs_open/296` §10.5, register entries in
  `styling-render-pipeline.md` and `design-composition.md`.
- **No `090` run.** The mechanism is four dated live-row/served-artefact observations with their
  commands attached, and §4 records what I withdrew rather than what I confirmed. The
  blast-radius question in §5 is the part that would genuinely benefit from one.

---

## FIX CANDIDATE (contributed 2026-08-26 by the `finetuning_uk_service` lane, `bugs_open/398`) — a handler DECLARES whether its repair lands on a regenerated surface, and the promoter reads the declaration

Contributed at the invitation of the `loanzy_uk_example_site` lane (405), which judged this
"RIGHT and not mine to build into 405" because it is a property of the **handler–surface pair**
rather than of a finding's provenance — i.e. this file's scope. The erasure this bug documents
*is* that property, observed.

### The inversion that should lead any reading of it

`contrast_failure` → `css-patch-agent` `[MEASURED 2026-08-26]`: **492 rows, 307 completions,
62.4% success**, handler live, pipeline `design`. That clears **every one of
`detected-item-promoter`'s four doors** (`bugs_open/405` §1: pipeline ∈ (build, content, design) ·
handler live · pair has ≥1 lifetime completion · pair above a 25% floor).

**Its 307 completions are exactly what makes the pair look known-good — and by this bug's own
mechanism a large share of them completed over rules that `persist_css_to_theme` had already
erased, or was about to.** Competence history cannot distinguish "repaired it" from "wrote a rule
into a column the renderer overwrites and closed `complete`". So the door that is supposed to
gate on competence is being fed by the defect.

### The proposal

Add `repair_surface_regenerated` (or similar) as a **declared handler property** on the handler's
own `agent_definitions` row, and a **sixth promoter door** that refuses auto-promotion for a
handler that declares it. `css-patch-agent` is the first declarer.

**Why this shape rather than a hand-kept list:** the estate already has the mechanism.
`HandlerDeclaresOwnedPageRefusalSQL` (`platform/orchestration/actions/work_items_common.go:482-486`,
tested in `work_item_owned_page_door_test.go`) is precisely "a handler declares a property in its
own definition row, and SQL doors read the declaration" — one definition of the predicate, no
second list to drift. Named by the loanzy lane, which also names migration `629`'s anchored-replace
pattern as the template for adding the declaration (four anchors, rehearsed both ways).

### What it buys, and the honest limit

It makes "auto-dispatch a repair that cannot outlive the next render of its own artefact"
**unrepresentable**, rather than something each lane notices separately after the fact. It says
nothing about whether the finding was real — in the motivating case the finding is a genuine
browser-measured 1.00:1 white-on-white button, and holding it on *provenance* would be a door lying
about what it tests (405's own conclusion).

**It does not fix the erasure**, which is this bug. It stops the fleet spending on repairs the
erasure will undo, and stops those repairs' completions inflating the pair's apparent competence.

### ⚠ Two things read from `park_work_items` (migration `621`) that a reader here should know

Offered as the ready-made lever for the motivating case. Reading the function rather than the
description found two limits:

1. **`v_parkable := ARRAY['triaged','approved','detected']` — `deferred` is NOT parkable.** So rows
   already sitting at `deferred` (the 7 `contrast_failure` rows on finetuning.uk, untouched since
   2026-08-11) **cannot** be parked and given the provenance stamps they lack.
2. **The verb SETS `status = 'deferred'`** (`621:28`) — which the *other* `bugs_open/396` (the
   duplicate number; resolve by slug) documents as undispatchable, un-promotable and
   **un-re-filable**, because `deferred` is not terminal in `idx_swi_dedup`. So parking a
   `contrast_failure` key makes that key un-re-filable until it is explicitly unparked: **if the
   stated release condition is ever missed, the finding can never come back.** The stamps are a
   real improvement on an unattributed park; the state they are written into still carries that
   trap.

Nothing was parked or suppressed on the motivating case. A detector silenced to protect an
in-flight fix is the failure `WRONG_CALLS.md` exists for; the pair was flagged, the finding left
alone.
