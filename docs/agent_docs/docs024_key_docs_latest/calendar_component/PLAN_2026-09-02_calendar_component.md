# PLAN — calendar_component

Scope, decisions, and the open architectural question this lane inherits.

## 0. What "calendar" means on this platform, and what it does NOT mean

Established by research (2026-09-02) before this lane existed. Three distinct "calendar"
things live in the codebase and get confused with each other by name alone:

1. **`period-calendar`** — a generic PAGE COMPONENT (VIZ-017, migration
   `605_period_calendar_component.sql`). An ordered list of NAMED PERIODS
   (months/quarters/seasons/weeks), no dates, no numeric fields, a real `<ol>`. **This is
   what "calendar development" means on this platform, and this lane owns it.** Built as
   one of three sibling components (`checklist`, `comparison-table` — NOT this lane's
   scope) in `bugs_closed/381`.
2. **The `editorial_design_uplift` lane's Phase E "timeline"** — explicitly NOT this.
   Fact-fed, dated, cited real-world events. Boundary agreed 2026-08-25 (a year in a
   period label means you want theirs, not `period-calendar`). If a request needs dates
   or citations, route it to that lane, not here.
3. **"Calendar" as editorial subject matter** (`darts-calendar-density`, PDC tournament
   calendars on dartsonline.com) — an article topic, rendered through the unrelated
   `evidence-timeseries` chart component. Nothing to do with the `period-calendar` UI
   component. A grep for "calendar" will keep surfacing this — don't conflate it with (1).
4. **A "tool" page named for a calendar that has no calendar in it** —
   `boxingonline.com`'s `/tools/fight-calendar/index.html` (found 2026-09-02, see NOTES).
   Built through the tool-page pipeline (`page_type='tool'`), not `build-site-planner`'s
   component composition. Genuinely broken, genuinely this lane's business — see §4.

## 0a. Why (4) cannot be fixed by (1)

`period-calendar` refuses dates and numbers by design (605's rule 1 — a period is a
recurring NAME, never a date). A fight calendar is the opposite shape: one-off,
dated, real-world events. Placing `period-calendar` on `/tools/fight-calendar/` would
be wrong, not just insufficient. The fix is a different mechanism, not this one
reused — see §4.

## 1. Current state (inherited, verified fresh 2026-09-02)

- `period-calendar` is LIVE, council-approved (`c134b0e9`), registered VIZ-017.
- Adoption, queried live 2026-09-02: **2 placements** — `homegarden.uk` `/index.html`
  (landing, 2026-08-26) and `farmerinsurance.uk`
  `/blog/seasonal-farm-insurance-risks.html` (blog-post, 2026-08-31). The second
  placement is new and unrecorded anywhere before this session — see NOTES.
- Its two siblings remain at **zero** placements each (`checklist`, `comparison-table`),
  8+ days and multiple site builds after launch. Real, current, unexplained — not this
  lane's components, but a comparison point worth carrying forward.

## 2. The open item this lane inherits, unowned until now

`homegarden.uk`'s build (2026-08-25) surfaced something neither the `bugfix_381` lane
nor the `bugfix_206` lane claimed: told (informally, via the mission brief) to cover a
gardening year "month by month" with no calendar named explicitly, the planner satisfied
the promise at the SITE level — 17 pages, one per month plus a "this month" index, wired
into nav — rather than at the PAGE level with the `period-calendar` component built to
satisfy exactly that promise. Recorded by the `loanzy_uk_example_site` lane (its NOTES,
2026-08-25, "the interaction neither lane predicted") but explicitly not filed —
"belongs to neither bug".

This is a real design question for a calendar-development thread to own:

- Is a 17-page month-by-month structure a legitimate alternative to one page carrying
  `period-calendar`, or is it the planner reaching for a heavier, harder-to-maintain
  shape when a lighter component would satisfy the same promise more cheaply?
