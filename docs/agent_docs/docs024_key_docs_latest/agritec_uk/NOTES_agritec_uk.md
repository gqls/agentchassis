# NOTES — agritec.uk

Append-only, newest at the bottom. Technical log: what was tried, what the system actually said,
and every misstep. The missteps are not an appendix — they are the point.

---

## 2026-08-21 — session 1, scoping and measurement

### What the owner asked for

Rebuild `agritec.uk` inside the framework. An adoption, but rebuilt from scratch so the tools and
guides are deconstructed and fully managed. Full mission, spec and plan. More imagery and
infographics. News, editorial and directory listings later. Same hosting, replacing what is there.
Two clarifications arrived mid-session: **everything eventually migrates**, and the copy is
**written afresh, to the same level of detail and greater**.

### MISSTEP 1 — I took the repo copy as the inventory. It is stale.

`/home/ant/projects/domains/agritec.uk/01/` has **6 tools and no `/deepdives/` directory**. The
live site has **13 tools and a seven-part deep-dive series**. I only found this because I fetched
`https://agritec.uk/` to check hosting, and the returned HTML had a nav link (`/deepdives/`) that
does not exist anywhere in the repo.

Had I not fetched it, the plan would have silently scoped out **14 of 26 pages** — more than half
the site — while appearing complete, and the subject ledger would have certified that as
"everything".

**The check:** for any adoption, inventory from the **live site**, not from a local copy. A repo
directory is a snapshot of when someone last synced it; here that was seven months and one whole
content section ago. Cheap: one `curl` of the homepage and a diff of the nav links against the
local tree.

### MISSTEP 2 — my first crawl double-counted pages.

The recursive link-follow emitted `/deepdives/../tools/lens-calculator.html` alongside
`/tools/lens-calculator.html`. Both fetch 200, so both looked like real, distinct pages, and the
first inventory over-counted. Normalise paths before counting. Recorded in `RUNBOOK` §5.

### MISSTEP 3 — I nearly read a nav census as a fleet-wide orphan rate.

I ran a census of pages with no `site_nav_items` row and got, e.g., finetuning.uk 21 of 50,
robot-hands.com 27 of 45. Read naively that says half the fleet's pages are orphaned.

**It does not.** `site_nav_items` is the *top nav*. A page absent from it can be perfectly well
linked from a section-index hub. The measurement answers "is it in the nav bar", not "is it
reachable", and those are different questions.

**The check:** reachability is measured at the rendered page — count inbound hrefs — not at the
nav table. That is what `RUNBOOK` §4 does, and it is what produced the one finding that is real
(below).

### Column names I got wrong, cheaply (recorded so the next session does not repeat them)

- `site_specs.spec_type` → the column is **`aspect`**.
- `site_nav_items.nav_group_id` → **`group_id`**.
- `site_nav_groups.label` → does not exist; the label lives on the item.
- `site_publish_checks` is **not** a check registry — it has two columns, `site_id` and
  `last_checked_at`. The checker vocabulary lives in `site_work_items.item_type`.

### What is actually true, measured

Confirmed the owner's recollection that not all articles are linked from the index, and it is
worse than remembered:

- `/guides/vapor-pressure-deficit.html` — **0 inbound links from any of the four index pages**.
  Reachable only from the VPD calculator.
- `elms-calculator`, `insect-waste-converter`, `seaweed-carbon-est` — on the home page, absent
  from `/tools/`. The tools index lists 10 of 13.
- `ai-labor-calc` — the inverse: on `/tools/`, absent from home.

Two further findings the owner had not flagged:

- **The market ticker was fabricated.** `data-collector/v1/cmd/updater/main.go` is a `rand`
  simulator that stamps its own output `"Source": "Simulated Exchange / National Grid"` and writes
  a `market-ticker.json` the site never reads. The homepage hardcoded the values instead. Someone
  has already commented the ticker out on the live site.
