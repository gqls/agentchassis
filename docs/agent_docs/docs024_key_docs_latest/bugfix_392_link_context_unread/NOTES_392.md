# NOTES — bugfix_392 (append-only, newest at the bottom)

## 2026-08-25 — session `bugs_open/387` opens the lane

- **Ownership before touching anything.** `who-owns.py 392` → no owning workstream; no
  `bugfix_392*` dir; `git log --since='45 minutes ago'` showed 104 fleet commits and **zero**
  touching 392/393/394 or `prepare_link_context`. The 358 lane that filed it has declared itself
  complete ("Nowhere, by design — this lane's work is complete"). Routed at
  `bugfix_092_writer_link_constraints`, quiet since 2026-07-31. ⇒ unowned, resumed.
- **No SUMMARY today, deliberately.** The five headings would say "we have just started", which
  is the cadence rule's own test for not writing one. First SUMMARY when the canary is proven.

### What the evidence says (all `[MEASURED 2026-08-25]`, queries in RUNBOOK)

- 2 `LINK_CONTEXT_UNAVAILABLE` rows ever. **Both already healed or never published** — so the
  bug's stated severity ("three pages… medium-high") rests on damage that no longer exists. This
  is a latent class defect. Said plainly in the bug file's correction block, not softened.
- 411/736 deployed pages carry zero writer-authored prose anchors; 187 on prose page types;
  **48 of those 48 owned pages are link-less**, which is a striking enough ratio to be its own
  finding and belongs to the 396/333 lanes, not here.
- `orchestration_id` on the log row resolves to the exact page — verified end to end on **both**
  rows. The bug file says only `site_id` is recorded, which is true of `context` and false of the
  row.
- `page-rerender` neither spawns the writer nor knows `edit_live`; `page-content-writer` is the
  only agent fleet-wide running `prepare_link_context`. **392's own fix candidate 1 is wrong.**
- `internal-linker` is LIVE (7 completions to 2026-08-24) while `check_orphan_pages.go:14` still
  says "(future handler)".

### MISSTEP (mine), caught in-session, and it is the reason the census exists

I first measured link-lessness on `page_components.rendered_html` and got **140 of 737**. That
instrument counts template nav, hero and CTA links, so a page whose writer emitted nothing still
reads 2–3. Measured on the writer's own output (`content_data`) the answer is **411 of 736** —
nearly triple, and it includes 31 blog-posts that the first instrument called perfect (0/164
there). I noticed only because the per-page-type split had no mechanism behind it: "every blog
post is fine" is not a thing a writer does. Logged in `WRONG_CALLS.md`.
**The check:** before trusting any "does this page contain X" census, open one known-positive and
one known-negative row and look at where X actually lives. Here, hero/CTA links live in
structured fields (`cta_url`, `link_url`) and carry no `href=` at all.

### Three claims of mine REFUTED by the third fable planner, before any code was written

I ran three planners: framework-first, reuse-first, and one asked only to stress the design. The
third earned its cost by refuting the other two.

1. **My canary induction was self-blinding.** I had settled on setting an unparseable `site_id`
   in the step config, having read `resolveLinkContextSiteID` and confirmed "an explicit config
   value wins". It does reach the degrade — but the row is then written with the garbage site
   context, and the reader arm selects by `site_id`. **The induction would have proved the
   writer and silently never exercised the reader**, while looking like a clean end-to-end pass.
   Same shape as the missing control that filed bug 387. Replaced with a site-scoped opt-in hook
   that shrinks the query's context deadline *after* site resolution.
2. **I had the owned-page door backwards.** I wrote that `writeWorkItem` would park items filed
   against owned pages — an argument for excluding them as tidiness. In fact neither
   `internal-linker` nor `page-build-handler` declares `refuse_owned_page`, so the door does not
   fire at all; the refusal comes late at `SavePageSectionsAction`'s `OWNED_PAGE_GUARD`, **after
   the LLM spend**, terminating `wont_fix`. Excluding owned pages at the query is therefore
   load-bearing, not hygiene.
3. **My item_key would have collided.** I proposed sharing `internal-linker`'s
   `internal_link:<page>` namespace and called the co-dedup a feature. `check_orphan_pages`
   already mints `needs_links:<name>:<site>`, and `idx_swi_dedup` is UNIQUE on
   `(site_id,item_key)` with **no item_type column** — so an inbound-orphan finding and an
   outbound-absence finding on one page would share a slot and one would vanish. They are
   different defects and both must stay actionable.

