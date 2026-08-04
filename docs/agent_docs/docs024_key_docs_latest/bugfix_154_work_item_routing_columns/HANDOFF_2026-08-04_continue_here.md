# HANDOFF 2026-08-04 (rev 4) — fix LIVE and pod-verified; council verdict + root-cause diagnosis both in flight, unread; no live repro exists to test against

**This revision replaces rev 3** (same file, same date). Rev 3 shipped the
component-identity-drift fallback as source+tests only, submitted to
council, and left build/deploy/verify as the top open item. That is what
this revision reports on, plus two background runs now in flight. Read
`bugs_open/178`'s own file in full first — it is still the authoritative
account, six updates deep now.

## What happened since rev 3, in order

1. Owner rolled a fresh chassis build. **Verified at the pod, not the tag**:
   `v1.0.1251`, both replicas (`agent-chassis-5455ddcdcc-crnb6`,
   `agent-chassis-5455ddcdcc-gpr92`), `strings /app/agent-chassis | grep -c
   "single-unmatched-prose-slot"` = 1 on each, plus the existing `"SECTION
   SHRINK"` positive control = 2 on each. **The fallback fix from rev 3 is
   live in production right now.**
2. Checked the council submission from rev 3 (`8a3e0315-…`) — **it never
   ran**. Zero `orchestration_states` rows by exact id or by payload search,
   8.5h after submitting, far past the ~29min this repo's own docs cite as
   normal queue latency. Likely cause: `kcat -P`'s known silent-drop shape
   colliding with a connectivity timeout hit at submission time. **Do not
   trust a printed `SUBMISSION_CORR` alone** — check for an orchestration
   row within a few minutes of submitting, not just once, before moving on.
