# HANDOFF 2026-08-24 — `garden-tools.uk` finished. 7 of 12 pages serve. What must be fixed before the next domain.

**Supersedes `HANDOFF_2026-08-23c_...` (build in flight) and `..._23b_...` (wrongly said the build was
dead — kept and bannered as a worked example).** The pre-run `HANDOFF_2026-08-23_garden_tools_continue_here.md`
is still the reference for the **pre-flight recipe** and the DNS/zone setup; nothing else in it.

## 1. Final state `[MEASURED 2026-08-24 09:05Z, at the served pages, cache-busted]`

Site `16784842-f7d8-4467-bb5b-eb1fb5c1caba`. Two submissions: the first (17:17Z) **died** at hop two;
the second (19:23Z) completed. **Do not read the first submission's dead items as current.**

| serving (7) | bytes | | never built (5) | why |
|---|---|---|---|---|
| `/index.html` | 66,395 | | `/buying-guides/index.html` | section-index — no sections ready |
| `/how-we-assess.html` | 67,362 | | `/brand-directory/index.html` | entity-directory — same |
| `/seasonal-planner.html` | 66,999 | | `/entities/brand-profile.html` | entity-page — same |
| `/about.html` | 65,188 | | `/blog/buying-guide-post.html` | blog-post — same |
| `/care.html` | 65,042 | | `/tools/finder/index.html` | tool — owner-gated to human review |
| `/contact.html` | 57,333 | working 2-input form | | |
| `/affiliate-disclosure.html` | 54,020 | | | |

**Dead links: 9 instances, 4 distinct targets.** Home page carries **3**.
`/tools/finder/` ×4 · `/buying-guides/` ×3 · `/brand-directory/` ×1 · `/blog/buying-guide-post.html` ×1.
`brand-profile` is unbuilt **and unlinked** — the orphan case, a separate milder defect.

**Open review items:** 4 × `needs_page` no-op, 1 × `owned_page_review` (tool-finder),
1 × `needs_section_data` (contact email), 8 × `unresolved_cta` (**at least one is stale — see §3**).

**⚠ The site is deliberately UNREPAIRED.** It is a clean measurement of the unaided route. Repairing
it ends that. If the owner wants a working garden-tools site, that is a different job — say so
explicitly and start a new lane doc, do not quietly fix this one.

## 2. WHAT MUST BE FIXED BEFORE THE NEXT DOMAIN BUILD — ordered by what stops a site being usable

### 1. `bugs_open/376` — a refused exemplar can kill the build outright. **UNOWNED. Highest priority.**
`vertical-exemplar-researcher` crawls three LLM-nominated exemplars; Firecrawl refuses some hosts
outright; the crawl steps have **no `on_error`**, so one refusal kills the orchestration and discards
the crawls that succeeded. `create_next_item` — the sole estate-wide producer of `needs_strategy` — is
the last step, so an exhausted 3-attempt budget is **terminal**. Submission 1 died exactly this way.
The refused host appeared in **4 of 5 draws**, so retry escapes sometimes and usually does not.
**Cheapest real fix:** `on_error` tolerance on each crawl step (N-of-3 is research, not a
transaction), with a stated floor. **The fix that pays:** persist refused hosts and exclude them at
selection — the only candidate that gets cheaper over time.
⚠ Verifying it: the refused crawl's step record reads `"success": true` (a dispatch receipt). Join on
`request_id`; verify at the artefact. In `LANDMINES.md`.

### 2. `bugs_open/206` — container pages have no builder. **OWNED, approved fix, mid-flight.**
Four of the five never-built pages share one error. `entity-directory`, `entity-page`, `blog-post`
and `section-index` are all roles whose content is *other content*; on a greenfield build that
content does not exist. **A third of the plan is undelivered and the front page breaks.**
The owning lane's approved fix **deliberately excludes `section-index`** (the guardian showed the
two-producer divergence was a silent no-op). I have contributed the measurement that qualifies its
cost sentence: leaving `section-index` parked is not "no worse off" — it is **3 dead links from 3
live pages including the home page**. Also flagged: `blog-post` may be a fourth uncounted type.
**Do not compete with that lane.** Contribute; they are six council rounds in.

