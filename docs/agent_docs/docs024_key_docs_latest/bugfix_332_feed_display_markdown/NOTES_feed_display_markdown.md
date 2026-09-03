# NOTES — bugs_open/332 feed display markdown

Append-only, newest at the bottom. The missteps are not an appendix — they are the point.

---

## 2026-09-03 ~16:00Z — picking the bug up

`scripts/who-owns.py 332` says OWNED-or-recently-active, pointing at
`site_delivery_and_editor` (119 commits/14d). Read the handoff before assuming that means a
live thread: §3 lists 332 as **unstaffed**, and `MEMORY_workstreams` records `bugfix_184`
closing with *"NOTHING OWED unless 332's trigger fires"*. It has fired (09-02 addendum). No
open `site_work_items` row touches it either — the only markdown-ish hits are `page_rerender`
items on a page literally named `tool-markdown-tables`, which is a different thing entirely.

So: unowned, resumed here. `who-owns.py` reads commits, so it cannot see a session mid-fix —
the handoff's own status line was the better instrument.

## 2026-09-03 ~16:10Z — the bug is still valid, fleet-wide, at the artefact

Five news pages serve literal markdown, each with a per-host 404 control:
boxingonline.ugg2.com 5 `](http`, fundamentallyai 3, robot-hands 2, ai-agent-orchestration 2,
idea.uk 2.

> **MISSTEP, caught immediately.** The first loop used one shared `page.tmp` for every host.
> `curl -o` only overwrites on success, so when fundamentallyai returned 000 the script
> reported idea.uk's counts under fundamentallyai's name — a *plausible* number, which is the
> dangerous kind. Redone with per-host filenames, `rm -f` before each fetch, and an explicit
> `NO MEASUREMENT (fetch failed)` branch. **The check that would have caught it is the one
> that makes a failure look different from a result.** → WRONG_CALLS.

## 2026-09-03 ~16:15Z — the shape of the defect, and the control that proves the strip works

Every one of the 14 live occurrences is an **unclosed** half-pattern. And the same pages carry
**zero** ATX headings while 1,177 feed rows have them.

That zero is the whole diagnosis. The strip is not missing and it is not disabled — it is
*blind*, because `MDLinkRe` requires a closing `)` and `MDBoldRe` a closing `**`. Without the
heading control I would have had "markdown reaches the page" and no way to tell a broken strip
from a blind one.

## 2026-09-03 ~16:25Z — answering the addendum's [UNVERIFIED]

The 09-02 addendum measured the *stored shape* (≤200 chars, trailing ellipsis, 54% unclosed)
and said plainly it had not read the writer. The writer is
`internal/adapters/websearch/providers/firecrawl.go:143-150` — `snippet[:197] + "..."`.

Ours. Hardcoded. No config key. So the sentence in 332's own fix candidate — *"strip BEFORE
truncate — a link cut mid-URL is a half-pattern nothing can match"* — describes a hazard we
create, one layer above where the file was looking for it.

Also: `[:197]` is a **byte** slice. 2 rows already carry U+FFFD.

## 2026-09-03 ~16:35Z — a Postgres limit that reads like a syntax error

A census failed with `ERROR: invalid regular expression: invalid repetition count(s)` and I
bisected eight regexes looking for a bad escape. The culprit was `{0,300}`: **Postgres caps
repetition counts at 255.** It is a limit, not a syntax error, and the message does not say so.
→ RUNBOOK, → WRONG_CALLS.

## 2026-09-03 ~16:45Z — the finding the bug file does not contain

`/data/news-archive.json` — public, 200, 20 items — carries 7 headings, 4 complete markdown
links, 5 truncated links, a list marker, an image and a bold marker. Completely unstripped.
`loadNewsItems` takes title and summary raw.

**The 7 headings are what makes this decisive.** Headings are the pattern the existing strip
handles perfectly. Server HTML: zero. JSON: seven. Same query, same rows, one reader
sanitising and one not. There is no reading of that pair except "the JSON path never learned".

## 2026-09-03 ~16:50Z — I got the severity of that wrong, in the safe-looking direction

I grepped the served page HTML for `news-archive.json`, got **0**, and wrote that the
client-side overwrite was *"latent, not live"*.

Wrong. The script is an external `<script src="/tools/assets/news-listing.js">`, so its code is
not in the page — a grep of the HTML can never find the fetch. Fetching the asset settled it in
one command: 200, 3,587 B, and it does `container.innerHTML = html` unconditionally on a
successful fetch (`hasServerRenderedItems` only guards the empty-feed and fetch-failed
branches). Live on all five news hosts.

**The lesson, and my own memory already held it** — *"a client-side absence is not an
absence"*. I read the index line and still made the mistake, because the grep returned a clean
zero and a clean zero feels like an answer. → WRONG_CALLS.

So the server-side strip is *cosmetic* for any JS-enabled visitor: the JSON wins. And it means
`sweep_site_defects.sh` §1.4, which greps served HTML, under-reports this defect **by
construction** — the same blind-check shape the bug is about.

## 2026-09-03 ~17:00Z — the number that removed a whole workstream from the plan

