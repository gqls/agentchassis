# agritec.uk — subject ledger

Created 2026-08-21. **This file is the completeness contract for the rebuild.**

Every page on the live hand-built site has a row here. A row records the *subject* — what the
page teaches — not its text. The copy is written afresh by the framework; nothing is ported.
See `PLAN_2026-08-21_agritec_uk.md` §"written afresh".

The ledger does three jobs:

1. **Nothing is dropped.** Owner, 2026-08-21: every subject the existing site covers is
   eventually covered by the new one. A subject with no destination page is visible here as an
   unfinished row. (The current site's own indexes lost four pages exactly this way — see
   §"Measured defects" below.)
2. **Nothing comes out thinner.** Owner, 2026-08-21: *"to the same level of detail and greater."*
   The floor columns are the measured depth of the live page on 2026-08-21. The rebuilt page must
   meet or exceed every one of them.
3. **Nothing gets ported.** The ledger holds subjects, not prose, so there is no source text
   sitting here inviting a copy-paste.

**Status vocabulary:** `not-started` → `evidence-sourced` → `built` → `deployed` → `verified`.
`verified` means the depth check AND the reachability check have both been re-run against the
live rebuilt page, not that it looks finished.

---

## How the floor was measured

Each live page fetched 2026-08-21 and counted with scripts and styles stripped:

    words   body text word count
    H2      count of <h2 ...
    tbl     count of <table
    eqn     count of equation-box blocks
    fig     count of <figure | <img | <svg

The command is in `RUNBOOK_agritec_uk.md` §2 so the same measurement can be re-run against the
rebuilt site. **Re-run it — do not eyeball the result.**

**Depth target, not just floor.** Measured on four mature framework sites the same day:
`page_type='blog-post'` averages **1,568–1,876 words**; `page_type='guide'` averages **511**
(one site, n=5). The old agritec explainers sit at 315–453 — the `guide` shape. So the explainers
are built as `blog-post` and target **~1,600 words**, not merely the floor. Building them as
`guide` would reproduce the depth we were asked to exceed.

---

## Phase 1 — agri / CEA cluster

### Tools (6)

| # | Subject | Live source | Inputs | Floor (words) | Dest | Status |
|---|---|---|---|---|---|---|
| T1 | Convert a crop's DLI target and photoperiod into required PPFD, fixture load and electrical running cost | /tools/vertical-energy-calc.html | 5 numeric | 439 | | **evidence: part** — 5 ONS non-domestic electricity price facts registered 2026-08-22, all writer_line-scoped to their quarter. Still needs: LED efficacy by fixture class, grid carbon intensity, crop DLI ranges |
| T2 | Vapour pressure deficit from air temperature, RH and leaf-temperature offset | /tools/vpd-calculator.html | 3 numeric | 517 | | not-started |
| T3 | Stock-tank dilution to move a reservoir from current EC to target EC without precipitation | /tools/nutrient-dosing.html | 3 numeric + 1 select | 612 | | not-started |
| T4 | Black soldier fly mass balance: wet waste to larvae, protein, frass, and rearing area | /tools/insect-waste-converter.html | 2 numeric + 1 select | 524 | | not-started |
| T5 | Macroalgae carbon estimate separating cycling from sequestration, with credit valuation | /tools/seaweed-carbon-est.html | 2 numeric + 1 select | 499 | | not-started |
| T6 | Model SFI revenue by stacking compatible actions across a farm's area and boundaries | /tools/elms-calculator.html | 9 numeric + 8 toggles | 801 | | **evidence: DONE** (72 attested SFI26 facts, 2026-08-22) · **spec written**: `TOOL_SPEC_sfi_stacker.md` · blocked only on the site existing (Phase 3) |

#### T6 is a redesign, not a rebuild (found 2026-08-22, Phase 2 run 1)

The scheme the existing calculator models has been replaced, and the change is not limited to the
management payment:

- **The management payment is gone.** *"the SFI management payment has been removed for SFI26
  agreements"* — gov.uk SFI26 scheme rules. The live calculator pays it as a headline line item.
- **The action list itself has changed.** SFI26 carries **71 actions, against 102 in SFI24**
  (registered fact, gov.uk). The live calculator hard-codes eight action codes — SAM1, SAM2, SAM3,
  IPM1, IPM4, NUM1, HRW1, HRW2 — with rates from the older offer. **Every one of those codes and
  rates has to be re-sourced against SFI26; none may be carried across.**
