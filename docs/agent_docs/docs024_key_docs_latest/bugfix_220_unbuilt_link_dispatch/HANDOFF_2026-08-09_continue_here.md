# HANDOFF 2026-08-09 — bugfix 220 lane: ⛔ CLOSED, NOTHING OWED. Do not action this file.

> **CLOSED 2026-08-09 15:14 UTC. The convergence proof this handoff was written to
> chase LANDED, and the whole run completed: 10/10 items `complete`, 10/10 via
> disjunct (a), zero failures. Bug 220 is fixed, live and proven end to end.**
>
> **Nothing below is a live instruction.** STEP 0 is answered, the "if it has NOT
> landed" branch is dead, and the watcher/re-fire advice is spent. It is kept only as
> the record of how the proof was chased.
>
> **Two of this file's own statements are now known WRONG — do not act on them:**
> 1. **The PASS table's `saved_page_id` = `deployed_page_id` leg used a jsonb path one
>    level too shallow** (`response→deploy_result→rendered_page`; the real shape
>    interposes another `response`). It returns empty on **every** row, so that leg
>    could never pass. **The control this file tells you to validate against could not
>    catch it**, because the column was *expected* empty there. Corrected query, and a
>    control that reads NON-empty, are in `RUNBOOK_unbuilt_link_dispatch.md`.
> 2. **"The four items that are EXPECTED to fail" is REFUTED.** All four converged via
>    disjunct (a), on two distinct section-index pages built from zero components.
>    **Candidate 4's demand signal is therefore absent, not merely weak.**
>
> **Where to go instead:** `SUMMARY_2026-08-09_unbuilt_link_dispatch.md` for the
> read-out; the bug file's § "FINAL LEDGER 15:14Z" for the evidence and the candidate
> 4 ruling; `NOTES` tail for the missteps. Still genuinely open (low priority, and
> self-correcting): the mortgagecalculator.co.uk residue, item 1 under "Also open".

## ~~STEP 0 — has the proof already landed?~~ ANSWERED: yes, at 14:24–15:14Z. Historic only.

A run was fired at 13:10 UTC on 2026-08-09 (corr
`576f0ab9-5a17-4449-9bbc-ee1983576433`, `PUBLISH_OK` receipt) against dartsonline.com
(site `5fe8785b-223d-41a3-88ee-c07187622381`). It re-minted **10**
`unbuilt_internal_link` items. By the time you read this they may have dispatched.

```sql
SELECT left(w.id::text,8) AS item, w.status,
       w.spec->>'page_name' AS container, COALESCE(t.name,'?') AS target,
       w.result->'response'->'sections_saved'->>'page_name'              AS saved_page_name,
       left(w.result->'response'->'deploy_result'->'rendered_page'->>'page_id',8) AS deployed_page_id,
       w.result->'_verification'->>'status'                             AS verif,
       left(w.result->'_verification'->>'detail',90)                    AS detail
FROM site_work_items w LEFT JOIN pages t ON t.id = w.page_id
WHERE w.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
  AND w.item_type='unbuilt_internal_link' AND w.created_at > '2026-08-09 13:10:00+00'
ORDER BY t.name, w.created_at;

-- and the target itself:
SELECT build_status, COALESCE(deployed_at::text,'NEVER') AS deployed,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id) AS comps
FROM pages p WHERE p.id='769e3b72-e9fa-4177-aa59-8c8d068e33f9';  -- grip-styles
```

**State at 13:57 UTC when this was written:** all 10 still `triaged`, grip-styles
`planned` / never deployed / 0 components, **11 items still ahead of them in the
queue** (down from 20 at 13:20 — it is draining, roughly one dispatch run every 8
minutes, 5 items per run).

### What a PASS looks like — all four legs, one row

For any of the six items whose target is **grip-styles**:

| leg | required value |
|---|---|
| `saved_page_name` | **`grip-styles`** — the TARGET. This is mig 342's leg; `beginners`/any container here means 342 regressed |
| `saved_page_id` = `deployed_page_id` | both `769e3b72…` (grip-styles) |
| `verif` | `verified` |
| `detail` | must say **"target page … has shipped"** — disjunct **(a)** |
| `pages.deployed_at` (grip-styles) | non-NULL |
| `curl -s -o /dev/null -w '%{http_code}' https://dartsonline.com/blog/grip-styles.html` | **200** |

