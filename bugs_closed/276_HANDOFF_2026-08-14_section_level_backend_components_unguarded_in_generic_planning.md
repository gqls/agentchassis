# 276 — a backend-requiring SECTION component can still be planned onto a backend-less site: VMB-010's planner half was never built

## STATUS: FIXED AND LIVE 2026-08-15 — three call sites gated, not the one this file names

**Filed 2026-08-14, webdesign_uk_build_service lane**, at the council's
direction — the REVISE round on migration 406 (corr
`c78ed496-a6f4-4ebc-a6c3-1fc4a9221546`) objected, correctly, that the tool
half of VMB-010 got the rigorous gate while the section half stayed a register
note: "disclosed exposure is still exposure — a human should confirm the
section-level half is tracked with a concrete follow-up." This file is that
follow-up. Not run through 090 — stated substitution: the gap is asserted by
the concept register itself (VMB-010, "designed, never built", since
2026-06-11) and re-verified against live config today; evidence inline.

## The gap

`content_components` rows tagged `requires-backend` exist at TWO levels:

```sql
SELECT name, component_level FROM content_components
WHERE COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend';
-- intent-probe     | section
-- chat-input-box   | tool
```

- The **tool** path is gated since 2026-08-14: tool-suggester's
  `load_library_tools` only offers such a tool where
  `sites.deploy_config->'capabilities' ? 'backend'`
  (`sql_for_agents/406_tool_suggester_requires_backend_gate.sql`).
- The **section** path is NOT gated: nothing in generic section planning
  (`plan_sections` / the component library load) reads the tag. VMB-010's
  designed remedy — `load_components` gains
  `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` so such
  components are opt-in via roadmap `section_types` only — was never applied.
  A static B2-hosted site can be planned an `intent-probe` section whose
  frontend POSTs to a backend the site does not have.

## Why this is the documented partial-fix shape

