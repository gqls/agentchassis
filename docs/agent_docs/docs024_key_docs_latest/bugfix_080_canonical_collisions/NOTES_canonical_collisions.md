# NOTES — bugfix_080_canonical_collisions (append-only, newest at the bottom)

## 2026-08-03 (evening) — lane opened

Picked 080 as the next unowned open bug. Ownership: `who-owns.py` unowned for 080 and 161 only;
161's memory entry says fixed-at-source awaiting redeploy, so 080 is the real candidate. Checked
LIVE transcripts too (who-owns reads commits and is lagging): sessions active on 178, 163, 188,
123, 122, 186, 072/145, 029/033 — none on 080.

Validity re-check found the bug UNDERSTATED:
- robot-hands `/news` pair still live (both 200, identical <title>, orphan has self-canonical).
- **`/gripper-catalog` is a second fully-live pair the bug file never mentions.** Found by
  keying the census on URL shape instead of page_type (the filter that blinded the original
  survey — that lesson is already in WRONG_CALLS from 07-27; the census query is in RUNBOOK).
- webdesign.co.uk `news` row moved planned→deployed since the file was written.
- dartsonline has 3 stem-collisions where the flat side is planned+archived (benign today; its
  `/guides.html` 404s on the wire).

**MISSTEP (recorded per the marker discipline):** my first census draft filtered liveness by
`build_status='deployed'` — the exact blindness `bugs_open/185` is open about, in a session that
had 185's memory entry loaded. Caught before any conclusion rested on it, by the explore agent
flagging `PageHasShippedPredicateFor`; re-ran under the shared predicate. Same 6 groups — but
only because this population happens to carry no shipped-not-'deployed' rows. The check that
would have caught it cold: grep your census SQL for `build_status` before trusting it.
→ Logged in WRONG_CALLS.md (2026-08-03 entry) when the batch of doc edits lands.

