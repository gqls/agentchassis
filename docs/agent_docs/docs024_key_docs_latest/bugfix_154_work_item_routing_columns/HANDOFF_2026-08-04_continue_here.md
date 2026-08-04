# HANDOFF 2026-08-04 (rev 3) — component-identity-drift gap: measured, fallback shipped for the unambiguous case, council submitted; NOT yet built/deployed

**This revision replaces rev 2** (same file, same date — rev 2 itself
replaced a morning version that had two wrong claims, both corrected in
place; see `bugs_open/178`'s own file for the full corrected account). Rev 2
left one open task at priority 1: design and implement a fix for the
component-identity-drift gap, after measuring how often it actually
happens. That is what this revision reports on. Read `bugs_open/178`'s own
file in full before doing anything further — it is still the authoritative
account, now with a new final update covering everything below.

## What happened since rev 2, in order

1. Checked nobody else had touched this ground first (`git log` on
   `bugs_open/178`'s file and `load_current_section_content_action.go` —
   clear since `71ecbb013`).
2. Went to verify rev 2's own hypothesis ("the selector can evidently pick a
   different [fallback component] build to build") before trusting it as the
   basis for a fix. **It was wrong as stated**: `article-body` and
   `generic-text-block` each have exactly one active `content_components`
   row under their own distinct `section_type` — there was never a choice
   between competing candidates for this specific instance. The real
   mechanism that produced a stored slot (`generic-text-block`) disagreeing
   with a plan that has only ever said `article-body` (single plan, unchanged
   since 2026-07-17) is still not reconstructed — flagged as `[UNVERIFIED]`
   rather than asserted.
3. Measured the RATE at one remove from the unresolved mechanism: not "how
   often does resolution flip" (undefined without step 2's missing piece)
   but "how many pages right now have a stored slot the current plan doesn't
   name at all" — the condition that actually determines whether the
   exact-name join fails, regardless of cause. **Result: 3 of 127 pages
   checked (2.4%)**. Query and the three cases are in `bugs_open/178`'s new
   update and `NOTES_work_item_routing_columns.md`'s 2026-08-04 entry.
4. Implemented fix candidate 1 from `bugs_open/178`'s original list, for the
   UNAMBIGUOUS case only: `load_current_section_content_action.go` now falls
   back to attaching a page's one remaining prose-sized, unclaimed
   `page_components` row when exactly one ready section misses the exact
   `slot_name` join. Two-or-more on either side: left unmatched, exactly as
   before — no guessing. Recorded as `edit_live_meta.fallback_matched`,
   which doubles as ongoing fleet-wide instrumentation of this drift class.
5. Three new test cases, mutation-tested against the pre-fix code (`git
   stash` of just this file; all three failed as expected; restored,
   verified green). Full package `go test ./platform/orchestration/actions/...`
   green throughout, including the `bugs_open/192` regression tripwire.
6. Submitted to the council gate:
   `Council-Submitted: 8a3e0315-4576-4829-bf42-c0c8cdfc4e3a` — dispatched
   (`RUN_ORCH_ID=e86e3e21-825e-4e46-9f0c-d911ce40ff3a`), no
   `orchestration_states` row visible yet at submission time. Not treated as
   a dropped dispatch on that alone (this repo's own note: publish→run start
   has measured up to ~29 minutes under normal load) — check again before
   assuming it stalled the way the FIRST 178 submission (`97ebadcf`) did.
7. Committed source + tests + `bugs_open/178` update + workstream docs in
   this pathspec (see git log for the exact hash — not filled in here to
   avoid the same "commit hash guessed before it exists" trap this repo has
   hit before).

## State

Source-and-tests only. **Not built, not deployed, not pod-verified, not
live-verified against a real page.** This matches the pattern the rest of
`bugs_open/178`'s history has followed (image builds bundled into a later
whole-fleet release) but means: until an image is built and rolled, this
fix changes nothing about what the fleet actually does. Do not report this
as "fixed" or "closed" on the strength of the commit alone —
`bugs_open/178` stays OPEN.

## OPEN — in priority order