### 3. `bugs_open/328` — an unbuilt page stays linked. **Confirmed live on a greenfield build.**
This is how 376 and 206 become *visible* damage rather than missing pages. Note the narrowing this
build supports: **only unbuilt pages that something links to cause harm** (3 of 5 here). The other
two are orphans — a plan promising pages nothing references, which costs a build slot and a review
item but no visible breakage. **A fix for one does nothing for the other.**

### 4. Status columns are unreliable in BOTH directions. **Not filed as new — see `bugs_open/315` family.**
On this one site: **5 `page_rerender` items `complete`** against pages that are `planned`, have NULL
`deployed_at`, and 404; **and** 8 `unresolved_cta` items still `needs_human_review`, at least one
provably stale (the CTA it names now renders and resolves to a live 200). The queue says "12
rerenders done, 8 CTAs broken"; the artefacts say 7 real and at least 6 fine. **The usual mitigation
— distrust the optimistic direction — does not work when the pessimistic direction is also wrong.**
Transferable form written up in `016b` §9.

### 5. Claims: a live commission relationship that does not exist. **Mild; route to claims gating.**
`/affiliate-disclosure.html`: *"Some of the links on this site earn us a small commission… When you
click through to a retailer, **Amazon among them**…"* Present tense, named company. There is no
Amazon Associates relationship and **no affiliate links on the site at all** — the guides that would
carry them never built. Same shape as loanzy's lender panel, far milder, squarely what `evidence_base`
gating is for.

### 6. `bugs_open/327` — the trigger exits 0 whether or not it published. **Procedural mitigation only.**
Unchanged since 2026-07-30. Always verify the `needs_domain_research` row; never the exit code.

**NOT blockers, recorded so nobody re-chases them:** `bugs_open/326` is **fixed for the build chain**
(migration 572; proven live here — a re-submission at 2h05m51s inside the old 3h brake queued work).
`bugs_open/337` did not fire. `bugs_open/311`'s guard was **NOT EXERCISED** — 0 scoped components, 0
diversions, because the only tool page was owner-gated. Report that as "the guard never ran", never
as "no collision occurred"; the clean result carries no information about it.

## 3. Things measured here that will save the next session time

- **Time-to-first-agent = queue depth ÷ ~90s** (24m52s busy, 81s quiet). `build-pipeline-trigger`
  picks ONE site per tick, `ORDER BY wi.created_at ASC, wi.priority ASC` — **FIFO by item age;
  priority only breaks ties inside one timestamp.** Recipe in the RUNBOOK.
- **A site is serialised to one in-flight item** (`NOT EXISTS … status='claimed'`).
- **The classifier is reproducible** — two independent runs, identical structured verdict including
  confidence to 2dp; only free-text `industry_tags` drifted. Fixture-safe for **structured fields
  only**.
- **The rerender DOES restore gated CTAs, and points them at live pages** (`about` → `/how-we-assess.html`,
  `care` → `/seasonal-planner.html`, both 200). Build order costs a page nothing permanent provided a
  rerender follows. The "build order silently degrades early pages" fear is **refuted**.
- **Re-pin the `311` md5 baselines yourself before any run.** All eight moved 2026-08-20 under
  `bugs_open/283`. A handed-down pin is not a baseline. LANDMINES.
- **`orchestration_states` reaps on an exact sliding 24h clock** — never count it for history. Two
  sessions got this wrong within an hour. Use `site_work_items` ∪ `site_work_items_archive`.
- **`kubectl` klog lines are LOCAL time; the DB is UTC.** Stamp `date -u` in the same command.
- **The after-test harness** is in this session's scratchpad (`after_test.sh`) — **promote it into
  this directory on next use.** ⚠ v1 printed three clean checks that had not run (`pages.is_archived`
  does not exist). v2 asserts row counts, refuses error-text-as-data, uses `pages.url` not
  `pages.name`, and prints every match rather than a count. **Validate an instrument on live data
  before trusting a clean run from it.**

## 4. Falsifiers for this handoff

- `376` or `206` closing, or the crawl step gaining an `on_error`.
- Any of the five parked pages beginning to serve (someone acted, or a fix rolled).
- The apex serving anything other than the current 66,395-byte index.
- Firecrawl beginning to support `thespruce.com` — the whole 376 hazard rests on one host's blocklist
  status, which is not ours and can change without notice.
- Anyone repairing this site by hand, which ends its value as a measurement.
