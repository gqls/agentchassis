# NOTES — calendar_component

Append-only. Newest at the bottom.

## 2026-09-02 — lane opened

Session renamed "calendar" by the owner; asked to research everything discussed about
calendars fleet-wide and take ownership of calendar development if no dedicated thread
already existed.

**Grepped fleet-wide** (docs, bugs_open/closed, concept register, Go code, memory).
No bug, no workstream directory, no concept-register entry is calendar-specific in the
sense of a dedicated lane. "Calendar" hits fell into three unrelated buckets (see PLAN
§0): the `period-calendar` component, the `editorial_design_uplift` lane's
explicitly-distinct fact-fed timeline, and "calendar" as editorial subject matter
(dartsonline PDC pages, using an unrelated chart component). Only the first is
calendar-shaped work in the UI/component sense.

**Traced `period-calendar`'s history.** Built 2026-08-24/25 as the second half of
`bugs_closed/381` (the "planner composes pages from components that cannot express the
page it planned" bug), alongside two unrelated siblings (`checklist`,
`comparison-table`). Closed 2026-08-25 after a false start — an earlier draft of the
closure claimed "the calendar SHIPPED" on `homegarden.uk` when that domain was still a
parked stub returning HTTP 200 for every path (`WRONG_CALLS.md`, 2026-08-25 entry). The
final closure re-verified with a proper invented-path 404 control before landing.

**Checked current adoption, live DB, 2026-09-02** — see RUNBOOK for the query.
**2 placements**, not the 1 the closing lane last recorded on 2026-08-25:
- `homegarden.uk` `/index.html` (landing), 2026-08-26 — the build that closed the bug.
- `farmerinsurance.uk` `/blog/seasonal-farm-insurance-risks.html` (blog-post),
  2026-08-31 — **new, nobody had written this down before now.** First adoption on a
  second site, and the first on a `blog-post` rather than a `landing` page.

`checklist` and `comparison-table` both confirmed at **0** placements (explicit LEFT
JOIN, not a silent GROUP BY omission — see RUNBOOK for why that distinction matters
here). Not this lane's components, but worth carrying forward as a comparison point:
whatever makes `period-calendar` get chosen and its siblings not is worth understanding
before assuming the pattern will hold as more sites build.

**Found the open architectural question.** The `loanzy_uk_example_site` lane recorded,
on 2026-08-25, that `homegarden.uk`'s planner satisfied a "month by month" brief at the
SITE level (17 separate month pages) rather than reaching for the PAGE-level
`period-calendar` component built for exactly that promise — and explicitly declined to
file it, since it belongs to neither the 381 nor the 206 bug. Nobody has picked it up
since (checked: no later mention anywhere in the fleet as of today, `grep -rn
"calendar-shaped"` returns only the three hits already accounted for). This is now this
lane's first open item — see PLAN §2.

**Not done yet:** read the current `build-site-planner` prompt text for
`period-calendar` to see whether it already nudges toward the lighter component, before
proposing any change.

## 2026-09-02 — a fourth "calendar", and this one is broken: boxingonline.com's `tool-fight-calendar`

Asked to correspond with the `boxingonline.com` session about a discussion they'd had on
the research-agent / news-editorial / news-agent relationship. Before writing to them,
checked what "calendar" means on their site — and found a fourth distinct calendar
concept this lane hadn't catalogued (PLAN §0 named three).