3. Resubmitted the identical plan. New correlation, confirmed genuinely
   in-flight this time: **`Council-Submitted: 56f9a5a2-4d37-4114-9442-239861acd36e`**.
   As of this revision it has progressed through
   `review_editquality` → `review_constitution` → `review_prior_art` —
   still running, **verdict not yet read**. Check first:
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='56f9a5a2-4d37-4114-9442-239861acd36e' AND kind='council_report'
   ORDER BY created_at;
   ```
   If APPROVED: commit trailer `Council-Reviewed: 56f9a5a2-4d37-4114-9442-239861acd36e`
   on a follow-up commit (forward-only — do not amend the original). If
   REVISE: read the objections (`SELECT body FROM doc_notes WHERE
   categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1`) and decide
   whether to answer them or leave as advisory-only, same as this bug's
   first submission was left.
4. Fired the root-cause diagnosis this file has been carrying as an open
   item for two revisions: why does a build's resolved component name ever
   disagree with what an earlier build stored, given `plan_sections`' Path 1
   should match an existing correctly-named component directly? Fresh slug
   (`178-component-identity-drift-mechanism`, distinct from the earlier,
   already-answered `178-crosslink-regenerates-whole-section` run so it
   doesn't dedupe against stale results). **Run correlation:
   `167d2cc2-0b98-405c-a1d7-d54d80ed37c9`** (intake correlation
   `2bcf9359-4603-472f-ba00-4d1d5f33f6c8`, use the run one for artifacts).
   Claimed by `diagnose-dispatch-loop` within the 180s the trigger waited.
   **Not read.** Check:
   ```sql
   SELECT status FROM site_work_items WHERE item_key='needs_diagnosis:178-component-identity-drift-mechanism';
   SELECT body FROM doc_notes WHERE correlation_id='167d2cc2-0b98-405c-a1d7-d54d80ed37c9' ORDER BY created_at DESC LIMIT 1;
   ```
   Advisory note from the trigger: local HEAD was 50 commits ahead of
   `origin/087_towards_multiple_domains` at dispatch time, so the diagnosis
   read the pushed tree, not this session's own commit — irrelevant to this
   specific question (it concerns history predating the fix) but worth
   knowing if a future diagnosis run seems to be missing recent work.
5. **Caught and corrected my own bad plan before executing it**: had
   intended to live-verify the fallback against the 3 pages from rev 3's
   fleet measurement. On re-examination, none of them exercise the
   fallback — their mismatched slots aren't in their pages' current plans
   AT ALL (extra components attached via some other route), so
   `plan_sections` never puts them in `sections_ready` for any build, and
   there is nothing for the fallback to match against. The one page that DID
   show the real failure (`guide-independent-strategy`) no longer does,
   because the run that discovered it already rewrote the stored slot to
   agree with the plan. **There is currently no known live page that would
   exercise this fallback.** Did not manufacture one on production data.

## State

The fix is **live and pod-verified**, backed by unit + mutation-test
coverage, but has **no live end-to-end confirmation** (no known repro page
exists right now — see point 5 above). This is a real, acknowledged gap,
not an oversight to silently carry forward: don't upgrade this file's
confidence level about the fix beyond "deployed, tested in isolation,
unproven live" until either a natural occurrence is caught or a legitimate
test scenario is found. `bugs_open/178` stays OPEN.

## OPEN — in priority order

1. **Read the council verdict** (`56f9a5a2-…`, query above) and the **090
   diagnosis result** (`167d2cc2-…`, query above). Both were in flight, unread,
   when this revision was written — this is the first thing a fresh session
   should do, not new investigation.
2. **Watch for the fallback's first natural firing** rather than
   manufacturing a test:
   ```sql
   SELECT id, created_at, collected_data->'section_plan'->'edit_live_meta'
   FROM orchestration_states
   WHERE collected_data->'section_plan'->'edit_live_meta'->>'fallback_matched' = '1'
   ORDER BY created_at DESC LIMIT 10;
   ```
   Subject to `orchestration_states`' ~24h terminal-row retention — a clean
   result only means "not in the last ~24h of activity", not "never".
3. **The ambiguous case is still open** (two-or-more unmatched sections, or
   two-or-more candidate prose slots) — the fallback correctly refuses to
   guess there, which means the original bug is still reachable. Severity
   unknown; not yet observed.
4. **The root-cause mechanism** (point 4 above) may resolve this once read —
   don't re-investigate from scratch before checking its answer first.
5. **Watch list, unchanged for three revisions**: the shrink guard doesn't
   fire on a whole-slot rename (old slot gone, new slot appears) — a second,
   narrower gap in that guard's coverage, separate from everything above.

## Landmines specific to this lane (carry-forward + one addition)

- All landmines from the 2026-08-03 and rev-2/rev-3 handoffs still apply
  unchanged (shrink guard, dependency release rules, dispatch quiet-spell
  reading, `orchestration_states` retention, `output_field`
  shape-preservation, `build-dispatch-loop` IS scheduled,
  `bugs_open/087` vs `192` error-string collision, `content_components` has
  no site_id/page_id, "competing candidates" theories need a COUNT not a
  plausibility check).
- **A council/diagnosis trigger script printing a correlation ID is not
  proof the dispatch landed.** This session's first council submission
  printed a clean `SUBMISSION_CORR` and then silently never ran for 8.5h —
  `kcat -P`'s documented silent-drop behaviour, likely triggered by a
  connectivity blip at the exact moment of publish. Check for an
  `orchestration_states` row within a few minutes of any such dispatch,
  not just once right after triggering it, before treating the submission
  as "sent and queued."
- **A fleet-wide SQL measurement of "mismatched slots" is not automatically
  a set of live test cases for a specific fix.** Check that the mismatch
  actually reaches the code path under test (here: does the name appear in
  the page's CURRENT PLAN at all, i.e. would `plan_sections` ever put it in
  `sections_ready`?) before treating a measured row as a repro.

## Cold-start pointers

- `bugs_open/178`'s own file — still the authoritative account, six updates
  deep now, including this session's deploy-verification and correction.
- `NOTES_work_item_routing_columns.md`'s 2026-08-04 tail (three entries:
  192-lane verification, the measurement+fix, and this deploy-verification
  session).
- `SUBMISSION_2026-08-04_component_identity_drift_fallback.json` — the
  council submission (now on its second, successful dispatch attempt).
- Register entry PBP-028 (`docs026_concept_register/register/page-build-pipeline.md`)
  still has an open review question about matching by `slot_name` — still
  not folded in; the fleet measurement (2.4%) plus this session's correction
  about what it actually measures are both worth adding when someone next
  edits that entry.
- Commits: `08d0515f3` (178's original fix), `2b9d84072`/`71ecbb013` (192's
  fix + 178's correction), `4b3f9f89b` (this bug's fallback fix, rev 3).
  This revision (rev 4) made no code changes — docs only; check `git log`
  for its own commit if one was made after this file was written.
