# RESUME HERE — gripper dossier pilot

**Last updated 2026-07-27 19:20Z.** Read this first, then `NOTES_…md` (bottom
up) for the technical log and `README_where_we_are.md` for the owner's account.
Design of record: `DESIGN_2026-07-24_gripper_dossier_pilot.md` — **but read its
§2 CORRECTION before building anything public-facing.**

---

## State in one line

**The cluster half is BUILT, LIVE and PROVEN end to end.** All three DESIGN §6
fixtures pass on the real site, including the failing branch. The public-facing
half is not built and its design was rewritten on 07-26.

## What is proven, with the evidence

| fixture | result | evidence |
|---|---|---|
| 1 — success | **PASS** | `robot-hands.com/reports/d1a371be-04a5-4ee6-b744-d64c6fd9e7c4.html` HTTP 200, 43,049 B, carries the substituted formula literal `(2.5 × 12 × 2) ÷ (0.15 × 2)`; negative control `(9.9 × 99 × 9)` absent; sidecar `ready`; 0 refs on the homepage |
| 2 — honest no-match | **PASS** | `…/reports/29c3f8aa-3246-4a81-be8a-1e6b237cc467.html` HTTP 200, 41,670 B, item **`complete`** (a no-match is a SUCCESS), carries *"No gripper in this index meets the requirement"* verbatim, zero Match/Marginal |
| 3 — induced failure | **PASS** | `mass_kg="not-a-number"` → `score_grippers` hard error (never a guessed default) → `handle_failure` → `publish_failed` → `fail_workflow`; item **`failed`**; `…/reports/edd863e8-445b-494e-a5d7-5ebdeb6d68cb.json` serves `{"status":"failed"}` HTTP 200; **0 HTML pages created** |

Live on chassis **v1.0.1175**. Seeds applied: **204, 207, 209, 210**.

## Current switch positions — everything is OFF

- `report-dispatch` **disabled**, `report-request-pull` **disabled**. Nothing runs
  until someone enables them. Enable order is `report-dispatch` first, always.
- Three `source='manual-test'` work items remain (fixtures 1–3), plus **two live
  report pages left up for owner inspection**. **Cleanup is owed** once the owner
  has read them (DESIGN §6 / §8.9): delete the `manual-test` rows and the two
  pages. Do not clean up unasked — the owner asked to see them.
- **Seed 208 is committed but deliberately NOT applied.** Its `base_url` points at
  `…/api/gripper/v1`, a path the island Caddy allowlist 404s. Re-seed it to
  `https://tools.apis.uk/api/v1/tools/gripper` when the route exists.

## Next actions, in order

1. **Do NOT write `cmd/gripper-intake/`.** It would be the estate's fourth VM fork.
   The public half is a route group **inside the existing `tools-api`**
   (`internal/tools-api/`), which already has per-request CORS from the island's
   own `sites` table, a rate limiter, an input cap and a key. `tools-api` is the
   **gauntlet thread's**, and `bugs_open/083` is open against its error handling —
   coordinate before editing (`scripts/who-owns.py`).
2. **Before that, land two shared pieces** (`features_open/024` A2/A3), or the
   estate forks at exactly this point:
   - a **mailer** in `platform/` — `grep -rn "net/smtp" --include=*.go platform/
     internal/ cmd/` returns **nothing**; the only working one is idea.uk's VM app,
     outside the build;
   - **`platform/httpguard`** — one per-IP limiter, one CORS policy, one honeypot +
     timing gate. The public API's current limiter is the weakest of the three that
     exist.
3. **Then generalise `score_grippers`** into a config-driven engine with its rule
   table in `site_specs` (owner ruling 07-27: *finish the pilot first, generalise
   after*). Pattern to copy: `CHVerticalProfile`
   (`companies_house_vertical_profiles.go`) — a table, not Go per site. Nine of 296
   registry entries currently serve two of ~1,000 sites.

## Landmines that cost real time here

- **`scheduled_tasks.target_topic`'s column DEFAULT (`system.agent.generic.requests`)
  is a topic NOTHING CONSUMES.** 18 of 18 enabled tasks use
  `system.agent.scheduled.requests`. It fails **silently and looks healthy**: the
  scheduler logs *"Successfully produced message"* and *"Triggered task"* and stamps
  both timestamps — that is the **normal** fire-and-forget path
  (`cmd/scheduler/main.go:287-296`), so equal timestamps prove nothing either way.
  Discriminating evidence is downstream only: zero `orchestration_states` rows for
  the target agent type, and zero mention of the correlation_id in the chassis log.
- **`create_report_page` requires `request_id` to be a real UUID** — it becomes the
  page's public URL. An invalid one also silently disables the failure sidecar,
  because `handle_failure` builds it from the same field.
- **`complete` / `deployed_at` is not fetchability** (`bugs_open/098`). Fixture 2 was
  **404 for ~2 minutes** after the item said `complete`, then 200. Poll the URL.
- **Verify against the pod that is running NOW.** The chassis rolled twice during
  one session (v1.0.1173 at 13:45, v1.0.1175 at 18:00, both other threads). The pod
  I first pod-grepped no longer existed an hour later.
- **`bugs_open/029-hung-spawns` is live and roll-adjacent.** One run here hung at
  `spawn_handler` for 4m45s. Stopgap: mark the orchestration `FAILED`, reset the work
  item (`claimed_by=NULL, claimed_at=NULL, attempt_count=0`); it re-claims next tick.
  A hang has **`handler_spawned` ABSENT** in `collected_data` — check that field
  rather than sampling `current_step`, which shows the same value mid-flight.

## Owner-gated

- Anthropic key **issued 07-27** (capped per project, not per key — accepted for
  now). Gates only the island/public half.
- Cleanup of the two live fixture pages awaits the owner reading them.
- The soft-launch decision (unlinked → footer nav link) is still theirs.