**`boxingonline.com` has a page `/tools/fight-calendar/index.html`**, `page_type='tool'`,
role `tool-fight-calendar` — a fourth kind, built through the tool-page pipeline, not
`build-site-planner`'s component composition and not the news-editorial feature pipeline.
`[MEASURED 2026-09-02]` its two components: `hero-tool` (badge "Fight calendar",
headline "Every upcoming boxing fight, listed and dated", CTA "See the full calendar")
and `generic-text-block` (2,000+ characters of prose beginning "The calendar above pulls
together the fights worth building your weekend around… date, venue, fighters and how to
watch, in one place"). **There is no third component. There is no calendar.** The page
promises a fixture list and delivers a hero banner plus an essay describing a fixture
list that isn't there — the exact shape of `bugs_closed/381`'s defect class, on a page
type that bug never touched.

## 2026-09-02 — `theme kits` proposed `page_archetypes` as a fit for PLAN §2; traced it and it isn't, but it found a real adjacent gap

`theme kits` (session, plan `please-think-hard-about-starry-locket.md` §4) is replacing
`apply_gap_plan_action.go`'s hardcoded `defaultSectionsForPage` switch with a fleet/theme/
site-scoped `page_archetypes` table, and asked whether it fits PLAN §2's open question
(one page with `period-calendar` vs. homegarden's 17 month pages).

**Traced before answering, rather than taking the framing at face value.** It doesn't fit:
`defaultSectionsForPage` only fires when the LLM's own section list comes back **empty** —
a late fallback. `reconcile_site_plan_action.go`, which actually minted the 17 month pages,
reads sections purely from `site_plan_sections` (the LLM's own plan output) and has no
Go-side default of its own. So the 17-page shape was the planning LLM's judgment, not
anything `page_archetypes` touches. Different question, different code path — confirmed by
reading `reconcile_site_plan_action.go` end to end (no `switch`/`case`/default-sections
function in it) before replying, not inferred from the file's name.

**But it does intersect something real, currently live, in the exact file being ported.**
`defaultSectionsForPage`'s switch has no case for `page_type='section-index'` — only
`news-index` and `entity-directory` — so any section-index page reaching this fallback
(LLM omits sections, or `ensure_page_section_layout_action.go` fires because the page has
no layout — `bugs_open/206`'s own residual class, and one of `theme kits`' three rewire
targets) falls to the bare default `hero, generic-text-block, call-to-action`. That is
`bugs_closed/381`'s exact prose-only degradation, on a path 381's fix (the LLM-facing menu,
migrations 591-595) never reached — this Go fallback still cannot produce `checklist`,
`period-calendar` or `comparison-table` for any page type, and their planned parity test
would faithfully preserve that gap rather than close it. Flagged to them as a two-line
addition worth making while the switch is already being rewritten; left the call to them.

Declined to seed a `period-calendar`-carrying archetype now — that would be designing ahead
of PLAN §2's still-open decision. Told them it's the right eventual home for that entry
once (if) the decision lands, not before.

**Resolved.** `theme kits` checked real section-index pages before picking a case rather
than guessing: homegarden's month pages mostly serve the bare `hero, generic-text-block`
fallback (confirming the gap was live, not hypothetical); a few with real listings add
`content-listing`. Landing on `section-index` → `["hero","content-listing"]` — a real
improvement over today's served pages, not just a parity port. Nothing further owed from
this lane; their fix does not touch PLAN §2's still-open page-count question.

> **UPDATE 2026-09-02, later still — staffed.** A fresh session ("feed lane") appeared,
> independently verified the 427 write-up and the ownership gap (read the bug in full,
> ran `who-owns.py` on 427/316), and accepted the charter proposed to them: the raw
> ingestion pipeline (`content-feed-orchestrator`, `find_news_sites`, `feed-triage`,
> `content-feed-refresh`, `content_feed_items` health) — explicitly not `period-calendar`,
> not `news_editorial_features`, not the `d6d350ec` entity-directory page-role diagnosis.
> Starting on 427 fix candidate #1 now, `bugs_open/316` next, opening
> `docs024_key_docs_latest/news_feed_ingestion/` for their own standing five. This lane's
> business in 427 is done — the fix is theirs to build, and the tool-page-named-calendar
> question that started this thread has an owner for its root cause.

> **UPDATE 2026-09-02, later.** Filed as `bugs_open/427`, jointly with `boxingonline.com`.
> They independently re-derived both this lane's `444`→`20` correction (identical query,
> identical result) AND went further with a distribution neither of us had stated cleanly:
> folding the 34 no-row sites into the 20-that-have-a-row bucket counts, **37 of 54 sites
> (69%) hold zero facts and 42 of 54 (78%) hold five or fewer** — checked independently
> here too, same buckets, same numbers. That is the stronger form of the argument: 63%
> with no row alone could be explained as "those sites are young"; 78% at five-or-fewer,
> concentrated in one outlier, cannot be. `boxingonline.com` is the typical case, not the
> anomaly. Full writeup in the bug file §3. Ownership of the FIX is still open.

This isn't a NAMING gap this lane can shrug off as unrelated (contrast PLAN §0.3's
`darts-calendar-density`, which is genuinely a different thing). This one is load-bearing:
the site's own `design_intent.layout_preference` explicitly asked for "a dedicated
calendar section with clear event rows… a proper fixture list, not a generic table"
(quoted in `site_delivery_and_editor/COMPARISON_2026-08-31_boxingonline…md`, line 176),
and the `strategy` spec separately asked for `entity-directory` — "one page per major
upcoming fight… fighters, date, venue, broadcast, undercard" (same doc, line 144-146).
Neither shipped. That COMPARISON doc already flagged the general shape (§4/§6: "the tools
make the reader supply the data") but did not single out that the calendar tool carries
literally zero fixture rows, nor trace why — filed as item 7 on a to-do list, not
diagnosed.

**Why `period-calendar` (this lane's own component) cannot fix this, and why that matters
for the correspondence below.** `period-calendar` was built refusing dates and numbers
BY DESIGN (605's rule 1: "NO NUMERIC FIELD, STRUCTURALLY… `label` is a period NAME, not a
date"). A fight calendar is the opposite of that shape — dated, one-off real-world events,
not a recurring named cycle. So the fix is not "place my component here"; it is a
different, currently-missing mechanism. See the message to `boxingonline.com` for the
architecture opinion this produced.

## 2026-09-04 — a fifth and sixth "calendar", both live-broken on a paid site: `bugs_open/482`

`boxingonline.com` relayed an owner report: the fight-countdown counts to a stale date, the
fighter-comparator ships empty. Verified before filing rather than trusting the report —
comparator confirmed exactly (0 `<option>`, 0 `<select>`, direct DB read), but the
countdown's specific claim ("2026-09-02, two days stale") didn't match the stored
artefact: the real shape is worse — six hardcoded fights, **all** dated `year: 2025`, none
matching the real Canelo/Mbilli fixture that genuinely exists in `evidence_base`.
`boxingonline.com` independently re-verified against the live served page (no cache/publish
divergence) and traced their own error precisely (a regex hit never confirmed as the actual
field). Filed as `bugs_open/482`.

**The process lesson, worth keeping.** Filed 482 without re-reading `bugs_open/427`'s full
current length first — it had grown to 1,600+ lines since this lane last read it in full
(the "check in on progress" request two turns ago), and §22-23, added the same day, cover
the identical mechanism on a different tool: `tool-generator` fabricated 12 fixtures for
the fight-calendar tool itself, caught before deploy, with a three-layer checker plan
designed and awaiting the owner's go-ahead. Caught the gap myself before it became a real
duplicate — but the honest record is that it should have been caught by reading first, not
by luck. **Re-read a fast-moving joint bug's CURRENT tail before filing anything adjacent
to it, every time, not just when picking it back up.**

**What the near-miss produced, though.** Cross-checking 482 against §23's plan found a real
gap in it: neither of the two content-side checkers as scoped would have caught 482's
violation — layer 2 looks for an ISO-date-STRING key, this tool's dates are `year`/`month`/
`day` numeric fields; layer 3 validates `data-fact-id` attributes, this tool has none at
all. Added as §24 in 427, cross-referenced both ways. A genuine contribution the near-miss
paid for, not just a process wobble.

This lane's calendar taxonomy (PLAN §0) is now five things, not four — `period-calendar`,
the editorial timeline (not ours), "calendar" as subject matter, the hollow fight-calendar
tool (`bugs_open/427`, now essentially fixed), and this: tools that actively fabricate
dated content rather than merely omitting it. Not updating PLAN §0's list itself — it's
already served its purpose (disambiguation for a fresh reader) and this is 482's story to
carry, not a new permanent category to maintain here.

**Both bugs staffed, this lane's involvement closed out.** A fresh session named "427"
resumed the mechanism/fence work; a fresh session named "482" took the fix. Confirmed to
both: nothing built on this side, no objection to either taking it. "482" ran the fleet
census this lane had flagged as missing (§6) and sharpened it into a better bug than the
one filed here — 335 active tool components fleet-wide, `year:`-keyed dates → **1** (only
mine), `data-fact-id` anywhere → **0**, so layer 3 isn't just blind on my case, it has
**zero fleet-wide subjects.** Bigger finding: an existing, working fabrication gate
(`check_tool_fabrication_action.go`, `bugs_open/020`) already covers one of the three write
paths (`tool-recreation-handler`) and simply isn't wired to the birth path that built these
two tools — `bugs_closed/021`'s "a durable write guard covers one path only" class,
recurring. That reframing is strictly better than "a fourth calendar tool is broken", and
I said so. Nothing left for this lane here; watch `bugs_open/427`/`482` directly for
outcomes rather than this file going forward.
