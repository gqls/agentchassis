# HANDOFF 2026-08-26 — the canary is MID-FLIGHT. Start here.

**COLD-START for the finetuning.uk lane.** Supersedes `HANDOFF_2026-08-25b_continue_here.md` (keep
it for the 398 contrast detail and the older trap set). Technical log: `NOTES_finetuning_uk_service.md`
08-25 c/d + 08-26 a/b/c. Owner prose: `README_where_we_are.md` (his document — append only).

## ⚠ THE ONE THING TO KNOW FIRST

**A nine-page rebuild is HALF DONE on the live site, and it regenerates copy.** All 9 images are
made; **2 of 9 pages have rebuilt their copy, 7 have not.** The owner is reading the site while this
is in flight, so a page he complains about may be pre-rebuild, mid-rebuild or post-rebuild — **check
`page_components.updated_at` before treating any verdict of his as being about the new copy.** That
exact confusion has already happened once today (08-26, §"what he was reading").

**HIS STANDING INSTRUCTION, verbatim:** *"yes, rewrite them and we can do forward only corrections,
so you don't have to restore them we can keep rebuilding through the system until they are
acceptable."*
⇒ **FORWARD-ONLY. Never restore from a baseline.** `baselines/2026-08-26_pre_hero_rebuild/` and
`baselines/2026-08-26b_pre_index_faq/` are **diff bases, not undos.**

## State, 2026-08-26 evening

| thing | state |
|---|---|
| Canary imagery | **9 of 9 complete**, 16 assets created |
| Canary copy rebuild | **2 of 9 done** (`model-approach-selector` 19:05, `tool-ai-readiness-checker` 19:13). **7 `needs_page` still `triaged`** |
| Those 2 pages | **FAIL** the pre-registration: 4 and 7 comparison constructions (copy lane scored 3 and 5 — the gap is whether quiz answers count; both fail the ≤2 line) |
| Images visible? | **NO on those 2** — `hero-tool` had no image branch. Fixed tonight, see below |
| Owner's comparison rule | **LIVE in the site brief** (migration `648`, 19:03:22Z) — and the two pages that rebuilt at 19:05/19:13 **read it and produced the shape anyway** |
| Copy machinery | 627/628/629 live 08-25 21:11Z, 630 08-26 00:22Z. Site brief de-demonstrated by `646`/`647` (rather-than 7→0, not-just 4→0) |
| Truncation trial | Implemented by the copy lane in the **gate's repair prompt** (council `82b800e1`, commit `5a46b6470`) — **INERT until the next chassis roll** |
| Datasets | **All six built.** Voice sets are 26/13/16 rows (small — the corpus is 6,595 words, see `datasets/PROVENANCE.md`) |
| Open owner question | **Rebuild `index` + `faq`?** He keeps landing on the homepage, which is NOT in the nine. Baseline taken (`baselines/2026-08-26b_pre_index_faq/`). **He has not answered.** |

## What tonight established, and it is the most important finding

**A written instruction did not govern form.** The owner's rule went live at **19:03:22**; the two
pages rebuilt at **19:05** and **19:13** — *after* it — and still produced *"not a verdict"*,
*"rather than a guarantee"*, *"instead of the one a vendor happens to sell"*.

**But the rule is RIGHT** — every failure truncates cleanly with no loss:

| before | after |
|---|---|
| treat the result as a starting point for a conversation, **not a verdict** | treat the result as a starting point for a conversation. |
| treat the result as a guide **rather than a guarantee** | treat the result as a guide. |
| which approach fits your business problem **instead of the one a vendor happens to sell** | which approach fits your business problem. |
| honest trade-offs for each, **not just the "winner"** | honest trade-offs for each. |

Most of them discharge ONE requirement — the tool-fallibility mandate (kept per the copy lane's
ruling 3). The writer reaches for a comparison to say "this can be wrong". **So the fix has to be
mechanical, not asked for** — which is exactly what the copy lane built into the gate, and why it
matters that it is inert until the roll.

## Tonight's fleet change — `649`, at his explicit instruction

*"please make the components image capable, it can update across the fleet, all forty pages and
more, there is a lack of images everywhere and this will help."*

