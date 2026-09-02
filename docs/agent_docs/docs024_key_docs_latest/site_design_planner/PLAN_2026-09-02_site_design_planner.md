# PLAN — site-design-planner workstream

**Opened:** 2026-09-02, after the owner renamed this session "site design planner"
and asked to pick up the thread if one exists, or take responsibility for it if not.

## 0. What this workstream IS and is NOT

**Is:** the mechanism-level owner of the `site-design-planner` agent
(`agent_definitions.type = 'site-design-planner'`, `8b3cb270-ee0d-4df2-abb6-e266758d2747`)
and the composition-resolution stage it runs: layout / typography / palette
resolution, `install_site_composition`, the layout/palette/typography library,
and the renderer's theme-resolution cascade that consumes what it writes.
Concept register: DES-003 (pipeline shape), DES-006 (scope: "Choice B",
composition-only — writes exactly one spec aspect, `resolved_composition`),
DES-005 (install semantics), DES-013/014 (the palette/layout/typography split
and library), DES-032 (renderer cascade), DES-037 (scheme-aware layout
matcher). Full entries:
`docs/agent_docs/docs026_concept_register/register/design-composition.md`
(grep `DES-` — the register's own freeze date is 2026-07-13, so anything after
that, including everything below, is not in it yet).

**Is NOT:** `build-site-planner` (writes `sections`/`page_role` plan,
`write_site_plan_action.go`, `validate_site_plan`) or `webdesign-agent` (the
LLM design overlay that renders `styles.css`). Confirmed with the `bugs_open/427`
lane on 2026-09-02 (name collision risk flagged both ways) — they own
build-site-planner/428, this thread owns composition resolution. Do not act on
either of those from here.

**Is NOT** a per-site build thread. Several individual sites already have their
own dedicated sessions/threads (`loancalculator`, `ai-agent-orchestration`,
`mortgagecalculator.co.uk` all appear in `ListAgents` as named sessions, live or
recently offline). Where this thread finds a composition-mechanism defect that
happens to be visible on one of those sites, the finding is contributed to that
site's thread (per CLAUDE.md's "grep before you file" / `who-owns.py` norm), not
acted on unilaterally here — exactly the same rule bug 113's own notes state
repeatedly ("Do not repaint them from here").

## 1. State of the mechanism, as found 2026-09-02

**Mature and deployed.** The composition pipeline (domain-research-classifier →
site-design-planner → webdesign-agent) has been live since ~2026-04-19 and has
absorbed several generations of fixes: migration 025 (palette/layout/typography
split), the scheme-aware weighted layout matcher (DES-037, live 2026-06-25), and
most recently `bugs_open/113` (palette merge — dark sites inheriting light layout
literals for slots their palette didn't define), which is **functionally closed**
as of 2026-08-12 (fleet-wide fix live, last known instance —
`ai-agent-orchestration.com` — repaired at the served artefact) though the bug
file itself stays in `/bugs_open/` per the 2026-08-06 owner ruling (closed bugs
that are fixed-and-live still move to `/bugs_closed/`; 113 is kept open only
because its file still tracks one adjacent, undecided platform question — see §3).

**The "no re-resolve" platform gap that 113 exposed is fixed.** As of chassis
`v1.0.1290` (`fa078ab3d`), `install_site_composition` accepts a per-request
`allow_reinstall` flag (default false, read from the work item `spec`, not just
step config) — a single site can now be re-composed without touching the shared
agent definition. Proven behaviourally on `ai-agent-orchestration.com`.

## 2. Open items found, mechanism vs. site-specific

Three live `site_work_items` currently sit in this agent's territory
(`item_type IN ('needs_composition','needs_new_layout_candidate')`,
non-terminal status) — see RUNBOOK §1 for the query. None has been touched since
2026-08-20. **None require code changes** — they are all "someone needs to look
and decide", which is why they are stuck:

| site | item | status | what it's waiting on |
|---|---|---|---|
| `loancalculator.co.uk` | `needs_composition` | deferred | a 2026-08-12 site-owner decision to park it until the loancalculator framework rebuild is verified. **Has its own thread** (`loancalculator` in `ListAgents`) — un-parking is their call, not this thread's. |
| `adversecreditmortgage.co.uk` | `needs_composition` | unresolved | **Very likely stale, not live.** `site_specs.resolved_composition` was written 2026-08-25 and the site now serves a real composition (`palette-adversecreditmortgage-co-uk`, `tool-portal-light`, origin `adopted`) — three days after this item went `unresolved` (a two-strike anti-churn stamp, not a fresh failure — see `insertWorkItem` in `load_work_item_actions.go:1987-1997`). This site also has ~230 other `unresolved`/`deferred` items from an Anthropic-API billing outage on 2026-08-25/27 (`"credit balance is too low"`) unrelated to composition — this item is one drop in that flood, not a composition-mechanism defect. Not acted on here; the site's own audit/repair passes should sweep it. |
| `ai-agent-orchestration.com` | `needs_new_layout_candidate` | needs_human_review | **A genuine, by-design HITL signal** (DES-037: "library is missing a layout" — the matcher declined to force a bad fit). Filed 2026-08-12 when `site_tags=[]` at resolution time, so it fell back to `brochure-formal`. Sitting unresolved for three weeks. This is the one item in this table that is actually this mechanism's decision to make — see §3. |

## 3. The one real open question: does `ai-agent-orchestration.com` need a new layout?

**RESOLVED (the "why", not the site's own decision) — 2026-09-02, same day.** It
never got a real answer because the layout resolver never saw real tags: identity
data (`industry: "Technology Services"`) existed since 2026-05-01 and the layout
resolver's own extraction function couldn't reach it — the shared fallback that
would have (`readClassificationFromContext`) is used by the other three
composition resolvers and, it turns out, NOT by this one. Filed and fixed as
`bugs_open/431` / commit `bd8e45aba`. **Still genuinely open:** whether a
re-resolve (now that it would see real signal) lands on a real layout or a
second, better-informed `needs_new_layout_candidate` — that's an empirical
question for whoever triggers the re-resolve, deliberately not predicted here,
and deliberately not triggered from this thread (§4).

## 4. Decisions and their reasons

- **Scope kept narrow to the composition-resolution mechanism**, deliberately
  excluding build-site-planner and per-site build work, because (a) the owner's
  session name names this specific agent, (b) those other territories are
  already actively owned (confirmed via `ListAgents` + a direct message), and
  (c) CLAUDE.md's "before routing work AT an existing bug/item, check who owns
  it" applies exactly here — duplicating effort across sessions on a shared
  tree is the thing that norm exists to prevent.
- **One code fix made and committed** (`bd8e45aba`, §3 above) — a small,
  well-scoped, read-only-derivation change (deleting a private duplicate of an
  already-used shared helper), tested, council-submitted, filed as
  `bugs_open/431`. Judged safe to implement directly rather than only diagnose-
  and-hand-off: it only changes which layout candidates get scored, touches no
  DB writes, and three of its four sibling call sites already prove the pattern
  works.
- **Deliberately did NOT trigger a re-resolve on any live site.** The fix makes
  future re-resolution possible with real signal; whether to actually queue one
  for `ai-agent-orchestration.com` (or the other 3 affected sites) is left to
  their owning thread — `bugs_open/113`'s own repeated instruction ("do not
  repaint them from here") applies exactly here, and a session named
  `ai-agent-orchestration` already exists (offline) in this estate.