**The transferable lesson:** two planners asked to design agreed with each other and with me; the
one asked to attack found three defects in twenty minutes. Where a mechanism will be trusted to
run unattended, the adversarial seat is not a luxury round.

### Design, as it stands going into implementation

One new discovery check in `platform/orchestration/actions/discovery_checks/`, sibling to its own
inverse. No new service, no new image, no cron slot, no handler, **and no change to any live
agent** — the repair route (`content_rewrite` + `spec.mode='edit_live'` → `page-build-handler` →
`page-content-writer`) already exists and runs ~30/day, and because the repair re-runs the writer
it gets a fresh link allow-list and picks its own targets. Full design and the corrections block
in `PLAN_2026-08-25_392.md`.

⚠ `content_rewrite`'s 14-day health is 93 complete / 53 wont_fix / 45 failed / 21
needs_human_review — **~21% fail or are refused.** Any claim that a filed item equals a repaired
page is wrong, which is why acceptance is measured at the served page.

## 2026-08-25 (later) — two peer corrections, one of which refutes a correction of mine

- **The `webdesign-tool-rebuild` lane asked the right scoping question**, and my instinctive
  answer would have been wrong. My gate is `page_type IN ('blog-post','guide','content')`, which
  I assumed excluded tool pages. **It does not, fleet-wide: 103 deployed `tool-%` pages are typed
  `blog-post`.** Measured on their own site the gate still holds up — 62 owned + 3 generic typed
  `tool` are excluded, and of the 40 typed `blog-post` only 2 are link-less. I checked what those
  2 carry before calling them in-scope: slots `article-body, call-to-action, hero`, **no tool
  component**. They are prose guides about tools, which is the defect, not a false positive.
  Offered them a structural exclusion (anything carrying a tool-level component) if they disagree;
  not adding it speculatively, because here the structural and type rules select the same pages.
  Item key prefix settled at their request: **`no_outbound_links:<page>:<site>`** — self-describing
  and distinct from both `internal_link:` and `tool_crosslink:`.
- **MISSTEP (mine), caught by the `bugs_open/333` lane: my own correction #2 was wrong.** I wrote
  that the owned-page door does not fire for these handlers and that the refusal arrives late,
  after LLM spend. **`page-build-handler` declares `refuse_owned_page: true` and is the only live
  declarer fleet-wide**; the door fires at write time; 40 `content_rewrite` rows sit parked with
  the handler cleared, the earliest stamped `2026-08-24 19:19:12Z`. What my planner described was
  the world **before 08-24 19:19Z** — the ~83 late deaths are the census that MOTIVATED the door.
  I verified the declaration and the parked population myself before accepting the correction,
  having taken a planner's word once already today.
  **The check:** a fact about a mechanism is a fact about a *date*. Before repeating "X does not
  do Y", ask when the claim was true — this one had a five-day shelf life and I was inside it.
  **Consequence:** the design changes — do NOT exclude owned pages at the query; carry `page_id`
  so the door can see them; let the 48 park as recorded demand for `bugs_open/277` instead of
  vanishing into a query only I have run.
- **Correction to this lane's own README:** I called the 48/48 owned-page finding "a separate
  defect". 333's reading, which I accept, is that it is the measured shadow of the ownership
  guard working as designed — owned pages are excluded from the writer pipeline at selection, so
  writer-authored prose links structurally never land there. It is demand evidence for
  `bugs_open/277`, not a new defect. Fix the wording next time this lane writes prose; do not
  silently edit the README entry above.

## 2026-08-25 (third pass) — a stale pointer I would have shipped, and a count that aged in an hour

- **MISSTEP (mine, inherited then verified): I wrote `bugs_open/277` into the PLAN as the home for
  the 48/48 finding, and 277 has been CLOSED since 2026-08-22.** It is in `bugs_closed/`. I took
  the pointer from the 333 lane without checking a bug number's existence — the one check on this
  estate that costs an `ls`. They caught it themselves (their peer corrected them) and relayed it
  before I shipped code citing it.
  **The check:** a bug NUMBER is a claim about a file's location, and files move between
  `bugs_open/` and `bugs_closed/` daily. `ls bugs_open/ bugs_closed/ | grep ^NNN` before citing
  one, every time — and note the memory index already carries "a closed blocker keeps being
  obeyed" for the inverse failure (deferrals pointing at a bug that closed). This is that lesson
  with the arrow reversed: a POINTER written at a bug that has closed.
  **Correct home:** register entry **WII-028**
  (`docs/agent_docs/docs026_concept_register/register/work-item-integrity.md:402`), where the
  48/48 now sits as a dated demand-evidence block beside the 40 parked rows. There is no open bug
  for owned-page content repair.