- Those 17 pages depend on `directory-build-handler`/`section-index` builder coverage
  (`bugs_open/206`, still open) — if that population no-ops, the site's whole calendar
  disappears. `period-calendar` has no such dependency: it renders inline off
  writer-authored content, nothing else has to build first.
- Nobody has decided whether the planner SHOULD be nudged toward the cheaper component
  when both would satisfy the promise, or whether both are legitimate and the choice is
  content-dependent (a full 17-page month guide vs. a compact `period-calendar` list).

**Next action for this lane:** read the current `build-site-planner` prompt's
`period-calendar` description (see RUNBOOK) to see whether it already nudges toward the
lighter component, before proposing anything.

## 3. Decisions and their reasons

None yet — this is a fresh thread. Log corrections/decisions here as they're made, not
silently.

## 4. The hollow `tool-fight-calendar` — this lane's second open item, opened 2026-09-02

`boxingonline.com`'s Fight Calendar tool page has two components (`hero-tool`,
`generic-text-block`) and zero fixture rows — no dates, venues, fighters or broadcasters
anywhere on the page. The hero's own copy calls it "Every upcoming boxing fight, listed
and dated"; the text block refers to "the calendar above" that does not exist. Full
measurement in NOTES.

**My opinion, formed for the correspondence with the `boxingonline.com` session** (who
had separately been discussing how the research agent, the news-editorial pipeline and
the raw news-ingestion agents relate to each other):

None of the three currently-live mechanisms actually owns "keep a list of upcoming dated
events current", and the gap is structural, not a misrouting between them:

- **`research-agent` / `evidence-researcher` / `content-researcher`** do ONE-TIME,
  point-in-time research, registered as static `evidence_base` facts with fixed
  citations (dartsonline's PDC season counts are the worked example — verified once,
  2026-08-20, never revisited). Right tool for "how many events happened per season
  historically". Wrong tool for "what's confirmed for next weekend" — the fact changes
  weekly and the mechanism has no revisit step.
- **`news-editorial` (the NEWS-020 feature pipeline)** consumes `content_feed_items` to
  write EDITORIAL FEATURE PROSE about a current story — analysis, not structured records.
  It extracts `topics` (free-form concepts), not typed fields like date/venue/fighters.
  Right tool for "why this fight matters". Wrong tool for a fixture list, because it has
  nowhere to put a date it could check.
- **The raw news-ingestion agents** (`feed-triage` etc.) tag incoming articles with
  topics, credibility and source-tier at real volume (9,622 of 10,855 items) — this is
  where a promoter's fight announcement would actually first arrive as data. But
  `content_feed_items.entity_ids` and `.duplicate_of` are declared and **written by
  nothing** (news_editorial_features PLAN §3, "the three zeros", 2026-08-19 — still true
  as far as this lane has checked). So even the news side has no path from "an article
  confirms a fight" to a structured event record.

**So the fixture list boxingonline needs is a fourth thing, not a routing fix between the
three.** It wants what `content_feed_items.entity_ids` was declared for and never given:
extracting a STRUCTURED event (date, venue, fighters, broadcaster) out of a confirmed-fight
news item, the way `derive_card_asset_action.go` extracts an image out of a hero rather
than generating one from nothing. Its natural home is closer to the news-ingestion side
than to `research-agent` (because dates get corrected as news breaks, not fixed once at
build time) — and its natural OUTPUT shape is closer to the `entity-directory` page role
the boxingonline strategy already asked for (one page per fight, with full details) than
to any content COMPONENT, `period-calendar` included.

**Not this lane's build** — no news-ingestion or entity-directory machinery is calendar
lane territory — but the DIAGNOSIS is, because "why is the thing named calendar empty"
is exactly this lane's question to answer. Contributed to `boxingonline.com` /
`site_delivery_and_editor` rather than filed solo; see whether they already have a
sharper account from their own discussion before this becomes a bug file.
