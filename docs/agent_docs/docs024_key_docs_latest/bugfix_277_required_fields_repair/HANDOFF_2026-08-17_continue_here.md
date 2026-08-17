# HANDOFF — 2026-08-17, fresh chat starts here: 083 APPROVED and hardened; 277 close-out is a clock, not work; two owner decisions PENDING; router_engine still unstarted

**Supersedes `HANDOFF_2026-08-16_continue_here.md`** (read this one; that one's §2 priority list
is done or reassessed). Everything below measured 2026-08-17 ~11:00–12:45Z against chassis
`v1.0.1305` unless dated otherwise. Read this file FROM DISK, then `NOTES_…` from the bottom.

## 1. What is LIVE (verify, do not trust)

| thing | state | how to re-verify |
|---|---|---|
| chassis | `v1.0.1305`, 2 pods, OCI `revision=6a782274b`; carries `3c6354059` (born-`detected` producer) | `docker image inspect docker.io/aqls/agent-chassis:v1.0.1305 --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'`, then `git merge-base --is-ancestor 3c6354059 <rev>`. **The `build provenance` log line has scrolled — absent from `--tail=100000`. Empty ≠ unstamped.** |
| `required-fields-missing-handler` (seed 410 v3, CQ-023) | live, 1 active row, 8 routes — **but only 3 have ever been taken** | route census in §3 |
| `detected-item-promoter` (seed 430 + **444**, SCH-026) | live, 900s, **hardened 2026-08-17**, council **APPROVED** | `SELECT enabled, last_triggered_at FROM scheduled_tasks WHERE name='detected-item-promoter'` |
| migration `444` door-closers | applied + ledger-recorded, `_ROLLBACK.sql` alongside; both doors hold **0 rows** | `SELECT pre_query FROM scheduled_tasks WHERE name='detected-item-promoter'` — must contain `wi.pipeline IN ('build', 'content', 'design')` and `0.25 * (c + f)` |
| council `05a3d1c8` | **APPROVED round 2**, 12 seats, 3 abstained, 2 advisories (both answered) | `SELECT metadata->>'decision' FROM diagnosis_artifacts WHERE correlation_id='05a3d1c8-39c1-484d-85a8-11a47f4b07f3' AND kind='council_report' ORDER BY created_at` |
| council `7b0e2833` (the router) | REVISE ×4, **deliberately not resubmitted** — see §4 | — |
| RFC_030 (router engine) | RULED + SCHEDULED, lane at `docs024_key_docs_latest/router_engine/`, **nothing built** | its own `HANDOFF_2026-08-15_continue_here.md` |

## 2. TWO OWNER DECISIONS ARE PENDING — read `README_where_we_are.md` (2026-08-17 entry) first

Both are written up for the owner in plain prose there. **Do not action either without an answer.**

1. **Should the router's two content-rebuilding arms (`file_rewrite`, `file_recreate`) be gated
   until someone has run one by hand?** They have NEVER fired (§3). `bug_historian` objected at
   HIGH on the regeneration risk (closed case 056; `missingkey=zero`). The estate's own remedy is
   the 2026-08-02 §2 ruling (opt-in field, unsafe default OFF — and RFC_022 confirms that shape is
   *not* architecture-scope, so it would go to the normal gate). Not taken unilaterally: it
   changes the safety posture of an owner-blessed mechanism, and the arms are
   `conditional_branch` steps whose `then_step` is data, so "opt-in" means either a new expression
   form or redirecting to a park step — a design call.
2. **Who canaries a held pair, and how long may one sit?** 5 rows held (4 ×
   `page_component_status_drift → component-template-fixer` since 2026-08-10; 1 ×
   `placeholder_contact → page-build-handler` since 08-16). The promoter deliberately holds an
   unproven pair until a human runs one; nothing prompts anyone to be that human, so a held
   finding can sit for ever — a miniature of the bug this lane exists to fix.

## 3. The measurements a new session would otherwise redo

```
-- the router has only ever taken 3 of its 8 routes, and produced ZERO child items
no_content_data 35 · asset_sourced 1 · no_plan_owned 1 · file_rewrite 0 · file_recreate 0
SELECT COALESCE(result->>'route','(none)'), status, count(*) FROM site_work_items
 WHERE item_type='required_fields_missing' AND result ? 'route' GROUP BY 1,2;
SELECT count(*) FROM site_work_items WHERE parent_item_id IN
 (SELECT id FROM site_work_items WHERE item_type='required_fields_missing');   -- 0

-- the detected pile is NOT the promoter's meter (see the criterion-1 correction in bugs_open/083)
82 detected = 77 flag-only with NO handler (image_url_404 41, head_essentials_missing 36)
           +  5 with a handler, of which 0 pass the known-good rule
-- 40 of the 77 were restored to 'detected' DELIBERATELY by bugs_closed/284's migration 442.

-- 033's revalidator now sweeps this type and agrees with the router row-for-row
no_content_data -> unknown (29) · no_plan_owned -> unknown (1) · asset_sourced -> still_holds (1)
-- plus 56 rows of the type auto-closed 'resolved'. The backlog drop is mostly 033's work,
-- not this router's; the router's contribution is that the survivors carry a classification.
```