I was costing a repair campaign for the 9 baked-in components. Then:

```
robot-hands  news-listing  15:20:54  age 00:59
robot-hands  latest-news   15:19:45  age 01:00
… all 9 within 19 hours, three within the hour
```

The feed cycle rewrites those slots continuously and `queueNewsPageRerenders` re-resolves the
query each refresh. **A producer-side fix repairs every affected page on its own within about
a day** — no work items, no promoter-floor exposure, no campaign. That single query deleted
more planned work than any design decision did.

## 2026-09-03 ~17:30Z — two review passes disagreed, and the disagreement earned its cost

Ran a design pass and an adversarial pass. They disagreed on four points. Each was settled by
a measurement rather than by argument, and **three of the four went against my first design**:

1. **The truncated-link strip would have made the page prettier and less truthful.** My pattern
   ended `[^)\s]{0,200}$`, which swallows Firecrawl's trailing `...`. So
   `…and [lost in the ninth round](https://…-...` would have become *"…and lost in the ninth
   round"* — a grammatical, complete-looking sentence that is **not what the source said**,
   with nothing to signal a cut. `TestStripNeverInserts` only checks length, so no test would
   have caught it, and no served-artefact grep would either. **This is the single most
   valuable thing either pass produced.** Fix: re-emit the trailing `...`, with a dedicated
   test.
2. **The image detection pattern buys nothing.** `MDLinkRe` already fires on the inner
   `[alt](url)` of every letter-alt image. A new `md_image` name adds zero detection and can
   only perturb `transformRouteSlot`'s routing and the exact-pattern-set test. Dropped — strip
   order only.
3. **Unclosed bold and list markers do not belong in the shared primitive.** Confined to a
   feed-display tier that no detector reads.
4. **The adversarial pass attacked a projection I was not proposing** — it costed a full
   unification (escaping, ordering, attribution, dates). Mine leaves all of those with the
   caller. That one I did not concede; but its list of six divergences is now written into the
   plan as explicit non-goals, which is better than my having merely known them.

**Both passes independently said do not edit `firecrawl.go`** — irreversible, unreachable by
the kill switch, on the shared web+news path, makes `cmd/reasoningset`'s corpus bimodal, and
held by a live lane. The owner had chosen "display + ingest"; I am routing the ingest half to
its owner with the measurement attached and saying so plainly rather than quietly dropping it.

Also of note: the first two review agents were dispatched on Fable and both died on a usage
limit mid-run. Re-run on Opus. No result was lost because neither had produced one.

## 2026-09-03 ~17:45Z — Gate A, and why its control mattered

Gate A asks whether a new pattern would co-fire with `code_span` and revoke the section-editor
repair route. It returned **zero rows**.

A zero from a query nobody has proved can match is not evidence — so, in the same breath:
`SELECT count(*) … rendered_html ~ '`[A-Za-z0-9][^`]{0,80}`'` → **121**. The query
discriminates; the zero is real. The ported-page population it protects (`Ported Page
(webdesign.co.uk)` 115 components, `Ported Prose Block` 63) matched nothing.

## 2026-09-03 ~18:00Z — the components lane answered the JS question better than I asked it

I flagged the `innerHTML` defect to them and asked whether anything else in the library does
the same. Their census, quoted because it changes what I would have written:

**23 active components contain `innerHTML`. Only 2 are defects — mine.** The split is the
finding, and the raw total would have mis-sized the work by an order of magnitude:

- **2 DEFECT** — fetch JSON, accumulate with `html +=`, assign, no escaping: `news-listing`
  (13 accumulations), `latest-news` (6).
- **12 SAFE, and the best pattern in the library** — the directory/tracker family use
  `innerHTML` **only to clear** (`container.innerHTML = ""`), and route all data through
  `textContent` via an element helper. Zero concatenation.
- **1 SAFE, the fallback pattern** — `webdesign-couk-header` genuinely fetches and
  concatenates, but has a complete `esc()` covering `& < > " '` applied to *every* interpolated
  value. They checked specifically for a gap, since a partial escape helper is the usual shape
  of this bug. There isn't one.
- **8 SAFE, not data-backed** — calculators interpolating locally computed numbers, or user
  input through an `escHtml` built on `createTextNode`.

They also checked a surface I had not thought of: whether any `js_content` carries Go template
placeholders, which would bake DB text into the script body itself and be invisible to a
script-level review. One component has one (`contact-block`) and it has no `innerHTML` at all.

**Their recommendation, adopted:** copy the directory family — build elements and assign
`textContent`, keeping `innerHTML` for clearing only — rather than adding an escape helper.
It is already the majority pattern, it removes the class instead of filtering it, and it
cannot regress the way an escape helper does when someone later adds a fourth interpolated
field and forgets to wrap it.

**Their caution, adopted:** quote the split, never the total, or the next reader sizes the work
at 23. And state *why* the exposure number is low — the RSS ingest path calls `stripHTML`
(`feed_actions.go:248`) and the search path does not — so that a future non-zero reads as the
same bug arriving rather than as a new one. That asymmetry is the thing that can change
without anyone noticing, because a new ingest source inherits whichever path it is wired to.