- **Three constraints exist that the calculator does not model at all:** a £100,000 annual
  agreement value cap, a three-hectare minimum to be eligible to apply, and a limit of 25% of the
  farm's agricultural area on any combination of limited-area actions. A stacking tool that
  ignores all three can produce a total the scheme would never pay.

So T6's real work is: re-source the SFI26 action set and rates, then design a tool that models the
caps as well as the sums. **It needs its own evidence run for the action rates** — the run so far
covered scheme structure, not the per-action figures.

**Every tool additionally owes:** a Tier-4 headless-Chromium acceptance run, and — per
`bugs_open/288` — every constant it encodes registered as an evidence fact *and* asserted in the
page's visible copy, because the claims scanner cannot see inside JavaScript.

### Explainers (6)

| # | Subject | Live source | words | H2 | tbl | eqn | fig | Dest | Status |
|---|---|---|---|---|---|---|---|---|---|
| G1 | Why photometric units fail for plants: PAR, PPFD, DLI, the integration formula, and fixture efficacy as a double cost (electrical + HVAC) | /guides/physics-of-light.html | 453 | 4 | 1 | 1 | 2 | | not-started |
| G2 | How ELMS replaced area payments, and how SFI options layer without losing yield | /guides/elms-stacking.html | 441 | 3 | 0 | 1 | 1 | | not-started |
| G3 | The insect as a bioreactor: bioconversion rate, frass as the second product, metabolic heat as the density constraint | /guides/insect-bioconversion.html | 409 | 4 | 0 | 1 | 0 | | not-started |
| G4 | Why PPM is a fiction: EC against TDS conversion scales, and the A/B tank rule | /guides/hydroponic-chemistry.html | 324 | 3 | 1 | 1 | 0 | | not-started |
| G5 | Carbon cycling against sequestration, the stoichiometry, and the UK regulatory position | /guides/seaweed-carbon.html | 319 | 3 | 0 | 1 | 0 | | not-started |
| G6 | VPD as the driver of transpiration: the kPa target band and the leaf-temperature offset | /guides/vapor-pressure-deficit.html | 315 | 4 | 1 | 1 | 0 | | not-started |

#### DLI evidence carries a GROWTH-STAGE caveat (2026-08-22, Phase 2 run 6)

Four DLI facts registered, from Purdue and Virginia Tech extension services. Sound sources, and
DLI transfers across borders in a way SFI payment rates do not — a lettuce plant's light
requirement is physics and biology, not policy, so US extension figures are legitimate here in a
way US subsidy figures never would be.

**But three of the four are for TRANSPLANT production specifically**, which is a distinct growth
stage with a lower light requirement than a mature or fruiting crop. Purdue gives 15–20 for tomato
*transplants*; the retired site's table gave 20–30 as the tomato optimum, presumably for fruiting.
Both can be right. **A writer who drops the stage qualifier turns one into the other and
under-lights somebody's crop.**

The control already holds: three of the four writer_lines say "transplant production" explicitly.
The exception is CIT-a6eef3fe8aef9044 (Virginia Tech, "lettuce grows well at a DLI close to 15"),
which carries no stage. Treat that one as unqualified and say so, or pair it with a stage-specific
source before using it.

Note also what the extractor did RIGHT here and got wrong on the LED facts: for the three ranges
it left `value` null and put the range in the writer_line, rather than picking an end of the range
as "the" figure. That is the correct handling of a range, and it is inconsistent with run 5 — so
it is a behaviour to verify per run, not to rely on.

#### DEPTH GATE: PASSED, measured 2026-08-24

All six built and `deployed`, three components each. Word counts of rendered components against
the floor taken from the live retired site on 2026-08-21:

| # | subject | floor | achieved | ratio | destination | status |
|---|---|---|---|---|---|---|
| G1 | physics-of-light | 453 | **1803** | ×4.0 | `/blog/the-physics-of-horticultural-lighting.html` | deployed |
| G2 | elms-stacking | 441 | **1648** | ×3.7 | `/blog/stacking-agricultural-scheme-actions.html` | deployed |
| G3 | insect-bioconversion | 409 | **1400** | ×3.4 | `/blog/insect-bioconversion.html` | deployed |
| G4 | hydroponic-chemistry | 324 | **1726** | ×5.3 | `/blog/hydroponic-solution-chemistry.html` | deployed |
| G5 | seaweed-carbon | 319 | **1498** | ×4.7 | `/blog/seaweed-and-the-carbon-question.html` | deployed |
| G6 | vapour-pressure-deficit | 315 | **1717** | ×5.5 | `/blog/vapour-pressure-deficit-and-transpiration.html` | deployed |

