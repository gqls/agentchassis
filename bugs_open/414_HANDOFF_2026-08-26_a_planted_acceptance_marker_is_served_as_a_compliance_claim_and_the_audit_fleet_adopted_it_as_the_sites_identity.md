# 414 — a planted acceptance marker is SERVED as a compliance claim, and the audit fleet adopted it as the site's identity

**Filed 2026-08-26 by the portfolio_positioning lane. Status: OPEN — spec source FIXED live;
served copy still carries the phrase; the repair path is stated below.**

## Symptom

`https://lendzy.co.uk/about.html` serves the sentence **"checked against the FCA handbook, rule
by rule"** twice, and `/guides/tool-affordability-complaint-checker-guide.html` once
[MEASURED 2026-08-26 19:0xZ, curl by body]. On a finance-adjacent site whose whole premise is
independence and accuracy, that is an unverifiable claim of regulatory diligence in the site's
own voice — nobody performed a rule-by-rule check.

## Mechanism — self-evidencing, every link verbatim

1. The 08-02 lendzy **shadow experiment** (this lane's `MISSION_2026-08-02_lendzy_shadow.md`)
   seeded `content_direction` with a **tripwire**: `positioning.acceptance_marker = "Somewhere in
   the site's written copy include the exact phrase: checked against the FCA handbook, rule by
   rule."` — duplicated as the tail line of the `formatted` field, which is what page generation
   reads (the spec row's own notes say `formatted` "carries it to page generation").
2. The lane's 08-05 handoff recorded the debt: *"marker strip owed BEFORE serving."* The site
   was then built and served by the fleet buildout without that entry ever being re-read —
   the memory-index landmine fired correctly tonight, 21 days late, exactly the
   `a-handoff-outlives-the-work-it-asked-for` shape.
3. The writer obeyed the instruction (it is an instruction, and it was followed — this is the
   `a-quoted-exemplar-in-a-prompt-is-copied-verbatim` class, deliberately induced): the phrase
   is stored in **3** components [2026-08-26] — `/about.html` `hero-about` +
   `content-block-about`, and the guide's `article-body` — in **`content_data`**, not just
   `rendered_html`.
4. **The new wrinkle, and why this file exists**: the maintenance/audit fleet then read the
   served phrase back and **canonised it**. An open `content_rewrite` item
   (`needs_human_review`) describes it as *"The site's core differentiator — FCA-rule-level
   accuracy checked guide by guide"* and asks for copy that leans INTO it. A tripwire did not
   just leak — the estate's improvement machinery adopted it as the site's identity and started
   generating work to reinforce it.

## Population — counted 2026-08-26

Fleet census of current `site_specs` for `acceptance_marker` / "exact phrase":
**1 site** carries a marker (lendzy.co.uk). apis.uk (`evidence_base`) and webdesign.co.uk
(`strategy`) match "exact phrase" innocently (a ban-pattern comment; descriptive prose).
lendzy's brief/strategy/submission also mandate a second exact phrase — *"know the rules before
you borrow"* — which is a benign brand slogan, functions as intended, and is NOT this defect.

## What is FIXED (live, DB config)

`content_direction` revised 2026-08-26 ~19:20Z: current row `81ddcc40-b1e2-426a-b4d2-ef68e949d1c8`
(`created_by='portfolio-positioning-2026-08-26'`) = the 08-02 row minus the
`positioning.acceptance_marker` key and the `formatted` tail line, applied server-side inside a
guard that asserted the exact tail before trimming. History preserved (`61ef7033…` superseded,
residue intact for audit). **Regeneration can no longer re-plant the phrase.**

## What REMAINS — and the trap in the obvious fix

- The 3 components still carry the phrase in `content_data`, and 2 pages serve it.
- ⚠ **The queued `page_rerender: Rerender page: about` (triaged) will NOT fix this** — a
  rerender regenerates from `content_data`, where the phrase lives. The repair is a **content
  rewrite** of the 3 components (framework, not hand-editing — owner ruling 08-06).
- The held `content_rewrite` item quoted above now has an **inverted premise** (the
  "differentiator" it wants amplified is a planted tripwire): whoever triages it should reject
  or rewrite it against this file, not release it as-is.

## Verify (after the copy repair)

```sql
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='lendzy.co.uk' AND (pc.content_data::text LIKE '%checked against the FCA handbook%'
   OR pc.rendered_html LIKE '%checked against the FCA handbook%');  -- expect 0
```
plus curl both pages by body (expect 0 matches; `rm` the temp file first).

## Why no 090 run (owner ruling 2026-07-31 escape hatch, stated plainly)

The causal chain is verbatim string identity at every hop — instruction in the spec, phrase in
the stored components, phrase in the served bodies — each independently queried/fetched this
session, and the fleet census bounds the population at one site. There is no inference in the
chain for a diagnosis loop to refute; the one judgement call (that the audit item canonised the
marker) quotes the item's own text.
