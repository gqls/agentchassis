# PLAN 2026-09-04 — farmerinsurance.uk gets its own lane

Site: `farmerinsurance.uk` — site_id `99cae989-2413-430d-b026-59dfeeb638c0`.
Opened 2026-09-04 on the owner's instruction ("please pick up the farmer insurance thread here").

## Why a lane, and what it is NOT taking from anyone

Farmer was built by the **loanzy lane** as the second run of the greenfield route, and that
lane still owns the ROUTE. It is the same split the estate already made for
`lendzy.co.uk` on 09-02 ("this is lendzy's own lane now"): the route is one workstream, the
site is another. This lane owns **the site as a served artefact** — what a visitor sees, its
queue, its specs, its content — and nothing about the greenfield route's machinery.

Explicitly NOT claimed here (each has a live owner, and this lane contributes rather than competes):

| aspect | owner | this lane's part |
|---|---|---|
| greenfield route, record-mode council, growth posture | `loanzy_uk_example_site` | consume; report site-level effects |
| stage-2 copy edits + banned register | `copy_quality_two_stage` | farmer's 14 pending proposals are THEIR batch; this lane chases the review, does not apply it |
| news feed region default | `bugfix_316_news_feed_ordering` | supply farmer's live evidence |
| entity-directory build handler | `bugfix_206_directory_build_handler` | supply farmer's live evidence |
| carousel default / component library | `staged_component_build`, offer-analysis lane | supply farmer's measurements |
| evidence register method (FCA §8) | `lendzy_co_uk` | farmer's register is filled (7 facts / 5 bans, 09-02) |

## Where the site actually is (measured 2026-09-04, this session)

`[MEASURED 2026-09-04 12:xx BST]` 18 active pages, all serving 200, with an invented-path
control returning 404 — so HTTP readings on this domain are informative (the parked-domain
landmine does not apply). 21 pages archived by the 08-31 tool cull; all 21 still 404.
Every one of the 27 distinct internal link/asset targets across those 18 pages resolves 200.

## Phase 1 (this session) — separate the site's REAL defects from its stale queue

The queue reads as ~274 open work items. Most of the volume is not a live defect:

1. **Stale/moot buckets, measured against the artefact** — see NOTES for the per-bucket
   evidence. Decide, per bucket, whether the row is FALSE (target fixed), MOOT (page
   archived) or TRUE, and whether the producing detector can retract it at all.
2. **The two defects a visitor can see today**, both on the homepage:
   - the **health-insurer directory** — a UK private-medical-insurance directory on a farm
     insurance hub, root-caused this session to one map entry (bug filed; 090 fired);
   - the **US news feed** — the owner's 08-31 finding #3, still live four days later.
3. **The 14 stage-2 copy proposals** parked for the owner's batch review since 09-02.

## Phase 2 (next) — the owner's six 08-31 findings, re-measured rather than re-asserted

`OWNER_REVIEW_2026-08-31_farmerinsurance_first_review_and_routing.md` (loanzy lane) routed six
findings to five lanes. Four days on, nobody has re-measured farmer to see which landed. This
lane's standing job is that re-measurement, at the artefact, on a stated cadence — and to say
so plainly when a routed ask has not moved.

## Decisions this lane expects to put to the owner

- **The health-insurer directory**: retract it, or replace it with the agricultural-broker
  directory the growth_path actually specified (a new directory KIND — bigger work).
- Whether farmer's growth posture stays UNSET (it is neither held nor released today).