**Every one clears its floor by roughly four times**, and all six land in the 1,400–1,803 band
the `blog-post` measurement predicted (~1,600). The owner's instruction — "the same level of
detail and greater" — is met on the measure it was set against, and it is the `page_type` choice
that did it: the `guide` shape measured ~511 words.

Two of them only built after **two of my own `banned_claims` patterns were narrowed** — the
ticker ban was matching the citation "Carbon Brief, May 2025", and the management-payment ban was
matching the honest past-tense sentence it was written to permit. Both failures presented as a
page refusing to build, which is the good failure mode; a ban that quietly suppressed the sentence
would never have surfaced.

**Every explainer additionally owes:** at least one code-rendered infographic (the whole current
guide set has exactly one diagram between them), a sourced figure for every number, and the
equation/table count preserved or exceeded.

### Hubs (4)

| # | Subject | Live source | Status |
|---|---|---|---|
| H1 | Home — what the site is and the way into both clusters | / | not-started |
| H2 | Tools index — must list **every** tool | /tools/index.html | not-started |
| H3 | Explainers index — must list **every** explainer | /guides/index.html | not-started |
| H4 | Deep-dive series index | /deepdives/index.html | phase 2 |

---

## Phase 2 — IoT / machine-vision cluster

The engineering build-log for a distributed optical bio-monitoring system. Agricultural, not
off-topic: it is a crop-monitoring rig. Rebuilt after Phase 1 is verified live.

### Tools (7)

| # | Subject | Live source | Inputs | Floor (words) | Status |
|---|---|---|---|---|---|
| T7 | Machine-vision lens selection from sensor format, working distance and target width | /tools/lens-calculator.html | 2 numeric + 1 select | 518 | phase 2 |
| T8 | Edge buffer sizing from capture mode, resolution, interval and required survival time | /tools/edge-buffer-calc.html | 3 numeric + 2 select | 386 | phase 2 |
| T9 | Image-quality gate simulator: focus and exposure against an acceptance threshold | /tools/image-quality-sim.html | 3 range | 816 | phase 2 |
| T10 | IoT service config and systemd unit generator | /tools/config-generator.html | 5 text | 374 | phase 2 |
| T11 | Serverless cost projection for camera fleet ingestion, storage and inference | /tools/cloud-cost-calc.html | 5 numeric | 509 | phase 2 |
| T12 | Operational SQL builder for the monitoring dataset | /tools/sql-query-builder.html | 4 select | 728 | phase 2 |
| T13 | Human-in-the-loop review labour and cost from model success rate | /tools/ai-labor-calc.html | 4 numeric | 218 | phase 2 |

### Deep dives (7)

| # | Subject | Live source | words | H2 | tbl | eqn | fig | code | Status |
|---|---|---|---|---|---|---|---|---|---|
| D1 | Distributed optical bio-monitoring: the system architecture | /deepdives/iot-system-architecture.html | 435 | 3 | 1 | 0 | 2 | 0 | phase 2 |
| D2 | The physical edge: hardware specification and Linux configuration | /deepdives/iot-hardware-spec.html | 434 | 4 | 0 | 1 | 4 | 2 | phase 2 |
| D3 | Edge software, part 1 | /deepdives/iot-edge-software-p1.html | 604 | 5 | 0 | 2 | 4 | 4 | phase 2 |
| D4 | Edge software, part 2 | /deepdives/iot-edge-software-p2.html | 605 | 4 | 0 | 0 | 4 | 2 | phase 2 |
| D5 | Cloud ingestion | /deepdives/iot-cloud-ingestion.html | 581 | 5 | 0 | 1 | 2 | 4 | phase 2 |
| D6 | The AI model: development and lifecycle | /deepdives/iot-ai-model.html | 494 | 5 | 0 | 0 | 2 | 9 | phase 2 |
| D7 | Data reporting | /deepdives/iot-data-reporting.html | 577 | 4 | 0 | 0 | 4 | 5 | phase 2 |

---

## Measured defects in the live site (2026-08-21) — do not reproduce these

Fetched and counted, not recalled:

- **`/guides/vapor-pressure-deficit.html` is linked from no index page at all.** Home, `/tools/`,
  `/guides/` and `/deepdives/` each return 0 matches for it. Its only inbound link is from the VPD
  calculator. The owner remembered this; the measurement is worse than the recollection, because
  the repo copy and the live site are orphaned differently.
- **`elms-calculator`, `insect-waste-converter`, `seaweed-carbon-est`** appear on the home page and
  are **absent from `/tools/`** — the tools index lists 10 of 13.