- **MISSTEP (mine): my "11 named-handler deferred rows are pre-door owned parks" was wrong on
  cause, and the number was stale when I sent it.** Provenance is `voiceh-rollout` ×9 on GENERIC
  pages (08-08) and `apis-uk-bees-lane` ×2 (08-24, no live page row) — nothing to do with the
  door. It was still worth sending: it gave 396 a `created_by` lead on 11 of its 114 untraceable
  rows. But re-measured a few hours later the same population is **52**, because
  `required-fields-missing-handler` (28) and `tool-generator` (13) filed in between.
  **The check:** a count of a live queue is a count of a MOMENT. The estate's 2026-08-22 ruling
  says a census carries the date it was taken; on a queue that moves this fast the date is not
  enough — carry the timestamp, and re-run before quoting rather than forwarding your own figure.
- Both of these were caught inside an hour by a peer lane rather than by me, and both were in
  material I had already committed. That is two of my last three errors found by someone else's
  correction rather than by my own check.

## 2026-08-25 (fourth pass) — the 52 was mine, not the world's, and the cross-lane thread closes

- **MISSTEP (mine, the fourth today, and the third made while correcting someone else): my
  "52, not 11" was a conflation, and the truth is still exactly 11.** `handler_agent` is
  `NOT NULL DEFAULT ''::text` — a cleared handler is the EMPTY STRING — so my
  `handler_agent IS NOT NULL` filter swept the owned-page door's own parked rows into a count of
  named-handler rows. Split with the discriminator printed rather than filtered:
  `required-fields-missing-handler` 28 and `tool-generator` 13 are door-parked (handler `''`,
  `error` leading `OWNED_PAGE_GUARD`); `voiceh-rollout` 9 and `apis-uk-bees-lane` 2 are genuinely
  named. `bugs_open/396`'s 114 is not drifting, and the warning I asked the 333 lane to relay was
  itself the error.
- **⚠ I had the fact in my own output and overwrote it.** I ran
  `coalesce(handler_agent,'(cleared)')` on those rows and it printed *blank* rather than
  `(cleared)` — only possible for `''`. I then told another lane "all 40 have `handler_agent`
  null" from a query whose output said the opposite, because the word "cleared" had primed me for
  NULL. The instrument answered; the reader substituted the primed word.
- **Second mechanism, and the more useful one:** the door clears the handler but PRESERVES
  `created_by`, so a parked row wears the name of whatever filed it and reads as live, named work
  in any `GROUP BY created_by`. **A population defined by an ABSENCE cannot be censused by a
  field that is still present.** Discriminating predicates now in RUNBOOK §8.
- **And the rule that would have caught it earliest:** "re-running the same query" must mean
  re-pasting the same TEXT. Mine was re-typed from memory of *what it measured*, which is where
  the filter went — and a remembered query is a NEW query wearing the old one's authority, so a
  changed result reads as a changed world rather than a changed instrument.
- **The landmine already existed and did not reach me.** The column fact sits in the
  `mistyped_deployed_page` entry, whose footprint was widened 2026-08-02 *because it had already
  bitten a second lane*. Mine is the third occurrence. It missed me because I greped the
  situation (parked `deferred` rows) and the entry is titled after someone else's situation — a
  representation fact is discoverable only by the COLUMN's name, and a table-and-column footprint
  can never match the SessionStart hook's dirty-path test. Filed the new manifestation
  (`529a31c9a`, verifier armed) and added this as a fifth trigger to the
  `grep-landmines-for-your-symbols` memory.
- **Cross-lane thread with `bugs_open/333` CLOSED.** They widened the footprint of the door entry
  where the parked row's identity lives to carry `handler_agent`/`created_by`, with a dated
  third-occurrence note and a cross-pointer to my census entry; `landmines-sync.py --check`
  reports **in sync**, 829 entries, 0 orphaned. Four corrections between the two lanes today,
  both directions, every one caught before a doc inherited it.