- **The data layer is dead.** `fetch(` count across all six agri calculators is **0**. The six
  `/data/*.json` files are read by nothing; every number is hardcoded per tool. The files are
  still publicly served — `/data/crop-dli-table.json` returns 200.

That last point matters for the cannabis ruling (owner, same day): the cannabis rows are in a
publicly-reachable JSON that no page fetches. **A client-side absence is not an absence** — a
server-side reader uses the same public URL. `b2 sync --delete` at cutover is what actually
retires it, not the fact that nothing links to it.

### The measurement that decided the build shape

The owner asked for "the same level of detail and greater", so I measured the existing depth
rather than guessing at it: explainers 315–453 words, deep dives 434–605, and exactly **one**
diagram across all six explainers.

Then I measured what the framework itself produces, because "greater" needs a destination:

| page_type | Sites | Pages | Avg words |
|---|---|---|---|
| `blog-post` | 4 | 40 | 1,568–1,876 |
| `guide` | 1 | 5 | 511 |

**`page_type` alone decides whether an explainer lands at ~1,600 words or ~500.** The old agritec
explainers sit at the `guide` shape. Building them as `page_type='guide'` — the obvious choice,
since the site calls them guides and the URL is `/guides/` — would have reproduced the exact
depth we were asked to exceed, while looking like the natural mapping.

`[MEASURED, sample noted]` The `guide` figure is one site, n=5. Strong signal, not a law.
Re-measure before leaning on it again.

### What has NOT been done yet

Nothing has been dispatched at the cluster. No `sites` row, no specs, no submission. The next
session's first action is the seed SQL (PLAN §4 Phase 1), and **the seed must land before the
first page is written** or the claims layer silently no-ops for the whole build.

---

## 2026-08-21 — session 1 continued: the seed, and a claim I had to withdraw

### Written

`SEED_2026-08-21_agritec_site_and_specs.sql` — sites row, `evidence_base`, `imagery_style_guide`.
**Not applied.** Validated locally only.

### Validated before it goes near the cluster

- Both dollar-quoted JSON blocks parse (`json.loads`): `evidence_base` 8,716 bytes, 0 facts,
  19 banned patterns, 28 allowed entities; `imagery_style_guide` 2,679 bytes with `content_hero`
  and `infographic` kinds.
- All 19 banned patterns carry a reason.
- All 19 compile **under Go's RE2**, not just Python's engine — the seed comment says the gate
  compiles them case-insensitively, so they were tested as `(?i)`+pattern, which is what
  `claims.go` actually does. Python passing is not evidence about the engine that runs them.
- The cannabis ban (owner ruling D8) tested on **both arms**: it fires on "Cannabis (Flower)
  needs a DLI of 30-45" and does **not** fire on "Leafy greens sit around DLI 12-17". A ban only
  tested on the positive arm can be a pattern that matches everything.

### MISSTEP 4 — I wrote into the plan that the seed was in council scope. It is not.

The approved plan said, in its own words, *"The seed SQL is an appliable DB migration and
therefore in scope for the council gate."* A session following that line would have submitted and
been refused client-side.

The error was reasoning from the **widening** (2026-08-19, `bugs_open/314`: migrations are in
scope now) to **this file**, without checking that this file is a migration. It is not. It is
per-site setup applied out of band with `psql -f` — exactly what the oufe seed's own header says
it is — and it lives in a workstream directory, not `sql_for_agents/`.

**The check**, and it took one command: source the single-source scope file and run the same
predicate `097` runs.

    . scripts/council-scope.sh
    jq -n --arg f "<path>" --arg code "$COUNCIL_SCOPE_CODE_RE" --arg mig "$COUNCIL_SCOPE_MIGRATION_RE" \
      -r '(($f|test($code)) or ($f|test($mig)))'

    false  <- docs/.../agritec_uk/SEED_2026-08-21_agritec_site_and_specs.sql
    true   <- docs/agent_docs/sql_for_agents/499_example.sql
    true   <- platform/orchestration/actions/x.go

