# 346 — twelve sites still serve the pre-252 page head, and nothing is scheduled to rebuild them

**Filed 2026-08-21** by the `bugfix_252_og_lang_assembly` lane, on the owner's instruction, as the
residual of `bugs_closed/252` (og/lang slug — **not** the disk/scheduler 252). **Not a defect: a
damage-tracking item.** Nothing here is broken. Every mechanism works and is proven.

> **Read this first, so nobody re-diagnoses a fixed bug.** The defect in 252 is **dead**: an assembled
> page now states its own Open Graph identity, agreeing with its canonical, and declares its site's
> language. Proven at three artefacts on chassis `v1.0.1320`/`v1.0.1321` (symbols probed in the running
> binary on both replicas, with positive and negative controls). Every site's **stored** head is
> repaired — 22 of 24 carry a language, none bakes the homepage `og:url`, and the four duplicated-tag
> heads are clean.
>
> **What this file tracks is the difference between a repaired stored head and a repaired PAGE.** A page
> picks up the corrected head only when it next re-assembles. **The owner ruled on 2026-08-21 that we do
> NOT force those rebuilds** — the cost was ~500 `page_rerender` items into a queue whose drain half is
> `bugs_open/083`, which is other lanes' exposure too. That ruling stands and is not the question here.

## What is outstanding, measured 2026-08-21

Signal: a page carries the fix iff `GREATEST(deployed_at, last_built_at)` is later than
**`2026-08-20 17:30Z`**, when the first corrected chrome landed.

⚠ **Pin that timestamp — do NOT compare against `site_components.updated_at`.** That column moves every
time chrome re-renders for any reason, which silently RE-CLASSIFIES already-fixed pages as stale. It
cost me a reading that went backwards (252 "fixed" at 09:00, 217 an hour later, with no page having
regressed). The moving comparison answers "is this page's head current"; only the pinned one answers
"does this page carry the 252 fix".

| | |
|---|---|
| assembled pages fleet-wide | **727** |
| carrying the fix | **225** |
| **still serving the old head** | **502** |
| real sites at **zero** pages carried | **12** (+2 `*.internal` pool sites, excluded — they serve nobody) |
| pages on those twelve sites | **278** |

**The twelve, largest first:** finetuning.uk (49), loancalculator.co.uk (43),
leopardessconsulting.co.uk (40), mortgagecalculator.co.uk (30), idea.uk (24), loancash.co.uk (22),
lendzy.co.uk (20), loanzy.uk (14), noted.co.uk (12), oufe.com (11), webdesign.uk (8), cookly.uk (5).

**Natural rebuild rate, excluding this lane's own dispatches: ~1 page/hour fleet-wide, and bursty.**
So the active sites will drain in days and **these twelve realistically will not drain at all** —
remortgagecalculator.uk was on this list yesterday and has since healed, which is the mechanism
working, but it is a four-page site.

## RE-MEASURED 2026-08-21 evening — it is already draining

| | at filing (morning) | now |
|---|---|---|
| assembled pages | 727 | 727 |
| carrying the fix | 225 | **240** |
| still stale | 502 | **487** |
| **real sites at zero** | 12 | **10** |

**Two sites healed on their own within the day**, with nothing dispatched at them. That is the
owner's "let rebuilds carry it" ruling working, and it is the reason this file is a tick-list rather
than a work queue. It also sets the expectation honestly: the sites that heal are the ones being
worked on, and the quiet ones will still be here in a month.

**Also settled since filing:** every stored head is now repaired — 24 heads, 23 carrying a language
(the 24th is `loanandmortgagecalculator.co.uk`, `permanent`-locked hand-authored chrome, correctly
untouched), **zero** still baking a homepage `og:url`, **zero** blank duplicate `og:title`, and
**zero** unlocked heads missing brand tags. So nothing upstream of the pages is outstanding: every
page that re-assembles from here is correct.

## What a visitor actually sees on an unhealed page

- Sharing any inner page shows the **homepage's** URL, not the page's — beside a `rel="canonical"`
  that correctly names the page. The page contradicts itself.
- On four of the twelve (the `head-seo-standard` family: finetuning.uk, leopardessconsulting.co.uk —
  plus gaswholesalers.com and ai-agent-orchestration.com, both already healed) the page also carries
  **two `og:title` tags**, one blank.
- `<html lang="en">` rather than `en-GB`, which is what screen readers use to choose pronunciation.

Modest per page; the reason it is written down is that it is 278 pages with no end date.

## How this closes — deliberately cheap

**It heals for free the next time anyone rebuilds one of these sites.** There is nothing to fix and no
code to write. Any of these does it:

- a content or copy change on the site (the ordinary path);
- a chrome refresh: `rerender-pages` with `spec.refresh_site_components: true` for that site;
- a single page: `049b_deploy_single_page.sh <page_id> <site_id> <domain>` (assemble-only, no reason).

**⚠ Send the full dispatch envelope.** An orchestrate message missing these five Kafka headers is
accepted, recorded, and marked **COMPLETED having run nothing** — it cost the 252 lane a retracted
finding against two working agents:
```
-H action=orchestrate -H sender_agent_type=cli -H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses -H timestamp=$TS
```
⚠ And a `while read` loop with `kubectl run -i` inside exits after ONE iteration (`-i` eats the loop's
stdin) — write the payload to a file and redirect it. It reported success having done 1 of 20.

**So the suggested disposition is: no scheduled work. Whoever next touches one of these twelve sites
for any reason gets this for free — and can tick it off here.** Re-measure with the pinned query above
rather than assuming; the list only shrinks.

## Verify at the artefact, never at the status

```bash
curl -s "https://<domain>/<inner-page>" \
  | grep -oE '<html[^>]*>|<meta property="og:[^>]*>|<link rel="canonical"[^>]*>'
```
Healed looks like: exactly one `og:title` carrying the **page** title, `og:url` == the page URL == the
canonical, and `lang="en-GB"`.

⚠ **On a HOMEPAGE, `og:url` cannot tell you anything** — the bare `/` is correct both before and after.
Use `lang`, or test an inner page.

## Two results on this list that are NOT this bug

- **webdesign.uk vs webdesign.co.uk.** Only `webdesign.uk` (8 pages) is on this list.
  **`webdesign.co.uk` (117 pages) is a different case and will never gain `lang` from a rebuild** — its
  head component is a bare fragment with no `<head>` open tag to carry the attribute. That is
  `bugs_open/347`, being fixed separately.
- **Many pages will legitimately have no `og:description`.** Correct-or-absent by design; the column is
  empty on a large minority of pages (`bugs_open/320`, backfiller now scheduled). An absent tag there
  is the mechanism working.

## See also

- `bugs_closed/252_…og_tags_and_hardcodes_html_lang_en…` — the defect, its fix, and the full evidence.
- `docs/agent_docs/docs024_key_docs_latest/bugfix_252_og_lang_assembly/` — PLAN / RUNBOOK / NOTES /
  README / FINDINGS / DECISIONS. `DECISIONS_2026-08-21_owner.md` is where this item was decided.
- `docs026_concept_register/register/seo.md` **SEO-005** — the mechanism, and the standing threshold
  that a fifth one-producer head fix must raise an RFC on SEO-003 rather than take a fifth patch.
- `bugs_open/322` item 4 — the guard that made 252 possible is still open; a future per-page tag added
  to that block reproduces 252 exactly.