Explorations (3 agents): the machinery map is in the plan. Key facts settled:
- The 080 gap-planner arm is FIXED at HEAD and live (pod-grep 3); the ON CONFLICT is now
  DO NOTHING + resolveNewPageConflict (081/175's work), so 080's "page_type absent from DO
  UPDATE SET" description is historical for that file.
- 081 shipped a REFUSAL not a repair; retype-deployed and retract-live both have no mechanism
  (098 veto → RFC_011). So 080's residual is a decision item, not a code fix.
- Un-canonicalised writers remaining: create_blog_posts (080's exact shape), deploy_tool (TL-010),
  create_tool_component's fallback+guide, adopt_verbatim (deliberate), webdesignport (deliberate),
  create_report_page (out of class — UUID names).
- Candidate-3 detector: unbuilt, unclaimed (phrase appears once in repo — in 080 itself).
- Detector needs BOTH signals: canonical-name grouping misses gripper-catalog (content
  canonicalises to itself); URL path-key catches it. Verified against live rows.

User rulings this session: (1) detect and file only, no live mutation; (2) address every
un-canonicalised writer — after the requested recheck, honoured per-surface (adopt_verbatim would
trip LANDMINES:2476; naive deploy_tool canonicalisation would mint 12 duplicates).

## 2026-08-03/04 (overnight) — implementation

All surfaces + detector written, tests green (`go test ./platform/orchestration/actions/...
./platform/orchestration/datahelpers/...` → ok, includes 5 blog tests, 4 tool-identity tests,
2 convergence tests, 13 detector tests):

- A `create_blog_posts_action.go`: identity via CanonicalisePage, fail-loud empty triple
  (`canonicalisation_failed` counter in the result). Byte-identical for clean blog-post slugs.
- B `deploy_tool_action.go`: `resolveToolPageIdentity` — existing row under either name wins
  (verbatim, no re-key); new tools canonical. `companionGuideIdentity` shared by both tool
  surfaces (byte-convergent, enforcement only).
- C `create_tool_component_action.go`: the silent flat-URL fallback now refuses + rolls back the
  component; guide identity through the shared helper.
- D `adopt_verbatim_identity_convergence_test.go`: verbatim names pinned as FIXED POINTS of
  CanonicalisePage (URL half untouched — that divergence is the feature).
- F `discovery_checks/check_page_canonical_collision.go`: union-merged two-signal grouping,
  ≥2-active filing rule, prior-wont_fix suppression, RFC_010 retraction of stale open items,
  verifier with site-existence guard. **A fable design-review agent found 8 issues before
  implementation** — the load-bearing ones: one-item-per-signal would have filed 3 items on
  robot-hands not 2 (union-merge fixes it); the verifier cannot see item_key so the spec carries
  the group signature; a wont_fix would otherwise resurrect the item every sweep, for ever.

**Found while wiring the 220 lockstep: the LIVE claim-timeout exclusion list is TWO behind the
declared one.** The 151 lane added `content_duplication` to 220's file (ec8ad7959) but no
targeted-replace ever reached `scheduled_tasks.pre_query` — verified live 2026-08-03: 4 entries,
theirs absent — and their check went live via seed 296 the same day, so their items CAN dispatch
and the sweep could auto-complete them past their verifier. Seed
`302_claim_timeout_exclusions_catch_up.sql` carries both their declared entry and mine, stated
loudly in its header. **→ 151/brochure lane: read 302's header** (this is the "consumers told"
half; also flagged here because that lane reads who-owns on 156/151).

MISSTEP (caught by my own re-read, logged for WRONG_CALLS): the first draft of 302's verify
block had a stray newline inside the old-list LIKE pattern — a verify that could never fail,
the exact disconfirmability class WRONG_CALLS tallies. The check: read every LIKE pattern
end-to-end before trusting a DO/RAISE block, and ask what result would make it fire.

Seed `303_enable_page_canonical_collision.sql` written (image → seed order; expected first
sweep: exactly 2 items, both robot-hands). NOT applied — inert until my image is pod-verified.

Another lane rolled a fresh chassis while this was in flight (user note ~21:00). My changes are
uncommitted so they are NOT in that image; council submission timed after the roll settles
(a roll kills an in-flight council run).

## 2026-08-04 (morning) — committed, APPROVED r1, seed 305 live; blocked on the whole-fleet release

Commits: `9595c43fc` (surfaces A–D + tests), `96dd3015c` (detector + verifier + PLAN-047 register
entry same-commit + seeds 305/306), lane docs. Clean `git archive HEAD` builds and all tests pass.
Seed numbering collided mid-task (another session minted its own 302/303) → renumbered to 305/306.
Local docker build of chassis at v1.0.1248 proved HEAD dockerises; NOT deployed — releases are
whole-fleet, owner-run (memory ruling 2026-08-03). **Seed 305 APPLIED + verified live** (6-entry
exclusion list confirmed by its own fenced replace).

**Council: APPROVED round 1** (corr `83710c81`, 6 advisory objections, none high, 4 abstained).
The four checkable objections, checked:

1. **"Which upsert helper does each surface feed — three helpers, opposite collision policies"**
   (editquality/bug_historian/reuse/guardian/constitution/debug_historian, the recurring one):
   - `create_blog_posts` → its own `ON CONFLICT (site_id, name) DO UPDATE SET title, page_type,
     sections, page_spec` (`:241`). Two runs producing the same canonical name now collide-and-
     update — the intended policy for a plan-authority surface. Compatible.
   - `deploy_tool` → `UpsertPageForRole` (`:386`, `:531`), the refuse-shipped/adopt-unshipped
     seam. `resolveToolPageIdentity` runs BEFORE it: an existing row under either name is reused
     (refresh path, no second row possible); a fresh canonical name creates. A SELECT→upsert race
     lands in UpsertPageForRole's explicit conflict branch, which refuses rather than duplicates.
     Compatible — and NOT a fourth helper: it resolves *which identity to ask for*, then routes
     through the existing seam; the three helpers answer *what to do on collision*.
   - `create_tool_component` → plain INSERT, no ON CONFLICT (`:289` region): a collision errors
     loudly and rolls back the component — no silent second row. Compatible.
2. **debug_historian: hand-rolled `status='active'` claimant predicate.** The enumerate-first
   discipline was done before the predicate was chosen — `GROUP BY status` → exactly
   {active 561, archived 27} (RUNBOOK). "Claimant" is a different question from "has shipped"
   (`PageHasShippedPredicateFor`): an archived-but-still-served page (098's class) is deliberately
   NOT a claimant here — its remedy is 098's retraction question, not a which-row-survives
   decision. Check and verifier share the predicate, so they cannot disagree.
3. **prior_art_librarian: the absence claim needs a directory check, not a grep.** Done before
   building: the explore pass enumerated all 85 files in discovery_checks/ and every Register()
   call site (52 checks); no name/URL collision detector exists under any name. The
   scheduled_tasks claim is now self-verifying: 305's fenced pre-check REQUIRED the exact 4-entry
   list and passed.
4. **guidelines: recurrenceExpected on the shared insert path.** `WorkItemSpec` carries no such
   field (remit.go records the limitation). Two-strike counts only complete/failed; our items are
   needs_human_review → wont_fix (suppressed by the check's own arm) or complete-via-human. Two
   human completes within 7 days would born-unresolved a third file — acceptable, documented here.

guardian's containment note (three surfaces + check in one plan) is partially answered by the
commit split: surfaces in `9595c43fc`, detector in `96dd3015c` — independently revertable.

**STILL OWED to close 080:** owner's whole-fleet release at a tag whose HEAD ≥ `96dd3015c` →
pod-grep every replica (`page_canonical_collision` ≥1, negative control `"/tools/%s.html"` = 0)
→ apply seed 306 → induce completeness sweep on robot-hands → expect exactly items
`page_canonical_collision:/news` + `:/gripper-catalog`, re-run files 0 new → move 080 to
bugs_closed (both paths on the commit, `git ls-tree` check). With the seed, also add
`page_canonical_collision` to `liveConfiguredChecks` in
`discovery_checks_registration_test.go:42` (its contract is to pin LIVE config, so it gains the
entry when the config does, not before).

## 2026-08-04 (late morning) — the owed chain ran; 080 CLOSED

Owner rolled **v1.0.1250** (whole-fleet). Every step of the owed list above executed, in
order, all green:

- Pod-grep BOTH replicas (`-4dzzx`/`-5z5sn`): `page_canonical_collision` 5,
  `collisionCanonName` 1, negative `"/tools/%s.html"` **0**, control `request_browser_run` 5,
  discriminator 0. (Gotcha for the runbook: five sequential `grep -acF` over the binary in one
  exec exceeds a 2-min timeout — split into ≤2 per exec.)
- Seed 306 applied; checks list verified — 32 entries, neighbours intact, snapshot taken.
- Fixture entry committed (`9ec60ccd2`) — also caught `content_duplication` missing from the
  same fixture since 08-03 (declared in the message).
- Induced sweep corr `10bafdf5` → **exactly** `page_canonical_collision:/news` +
  `:/gripper-catalog`, needs_human_review, no handler, nothing else fleet-wide.
- Second sweep corr `239fdfb8` → COMPLETED, **0 new rows** (2 rows / 2 keys, newest
  created_at unchanged) — dedup + prior-item suppression proven by induction, not asserted.
- 080 closing banner written; file moved to `bugs_closed/` with BOTH paths named on the
  commit; `git ls-tree` check in the closing commit's message. 016b §10 row flipped to CLOSED.

Residuals recorded where they live: the two decision items (the queue), retirement
(098/RFC_011), queue surface (033). **This lane is done.**

## 2026-08-04 (midday) — OWNER RULED on the two decision items; execution via RFC_011's mechanism

**Ruling: "keep the page that the spec would ordinarily produce."** So: `/news` keeps
`news-index` @ `/news/index.html` (re-typed → `news-index`, the relojistas repair);
`/gripper-catalog` keeps `gripper-catalog-index` @ `/gripper-catalog/index.html`. Strays
(`/news.html` `18d681af`, `/gripper-catalog.html` `64fab29e`) retire.

**RFC_011 was DECIDED (option B) the evening of 08-03** — my "retirement has no mechanism"
line above is superseded: `retract_page_deployment` is legitimate, proven (learning-center
200→404 by the 098 lane), fired via `sql_for_agents/216_TRIGGER_page_retraction.sh`
ALWAYS with `PAGE_IDS`. **Retraction only considers NON-ACTIVE pages**, so the strays were
archived first.

Executed so far:
- ruling recorded in both items' `spec.decision` (+`decided_by`);
- both strays `active` → `archived`;
- kept news row re-typed `section-index` → `news-index` — spec triple now exact.

**OWED:**
- the 216 dispatch itself — **classifier-blocked for a session; the owner fires it**
  (command in README/chat), then read `collected_data.retraction_audit` for refusals
  (a still-linked page is refused with referrers named — that is the guard working);
- two-part acceptance: both urls 404 now AND still 404 after the ~20:0x news refresh;
- then the next completeness sweep closes both items via the check's own Resolved arm
  (group loses its second active claimant) — induce one and watch, or let the schedule.