One call site of a shared judgement ("can this site support a backend
component?") now has the rigorous fix while the sibling stays unguarded —
016b §9's class, and `bugs_open/093`'s shape (one guarded call site, the other
path unchecked). The OUTCOME class already has a live instance in
`bugs_open/228` (contact-block fakes success with no transport;
`bugs_closed/017` is the ancestor): a visitor-facing component that silently
does nothing. This file is about the GATE that should make that outcome
unrepresentable for tagged components.

## Fix candidates (ranked by what closes the door)

1. **Apply VMB-010's designed filter** to the generic component-library load
   used by section planning, with the same capability escape hatch as 406:
   excluded UNLESS the site carries `capabilities:['backend']` (or the roadmap
   explicitly names the section type). One query change if the load is config
   (find the live loader first — `load_component_library_actions.go` and the
   planner's library step both filter by level; the gate belongs where
   section candidates are enumerated). Mirror 406's discipline: id-scoped
   UPDATE, pre-state gate, DO/RAISE verify, disagreeing-pair proof.
2. **The audit half** (VMB-010's third piece): a discovery check comparing
   placed components' `requires-*` tags against site capabilities, filing
   `site_work_items` findings — catches instances that pre-date the gate
   (any intent-probe already placed on a static site) and any future writer
   that bypasses planning.
3. Do nothing; keep the register note. Rejected — that is the state the
   council objected to.

## How to verify a fix

Disagreeing pair, same method as 406: run the section-candidate enumeration
for a `capabilities:['backend']` site and for a static site; `intent-probe`
must appear only for the first. For candidate 2: place (or find) an
intent-probe on a static site and confirm a finding is filed.

## Cross-references

- Concept register `vm-backend-sites.md` VMB-010 (the design, and today's
  status note naming this file).
- `sql_for_agents/406_tool_suggester_requires_backend_gate.sql` — the tool
  half, worked example of the migration discipline.
- Council REVISE note for corr `c78ed496-a6f4-4ebc-a6c3-1fc4a9221546`
  (bug_historian objection, 2026-08-14).
- `bugs_open/228`, `bugs_closed/017` — the outcome class this gate exists to
  prevent.

---

# TAKEN UP AND CLOSED 2026-08-15 (workstream `bugfix_276_requires_backend_section_gate`)

Checked `who-owns.py 276` and the owning `webdesign_uk_build_service` lane's own 2026-08-15
handoff first — explicitly flagged "for pickup (either lane, or a fresh thread)". Confirmed
not competing with any live session before starting.

**Correction to fix candidate 1 above: it names one call site. There are three.** A
fleet-wide sweep of every `agent_definitions` step whose query touches `content_components`
(14 hits, all classified) found three section-candidate menus, not one:

| agent type | step | dispatches (30d, measured 2026-08-15) | migration |
|---|---|---|---|
| `content-gap-planner` | `load_available_components` | **131, most recent today** | 418 |
| `build-site-planner` | `load_components` (the one this file names) | 2 | 419 |
| `site-planner` | `load_available_components` | **0, ever** (unbounded check) | 420 |

`content-gap-planner` — never named in this file or in VMB-010 before today — is the
overwhelmingly dominant real placement path. Fixing only `build-site-planner` (candidate 1 as
written) would have left the actual live risk wide open.

**What shipped:**
- `sql_for_agents/418_content_gap_planner_requires_backend_gate.sql` — capability-checked
  gate, `$1 = site_record.site_id` (binding proven: two sibling steps in the same linear
  workflow already resolve it; 131/131 recent orchestration rows confirm it live-resolves).
- `sql_for_agents/419_build_site_planner_requires_backend_gate.sql` — candidate 1's own
  target, same clause, reusing the `$1` migration 407 already proved for this step.
- `sql_for_agents/420_site_planner_requires_backend_gate.sql` — **unconditional** exclusion,
  no capability check. `site-planner` has no `ensure_site_record`-equivalent step anywhere in
  its workflow, so no site-id binding could be proven; an unproven guess risks a nil-param
  hard-fail (outage-class) on code that has never once executed, for no present benefit.
  Unconditional removal closes the door without that risk.

All three are id-scoped with a byte-exact pre-state gate (concurrent-edit protection — the
`content-gap-planner` row was touched twice by an unrelated session, migrations 413/414, in
the hours around this work), dry-run tested (COMMIT swapped for ROLLBACK, both verify `DO`
blocks passing), and the pre-state gate was mutation-tested against a deliberately wrong
literal to confirm it can actually fail. Each carries a `_ROLLBACK.sql` sidecar.

**Verified live**, disagreeing pair (real site ids: `relojistas.com` = backend-capable,
`gamesdesign.co.uk` = static control): both capability-gated migrations return identical row
sets for the backend site and exactly one fewer row (`intent-probe`) for the static site,
before AND after apply; the unconditional migration removes exactly `intent-probe` with no
site id involved.

**Fix candidate 2 (the audit check) stays explicitly deferred** — VMB-010's own separate
"Later" item; nothing live to audit today (intent-probe's sole placement is already
backend-capable); it is Go discovery-check work (image rebuild) outside this config-only
change.

Council-submitted with `FORCE=1` (config-only, no Go bytes changed — same reasoning migration
406 itself used successfully): `SUBMISSION_CORR=d1ae766d-b1e4-4dec-a8e4-efd9ef064137`.
Applying did not wait on the verdict, per house norm; commit carries `Council-Submitted:`.

Concept register `vm-backend-sites.md` VMB-010 updated (appended, not rewritten) with the
same three-call-site correction. Full record: `docs/agent_docs/docs024_key_docs_latest/bugfix_276_requires_backend_section_gate/`.

**Closure bar met**: all three migrations applied, verified live via disagreeing pair,
register updated — config changes are live immediately on `apply` (no image rebuild
dependency), so "fixed AND live" holds as of 2026-08-15.
