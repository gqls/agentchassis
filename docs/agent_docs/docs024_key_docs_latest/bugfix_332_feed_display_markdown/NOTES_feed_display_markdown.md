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

## 2026-09-03 ~18:00Z — the feed lane took the ingest half, and declined the contested one

I routed the firecrawl findings rather than editing their file. They came back having **fixed
the producer side themselves** (`6f0a246de`): rune-safe cut, and the cut now backs off a
genuine link opening so the producer stops manufacturing the shape my strip is blind to.

Three things from their reply worth keeping:

- **They verified my measurements first-hand before acting, and one differs**: 943 rows at
  exactly 200 bytes where I had 941. That is the feed running between two reads, and it is the
  direction that makes sense. 288 unclosed tails and 2 U+FFFD match exactly. A census that
  moves by 2 in an hour is the same census; one that moved by 200 would not be.
- **They declined `SafeCut` for a layering reason I had not considered and accept**:
  `datahelpers` imports `database/sql`, and nothing under `internal/adapters` does. Making a
  small adapter binary depend on `platform/orchestration` to borrow four lines is the wrong
  trade when `unicode/utf8` is stdlib and is the whole of `SafeCut`'s body.
- **They disagreed with my "the cheaper instrument for a budget problem is the budget"**, and
  they are right: *any* budget can land mid-link, so a config key fixes frequency while the
  boundary rule fixes the class. They also noted `web_search` has no `ActionInputSpec`
  registered at all, so an optional key there would land where the RFC_022 audit reports "NOT
  COUNTED, unknowable" rather than under budget.

They declined strip-before-cut for exactly the reasons I gave, having reached them
independently. Two lanes arriving at one seam from opposite ends and agreeing is the strongest
evidence available that the reasoning is not just mine.

## 2026-09-03 ~18:20Z — my own new sweep check was blind, and only the motivating case caught it

Added the `/data/*.json` arm to `sweep_site_defects.sh` §1.4, ran it on boxingonline, and it
reported **0 fields carrying markdown** on a file I had measured half an hour earlier as
carrying 7 headings and 9 links.

The JSON is `json.MarshalIndent` output, so every key is followed by a colon **and a space**.
My `'"(summary|title)":"[^"]*"'` could not match anything, on any file, ever.

**A green result from a new check reads as a fix working.** Had I not had an independent
measurement of the same file in front of me, "0 fields" would have been indistinguishable from
success — and the check would have gone into the fleet's acceptance sweep certifying every
site clean for ever. → WRONG_CALLS. The rule I now keep: **a new detector's first run belongs
on the case that caused it, with a positive control.** Post-fix: latest-news.json 5,
news-archive.json 10, both flagged.

The site_delivery lane also caught something mechanical I would have got wrong: use the
script's own `blind()` helper rather than printing the word. `blind()` increments `BLIND`, and
line 312 exits non-zero on `FINDINGS+BLIND` — so a hand-rolled print would have left the
script exiting 0 on a site with `rss_feed` off, reproducing this bug's own shape one layer
down.

## 2026-09-03 ~18:40Z — a guard test that failed for the wrong reason

`TestFeedBoldTailGuards` asserted `"in Python use **kwargs and **args here"` passes through
unchanged. It failed. The obvious response is to tighten the new bold-tail rule.

That would have been wrong. A four-line probe applying the two regexes separately showed the
strip came from the **pre-existing complete-bold pattern** — the text between the two `**`
pairs, `kwargs and `, is itself a valid `**…**` match. My new rule never fired. Tightening it
would have weakened a correct new pattern to fix an unrelated live one, and left the actual
residual untouched. → WRONG_CALLS.

The residual the rule genuinely has is now **stated in the test**: `"pass **kwargs to the
function"` does fire. Measured over all 5,112 live rows, it occurs in **none** of them, and
the rule changes exactly 7 rows — every one a genuinely truncated bold opening.

## 2026-09-03 ~19:00Z — council APPROVED, and two of the four objections were right

`803f0d81-02be-4bb6-9e65-363439ff87ba` — **approved, 4 advisory objections, none high**. I
acted on all four rather than banking the approval, because the code was already on the shared
branch.

**`reuse_agent`, MEDIUM, and the sharpest one:** *"why does `feed_normalize` call
`TruncateString` rather than the `FeedDisplaySummary` this plan just built? This is precisely
the question this seat exists to force."* Answered in the code, because it is a real design
boundary: the projection is gated on `DISABLE_NEWS_MARKDOWN_STRIP`, a **display** kill switch,
so calling it on an **ingest** path would put a display lever in charge of what gets written to
the database — the exact irreversibility I refused at firecrawl. Its premise was also wrong in
a useful way: `TruncateString` is not a third primitive, `data_helpers.go` defines it **as**
`SafeCut` plus the ellipsis.