## 4. Why trail `7b0e2833` gets NO round 5 (this reverses the last handoff's item 2)

Round 4 was gated by **`editquality` HIGH: "a no-op dressed as an edit"** — four edits were
already-committed work re-listed as pending. The suggested "short round 5 citing the owner
rulings" has *no real edits at all* and would be gated identically. **A trail cannot be closed by
narration.** Meanwhile most of round 4's objections are answered by shipped code rather than
argument (the promoter exists; the producer is born-`detected` and live; `prior_art_librarian`
approved the sole-carrier premise itself on 2026-08-17). The surviving residuals are decision 1
above and the RFC_030 seats' real ask — *"acceptable only if RFC_030 is genuinely a hard gate on
a 4th router, not aspirational"* — which belongs to the `router_engine` lane, not to another
round here. Record and leave, as the last handoff itself permitted.

## 5. Owed work, in priority order

1. **Get answers to the two decisions in §2.**
2. **277 close-out is a CLOCK, not work.** Re-check ~**2026-08-22**: (a) churn guard — at day 2,
   exactly **one** new row since the 08-15 assignment, born-detected → promoted → routed → parked,
   **zero `unresolved`**; (b) the two cancelled conversions re-raising (no `cancelled` rows of the
   type remain, so this depends on discovery rotation; if unseen by 08-22, re-file by hand).
   Then 277 → `bugs_closed/` (**both paths on the commit** — LANDMINE).
3. **083 → `bugs_closed/`**: criterion 3 MET (4/4 at the served page), criterion 1 holds under its
   corrected wording, criterion 2 met but **non-discriminating** (already true 6 days before the
   fix — do not bank it). Remaining: let `444`'s doors sit a week and confirm they still hold
   nothing they should not. Note the file also carries a **named open residual** — guardian's
   point that a new pipeline value silently stops being promotable, whose cheap control is to have
   the pre_query return a count of rows a door held so the scheduler log shows it.
4. **Start the `router_engine` lane** (RFC_030) — its own cold-start handoff exists. This is now
   the largest real piece of work in the area, and §4's residual is waiting on it.
5. **Two sibling born-triaged producers** (`check_integrity.go` ×2 sites,
   `check_tool_acceptance_due.go` ×1) — their lanes' call; both pairs are known-good so the
   promoter would carry them. Mention, don't do.
6. **Raise the council gate's config blind spot** as its own item: `097` scopes on
   `platform/`/`internal/`/`pkg/`, so a mechanism shipping as `agent_definitions` /
   `scheduled_tasks` config cannot be submitted at all. Round 1 passed only because a Go file rode
   along; round 2 needed `FORCE=1`. A large fraction of this estate's behaviour is config.

## 6. Landmines this lane hit (all now in LANDMINES.md / WRONG_CALLS.md — grep before repeating)
- **`failed` rows carry NO `completed_at`** (0 of 265, vs 5882/5921 completes). Any pair-health or
  success-rate query keyed on that column returns a uniform 100% and *cannot* come out otherwise.
  Key on `updated_at`. This nearly killed 444's whole finding. **New LANDMINES entry.**
- **A verify criterion is a figure and goes stale like one.** 083's criterion 2 was carried
  forward twice while already satisfied. Re-measure criteria, not just evidence.
- **A zero from a detector you just wrote needs a demand control.** Three pages that still carry
  the defect are what made "0 hits" mean something.
- **Fetch the URL the row literally names** — the rows store `/guides/index.html`; requesting
  `/guides/` 404s on that site. A whole sample failing is evidence about the instrument first.
- The loop's `mark_complete` REPLACES `result`; `landmines-sync.py --apply` alone consumes the
  new-entry status (use `landmines-verify-dispatch.sh`); the scheduler GATE requires
  `pipeline='build'` while the loop's loader does not filter; never `--apply` the migration runner
  unscoped (use `--record-only <file> --note`).
- **New seam nobody designed:** 033's revalidator now writes `result` on rows this router parked.
  They compose correctly today (`route` and `revalidation` side by side); nothing guarantees it.

## 7. Session-start checklist
`git log --oneline -10` · re-read this file from disk · `scripts/who-owns.py 277` and `083`
(by SLUG — 083 is an ambiguous number) · grep live `.jsonl` for `router_engine|RFC_030|7b0e2833`
· re-measure §1 and §3 · then §5 item 1.
