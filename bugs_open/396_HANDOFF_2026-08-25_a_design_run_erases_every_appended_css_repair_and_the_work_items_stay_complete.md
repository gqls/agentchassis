# 396 — a design run erases every appended CSS repair, and the work items stay `complete`

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