**`bug_historian`, MEDIUM, and I had genuinely not done it:** *"the corpus validation is scoped
to `content_feed_items`; neither check enumerates and re-validates EVERY caller of the shared
function against its own corpus. This is the exact recurring shape where one call site of a
shared judgement gets the rigorous fix."* Fair — and pointed, given that is the shape this bug
is *about*. Done: **40,318** `content_data` string leaves from every unlocked page_component
fleet-wide, run through the widened strip and diffed against the verbatim pre-332 body.

```
0 blanked · 0 contract breaks · 15 values newly changed
```

All 15 are truncated markdown links inside stored news summaries — the defect being fixed. No
authored prose touched.

**Its LOW — "pull the precedent, don't cite it from memory" — changed what I could claim.** The
`render_guardian`'s actual words in `060bcc0a` are that *"a strip that reduces a field to empty
(e.g. a bare image-markdown token) could slip through as silently-blank content"*. The concern
was **emptiness**, not the stray `!`. So `!alt` → `alt` preserves exactly the property the seat
asked for. My argument was right; it is now grounded in the text rather than my recollection.

Reading it also surfaced a **second** low from that round I had not carried forward: no check
that a stripped result still clears the assembler's ≤10-visible-char section-drop floor. My
widening strips more, so it inherits that. Measured on the same 40,318: **4,697** values
already sit at or under 10 chars and **zero** are newly pushed under.

`tooling_provenance`'s low — query `doc_notes` for a news-feed subject before restructuring
three producers — returns only landmine-sync rows about the news-feed **seeder**, a different
seam. An absence, but now a measured one rather than an assumed one.

`editquality`'s "missing" note asked a human to confirm the deferred innerHTML fix is actually
tracked. It is: `bugs_open/472`, with the components lane's full library census.

## 2026-09-03 ~19:30Z — the 472 migration, and why it is held

Wrote `758_..._HOLD.sql` converting both news scripts to element construction. Rehearsed under
`BEGIN/ROLLBACK` on the live rows (2 UPDATEs, 1 row each, all three post-conditions true), then
**induced the verify to fail** by removing the second UPDATE — it aborted with *"1 component(s)
still concatenate markup"*, exit 3. So the guard bites rather than decorating.

**`_HOLD` rather than runner-queued, for a specific reason:** this rewrites DOM construction on
five live customer sites, and the usual verification **cannot see the result** — a static curl
reads the server HTML, which this very script replaces on load. Every check available to an
automated apply would pass on a page that renders empty. My own LANDMINES entry says so, which
is exactly the kind of moment an entry is for. Council `17a61f16-852d-47bd-947f-b0046e565abf`.

## 2026-09-03 ~20:00Z — the council APPROVED the version with the regression, and that is my fault rather than the gate's

Timeline, which is the point:

```
18:02:32  758 round 1 submitted
18:06:57  council_report — APPROVED
~18:30    components lane, on review, finds a LIVE REGRESSION in the approved version
18:13:54  round 2 submitted with the fix
```

The approved plan would have turned the "More insights" footer link into `href="#"` on every
site with a news index. Five seats read it and none saw it.

**That is not a failure of the gate, and it would be a comfortable mistake to record it as
one.** The council reviews a plan **against the evidence the plan supplies**. My submission's
sketch showed `safeHref(item.url)` and `safeHref(data.insights_url)` side by side — the defect
is visible there *in principle* — but nothing in my rationale said what shape `insights_url`
has. Without knowing it is filled from `pages.url` and is `/news.html` on every live site, the
two call sites look identical and the guard looks correct. **The gate cannot discover evidence
you did not give it.**

The components lane found it because they **ran the query**. They read
`render_news_section_action.go:213-218`, then went and looked at the live values. That is not a
better review process, it is a different one: seats reason over a submission, a peer reasons
over the system.

**What I actually owe from this, and it is a submission-writing rule rather than a code rule:**
when a plan applies ONE transform to TWO inputs, the submission must state the **shape and
provenance of each input**, not just name them. "item.url and data.insights_url" is a list.
"item.url is third-party and absolute; insights_url is internal and site-relative, from
pages.url, `/news.html` on every live site" is a submission a seat could have objected to.

Round 2 states it, and states this failure, in the rationale.

**The wider version, which is worth carrying past this lane:** an APPROVED verdict certifies
that the reviewers found no objection in what you showed them. It is not a second opinion on
the system. My round-1 risks block even named the wrong risk — I flagged "a typo in a class
name would silently unstyle the list" as the thing needing a reviewer, the class names were
clean, and the defect was elsewhere entirely.
