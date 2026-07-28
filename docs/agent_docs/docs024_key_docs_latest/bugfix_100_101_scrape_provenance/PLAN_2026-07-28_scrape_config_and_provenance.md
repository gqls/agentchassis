# PLAN 2026-07-28 — `bugs_open/100` + `bugs_open/101`: the scrape step lies about what it does, and cannot say where it went

**Thread:** "bugsearch", opened 2026-07-28 ~17:30 BST.
**Scope taken:** `bugs_open/100` (provenance read from the LLM) and `bugs_open/101`
(four inert `scrape_web` config keys), **as one bundle**. Both files independently
instruct this: *"bundle with `bugs_open/100`, same step, same roll"* (101 §Fix
candidates 2) and *"Bundle the `bugs_open/101` scrape-config fix into the same
round"* (100 §Fix candidates). They touch the same workflow step of the same agent.

## Why these two, and why now

Coverage checked before starting, three ways:

- `scripts/who-owns.py 100` → **no owning workstream**. `101` → names `vetcomparison`,
  which is the *blocked party*, not a fixer: its own memory line reads "P1: provenance
  STRUCTURALLY IMPOSSIBLE (100+101) ⇒ crawl restart BLOCKED on a Go change". Its last
  commit on this is 2026-07-27 and it has written no fix.
- `site_work_items` — **0 open rows** matching scrape/provenance/config-key/observation.
- No commits today (2026-07-28) touch either bug file.

Both re-grounded live before any code was written (see NOTES §1) — neither had moved.

## The finding that resizes the work

101 carries an `[UNSETTLED]` box marked **"READ THIS BEFORE IMPLEMENTING candidate 2 —
it may not be sufficient"**: does production's Firecrawl path strip nav/footers? If it
does, adding page fetches delivers nothing, because company numbers live in footers.

**It is now settled, and the answer is a second bug in shared code.**
`FirecrawlScrapingProvider.Scrape` (`internal/adapters/webscrape/providers/firecrawl.go:77-111`)
reads `only_main_content` into a bool and then adds the key to the payload **only when
it is true**. So `only_main_content: false` — an explicit request for the full page —
is dropped, and Firecrawl applies its own documented default of **`true`** (verified
against Firecrawl's `/scrape` API reference, not assumed: *"default: true … excludes
headers, navs, footers"*). The caller gets the exact opposite of what it asked for.

The `/crawl` path **in the same file** (line 338) gets it right — `if onlyMain, ok :=
config["only_main_content"].(bool); ok { … }` — so the two paths disagree and the
single-page one is wrong.

This is the same defect class as 101 itself (config that reads as live and is not), it
is in a **shared provider used by every scrape on the fleet**, and it is the gate on
101's candidate 2. It is fixed here.

## Decisions, and their reasons

**D1 — Fix the class, not just the four keys.** 101's candidate 1 ("reject unknown
config keys") is preferred by its own file because it stops the *next* instance. The
prior art agrees: `bugs_closed/042` (numeric config never read), `bugs_closed/127`
(`search_type` discarded), `bugs_closed/025` — this class recurs.

**D2 — but opt-in per action, because the survey says a fleet-wide allow-list is not
sizeable in one task.** Measured live: **1,155 steps carrying config, 228 distinct
actions, 811 distinct (action, key) pairs.** Declaring a complete, correct key list for
228 actions in one change would be a guess at scale, and a wrong allow-list that
*rejects* is far worse than the bug. So: an action opts in by declaring its config keys;
declared actions get checked, undeclared ones behave exactly as today. `scrape_web` is
the first declarer.

> This is deliberately "operators must remember X", which
> `order-fix-candidates-by-what-closes-the-door` warns is a defect in costume. Accepted
> knowingly, because the alternative is worse at this size, **and mitigated by making
> non-adoption visible**: the audit script reports which live (action, key) pairs are
> undeclared, so the gap is a number someone can drive down rather than an invisible
> absence. Same shape as the `098` unreviewed-commits report.

**D3 — Reuse `ActionInputSpec`, do not build a second registry.** 134 files already
register a spec via `datahelpers.RegisterActionInputSpec`. `GetActionInputSpec` has **no
callers** — the registry is populated and read by nothing but a parity test. Extending
it is strictly cheaper than a parallel mechanism, and it turns an inert registry into a
load-bearing one.

**D4 — Provenance is recorded by the fetcher and is never a model claim.** 100 §"Why the
obvious fix is WRONG" is explicit and this plan does not relitigate it: asking the LLM
for `source_url` makes provenance an assertion by the same call that produced the facts.
The fetched URL is threaded from the scrape result to the writer. The writer must also
**stop reading provenance out of the model object entirely**, or the old path silently
survives as a fallback.

**D5 — Honour the four keys against what the adapter genuinely supports.** Checked
downstream before promising anything: the adapter switches on `action` (`scrape`/`crawl`/
`map`/`extract`, `adapter.go:283`) and `Crawl`'s payload builder already reads `limit`
and `include_paths` (`firecrawl.go:280-305`). So `max_pages` → `limit` and `follow_links`
→ `include_paths` have a real home; they are not being re-advertised into a second void.

## Phasing

| # | change | file | live when |
|---|---|---|---|
| P1 | `only_main_content: false` becomes expressible | `providers/firecrawl.go` | web-scrape-adapter image |
| P2 | `scrape_web` honours its four keys | `actions/webscrape_actions.go` | chassis image |
| P3 | provenance from the fetch, not the model | `actions/business_intel_actions.go` | chassis image |
| P4 | declared config-key contract + unknown-key detection | `datahelpers/action_inputs.go`, `validation/workflow.go` | chassis image |

**Two images**, not one — P1 is the adapter, P2-P4 the chassis. 100/101 both said "one
roll"; that was written before P1 existed and is corrected here.

## How this gets verified

Per 101 and 100, **not** on a green run or a count going down:

- P1: the discriminating check is a *payload* assertion — a caller passing
  `only_main_content:false` must produce a payload **containing** `onlyMainContent:false`,
  not one omitting it. Unit-testable without Firecrawl; that is the point.
- P3: `source_url` non-empty **and** `raw_data ? 'source_url'` still **false** — the
  second column is what distinguishes "the fetcher recorded it" from "the model claimed
  it". A populated column alone proves nothing.
- P4: seed a definition carrying a bogus key on a declared action and watch it be
  reported. A silent accept is the bug restating itself.
- Spawned worker pods run `agent_definitions.image_tag`, **not** the deployed image
  (`bugs_open/066`) — any pod-grep must target the spawned pod.

## Status

- 2026-07-28 ~17:30 — opened, both bugs re-grounded, scope set.