1. **Build, deploy, pod-verify, and live-verify.** Pod-grep marker once
   built: `strings /app/agent-chassis | grep -c "single-unmatched-prose-slot"`
   would need the log-line string embedded — check what's actually
   grep-able post-build (the zap message text or a symbol name) before
   relying on a specific string. Live verification: dispatch a real
   `content_rewrite` item (with `spec.mode="edit_live"` patched in if it
   predates `08d0515f3`) against one of the three pages named in this
   session's fleet measurement (`bugs_open/178`'s new update names them),
   and confirm `edit_live_meta.fallback_matched=1` plus an unchanged
   `content_data` length (plus the inserted anchor) on the previously-
   mismatched slot.
2. **The ambiguous case is still open** — two-or-more unmatched sections, or
   two-or-more candidate prose slots on one page, and the fallback correctly
   refuses to guess, which means the original bug is still reachable there.
   Not yet observed in the fleet (no page in the 2.4% showed more than one
   unclaimed prose-sized slot at once) — unknown severity, not zero.
3. **The actual mechanism is still undiagnosed.** Why can a build's resolved
   component name for a page's section differ from what an earlier build
   stored? This session found the specific hypothesis in rev 2 was wrong,
   but did not replace it with a confirmed alternative. A `090` run against
   this specific question (not the handler-gate question the first `090` run
   for this bug already answered) would be the appropriately-scoped next
   step per this repo's diagnosis-before-debugging default — this is exactly
   the "cause is still non-obvious after a quick look" case it's for.
4. **Council verdict** for `8a3e0315-4576-4829-bf42-c0c8cdfc4e3a` — check
   before resubmitting or assuming stalled; the first 178 submission
   (`97ebadcf`) DID stall (reached `review_constitution`, never advanced,
   advisory only). If this one also stalls, that's now two-for-two on this
   lane and possibly worth its own mention to whoever owns council-gate
   reliability, not just a shrug.
5. **Watch list, unchanged from rev 2**: the shrink guard doesn't fire on a
   whole-slot rename (old slot gone, new slot appears) — a second, narrower
   gap in that guard's coverage.

## Landmines specific to this lane (carry-forward + one addition)

- All landmines from the 2026-08-03 and rev-2 handoffs (shrink guard,
  dependency release rules, dispatch quiet-spell reading,
  `orchestration_states` retention, `output_field` shape-preservation,
  `build-dispatch-loop` IS scheduled, `bugs_open/087` vs `192` error-string
  collision) still apply unchanged.
- **`content_components` has no `site_id`/`page_id` — it is a fleet-shared
  library.** A component minted via `needs_new_component` for one page's
  section is selectable by ANY other page whose plan names the same
  `section_type`, forever after. This is background for the whole
  drift-gap investigation, not specific to this fix, but easy to forget
  when reading a single component's `description` field and assuming it
  scopes the row to that page.
- **A "the selector flips between competing candidates" theory needs a
  COUNT, not a plausibility check** — `SELECT section_type, count(*) FROM
  content_components WHERE component_level='section' AND is_active AND
  forked_from IS NULL GROUP BY section_type HAVING count(*) > 1` before
  trusting it as a root cause for any specific instance. It was 1 for both
  components in this bug's own case, which is what falsified rev 2's guess.

## Cold-start pointers

- `bugs_open/178`'s own file — still the authoritative account, now five
  updates deep, including this session's.
- `NOTES_work_item_routing_columns.md`'s 2026-08-04 tail (two entries from
  today: the 192-lane verification, then this session's measurement+fix).
- `SUBMISSION_2026-08-04_component_identity_drift_fallback.json` — the
  council submission this session filed, for its `grounded_in` evidence if
  a verdict comes back with objections to answer.
- Register entry PBP-028 (`docs026_concept_register/register/page-build-pipeline.md`)
  still has an open review question about matching by `slot_name` — this
  session's fleet measurement (2.4%, three concrete pages) is evidence worth
  folding in when someone next edits that entry; not done this session.
- Commits: `08d0515f3` (178's original fix), `2b9d84072`/`71ecbb013` (192's
  fix + 178's correction), and this session's commit (see `git log -- \
  platform/orchestration/actions/load_current_section_content_action.go`).