**The detail text is the load-bearing part, not the `verified` stamp.** A `verified`
whose detail says *"href … is no longer rendered"* is disjunct **(b)** — the link
was removed rather than the target built. That is a legitimate resolution of the
work item but it is **NOT the convergence proof**, and it is exactly what happened
on the morning run (see "already proven" below).

**Validate your query against a control before trusting a pass.** Run the same
query against item `338deb27` and it must come out WRONG: `saved_page_name` =
`beginners` (the container), `saved_page_id` = `5009f5c8`, `deployed_page_id`
**empty**, `verif` = `verified` via disjunct (b). If that row does not read wrong,
your jsonb paths are typos and any "pass" you read next is meaningless.

### If it landed → close the bug

220 is then fixed-and-live end to end. **Per the owner's 08-06 ruling a finished bug
STAYS in `bugs_open/` — update the file head, do not move the file.** Record the
proof in the bug file, NOTES, and `README_where_we_are.md`, and write a SUMMARY (the
last one predates the whole acceptance arc, so the five headings would genuinely
differ — this is a real milestone).

### If it has NOT landed → do not force it

The queue drains on its own; independent scheduled dispatch runs hit this site every
5–15 minutes. Watch, do not re-fire: a second loop racing a scheduled one only
muddies the evidence. A watcher script is at
`scratchpad/watch_220_convergence.sh` in the 2026-08-09 session's scratchpad; if you
write your own, **key it on item ids and a FIXED timestamp floor — a
`created_at > now()-interval` window blinds itself as the items age.**

If it has genuinely stalled (no dispatch after ~2 hours, no scheduled runs
arriving), re-fire the loop with `scratchpad/fire_improvement_loop_dartsonline.sh`.
Two rules: payload in the container COMMAND (never stdin — `kubectl run -i | kcat -P`
drops ~4 in 5 silently at exit 0) and the command must end `&& echo PUBLISH_OK`. **No
`PUBLISH_OK` means nothing was published.** The shipped
`060improvement_loop/076_improvement_loop_trigger.sh` **cannot** be used: it
re-assigns SITE_ID/DOMAIN to robot-hands.com *after* parsing its own arguments.

### The four items that are EXPECTED to fail

Four of the ten target `section-index` directories (`brands-index` ×3 —
`69818add`/`0469f44f`/`6e1b562b`; `shop-index` ×1 — `b4184d0f`). These are **deferred
candidate 4's demand signal** and should fail LOUDLY rather than converge. `failed`
on those is the designed outcome, not a regression — record it as the demand signal
and leave candidate 4 deferred unless the owner wants it picked up.

## Already proven — do NOT re-derive any of this

- **All three config legs live** (re-checked 13:0x UTC): dispatcher
  `"page_id?": "current_item.page_id"`; `load_page_record.authoritative_page_id` =
  `input_data.page_id` (mig 340); `save_sections.page_name_field` = `page_record.name`
  (mig 342). Queries in the RUNBOOK — note `input_mapping` sits under
  `call_handler->'config'`, one level deeper than the step, and reading the step
  returns EMPTY, which looks exactly like a missing leg.
- **Binary carries the fix on the current pods** (`agent-chassis-5c5bbf8548-khpl4`
  /`-mkdjp`, created 12:23Z): `authoritative_page_id` 3, `unbuilt_internal_link` 7,
  invented negative control 0, both replicas. Re-grep after any roll.
- **Council APPROVED at round 2**, corr `def4441c`. 342's own submission was refused
  client-side by the docs-scope filter (config-only change) — recorded, not forced.