`DRY_RUN=1` on the trigger answers the same question for free. CLAUDE.md says the scope is
single-sourced in `scripts/council-scope.sh` and not to re-derive it — I re-derived it anyway,
from a summary of a ruling rather than from the file. Corrected in PLAN section 5, marked, with
the evidence inline.

**The generalisable shape:** a rule that recently *widened* is the easiest kind to over-apply.
"X is in scope now" invites you to skip asking whether the thing in front of you is an X.

### Measured while writing the seed, and it changes what the gate is worth here

The oufe seed records that `ScanUnregisteredNumbers` was near-inert on finance prose. I did not
carry that claim across — I read the code, because the vocabulary is domain-specific and agritec
is a different domain.

`businessClaimContextRe` (`datahelpers/claims.go:660`) contains **no agricultural or
physical-units vocabulary at all**: no hectare, yield, crop, tonne, payment, rate, efficacy, DLI,
PPFD, EC, kPa, kWh, larvae or carbon. And `isExcludedNumber` (`claims.go:849`) excludes any
number directly preceded by a currency symbol — the multibyte check at :873 catches the pound
sign explicitly — and excludes en/em-dash ranges.

So on this site: every SFI payment rate is excluded before the lexical gate is even reached;
every agronomic range ("12-17") is excluded as a range; and everything else fails the lexical
gate for want of a matching word. **The automated number scan is close to vacuous on agritec's
entire subject matter.** `banned_claims` and `writer_block` are doing the work, plus human review.
Do not read a clean claims report on this site as "no invented numbers".

Combined with `bugs_open/288` (the register guards copy, not code), the honest position is that
neither layer sees the constants inside the six calculators. That is why the seed's `writer_block`
carries an explicit paragraph distinguishing an input *default* (not an assertion) from a constant
the tool applies on our authority (an assertion, which needs a registered fact **and** a mention
in visible copy where `extractAssertions` can reach it).

---

## 2026-08-21 — seed APPLIED, and a measurement that changes Phase 4

### Applied

Seed applied to the live DB. Verified, not assumed:

    domain      email                        status  network_id
    agritec.uk  agritec@contactforsales.com  active  ...0002

    aspect               is_current  pinned  source  created_by
    evidence_base        t           t       manual  agritec-workstream-2026-08-21
    imagery_style_guide  t           t       manual  agritec-workstream-2026-08-21

    bans=19  facts=0  entities=28  writer_block_chars=2024  has_rule=t
    work_items=0  pages=0

**`work_items=0` and `pages=0` are the important two.** The seed is inert: it creates the site
and its guard rails and dispatches nothing. Nothing is running.

### Verified that the reader actually reads it

`EvidenceBase`'s struct tags (`claims.go:259`) are `audit_doc`, `governing_rule`, `facts`,
`banned_claims`, `allowed_entities`, and `BannedClaim{pattern, reason}` — exactly the keys seeded.
A key mismatch here would have produced a present-but-toothless gate that verifies clean.

`writer_block` is **not** on that struct. It reaches the writer through the agent's prompt, and
`TestRegulatedOnlyBaseIsNotSafeToWriteBack` pins the loss deliberately: `writer_block`,
`schema_notes`, `source.citation` and `fact.writer_line` are all destroyed by a parse+marshal
round trip. **No `ParseEvidenceBase` caller may persist the struct** or this site's 2,024-character
writer_block dies silently. The scheduled refresher is safe — it works on `map[string]interface{}`
and honours `writer_block_managed`, which this seed deliberately leaves unset so the hand-written
block is left alone.

### MISSTEP 5 — I nearly recorded "tool-generator does not read the evidence register". False.

The first measurement was a text search of `default_config`, which said `tool-generator` never
mentions `evidence_base`. I was one keystroke from writing that down as "the tool path never sees
the register".

