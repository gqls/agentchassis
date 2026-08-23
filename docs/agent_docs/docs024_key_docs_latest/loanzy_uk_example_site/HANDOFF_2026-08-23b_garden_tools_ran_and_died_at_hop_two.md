# HANDOFF 2026-08-23b — ⚠ SUPERSEDED AFTER 35 MINUTES. ITS TITLE IS WRONG: THE BUILD IS ALIVE.

> **READ `HANDOFF_2026-08-23c_garden_tools_escaped_and_is_building.md` INSTEAD.**
> This file's central claim — that the build died terminally at hop two and that retrying is
> structurally incapable of escaping — **was refuted 35 minutes after it was written**, by the
> system, while I watched. At 20:02Z the second submission's `needs_vertical_research` re-drew a
> different exemplar set, cleared all three crawls, wrote a `vertical_landscape` spec at 20:05:45Z
> and queued `needs_strategy` at 20:05:55Z. The cascade is at hop three.
>
> **It is kept, unedited below, deliberately.** The lane's rule is that a wrong belief is not
> deleted — it is the only record of how the understanding moved, and this one is a worked example
> of converting four identical observations into "structurally incapable". Retraction and the
> transferable rule: `bugs_open/376` §4b and `NOTES_loanzy_uk_example_site.md` (20:06Z).
>
> **Still accurate below:** the pre-flight recipe, the 376 mechanism (no `on_error`; sole producer
> of `needs_strategy`), the 326 correction, the stale-baseline landmine, and the measured dispatch
> properties. **Wrong below:** every sentence asserting the build is dead or that retry cannot work.

---

# (original title) `garden-tools.uk` RAN. It died at hop two. Do not re-run it expecting a site.

**Supersedes `HANDOFF_2026-08-23_garden_tools_continue_here.md`**, which said "the build has NOT been
run" and is now falsified by its own §6 criteria (`sites`/`site_work_items` are non-zero). That file
stays for its pre-flight recipe and its bug-state table, both still good; ignore its §2 "GO".

## 1. What happened, in five lines

- Submitted **17:17:18Z**, nothing but the domain. No mission, no email, no seed. **No deviation.**
- The classifier did well: **unregulated editorial affiliate hub**, 12 pages, confidence 0.82, and
  it said in its own reasoning that a regulated direction did not apply. `CGV-032` never had to fire.
- **Hop two killed it.** `vertical-exemplar-researcher` crawls three LLM-picked exemplars; Firecrawl
  refuses `thespruce.com`; the crawl step has no `on_error`, so one refusal discards the stage.
- Three attempts, **1h37m**, then `failed`. **`bugs_open/376`** — filed, evidenced, severity HIGH.
- Re-submitted at 19:23 as the natural recovery. It queued (proving `bugs_open/326`'s fix live) and
  **died the same way again**.

## 2. State right now

| thing | state |
|---|---|
| site row | `16784842-f7d8-4467-bb5b-eb1fb5c1caba`, `active`, `build_status=pending` |
| specs | 2 generations of identity/classification/content_direction/design_intent; run 2 `is_current` |
| pages | **none.** `needs_strategy` was never created and **nothing can create it** (376 §2a) |
| apex | still the 9-byte `Not found` — nothing was ever deployed |
| work items | 2× `needs_domain_research` complete · 1× `needs_vertical_research` failed · 1 more cycling · 1 `site_unreachable` detected (inert) |

## 3. The one thing to know before you touch it

**Re-submitting will not build a site.** It will re-run the classifier, reach hop two, and die on
the same host. That is not a flake — it was measured four times (376 §4/§4a). Re-run it only if you
are testing the front door or gathering more 376 evidence, and say which.

**`376` must be fixed, or this domain built in a vertical whose exemplars are scrapable, before this
lane can measure anything past hop two.** Everything the lane exists to test — the 311 after-test,
the 260 after-test, tool pages, nav, the served artefact — is downstream of a hop that never runs.
**The after-test harness is written and validated and has never been used in anger:**
`after_test.sh` (this session's scratchpad; promote it into the lane dir when a build gets far
enough to need it). Its collateral check is proven discriminating — it returned CHANGED on stale
baselines and UNCHANGED on fresh ones the same morning.

## 4. What was found on the way, and where it lives

- **`bugs_open/376`** (new) — the exemplar-crawl kill. §4a has the deeper mechanism: the
  `competitors_found` branch has never fired, so site-specific input never reaches the selection.
  ⚠ **bounded, and the bound is that the question is UNANSWERABLE, not that the sample is small.**
  The 0/4 comes from `orchestration_states`, which reaps on a **24-hour clock** (oldest row measured
  at exactly 24h). The durable tables show **32 `needs_strategy` items across 27 sites since
  2026-04-02**, and the historical *selections* are reaped with the rows — so "has this branch ever
  fired?" cannot be answered from that table at any sample size. ⚠ **This handoff's own first
  version said "the agent has only ever run 4 times" — corrected 2026-08-23 19:40Z.** Before any
  `count(*)`/`min()`/absence claim here, `grep orchestration_states LANDMINES.md` (an entry has
  existed since 2026-08-02) and print `min/max(created_at)` beside the count.
- **`bugs_open/326` corrected and now FIXED for the build chain.** Its filed root cause (dedup index)
  was wrong; the real one is `writeWorkItem`'s 3h two-strike block. Migration **572** declares
  `recurrence_expected` on all five build-chain hops. **Never hand-rename `item_key`s** — that
  instruction is retracted in all four documents that carried it.
- **The 311 after-test baselines were stale before we started** — all eight moved 2026-08-20 under
  `bugs_open/283`. **Re-pin your own before dispatch; never use a handed-down pin.** LANDMINES.
- **Two measured route properties** (route handoff §10): time-to-first-agent is queue depth ÷ ~90s
  (24m52s here), FIFO by item age with `priority` only a tie-break; and a site is serialised to one
  in-flight item.
- **The classifier is reproducible** — two independent runs, identical structured verdict including
  confidence to 2dp; only free-text tags drift. Usable as a fixture for the structured fields only.
- Full account: `NOTES_loanzy_uk_example_site.md` (2026-08-23 entries) ·
  `SUMMARY_2026-08-23_the_route_dies_at_hop_two_on_a_website_it_is_not_allowed_to_read.md` ·
  `README_where_we_are.md` · route defects `HANDOFF_2026-08-19_fixing_the_one_shot_route.md` §9-10.

## 5. Falsifiers for THIS handoff

- `376` moving to `bugs_closed/`, or the crawl step gaining an `on_error`.
- A `needs_strategy` row appearing for this site (it would mean a producer exists that the
  2026-08-23 sweep did not find).
- The apex serving anything other than the 9-byte `Not found`.
- Firecrawl beginning to support `thespruce.com` — the whole failure rests on one host's blocklist
  status, which is not ours and can change without notice. **Re-check it before re-running:** the
  refusal is visible in `site_work_items.error` as `WEBSCRAPE_ERROR`.
- A second vertical exercising `vertical-exemplar-researcher` — that would widen §4a's evidence base
  beyond this one domain, in either direction.
