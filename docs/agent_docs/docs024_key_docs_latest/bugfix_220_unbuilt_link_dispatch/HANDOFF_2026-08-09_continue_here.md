# HANDOFF 2026-08-09 — bugfix 220 lane: continue here

Written for a fresh session with no context. Read this, then the bug file's
2026-08-09 addendum; everything else is reference.

## What this lane is

`bugs_open/220`: `unbuilt_internal_link` work items name their TARGET page in the
`page_id` column, but dispatch rebuilt the CONTAINER and completed green forever.
Fixed as three legs + one more found in acceptance:

1. **Mapping** (mig 340): `build-dispatch-loop.call_handler.input_mapping` +=
   `"page_id?": "current_item.page_id"`.
2. **Authority** (Go + mig 340): `load_page_record` gains `authoritative_page_id`
   (opt-in, RFC_010 §2); when supplied, lookup is by id, name ignored; no-row is
   FATAL; malformed is FATAL. Only page-build-handler's step opts in.
3. **Verification** (Go + mig 341): `VerifyUnbuiltInternalLinkResolved`,
   fail-closed, disjuncts = link removed OR target shipped
   (`NeverDeployedPagePredicate`); claim-timeout exclusion lockstep both halves.
4. **Identity chain** (mig 342, found 08-09): `save_sections.page_name_field` →
   `page_record.name` — it was the one step still reading `spec.page_name`, and it
   saved the TARGET's freshly-written copy ONTO THE CONTAINER.

Commits: `a60a13cbb` (main), `03433f4b5` (gofmt), `e55cbfa64` (council r1 catches),
`58c6555c2`/`70b5152db`/`0c7dcfd1f` (trail docs). Council: **APPROVED r2**, corr
`def4441c` (342's own submission was refused by the gate's docs-scope filter —
recorded in NOTES, not forced). Register **WII-012**. Migs 340/341/342 applied,
read back, recorded (⚠ "342" also names the thunder lane's unrelated file —
resolve by slug).

## Proven so far (2026-08-09, dartsonline acceptance, corr 110acf5a)

- Binary on both replicas carries all commits (pod-greps in NOTES).
- Routing: dispatch of item `338deb27` targeted grip-styles (the TARGET) — the
  deploy skipped honestly ("no component rows yet") instead of shipping the
  container's file.
- Verifier: completion carries `_verification.status='verified'` (disjunct b).
- The regression floor independently stopped the sibling's wrong-page save
  (`a8327624`, failed loudly).

## OPEN — in priority order

1. **Watch the repair land**: item `3cb732b1`
   (`needs_content_page:beginners:repair-338deb27`, priority 30, site
   dartsonline `5fe8785b-223d-41a3-88ee-c07187622381`). Beginners' `content_data`
   currently holds grip-styles' copy (contamination at 10:00:56Z, item 338deb27);
   its SERVED page is still the correct old render. Success = beginners'
   components hold beginners copy again and the page deploys. Until then do NOT
   let a beginners rerender run (two were cancelled: `47ba8f2c`, `3c10ab6c`;
   discovery may re-mint).
2. **The end-to-end convergence proof** (the one thing never yet observed): after
   discovery re-mints unbuilt items (the 08-09 ones were cancelled), one must run
   the whole chain under migs 340+341+342: writer writes the TARGET's sections,
   `sections_saved.page_name` = the TARGET, deploy renders it,
   `pages.deployed_at` set, item completes verified via disjunct (a), curl the
   target = 200. Candidate targets: grip-styles (plan: hero/article-body/CTA) or
   any row from the census query in the RUNBOOK. Fire the loop with the SAFE kcat
   pattern (payload in container COMMAND + PUBLISH_OK; the shipped
   `060improvement_loop/076_improvement_loop_trigger.sh` HARDCODES
   robot-hands.com after its arg parsing — do not use it as-is).
3. **Close consideration**: once (1)+(2) hold, 220 is fixed-and-live end to end.
   Owner ruling 08-06: a finished bug STAYS in `bugs_open/` — update the file
   head, don't move it.
4. **Deferred, on record**: candidate 4 (route unbuilt targets by page_type via
   `availableBuilders` — package-direction refactor). Demand signal: directory-
   or tool-typed targets landing `failed` after attempts.
5. **Adjacent, told not owned**: the dead_fragment lane's
   `VerifyDeadFragmentLinkResolved` still uses the LIKE-concatenation shape the
   council flagged (over-match on `_`); flagged in bug 220 § COUNCIL TRAIL and in
   the code comment.

## Gotchas this lane already paid for (do not re-derive)

- A watcher whose item query has a `created_at > now()-interval` window blinds
  ITSELF as items age — key on item ids.
- The verifier + the discovery check read STORED `rendered_html`, not served;
  they agree with each other, so no churn, but a served page can lag either way.
- `kubectl exec -i` inside a `while read` loop eats the loop's stdin —
  `</dev/null` it.
- Migration numbers, bug numbers, and "the census" all expire in hours; re-check
  at the moment of use, not at the moment of planning.

## Where everything lives

- Bug file: `bugs_open/220_HANDOFF_2026-08-08_...md` (mechanism, reproduction,
  census, council trail, 08-09 addendum).
- This dir: PLAN (design + reasons), NOTES (append-only technical log), RUNBOOK
  (queries + acceptance commands), README_where_we_are (owner prose),
  submission JSONs (r1, r2, 342).
- Memory: `bugfix-220-unbuilt-link-dispatch-workstream.md` + the
  MEMORY_workstreams line.