Then I read the deciding arm instead of the summary. `tool-generator`'s `load_brand_context` step
is `read_site_spec` **with no `aspect` in its config**, and `site_spec_actions.go:480` treats an
omitted aspect as *all aspects mode* — `SELECT aspect, data FROM site_specs WHERE site_id=$1 AND
is_current=true`. So it loads `evidence_base` along with everything else, and `generate_tool_html`
lists `site_specs` in its `input_fields`. **The register does reach the prompt.**

The check that caught it: the config-text search answers "does the prompt NAME this?", which is
not "does the agent RECEIVE this?". Two different questions, and the first one reads like the
second.

### What IS true, measured across the four writing agents

| agent | receives site_specs | prompt names writer_block | names evidence/fact/invent |
|---|---|---|---|
| page-content-writer | yes | **yes** | yes |
| tool-recreation-handler | yes | **yes** | yes |
| tool-generator | yes (all aspects) | **no** | no (5,189-char template: no evidence, no fact, no source, no invent) |
| tool-improver | yes (all aspects) | **no** | no |

So on the tool paths the evidence base is **present in context and unaddressed by the
instruction**. I have proven the prompt does not direct the model to it. I have *not* proven the
model ignores it — that would need a run, and it is not a claim I am making.

`[MEASURED 2026-08-21]` against the live `agent_definitions` rows, plus `site_spec_actions.go:436-484`
for the aspect-selection arm.

### Why this matters here more than on most sites, and what it changes

