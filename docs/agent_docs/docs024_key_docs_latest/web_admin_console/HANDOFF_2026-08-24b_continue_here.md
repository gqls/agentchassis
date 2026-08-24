# HANDOFF 2026-08-24b — Builds screen built; deploy ordering is the live constraint

Supersedes `HANDOFF_2026-08-24_continue_here.md` for STATE; that file's §1 table (edge
infrastructure proofs) and §2/§3 (owner box steps) remain the reference — read it after this.

## 0. What changed since the morning handoff

1. **The build-steps screen (§4 there) is BUILT and COMMITTED, not live.**
   - Backend `e6350e74b` (`internal/core-manager/admin/`): `GET /admin/workflows` gains
     `site_id` filter (indexed column, PLAN §6a) + per-row `has_step_error`
     (`collected_data ? '__step_error'`, exact per §6b) + `site_id`; terminate's
     `orchestrator_state` → `orchestration_states` (ADM-002 B2 — register entry updated);
     `HandleUpdateSiteSpec` evidence_base guard (PLAN §6h hazard: parse-before-write, 409
     `EMPTY_EVIDENCE_BASE` unless `confirm_empty`, stored counts + `superseded` returned).
     Six sqlmock tests. `Council-Submitted: 45b3c93f-7937-474d-8234-31c39bab033b` —
     **verdict unread as of this writing; read it and act on a REVISE** (find the run by
     payload: `collected_data->'input_data'->>'fix_correlation_id' = '45b3c93f-…'`; a chassis
     roll was in flight around submission time, so if NO orchestration row exists after the
     queue latency window, the dispatch may have hit the ~300s post-restart drop — that is
     the one case where resubmission with the same file + `RESUBMIT_CORR` is warranted).
   - Frontend `b3fbfdd02` (`frontends/admin-dashboard/src/App.tsx`): `BuildsView` — per-site
     stage timeline over `site_work_items` in the 082 cascade order with durations (§6d
     shape), other-activity rollup, orchestration drill-down surfacing `__step_error` in red
     whatever `status` says, resume/terminate behind confirms; §6g `⚠ overwritten ×N`
     badges (red when unlocked); §6h ENFORCED/advisory chips, counts echoed on
     evidence_base saves, 409→confirm→`confirm_empty` flow, prohibition nudge. Vite build
     proven via the Dockerfile builder stage.

2. **⚠ DEPLOY ORDERING — the one live constraint.** New SPA against old core-manager is the
   misleading combo (gin ignores the unknown `site_id` param → BuildsView shows the WHOLE
   FLEET's workflows labelled as the site's; no `has_step_error`; `confirm_empty` ignored).
   So: **(1)** wait until core-manager runs a build carrying `e6350e74b` —
   `kubectl -n ai-persona-system logs -l app=core-manager --tail=300 | grep -m1 'build provenance'`
   then `git merge-base --is-ancestor e6350e74b <stamp>`; **(2)** then
   `make admin-dashboard` (build+push+deploy; it builds the WORKING TREE — ensure App.tsx is
   at/after `b3fbfdd02` and bump `IMAGE_TAG`). Old SPA + new backend is harmless.

3. **Council verdict on the `/c/` prefetch guard (`6b1726ab…`): APPROVED round 1**
   (2026-08-23 12:07Z, one advisory objection, none high-severity). Morning handoff §5.2
   CLOSED. `0e9cb31ee`'s `Council-Submitted:` trailer is credited automatically by 098.

4. **LANDMINES gained the evidence_base wrong-shape silent-off entry** (footprint
   `site_specs`/`ParseEvidenceBase`/SQL seeds; the check is read-back-the-counts). Synced +
   verifier dispatched. The FILE is uncommitted — it carries the 333 lane's WIP entries too;
   whoever commits takes both, append-only, say so in the message.

5. **Falsifiers re-run this session:** `links.webdesign.uk` still does not resolve (morning
   §2 still owner-pending — the `/c/` move remains the top owner action);
   `customer_access_tokens` = 0; `www.apis.uk` 301s to apex **via the portfolio-sites-router
   Worker — NOT evidence the §3 Redirect Rule was applied** (NOTES 2026-08-24 correction;
   do not re-close that falsifier off a curl).

## 1. Open / owed (delta against the morning handoff's §5)

1. **Read the `45b3c93f` council verdict** (item 0.1 above) — the only owed action of mine.
2. **Deploy pair** per §0.2 when the roll lands — then eyeball `admin.apis.uk` → a site →
   Builds against the apis.uk chain (PLAN §6d table is the expected shape, ~67 min build).
3. Architecture boundary review when `links.webdesign.uk` goes live (morning §5.1) — unmet
   until the owner applies §2.
4. Morning §5.3–§5.7 unchanged (mail-scanner residual, HOLD ban, webdesign lane items, VPN
   parked, ADM-002 staleness — B2 now fixed-committed, B3/B4 still `[UNVERIFIED]`).

## 2. Falsifiers for THIS handoff

- Whether core-manager's stamp carries `e6350e74b` (the roll cadence is daily; the owner's
  fresh-chassis roll of 2026-08-24 afternoon may or may not have included core-manager —
  ask the pod, per service, never the fleet).
- The `45b3c93f` verdict may have landed (or the dispatch may have been dropped — §0.1).
- The dashboard image may already have been rebuilt by another session; check the running
  SPA for a "Builds" button before rebuilding.
- `customer_access_tokens` and handed_over counts — 0 as of 2026-08-24; any non-zero expires
  the "nothing at risk" claims inherited from the morning handoff.