### Standing tally for whoever picks this lane up

**Four wrong calls in one afternoon, three of them made while correcting someone else.** That is
not bad luck; correcting mode supplies confidence and suppresses the check. The design in
`PLAN_2026-08-25_392.md` is sound and is now peer-verified in three places — but it has needed
four correction blocks to get there, and **no Go code has been written yet**. Read the PLAN's
correction blocks bottom-up before trusting anything above them.

## 2026-08-25 (fifth pass) — the code is written, tested and committed

- **Shipped in `924079c94`** (6 files): the `missing_prose_links` discovery check + its tests, the
  lifecycle-posture declaration, the writer's `page_name`/`page_id`/`domain` enrichment + its
  tests, and the registry flip to `consumed`. Register entry **LNK-039** + index row in
  `073feba9d`. Council **SUBMITTED**, corr `7d923ff6-3810-4f2b-9000-e02df68a6b9e`, trailer
  `Council-Submitted:` — never `Council-Reviewed:` on a verdict not yet read.
- **Design as built differs from the PLAN in one place, and it is the right way round.** No
  `internal-linker` migration: the repair re-runs `page-content-writer`, which runs
  `prepare_link_context` itself, so the writer gets a fresh allow-list and picks its own targets.
  No live agent definition was touched at all.
- **Verification actually run, not asserted:**
  - `go test ./platform/orchestration/...` — green, and I checked the tests RAN (a `-run` regex
    matching nothing also prints `ok`; 14 of mine executed by name).
  - **11 mutations applied, compiled, tested and reverted** — 8 on the detector, 3 on the writer
    enrichment. None survived.
  - `verify-head-builds.sh --with ×5 --test`: HEAD itself fails 22 ways (Kafka down locally,
    DB-dependent suites, other lanes' build failures), so the suite is not a gate here. **The
    control is the diff**: failure set with my change is byte-identical to bare HEAD's once
    timings are normalised — nothing added, nothing removed. Post-commit,
    `verify-head-builds.sh` reports **HEAD 924079c94 builds**.
- **MISSTEP (mine), caught by my own runner: a mutation that did not COMPILE read as "killed".**
  My first M5 replaced the scope comparison with `false`, leaving `href` unused. `go test` exits
  non-zero for a build failure exactly as it does for a failing test, so the runner scored it as
  covered. Three of the writer mutations failed the same way (deleting an assignment leaves the
  variable unused). **The check:** a mutation must COMPILE or it measures nothing — build first,
  and treat a non-compiling mutation as invalid, never as a pass. Fixed in the runner and stated
  in the test file's header so the next reader does not inherit the false claim.
- **MISSTEP (mine), caught by the estate's own ratchet:** I did not declare the check's lifecycle
  posture, and `TestEveryPagesQueryingCheckDeclaresItsLifecyclePosture` failed the build until I
  did. That is `bugs_open/356`'s mechanism working — and worth recording that it caught a real
  omission on the first new check written after it existed. Declared `PostureArmed`: the remedy
  re-runs the writer over live prose, so a retired-but-still-serving page belongs to retraction,
  never to a rewrite.
- **Found while doing the baseline, and NOT mine: `cmd/config-key-audit`'s test package does not
  compile at HEAD.** `livedeclarations_test.go` names `livespec.LiveAuditOnlyDeclarations`, which
  exists only in an uncommitted working copy of `platform/livespec/livespec.go`. So the RFC_022
  budget counter and the cron parity test CLAUDE.md tells every session to run are unrunnable
  from a clean checkout. Reported to the `bugs_open/333` lane (they made the test-side commit);
  deliberately NOT fixed here, because committing another session's dirty file is the exact sweep
  the pathspec rule exists to prevent.

### What remains before this bug can close

1. Fleet roll, then prove the check is registered (`service_binary_capabilities`, positive +
   negative controls).
2. Add `missing_prose_links` to a live `run_checks.config.checks` array **by hand** — image
   first, then config.
3. Report-only read of the first pass, then the canary (PLAN's induction route — the site-scoped
   opt-in hook, NOT the bogus `site_id` that would corrupt the column the reader joins on).
4. Verify at the SERVED page with `scripts/probe-page-url.sh`, then close.
