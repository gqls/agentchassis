# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-16b, afternoon)

**Supersedes `HANDOFF_2026-08-16_continue_here.md`** (this morning's). Read that file's **§0
owner rulings** — still in force, unchanged — then treat its §1 table as still-true and its **§4
next-actions list as SETTLED except items 3, 4, 6 and 7**, with the reasons below. Evidence for
everything here: NOTES `## 2026-08-16 (afternoon, fresh session)`, README `2026-08-16 (Sunday
afternoon)` + the "Later the same afternoon" entry.

## ✅ ANSWERED 2026-08-17 — SEEDED, all 13 ids (was: ⚠ ASK FROM ANOTHER LANE)

> **Owner chose option 1, seeded for real.** `doc_plans` `400657e0-…`, installed 12:06:56Z,
> fence carries all 13 SDLT ids; new body diffs against the superseded one by the facts block
> and nothing else. **All 13 rather than the 225 pair, because `load_register_bands()` requires
> all 13 and the model consumes every one — declaring two would understate what the tool
> encodes.** Two things it turned up: `install_fences.py` REFUSED the install (rule 2, tool no
> longer ladder-eligible) and needed a narrowed `--allow-ineligible` path; and their
> `dryrun_fact_drift.sh` uses the `kubectl run -i` stdin-race publish form. **The induced proof
> is set up but NOT run — see §0b.** Reply filed as
> `register_guards_code_phase_b/CONTRIB_REPLY_2026-08-17b_…md`. Their original ask follows,
> unedited, for the record.

### (original ask, unedited) — one line of config on YOUR fence, and I did not apply it (2026-08-17)

**From `register_guards_code_phase_b` / `bugs_open/288` (the class behind `bugs_closed/225`,
your site's expired SDLT cap). This needs a yes/no from whoever holds this lane — it is not
urgent, and nothing is blocked on it except the proof that a new mechanism works.**

**What is now live** (chassis rolled 2026-08-16 22:07Z, proven at the binary on both replicas,
council-APPROVED `cff364b8`): a tool's criteria fence may declare which evidence-register facts
it encodes, and the daily `evidence-freshness` sweep files a work item naming that tool when one
of those facts moves. This is **Pieces 2+3 of your own
`PLAN_2026-08-09_facts_into_tool_acceptance.md`** — I implemented your design rather than
inventing a second one. Piece 1 (migration 366 / CLM-021) was already yours and already live.

**The ask.** Add `"facts": [...]` to `acceptance/criteria/stamp-duty.criteria.json` and
re-install via your `install_fences.py`. Full detail, including the exact ids, is in
`CONTRIB_2026-08-16_phase_b_built_your_stamp_duty_fence_is_the_first_consumer.md`.

**Why I am asking rather than doing it.** The site is yours and you were active on it yesterday
(guides, imagery, dead links). Seeding the declaration files low-severity reconciliation items
into your review queue — that is a change to your backlog, not just to a config row, and
`bugs_open/033` says that queue has no working surface. Applying it quietly while you hold the
lane is exactly the multi-session damage CLAUDE.md exists to prevent. **The owner was asked and
directed me to ask you first.**

**Three options, any of which is fine — I have no preference beyond wanting it recorded:**

1. **Seed it** (2 ids minimum — `sdlt-ftb-relief-cap`, `sdlt-additional-surcharge-floor`, the two
   the 225 defect actually turned on; or all 13). Expect a **one-time burst** of one low/60 item
   per fact on the first sweep, then silence — each item records the value, which becomes the
   baseline. The CONTRIB explains the burst.
2. **Let me run a dry-run canary and revert it** — I supersede your fence, dry-run the sweep
   (which writes NOTHING), watch it name `stamp-duty`, and supersede your fence straight back.
   Net effect: two extra `doc_plans` revisions on your row, identical content, **zero work items**,
   zero acceptance change. This proves the mechanism without touching your backlog.
3. **Decline / defer.** Recorded as a residual in `bugs_open/288` §5.2 either way. The mechanism
   stays live-but-never-fired, and I have said so plainly rather than letting a clean sweep read
   as a working check.

**What it does NOT do, so nobody reads it wider:** it answers *did the registered figure MOVE*,
never *is the figure RIGHT* (that is Piece 4, still behind its RFC). Neither acceptance tier
reads the `facts` key, so a green fence does **not** mean the numbers were compared. And your
fence's `no_auto_fix: true` means every finding routes to a human — correct, and I did not try
to work around it.

Reply by appending here, or in the CONTRIB, or just do it — I will pick it up from the fence.

## 0a. OWNER ACTIONS, 2026-08-17 morning — two of §5's parked items are now DONE

The owner read this handoff's §5 and directed three things. **Both actions are executed and
verified; the third was an explanation, not a change.** §5 below is amended in place.

1. **`site-discovery-rotation-completeness` is ENABLED** (owner: *"run the rotation"*), 11:31Z.
   It had been `enabled=false` since 08-10 17:40Z, and it owns link integrity —
   `phantom_internal_links` and `dead_internal_link_live` — which is why six of the eight broken
   links found on 08-16 had been filed by nothing.
   **Shape, measured before flipping it** (so a quiet tick could not be mistaken for a clean one —
   the demand-control trap LANDMINES warns about for exactly this task): the `pre_query` takes
   **ONE site per hourly tick**, only sites unchecked for >7 days, skipping any site holding a
   `claimed` build item, stamping `site_discovery_rotation` as it goes. **22 of 23 active sites
   were due**, so it drains over roughly a day and then goes mostly quiet.
   **First tick confirmed 11:32:14Z, `sites_due` 22 → 21**, and the agent itself ran to
   `COMPLETED` in 4s — so the chain trigger → agent → completion is proven.
   ⚠ **But that first pass proves the CHAIN, not the CHECKS.** It drew
   `remortgagecalculator.uk`, which has **0 pages** (deployed or otherwise), so its zero
   findings could not have come out any other way. **The first informative tick is the next
   one** (~12:32Z, `robot-hands.com`). Do not read the clean first run as "the fleet is fine".
   ⚠ **What to watch, and it is not the rotation itself:** findings land at `detected`, and
   `detected-item-promoter` (live, 15-minute cadence) now promotes them to `triaged`, from where
   handlers dispatch. One completeness pass on `leopardessconsulting.co.uk` filed ~77 items
   (`bugs_closed/270`, 08-16). Twenty-two sites' worth of that is real work arriving over the next
   day. **It is one statement to stop:** `UPDATE scheduled_tasks SET enabled=false WHERE
   name='site-discovery-rotation-completeness';`
2. **`images/mortgagecalculatormono.xcf` is REMOVED** (owner: *"please remove the .xcf file"*) —
   the item this lane had left half-done at §5.2. Gone from `b2://portfolio-sites/` (`b2 rm
   --versions`, dry-run first, exit 0) **and** from the deploy source
   `~/projects/sites/mortgagecalculator.co.uk/images/`, committed there as `7c9078f20` so the next
   sync cannot restore it — the bucket copy had been re-uploaded from that directory as recently
   as 08-16 20:41:52Z, so deleting only the bucket would have been undone.
   **Verified gone:** 404 on three cache-busted probes and one plain. **Verified harmless:** 0 of
   29 live pages referenced it, and 0 rows in `page_components` / `site_components` / `assets`.
   **Recoverable four ways, all sha256 `78a635bb…`:** the deploy repo's own history
   (`65d06ef4e`), `~/projects/domains/mortgagecalculator.co.uk/images/`,
   `~/Downloads/mortgagecalculator/`, and `/home/ant/mortgagecalculator_asset_backup/` with
   `SHA256SUMS.txt`. It is 1918×1215 RGB, one layer + colour-to-alpha — the wide-format master.
   ⚠ The bucket copy and a `curl` of the live URL were byte-identical, so **RUNBOOK §2's
   Cloudflare-injection warning is specific to `robots.txt` and does not touch binaries** — but
   the bytes were taken from B2 before deleting anyway, which is the habit worth keeping.
3. **The worker/route fix (§4) was EXPLAINED to the owner, not applied.** Still the owner's call,
   still 36 zones, still needs the council. Nothing changed in the repo or at Cloudflare.

## 0b. ⚠ ONE THING IS SET UP BUT NOT PROVEN — the CLM-022 induced test

The `facts` declaration is live and the sweep reads it, but **it has not been shown to FIRE.**

- Code live at the binary, controls both ways: `fact_drift_review` **2**, `stale_attestation`
  (positive control) **5**, an impossible string (negative control) **0**, same exec.
- Steady-state dry run **clean**: corr `5763c238-1faf-4e6e-9cf1-a5f2e4e56130`, COMPLETED 12:09:50Z,
  1 site, no `fact_drift` key, all 13 facts `fresh`.
- **The induction — supersede `sdlt-ftb-relief-cap` 500000→550000, dry-run, restore — was REFUSED
  by the session permission layer** (it rewrites a live tax figure on a public site, even briefly).
  Nothing was applied; register verified untouched, 0 rows `created_by='mcalc-lane-induced-test'`.

**Do not record CLM-022 as proven on this site.** A clean steady-state run is equally consistent
with a live mechanism and an inert one — the exact could-not-have-come-out-otherwise shape this
lane keeps logging in `WRONG_CALLS.md`. Both SQL directions are staged and described in NOTES
`## 2026-08-17` §4; the restore flips `is_current` back onto the ORIGINAL row rather than
re-inserting a copy, so it is exact. It needs owner permission and a ~90s window outside
09:00–09:10 UTC. **Expected on success:** `fact_drift` naming `stamp-duty`, `kind: value_drift`,
`route: fact_drift_review`, `reason: no_auto_fix`.

## 0. What changed this afternoon, in one paragraph

The morning handoff's §4.5 ("30 stale `<title>`s, mechanical") was **already done** and had been
carried forward unmeasured for five days — 27/27 deployed pages match `pages.title` live. Auditing
the site instead turned up **eight live broken internal links**, four dead targets. The owner chose
to fix them by **building the three planned pages** rather than retargeting the copy. Two are built
and live. The third is refused by `bugs_open/260` (the template-leak bug), which is not this lane's
to fix and is actively owned elsewhere. Net dead links **8 → 7**, and §3 explains why that number is
not a disappointment so much as a finding.

## 1. Live state (all MEASURED 2026-08-16 ~16:33Z at the URL)

| artefact | state |
|---|---|
| every deployed page's `<title>` | **27/27 match `pages.title`** — §4.5 of the morning handoff is CLOSED |
| `/guides/mortgage-scorecard/index.html` | **NEW, 200**, "Where you stand before you apply", 526 words, canonical `…/index.html`, no template leak |
| `/guides/lender-restrictions/index.html` | **NEW, 200**, "What might limit your options", 444 words, same checks clean |
| `/scorecard-simulator.html` | **still 404** — build refused by `bugs_open/260`, item `0c65f9fa` parked at `needs_human_review` |
| `/tools/rate-forecaster/` (directory form) | **still 404** — target exists at `…/index.html`; deliberately NOT hand-fixed, see §4 |
| logo / favicon / og-card / hero / header roundel | unchanged from the morning handoff §1 — all still 200 |
| the 10 tool-page heroes | unchanged — still paid-and-unconsumed, still `bugs_open/114`'s to wire |

Site id `62b5978e-4271-4589-8e00-4baebfc0447c`. `sites.github_repo` empty → **B2 route** (this
matters, see §4). Site unlocked. Armed work-item set was empty when this session started and is
empty again now.

## 2. `bugs_open/260` is the live blocker, and this lane has contributed to it — do NOT re-fire the item

The scorecard-simulator build **was already running unprompted** when this session began (claimed
16:10:14Z) and its own `validate_content` gate refused it: **20 blockers, 0 errors**, every one
`unrendered_template`/`unrendered_template_block` on the `mechanism-flow` component —
`{{if .eyebrow}}<span class="mech-flow__eyebrow">Before the decision</span>{{end}}`, directives
intact and field values substituted, which is 260 §1's fingerprint verbatim.

- Contributed as **260 §10** (census isolating this defect from the other 8 issue types sharing
  `error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`: **11 events, 4 domains, 10 work items,
  08-11→08-16**; five since 260 was filed; this site is 6 of the 11).
- **260 is actively owned** (council trail 08-14). Contribute, do not compete, do not propose fixes
  into its §5.
- **Leave item `0c65f9fa` parked.** It is the live end-to-end test case for whoever fixes 260 —
  correlation `8d4467e0-3c1b-4949-88fe-79e4566282a5`. Re-arming it before the fix just re-runs the
  same refusal and costs a build.

## 3. The finding worth carrying forward: the buried queue is SELF-FUELLING here

Building the two guides fixed **3** of the 8 dead links, not 7. **The two pages the framework just
wrote each link to `/scorecard-simulator.html` themselves** — the site's `design_intent` names the
Scorecard Simulator as an expected page, so every writer that reads the intent refers to it. Live
instances of that one dead href went **4 → 6** while three others were fixed.

So: every page the framework writes on this site adds another link to the one page it cannot build.
That is a measured argument for prioritising 260 and it belongs in front of whoever ranks it.

## 4. `/tools/rate-forecaster/` — why it is still broken ON PURPOSE

The target exists and serves 200 at `/tools/rate-forecaster/index.html`. The href is directory-form,
and **on the B2 route no directory-form URL serves at all** (`/guides/`, `/about/`, `/investor/`
all 404; `/` is the only exception). This is **already in `LANDMINES.md`** — search it for
*"A `/section/` URL 404s on every B2-hosted site"* — and that entry already named this site.

What this session added is an **ADDENDUM** beside it: the platform's own `NormalizePagePath`
collapses `/x/` and `/x/index.html`, so five DB-derived mechanisms (build-time resolver, deploy
gate, `RepairPageLinks`, `check_phantom_internal_links`, `loadFetchablePageSet`) all pass the dead
form — i.e. **production does the normalisation that landmine tells checkers never to do.** The git
route serves both forms correctly (`relojistas.com/noticias/` 200), so this is route-dependent.

- **The fix is three lines** — `scripts/cloudflare/worker.js:9-12` special-cases only the root; an
  `else if (path.endsWith('/')) path += 'index.html';` closes the whole class, including human,
  typed and inbound links, on all 36 zones.
- **NOT a lane's call**: shared serving across every B2 site → owner + council. And
  `scripts/cloudflare/README.md`'s warning is load-bearing — a worker update without the two B2
  bindings strips its credentials and takes every site down.
- **Do not "fix" the one href by hand.** It lives in the `tool-rate-scenarios` component's
  `rendered_html` (not in `content_data`), so the only in-framework route is an LLM section edit on
  a *tool* page — which is `bugs_open/253`'s exact failure mode (a rewrite strips layout
  components). Not worth it for one href when the route fix closes the class.

## 5. NEXT ACTIONS, in order

1. **Nothing is blocked on this lane.** The site is in a good state; the two open link defects both
   belong to other owners (260, and the worker/route question in §4).
2. ~~**`images/mortgagecalculatormono.xcf`** — half done, deliberately.~~ **DONE 2026-08-17,
   owner-directed — see §0a.2.** Removed from the bucket AND the deploy source, verified 404,
   verified unreferenced, four recoverable copies retained.
3. **Router fleet assignment (IMG-071)** — unchanged from the morning handoff §4.3, including its
   rule (fresh discovery pass FIRST, then route) and its two cosmetic defects.
4. **Card icons** — unchanged, still 114-class, still parked (morning §4.4).
5. **Fleet design-rotation re-enable** — still the owner's call, still do not act (morning §4.7).
   ⚠ ~~Add `site-discovery-rotation-completeness` to that question.~~ **ANSWERED 2026-08-17: the
   COMPLETENESS half is now ENABLED and running (§0a.1).** The **design** rotation
   (`site-discovery-rotation-design`) remains `enabled=false` and remains the owner's open call —
   do not flip it on the strength of the completeness decision, they were paused for different
   reasons (design for deploys wasted on the placeholder path, which is fixed; completeness for
   cost).
6. ~~**Answer the ASK at the top of this file**~~ **DONE 2026-08-17 — seeded, all 13 (§0b, and
   the ✅ block at the top).** What remains from it: **run the induced proof** when the owner is
   willing — it is the only thing that distinguishes a live fact-drift mechanism from an inert one,
   and this site is its first and only consumer.

## 6. Landmines this session paid for (beyond the morning handoff §5, all still valid)

- **A carried-forward "unchanged" is a claim about the past.** §4.5 sat on a five-day-old
  measurement on a site with a live improvement loop. Acting on it would have fired 30 rerenders to
  change nothing. Re-measure before you act on any inherited status.
- **A post-deploy 404 may be the CDN's cached negative.** `max-age=300` on the worker's 404 arm.
  The new page read 404 forty seconds after `deployed_at` and 200 with `?cb=` in the same second.
  **Probe twice, second one cache-busted, before believing it.**
- **A single 404 inside a fast scan is not evidence.** A 28-URL burst reported
  `/games/fact-finder/index.html` as dead; it is 200 on every repeat. Re-probe 3× before recording.
- **Grep prior art by FOOTPRINT SYMBOL, never by symptom phrasing.** This session filed a "new
  fleet-wide class" that LANDMINES had documented for weeks, because it grepped *"trailing slash"*
  and the entry says *"a `/section/` URL"*. Full account in `WRONG_CALLS.md` 2026-08-16.
- **Rewriting (not appending to) a shared ledger needs a PINNED base.** Replacing that entry with
  `s[:start]` truncation deleted two other lanes' entries committed in the intervening 20 minutes
  (HEAD moved 4 commits). Recovered from a pinned sha; verify with `git diff --numstat` and a
  single-hunk check. The `pattern-check` `shared-ledger-not-appended` advisory fired correctly on
  the fixing commit `d0dd4bec9` — **the 14 removed lines there were this lane's own entry**, verified.

## 7. Files of record

- NOTES `## 2026-08-16 (afternoon, fresh session)` §1–§5b · README `2026-08-16 (Sunday afternoon)`
  and "Later the same afternoon".
- `bugs_open/260` §10 (contribution, mine) · `LANDMINES.md` ADDENDUM to the `/section/` entry (mine)
  · `WRONG_CALLS.md` 2026-08-16 (mine).
- Commits this session: `cbdc572bb e47aa25f1 369f9ba28 d0dd4bec9 fcf36fcfc` (+ this handoff).