`hero-tool` and `case-studies-hero` had **no image branch at all**, so imagery filed against them
completed `complete` and orphaned the asset (`bugs_open/412` §7 — my error, and the one-query
pre-check is recorded there). Migration `649` gives both the branch their five siblings have, copied
from `use-cases-hero` rather than invented.

**`[MEASURED 2026-08-26]` 48 live instances across 13+ sites. 47 gain a LATENT capability and change
nothing today. Exactly ONE changes on its next render: `leopardessconsulting.co.uk/case-studies.html`,
which already carries `/assets/images/hero-case-studies.jpg` and could not display it.**
⇒ **That page is another lane's and was deliberately NOT re-rendered here.** A CONTRIB is owed to
the `leopardess` lane. finetuning's own `case-studies.html` verifies the change when its rebuild lands.
**No fan-out was fired** — 47 of 48 have no image to show, so a fleet re-render would be pure churn.

## Next session, in order

1. **Watch the 7 remaining `needs_page` items.** When pages change: **read them yourself and send
   that read to `copy quality two stage` BEFORE looking at their battery output.** Their P5: a clean
   battery with a failed read is a FAIL. Pre-registration:
   `copy_quality_two_stage/AUDIT_prompts/CANARY_2026-08-26_finetuning_nine_page_rebuild.md`.
   ⚠ **I have already seen their scores for the first two pages, so my read of THOSE is
   contaminated. The remaining seven are clean — keep them that way.**
2. **Report per page against the split:** **P2a** — the seven non-tool pages have a fully cleared
   demonstration stack, so prediction is 0 and any hit implicates the model prior itself. **P2b** —
   the two tool pages carry a bounded ceiling of 1 from `hero-tool`'s anti-fabrication guidance.
3. **Answer his `index` + `faq` question** if he has. Evidence to give him: the two pages that
   rebuilt AFTER his rule still failed, so rebuilding his front door tonight is not supported by
   evidence — the mechanical truncation lands with the roll.
4. **After the next chassis roll:** `--color-cta-bg-ink` goes live (bugs_open/398) — probe the binary
   with a present- and absent-control, then one `template_changed` fan-out for the CTA buttons.
   The truncation gate also arrives; a **pre/post pair across the roll is useful data**, so do not
   hold the queue for it.
5. **CONTRIB owed:** `leopardess` lane (their case-studies page gains an image), and see below.
6. **Three datasets want more material** — his emails, his copy, real customer replies. 117 words of
   attributable prose existed before he supplied the corpus; that is why the voice sets are small.

## Traps current for this lane

- ⚠ **Check `page_components.updated_at` before believing any verdict is about new copy.** The
  owner's "the copy has regressed" on 08-26 was a stale-page read — every phrase he quoted came from
  copy written 24–25 Aug, on pages outside the nine.
- ⚠ **A component can accept imagery it cannot display.** `SELECT cc.name, (cc.html_template LIKE
  '%hero_url%') FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE
  pc.page_id='<id>'` — run it BEFORE filing imagery. `hero-tool`/`case-studies-hero` were the two;
  both fixed by `649`, but the class of defect is open (`bugs_open/412` §7 candidate 4).
- ⚠ **A verify block only refuses what you told it to check.** `646` passed while 3 `not just`
  remained; `648` pinned a count at 1 that was really 2. Assert the full class, not the headline.
- ⚠ **`content_direction.formatted` is DERIVED** (`FormatContentDirection`). Editing it alone is
  erased on the next spec write — change the source keys AND formatted together.
- ⚠ **A template edited by SQL ships NOTHING** without a `template_changed` fan-out (`bugs_open/283`
  §13). Three `COMPLETED` rerenders served the old bytes on 08-26.
- ⚠ **The header is NOT ordered by `nav_order`** — a fleet-wide name-tier table decides it
  (`bugs_open/407`). Older sets: `HANDOFF_2026-08-25b`, RUNBOOK §7–§9.

## Bugs filed by this lane, open

`398` (cta_bg as a colour — fix live, CTA half awaits the roll) · `407` (header membership) ·
`412` (image requires a copy rebuild; §7 the orphan-imagery addendum).
