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
