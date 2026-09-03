# 473 — a stripped feed summary can still be the source site's NAVIGATION MENU, and no check can tell

**Filed** 2026-09-03 by the `bugs_open/332` lane, as the part of that bug's symptom which its
fix deliberately does **not** address. **Severity: LOW-MEDIUM — visible on a paid customer
site, but it is a content-quality defect, not a correctness one.**

## The mechanism, in plain terms first

A news item's summary is supposed to be a sentence or two about the article. On some items it
is instead a fragment of the source website's own **menu**, scraped in as though it were
article text.

Served on `https://boxingonline.ugg2.com/news/index.html`, 2026-09-03:

```
- Tennis (W)
- [NLL (Lacrosse)](https://www.espn.com/boxin...
```

```
- NFL
- [MLB](https://www.espn.com/boxing/story/_/id/49732339/...
```

`Tennis (W)`, `NLL (Lacrosse)`, `NFL` and `MLB` are ESPN's cross-sport navigation bar. They
are not about the boxing article they are attached to.

## Why `bugs_open/332`'s fix does not fix this, and must not be read as if it did

332 removes literal **markdown markers** from display text. Applied to the first example
above, it correctly yields:

```
Tennis (W)
NLL (Lacrosse)
```

Which is no longer broken-looking markup — and is still ESPN's menu, presented to a boxing
site's visitor as a summary. **A clean `](http` count on that page after 332 ships is
therefore not evidence this is fixed.** It is evidence of a different thing being fixed.

Worse, and worth stating plainly: stripping the `- ` markers **removes the visual signature**
that made this recognisable as scraped residue in the first place. The words now read as
deliberate editorial prose. That is a real cost of 332's fix, accepted knowingly, and it is
why this file exists rather than a note in a NOTES file.

## Where it comes from

`normalizeScrapeResults` **Strategy 2** (`feed_normalize_action.go:256-284`): when no usable
article links are extracted from a scraped page, the whole page's `markdown_content` is taken
as one item's summary.

`isNavigationLink` (`feed_normalize_action.go:294-320`) filters navigation out of the **links
array** on Strategy 1. Nothing filters navigation out of `markdown_content` on Strategy 2.

`[UNMEASURED]` how many live items came in this way. Strategy 2 is believed rare —
`source_type='scrape'` is 472 rows in 30 days with **empty** summaries [MEASURED 2026-09-03],
because Strategy 1 sets `summary: ""` — so the boxingonline instances may have arrived via the
`news_search` path with a provider snippet that was itself a scraped page. **Naming which
producer actually filed these is the first job**, and the census in `bugs_open/332`'s RUNBOOK
(source_type breakdown) is the shape to start from.

## Why it is NOT being fixed inside 332, stated so it is not re-litigated

Four reasons, and the first is the one that decides it:

1. **Its contract has no mechanical verifier.** 332's contract is *"no markdown marker
   characters reach a visitor"* — a regex over the served artefact settles it. This bug's
   contract is *"no meaningless summary reaches a visitor"*, and nothing can settle that
   automatically. Shipping an unverifiable transform inside a verifiable one is how the
   verifiable one stops being trusted.
2. **It is a different KIND of loss.** Stripping produces a character-subset of its input
   (`TestStripNeverInserts` pins that). Blanking a summary deletes the whole value, which is
   why the standing ruling calls blanking a *sanctioned remedy* for a specific case rather
   than a default, and why the council built the blank-manufacture guard.
3. **"This is nav, not prose" is a relevance judgement**, and relevance is explicitly another
   lane's seam (`feed_news_recommendation_action.go`, per 332's own §4). A markdown fix that
   quietly acquires a relevance heuristic is the scope creep the 2026-08-02 ruling exists to
   stop.
4. **Never LLM-rewrite a feed summary.** Standing ruling
   (`dartsonline_traffic/PREVENTION_2026-07-29`): *"authoring a summary for an article nobody
   read is fabrication."* Whatever the fix is, it is not "have a model write a better one".

## Fix candidates, ordered by what closes the door

1. **Filter navigation out of `markdown_content` on Strategy 2**, the way `isNavigationLink`
   already does for the links array. Closest to the cause, and the only one that stops the row
   being written. ⚠ The feed lane's own caution, given when they acknowledged this as theirs:
   a word-list would be *"a guess at every site's nav vocabulary"* — so this needs a
   structural signal (link density, repetition across items from the same host), not a
   vocabulary.
2. **A display-side quality floor**: after stripping, drop a summary that is only list
   fragments with no sentence. Cheap and safe to render — `{{if .summary}}…{{end}}` means an
   **empty summary renders nothing at all, cleanly**, so the card degrades to title + source
   with no gap. Recorded here because it is the enabling fact that makes this cheap to build
   later. But it treats the symptom and needs its own false-negative measurement (which
   legitimate summaries *are* short lists?) and its own kill switch.
3. **Do nothing and accept it.** Defensible while the volume is low — but the volume is
   `[UNMEASURED]`, which is candidate 1's first job.

## Owner

**The `news_feed_ingestion` lane**, who acknowledged it as theirs on 2026-09-03 and
deliberately did **not** take it that day, in their words because *"it needs a proper answer
rather than a quick filter"* and they would *"rather file it than half-fix it"*. Recorded so
nobody waits on it and nobody duplicates it.

## Relations

`bugs_open/332` §4 (which parks the relevance half) and its fix (which removes the markers and
leaves the words) · `feed_normalize_action.go` Strategy 2 · `isNavigationLink` ·
`feed_news_recommendation_action.go` (the relevance seam) · SITE_DEFECT_CATEGORIES §1.4, whose
own wording already names *"scraped navigation from the source site"* alongside the literal
markdown — the category saw this coming; only the markdown half had an owner.