- **Register WII-012**; migs 340/341/342 applied, read back, recorded (⚠ "342" also
  names the thunder lane's unrelated file — resolve by slug, never by number).
- **Routing + verifier proven live on the morning run** (corr `110acf5a`): dispatch
  of `338deb27` targeted grip-styles (the TARGET) and the deploy skipped honestly
  ("no component rows yet") instead of shipping the container's file. The verifier
  stamped `_verification` — but via disjunct (b), which is why the proof is still owed.
- **The beginners repair is DONE and proven at the artefact** (was priority 1 in the
  previous handoff): components rewritten 10:13:58 with beginners' own copy, page
  deployed 12:31 (after the repair), `curl .../blog/beginners.html` → 200 carrying
  both signature phrases. The two held rerenders (`47ba8f2c`, `3c10ab6c`) can stay
  cancelled; nothing is queued to republish the contamination. **The beginners →
  grip-styles link is back and still 404s, which is what keeps the acceptance test
  fair.**
- **`cancelled` frees the dedup slot** (`workItemTerminalStatuses`,
  `work_items_common.go:47`, migration 157) — that is why the cancelled morning items
  re-minted. If a future run mints nothing, check that lockstep FIRST.

## Also open, lower priority

1. **The residue** (filed 08-09 in the bug file, needs no action from you): the
   verifier is not retroactive. Of 6 pre-verifier `complete` items fleet-wide, one is
   live damage today — **mortgagecalculator.co.uk** serves a 200 page
   (`/guides/first-time-buyer/index.html`) linking to a 404
   (`/scorecard-simulator.html`) while its fixing item has read `complete` since
   08-05. **No repair was minted deliberately:** `complete` frees the dedup slot, so
   the next discovery pass over that site re-mints and converges honestly under the
   verifier. Whoever runs an improvement loop there gets a second, independent
   end-to-end proof for free.
2. **Candidate 4** — route unbuilt targets by `page_type` via `availableBuilders`
   (package-direction refactor). Deferred on record; demand signal above.
3. **Adjacent, told not owned**: the dead_fragment lane's
   `VerifyDeadFragmentLinkResolved` still uses the LIKE-concatenation shape the
   council flagged here (over-match on `_`). Flagged in bug 220 § COUNCIL TRAIL and
   in the code comment.
4. **Not this lane's file, but worth someone's ten minutes**:
   `scripts/trigger-landmine-verifier.sh:84` uses the unsafe kcat stdin pattern with
   no receipt (the ~4-in-5 silent-drop trap). A dispatch from it on 08-09 DID land,
   which is how the trap survives. `fire_improvement_loop_dartsonline.sh` is a
   working template for the fix.

## Gotchas this lane has already paid for — do not re-derive

- **Dispatch is ASCENDING priority at `max_items` = 5 per run.** Lower number goes
  first (measured twice: 5→8→10→10 this run; 35 before 45 before 80 on the morning
  run). One run takes a bite, not the queue. Reading `priority` as urgency inverts
  every estimate you make about when your item runs.
- **A page's served URL is NOT derivable from `pages.name`** — `beginners` serves at
  `/blog/beginners.html`, `guide-first-time-buyer` at
  `/guides/first-time-buyer/index.html`. A guessed URL 404s, and a 404 is exactly the
  signal meaning "never deployed", so the wrong guess reads as a finding. Read
  `pages.url`. (Filed as a LANDMINE 08-09; landmine-verifier returned STILL_VALID.)
- **A `complete` work item is not a repaired artefact.** The repair above read
  `complete` at 10:14; that proved nothing until the components, the deploy timestamp
  and the served bytes were each checked.
- **Ask by IDENTITY, not `count(*)`.** The residue census's count said "6 complete"
  and hid that 3 targets shipped later by unrelated work, 1 had its link removed, and
  only 1 was real damage.
- **The verifier and the discovery check read STORED `rendered_html`, not served.**
  They agree with each other, so no churn, but a served page can lag either way.
- **`kubectl exec -i` inside a `while read` loop eats the loop's stdin** — `</dev/null` it.
- **Migration numbers, bug numbers and any census expire in hours.** Re-check at the
  moment of use, not the moment of planning.

## Where everything lives

- **Bug file**: `bugs_open/220_HANDOFF_2026-08-08_unbuilt_link_dispatch_rebuilds_the_container_and_reads_green.md`
  — mechanism, reproduction, census, council trail, the 08-09 post-roll acceptance
  section, and the 08-09 midday residue section.
- **This dir**: PLAN (design + reasons), NOTES (append-only technical log, newest at
  the bottom — the cold-start read), RUNBOOK (every query that was hard to get right,
  including the acceptance assertion with its control), README_where_we_are (owner
  prose), submission JSONs (r1, r2, 342).
- **Memory**: `bugfix-220-unbuilt-link-dispatch-workstream.md` + the
  MEMORY_workstreams line.
- **Commits this session (2026-08-09 afternoon)**: `d2a72347d` repair proof + re-fire,
  `37f1a88ec` residue census, `599388b17` run ledger, `86f9bbe9b` + `e64b9c2ab`
  runbook, `21c69190b` dispatch-cadence facts. ⚠ the URL landmine's LANDMINES.md
  lines were swept into another session's commit `190ee4568` — nothing lost, both
  entries are at HEAD, and forward-only forbade an amend.
