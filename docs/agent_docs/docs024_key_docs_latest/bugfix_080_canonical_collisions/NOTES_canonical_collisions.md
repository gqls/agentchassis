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
