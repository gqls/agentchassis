# NOTES — vigilant designer + offer analyser

Append-only, newest at the bottom. Missteps are the point — record them.

---

## 2026-08-02 — programme opened

**What was done today (planning session):**
- Three exploration passes over the live system (design pipeline; checker/handler machinery;
  premise/offer substrate), two planning passes, four owner decisions taken. Full results
  distilled into PLAN_2026-08-02.
- Key discoveries that shaped the plan, each verified live today:
  - `AIService` is text-only — no vision path anywhere in the chassis. The screenshot critic
    needs a new seam (council-scope).
  - `render-audit-agent` findings stop in `collected_data` — the write tail is the named
    extension point (its 256 seed header says so deliberately).
  - `domain-strategist.create_next_item` unconditionally enqueues `needs_briefing` → a
    rebuild chain. Any premise-refresh automation without the B2 gate rebuilds live sites.
  - `bugs_open/115` is down to 2 open findings (finding 1 closed 07-27 with evidence);
    the rows DO carry a handler (`content-gap-planner` via the fallback rule) — what they
    lack is PROMOTION. The drain problem is a promotion problem plus routing precision.
  - The 3-pass cap's JSON path is `settings.maintenance_profile.audit_pass_count`
    (not `settings.audit_pass_count` as bugs_open/171's text implies).
  - 205 items at `detected` fleet-wide; promotion single-owner since migration 286 (08-02).
- [INFERRED] The fleet homepage-skeleton summary for the critic's distinctiveness judging can
  be computed cheaply from `site_plan_sections` — not yet proven against live data shapes.

**Corrections carried from planning (things first believed, found wrong the same day):**
- First believed "route or refuse-to-write" was needed for 016's item type — wrong; the rows
  route, they are never promoted. Promotion + category precision is the fix.
- First believed the checker enable path warns-and-skips unknown names — stale; since 149 B4
  an unregistered name is FATAL (`allow_unregistered_checks` is the escape hatch).

**Decisions:** see PLAN §Owner decisions + §Decision log (2026-08-02).

**Next:** Phase 0.1 (sweep topic migration), then 0.2 (convergence gate), then 0.3 (write tail).

## 2026-08-02 — A0.1: the "dead topic" premise was wrong, caught before it shipped

Wrote migration 290 (improvement-sweep → scheduler lane). The plan's premise, inherited
from the checker-gaps lane, was "generic.requests is a dead topic nothing consumes".
**Wrong**: the live chassis deployment consumes it as its MAIN lane (`REQUESTS_TOPIC`);
`scheduled.requests` is an EXTRA lane from the bugs_open/030 lane-split. The migration is
still correct — every working scheduled task uses the scheduler lane, and lane separation
is the stated design — but its header now argues from lane-separation, not topic-death.
The empirical no-run at generic (oneshot 07-26) stays honestly `[UNRESOLVED]`.
Check that caught it: read the deployment env (`kubectl get deploy -o jsonpath` on the
topic env vars) before repeating a topic-liveness claim. Correction appended to the
checker-gaps NOTES same hour.

Also learned: `run-migrations.sh --apply` takes EVERY pending file, and the tree carries
OTHER SESSIONS' uncommitted migrations (213/214, bugfix_029 lane). Apply strategy must
scope to 290 only (single `psql -f` + `--record-only`), never a blanket `--apply`.

## 2026-08-02 — A0.2 shipped (291): convergence gate live, one proof owed

- `page_components.content_hash` is a DEAD COLUMN (0/1,183 populated) — the plan's
  fingerprint input; caught by counting before building. Function hashes
  md5(rendered_html) instead.
- `enrich_news_feed` carries its error edge INSIDE config (`config.error_step`) — a fifth
  edge shape 288's dangling-edge checker does not cover. 291's guard covers all five.
- No `restore_agent_snapshot()` exists (checked pg_proc) — rollback recipe restores from
  the snapshot row directly.
- 291 applied alone + recorded; guard passed on live probe (fingerprint stable,
  audit_due=t on a never-audited site). **OWED → A0.4: one witnessed run through the
  gate** — the guard proves SQL, not the engine's parse of `audit_state.audit_due == true`
  or the two-param `record_audit_pass` binding. 171 annotated, closure deferred to that run.