Set this beside `bugs_open/288` (the register guards copy, not code — a constant inside a
calculator's JavaScript is checked by nothing afterwards). The tool path therefore has **neither a
clear instruction going in nor a check coming out**. Prose has both.

This site is six calculators whose SFI rates, LED efficacies and carbon fractions are exactly such
constants — and the owner's stated goal is that they be *evolved* through the framework, which is
`tool-improver`, the agent with the thinnest coverage of the four.

**Change to PLAN Phase 4:** the honesty constraint must be written into the tool brief we hand to
`tool-generator`, in the brief's own words. It cannot be delegated to the evidence register,
because the register arrives as an unaddressed blob. This is an addition to the existing control
(register the constant AND assert it in visible copy), not a replacement — that one guards the
check side, this one guards the instruction side.

**Not filed as a platform bug.** It is a cross-cutting structural claim, and CLAUDE.md's bar for
one of those is the 090 diagnosis loop or a stated substitute. Raised with the owner instead as a
decision: it is his call whether this lane widens into filing it.

---

## 2026-08-22 — Phase 2 run 1: the SFI calculator on the live site is paying a subsidy that no longer exists

### The finding

First evidence-researcher run (correlation `1e8c7735-b922-450c-b261-cbfac3e2d5d6`) registered **10
facts**, and four of them say the same thing:

- gov.uk SFI26 scheme rules: *"the SFI management payment has been removed for SFI26 agreements"*
- same page: *"You will not be paid: an SFI management payment for your SFI26 agreement"*
- defrafarming.blog.gov.uk, 2026-02-24: *"We will no longer offer the SFI management payment. It
  was intended as a time-limited payment to support farmers transitioning into the new scheme."*

**The live site still pays it, as a headline feature.**
`agritec.uk/tools/elms-calculator.html` opens with a green callout — *"You receive £20 per hectare
for the first 50 hectares… the first £1,000 of your SFI income is effectively guaranteed"* — and
its JS computes `Math.min(farmSize, 50) * 20` into a top-line "SFI Management Pmt" row. Next to a
GOV.UK link, which lends it authority.

This is exactly the `bugs_open/288` class stated in the abstract yesterday and met in the concrete
today: a legislated figure encoded in a calculator, checked by nothing, still running after the
legislation moved. The SDLT precedent ran an expired threshold for 16 months.

Four further SFI26 constraints landed that the old calculator does not model at all: a £100,000
annual agreement value cap, a 3-hectare eligibility floor, one agreement per farm business, and
limited-area actions capped at 25% of the farm's agricultural area.

### Verified first-hand, because a citation is not a reading

`verify_and_register` re-fetches the URL and rejects a claim unless the quote appears verbatim —
so the machine proved the words are on the page. It cannot prove we read them correctly
(`bugs_open/161`). So I fetched the GOV.UK page myself (HTTP 200) and confirmed all five quotes
present, plus the surrounding context: the removal sits in a list of SFI26 changes, unambiguous.

### Acted on it: five bans, added BEFORE any page exists

`SEED_2026-08-22_sfi26_bans.sql`. Tested on **both arms**, which is the part that mattered:

- Caught: the retired site's own sentences, "management payment is available", "farmers will
  receive an SFI management payment", "the management payment is £20 per hectare".
- **Left sayable:** "the SFI management payment has been removed for SFI26 agreements", "Under the
  SFI 2023 offer, an annual management payment of £20 per hectare was available", "You will not be
  paid an SFI management payment…", "DEFRA said it would no longer offer…".

That second list is the point. A ban that made the removal unsayable would suppress the single
most useful thing this site can currently tell an SFI reader. Two patterns failed their keep-arm
on the first attempt and were rewritten:
- `guaranteed[^.]{0,60}(£1,000|first 50)` required the words in an order the real sentence does not
  use — the site says "the first £1,000 … is effectively guaranteed". Now matches both orders.
- `(SFI)[^.]{0,40}management payment (of|is) £` caught the correctly-scoped past-tense form.
  Narrowed to present-tense verbs only (`is|remains|stands at|comes to`), so registered fact
  CIT-f88b5cd stays usable in the past tense with its scope attached — which is its only honest use.

All five compile under Go RE2.

### MISSTEP 6 — the supersede-then-insert as a single CTE cannot work, and it reads perfectly

First attempt wrapped supersede + insert in one statement with data-modifying CTEs. It failed:

    duplicate key value violates unique constraint "idx_site_specs_current"

All CTEs in a statement run against **one snapshot**, so the INSERT's uniqueness check never sees
the sibling UPDATE's supersede — the old row is still `is_current` as far as the partial index is
concerned. I invented that form because I needed to carry the existing document forward, and in
doing so diverged from the oufe seed, which uses sequential statements.

**It failed safely**: `BEGIN` + `ON_ERROR_STOP` rolled everything back, verified — bans, facts and
writer_block all unchanged afterwards. Rewritten as sequential statements in one transaction, with
a `DO/RAISE` guard that aborts unless it finds exactly one current row carrying exactly the 10
known facts. `DO/RAISE` rather than a `SELECT` verify block, because `ON_ERROR_STOP` ignores a
non-empty result and a `SELECT` cannot stop a `COMMIT`.

### Carry-forward proven, not assumed

    is_current  created_by                     bans  facts
    f           agritec-workstream-2026-08-21    19      0
    f           evidence-researcher              19     10
    t           agritec-workstream-2026-08-22    24     10

Facts still 10, writer_block still 2,024 chars, allowed_entities still 28. A lost `facts` array
would have looked exactly like success on the bans count alone, which is why both were checked.

---

## 2026-08-22 — Phase 2 runs 2 to 5: what the register keeps, and what it cannot reach

Six runs, **18 facts standing out of 27 registered**. Nine were removed on review. Every run
reported `COMPLETED`. The run status carries no information about the value of the output, which
is the single most important operational lesson of this phase and is now `RUNBOOK` §9.

| run | question | registered | kept | why |
|---|---|---|---|---|
| 1 | SFI rates and management payment | 10 | 10 | 9 primary gov.uk/DEFRA; 1 third-party, correctly scoped |
| 2 | "Ofgem average non-domestic" price + carbon intensity | 5 | 1 | 4 were DOMESTIC price-cap figures; wrong market entirely |
| 3 | DESNZ QEP industrial price (retarget of run 2) | 6 | 5 | 5 ONS non-domestic, writer_line-scoped by quarter; 1 metadata |
| 4 | grid carbon intensity | 4 | 0 | the figures exist only in .xlsx — unreachable, see below |
| 5 | LED photon efficacy | 3 | 2 | 1 was nine years old and said "now" |

### The three failure shapes, each caught by a different check

**Wrong market (run 2).** Ofgem's price cap is a *domestic* consumer protection. Every reader of
this site buys on non-domestic contracts. The facts were true, sourced, current and verbatim-
verified — and wrong for every reader. **Nothing about a true, well-cited fact announces that it
is about somebody else**, which is why this needs a deliberate check rather than a smell test.
One of the four also asserted "26.11 pence per kWh" from a two-column table whose other column
read 24.67; which column it was is unknowable from the register.

**Unreachable format (run 4).** Asked for grid carbon intensity; got a publication date, two
third-party methodology descriptions, and a GOV.UK correction notice about hybrid cars and hotel
stays. Cause measured, not guessed: the DESNZ publication page contains **zero** kgCO2e/kWh
figures — the factors ship only as three `.xlsx` files and a methodology PDF. `extract_claims`
requires a verbatim quote from page text, so a number in a spreadsheet cell can never satisfy it
while the landing page's prose easily can. **The property that makes the register trustworthy is
the same one that guarantees this silent miss.** No re-phrasing reaches it. Now in `LANDMINES.md`
with a free pre-check.

**Undated and time-sensitive (run 5).** "Many new LED fixtures **now** exceed 2.0 µmol/J" — from a
page dated 2017-07-03 in its own metadata, which the extractor recorded as `published: (none)`
while guessing `staleness_days`. So the refresh machinery was primed to measure drift from a date
it never captured. True when written; misleading now, after two DLC threshold rises. Dropped.
The two facts kept from that source do not age — mature HPS efficacy, and a physics ceiling — and
their 2017 date now sits in the `writer_line`, which is what the writer actually reads.

### Two things that worked better than expected

**`writer_line` is the real control, not `value`.** Several kept facts have a lossy `value` — 1.7
is one of two figures in its claim, 5.1 is the top of a range, 25.97 is a quarter-specific
figure ~20 months old. In each case the writer_line embeds `{value}` in a sentence that restores
what `value` drops: *"was {value} pence per kWh in Q4 2024, sourced from DESNZ QEP table 3.4.1
(as cited by ONS, May 2025)"*. A writer following that cannot state a stale quarterly figure as
today's price. This is why I left the `value` fields alone rather than "fixing" them.

**A valueless fact is not automatically junk.** Three facts carry no number and are the most
important in the register — the SFI management-payment removals. The test is whether a fact
carries an *assertion*, not whether it carries a digit. `value IS NULL` is a useful tell and a
bad rule.

### Decision recorded: LED efficacy is a user input, not a registered default

No citable *current* figure exists in a form this pipeline can reach. That is decision D4 working
exactly as written — the number is not published, it is asked for. It was already a user field on
the retired tool (read off the operator's own fixture datasheet), so nothing is lost. The three
LED facts serve the explainer instead, where a dated, attributed comparison of HPS against the
theoretical ceiling is the right teaching material.

Carbon intensity is **deferred, not failed**: no Phase 1 calculator consumes it. The energy tool
returns money, not emissions, and carbon intensity appeared on the retired site only in the
fabricated ticker and the dead data layer.

---

## 2026-08-24 — submitted, and the build ran further than expected while I was reading a peer's message

### Submitted (Tier-3), and the roadmap held

`TRIGGER_submit_tier3.sh`, correlation `84529075-bc81-4223-b52c-e3928555ad66`, COMPLETED. Both
briefs persisted (5,884 / 8,238 chars) — the whole reason for using that trigger over `082`. The
seeded, pinned `evidence_base` (95KB, 105 facts) and `imagery_style_guide` survived untouched.

The planner **obeyed the roadmap brief exactly**: index, about, tools, contact, a `section-index`
guides hub, and the six named explainers. Nothing invented.

### THE DEPTH DECISION PAID OFF — measured, not hoped

The four explainers that have rendered so far:

| page | words |
|---|---|
| hydroponic-solution-chemistry | 1,726 |
| vapour-pressure-deficit-and-transpiration | 1,717 |
| seaweed-and-the-carbon-question | 1,498 |
| insect-bioconversion | 1,400 |

Against the retired site's **315–453**. Roughly **four times** the depth, and within range of the
~1,600 the `blog-post` measurement predicted. The other two are still `needs_rebuild` and render
as 1 word, which is why they must not be read as failures yet.

### AND THE DEPTH DECISION CAUSED A DEFECT — the exact one we set out to fix

`/guides/index.html` has the sections I specified — `["hero", "guide-list"]`, so the
explicit-list-component requirement worked. The `guide-list` instance is `deployed` with 4,937
chars of rendered HTML.

**It contains ZERO anchors and zero cards.** Only the section furniture: a heading ("Explainers
behind the calculators") and a CTA. Three of the six explainers have **no inbound link from
anywhere on the site**.

The cause is my own decision. `guide-list` resolves `page_type='guide'`; I chose `blog-post` for
the word count, and it delivered the words. So the two halves of the same decision fought each
other, and I did not see it coming — the roadmap brief's "build the hub with an explicit list
section from day one" was necessary and **not sufficient**, because a list component and the page
type it resolves are a pair, and I only specified one of them.

**It will not self-heal.** Checked rather than assumed: `rerender_single_page_action` READS
`pages.sections` and assembles; nothing in the rerender path rewrites it. The 11 queued
`page_rerender` items will re-render an empty list, for ever.

**The fix, verified at the artefact before adopting it.** `blog-listing` is the component that
resolves blog-posts, and `fundamentallyai.com /platform-log/index.html` is the same shape as our
hub — a `section-index` with `["hero", "blog-listing"]`. That page is the subject of
`bugs_open/309` ("six unclickable cards so every article is orphaned"), so copying it blind would
have been copying a bug. I read its rendered HTML: **16 anchors, real `/blog/...` hrefs.** 309's
defect is not present there now. The component works.

### What is in flight and must NOT be mistaken for breakage

- **16 `unresolved_cta`** — every one is `secondary_cta_url` with "no real-page destination".
  Those point at tool pages that do not exist yet, because the tools are Phase 4. Expected.
- **11 `page_rerender` (triaged)** — in flight; will resolve the two `needs_rebuild` explainers.
- **5 `needs_page`** — at needs_human_review.

The site was submitted about three hours ago and the cascade is still working. Intervening on
things the pipeline is about to finish is how you end up fighting it.

### Peer message from the bugs_open/382 lane — checked, not relayed

They flagged migration 586 (`image-build-handler.call_variant_gen` now forwards `kind` and
`site_id`) and asked whether agritec had pre-fix per-page heroes.

**The obvious answer was wrong.** agritec has 17 assets created 12:12–12:30; 586 applied 13:46 —
so they all predate the fix, and twelve keys are literally `hero_about`, `hero_sfi`, `hero_vpd`
and so on. On the natural test the site looks squarely affected.

It is not. All 17 came from `needs_imagery` items, which route through `call_imagery_gen`; one
went `needs_hero_image` → `call_hero_gen`. **Nothing here touched `call_variant_gen`** — there are
no `unfulfilled_hero_variant` items on this site at all.

**The transferable point, sent back to them:** `asset_key` naming is not a reliable indicator of
which branch produced an image. `hero_<page>` says "per-page hero" and says nothing about whether
it came from plan-fulfilment or variant-repair. The discriminator is the work item type.

Corroborating on the unaffected path: all 17 are `banana/gemini-3-pro-image-preview`, none SDXL,
and 16 of 17 `origin_prompt` values carry this site's seeded palette hex verbatim — which appears
in no `site_plan_imagery.prompt` row, so the style guide was applied at generation time, not at
plan time. Seeding it before submission (PLAN Phase 1) did the job it was seeded for.