- **`ai-labor-calc`** is the inverse: on `/tools/`, absent from the home page.
- **The data layer is dead.** `fetch(` count across all six agri calculators is **0**. The six
  `/data/*.json` files are read by nothing; every number is hardcoded per tool. The JSON files are
  still publicly served (`/data/crop-dli-table.json` returns 200).

  > **CORRECTED 2026-08-22 — "dead" is right; "untrustworthy" was not, and the framing here
  > implied it.** Sourcing the crop DLI figures found their actual origin. Virginia Cooperative
  > Extension SPES-720, Table 3 publishes **Lettuce 12−17** and **Tomato 20−30** — exactly the
  > retired site's `crop-dli-table.json` values — attributed there to Dou et al. (2018), Faust et
  > al. (2005) and Pramuk. Read first-hand, HTTP 200.
  >
  > So that file is **UNCITED, not fabricated.** The distinction is load-bearing: an uncited figure
  > needs a source attached, an invented one needs deleting, and treating the first as the second
  > throws away work somebody did properly. All eleven ranges are now in the register by
  > attestation (`SEED_2026-08-22f_dli_table_attested.sql`).
  >
  > It does **not** extend to the market ticker, which stays fabricated and is separately proven
  > so — its feeder generates values with `rand()` and labels its own output "Simulated Exchange".
  > Two files in the same directory, two opposite verdicts. Check each; generalise from neither.
  >
  > What caught it: the run REJECTED those claims as `citation_lost`, whose own advice reads
  > "possible hallucination — discard". Following that advice would have produced both wrong
  > conclusions. The figures are in a table, and the separator is U+2212 MINUS SIGN — neither can
  > satisfy a verbatim-quote re-match. See `LANDMINES.md`.
- **The sitemap is a THIRD inventory, and it agrees with neither of the other two.**
  Found by the framework itself: within about four hours of the seed landing, the
  `check_site_structural_validity` discovery check ran against the live site and filed seven
  `sitemap_entry_dead_live` items. Verified by hand afterwards — 26 sitemap entries against 30
  live pages, and they diverge in *both* directions:
  - **7 entries are dead (404).** The sitemap lists the deep-dive series under an old
    `iot-01-architecture.html` / `iot-02-hardware.html` naming scheme. The real pages are
    `iot-system-architecture.html`, `iot-hardware-spec.html` and so on. Confirmed:
    `/deepdives/iot-01-architecture.html` -> 404, `/deepdives/iot-system-architecture.html` -> 200.
  - **11 live pages are absent from it**, including **every one of the seven real deep dives** —
    so the whole engineering series is invisible to search — plus `guides/elms-stacking.html`,
    `tools/elms-calculator.html`, `tools/insect-waste-converter.html` and
    `tools/seaweed-carbon-est.html`.

  Note *which* four agri pages those are: the same ones missing from the tools and guides
  indexes. The nav, the sitemap and reality each tell a different story, and the three failures
  overlap without matching. **This is the argument for the rebuild in one line** — the framework
  generates nav, sitemap and pages from one set of rows, so they cannot drift apart like this.

  No action needed on the old site: the items are `detected` with no handler, the check is
  mechanical and self-clears when a re-probe finds the entry live, and every URL involved is being
  retired at cutover anyway. Recorded because it is evidence, and because the rebuilt site must be
  checked for the same class rather than assumed free of it.

- **The market ticker was fabricated.** Its feeder,
  `domains/agritec.uk/data-collector/v1/cmd/updater/main.go`, is a `rand` simulator stamping its
  own output `"Source": "Simulated Exchange / National Grid"`. Already commented out on the live
  site. It does not come back.

## Standing content rules for this site

- **No cannabis content** (owner ruling, 2026-08-21). Live today in one place — a help-text hint
  under the DLI input of `/tools/vertical-energy-calc.html` ("Cannabis: ~30-45") — and in two rows
  of the publicly-served `/data/crop-dli-table.json`. Both go at cutover; `b2 sync --delete`
  retires the JSON along with everything else not in the new build. Because the framework's
  writers could reintroduce it from general knowledge of CEA crops, this is enforced as a
  `banned_claims` pattern in `evidence_base`, not left to prompt wording.
- **No unsourced figure is published anywhere** (owner ruling, 2026-08-21). Not in copy, not as a
  tool default, not in a spec.

## Repo copy warning

`/home/ant/projects/domains/agritec.uk/01/` is **stale** — 6 tools against 13 live, and no
`/deepdives/` at all. It is useful for reading the calculator maths at leisure. It is not the
inventory. The live site is.
