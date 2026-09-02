# 447 — the improvement loop's tool-suggester reads `identity`, never the brief, and planted the SIBLING'S tools by name on a site whose brief forbids tool pages — while the brief-fidelity auditor watched in record mode

**Filed 2026-09-02 ~21:45Z** by the `gamedesign.uk` lane, from the first improvement-loop run over
a FRESH-path editorial site. **Status: OPEN, UNOWNED** (instance held by this lane; class is the
improvement-loop / tool-suggester owners'). Related: `bugs_open/446` (the review that triggered the
loop), positioning's cross-TLD twin rule (P5), `bugs_open/439` (adoption carries the sibling's
brand — same collision family, other direction).

⚠ Number collision risk — 446 was the highest at filing; resolve by slug.

## 1. The one-paragraph version

gamedesign.uk's brief — v1 and v2, both in `mission_brief.text` — says *"It does not publish
calculators, simulators, tool pages or a guide library"* and *"where a piece would naturally point
the reader at a tool it links to the one on gamesdesign.co.uk rather than building its own."*
Positioning's GD2 row says the same: tools live on the sibling; duplication is the collision. The
owner ran the improvement loop over the site at 19:55Z. Its `design-discovery-agent` ran
`evaluate_tools`; **`tool-suggester` read `identity` (×8 references in its definition) and
`classification` (×2) — and nothing else** — decided a game design site needs game design tools,
and filed **eight `add_tool` items**, six of them the SIBLING's tools by name. `tool-deployer` then
planted **12 `planned` pages**, nav changes and "add tool reference to the index page" rewrites.
Meanwhile the `brief-fidelity-audit` filed, at 20:04:00, *"[verdict, not dispatched] The brief
states 'gamedesign.uk must not duplicate any of' the sister site's calculator and tool content"*
— it saw the violation as it happened and, being record-mode, did nothing. Dispatch-mode
suggester, record-mode guard.

## 2. Measured, 2026-09-02

`site_work_items`, site `8f17eb73-fc74-4718-8371-b3125bc4e414`:

| time | item | created_by → handler | what |
|---|---|---|---|
| 19:56:15 | `evaluate_tools` complete | design-discovery-agent → tool-suggester | the loop step |
| 20:02:15 | `tools` spec written | tool-suggester | *"gamedesign.uk serves senior, lead, principal… designers… wrestling with real studio problems"* → tools |
| 20:02:19–:27 | `add_tool` ×4 complete | tool-suggester → **tool-deployer** | Combat Balance Comparison Tool · Economy Sink & Faucet Flow Modeller · XP Curve Designer · Damage Formula Designer — **all four are `gamesdesign.co.uk` pages** (`tool-combat-balance-comparator`, `tool-economy-flow-modeller`, `tool-xp-curve-designer`, `tool-damage-formula-designer`, deployed there 2026-09-01) |
| 20:02:29–:30 | `add_tool` ×2 complete | tool-suggester → **tool-generator** | Design Role Scope Checker · Design Pipeline Friction Diagnostic (new) |
| 20:02:32–:34 | `add_tool` ×2 triaged | → tool-deployer | Stat Budget Allocator · Loot Table Balancer — **also the sibling's** |
| 20:04:00 | `needs_content_planning` deferred | brief-fidelity-audit | *"[verdict, not dispatched] The brief states 'gamedesign.uk must not duplicate any of' the sister site's…"* |
| 20:07:25–20:09:42 | 12 `pages` rows `planned` | tool-deployer | `tool-*` + `tool-*-guide` for each of the six |
| 20:07:25… | `content_rewrite` ×5, `needs_content_page` ×8, `nav_drift` ×N | tool-deployer / tool-generator | "Add Combat Balance Comparison Tool tool reference to index page", "Write companion guide…", "Nav membership declared for tool-combat-balance-comparator — rebuild nav" |

**Why the loop saw a tool gap at all:** the `acceptance-discovery-agent` filed
`structure_floor_unmet` at 19:57 — *"1 of 6 reader-facing structures delivered across 4 pages:
tool…"* — a structure floor that counts tools as a reader-facing structure every site should have.
On the practice seat of a twin pair, the floor is wrong by design.

**What the suggester reads** (`agent_definitions.default_config`, type `tool-suggester`, active):
`site_specs.specs.identity` ×8, `site_specs.specs.classification` ×2. **Zero** references to
`mission_brief`, `mission`, `strategy`, `content_direction` or `briefing`; the prompt contains no
occurrence of *sister*, *sibling*, *duplicate*, *positioning* or *must not*.

## 3. Held by this lane, reversibly (`SEED_2026-09-02d_hold_tool_suggester_plants.sql`)

19 + 7 + 2 items cancelled with the reason in `result`; 12 `tool-*` pages archived (never built,
never deployed, never linked); the two generated components remain in the library for the owner to
reinstate. **Not deleted** — 432's lesson. Cascade #2 (brief v2, corr `aab87c0c`) was already
claimed by the classifier when the hold went in; its plan must be read for the same shape.

## 4. Why it is a class, not a gamedesign.uk quirk

- **Every twin pair built under P5 has one seat that must NOT host the other's tools.** The
  suggester cannot know which seat it is on, because it never reads the document that says so.
- **The library deployer makes the collision literal:** `tool-deployer` copies the library
  component the sibling already uses, so the practice seat would serve the same calculator under
  the same name at a second domain — the SEO cannibalisation positioning's rule exists to prevent.
- **The guard that would have stopped it is record-only.** `brief-fidelity-audit` produced the
  right verdict 2 minutes after the plants began; a `[verdict, not dispatched]` row is not a control.
  Same finding as `bugs_open/446` §4a from the other side: the dispatching agents are the ones
  that do not read the brief, and the agents that do read it do not dispatch.
- **The structure floor is vertical-blind:** "1 of 6 reader-facing structures: tool" treats a tool
  as a universal floor. A law firm, a restructuring-finance journal (`oufe.com`) and the practice
  seat of a games pair all legitimately score zero.

## 5. Fix candidates, ordered by what closes the door

1. **tool-suggester reads the brief and refuses** when `mission_brief.text` (or `content_direction`)
   states the site publishes no tools — one input and one guard in its definition; the text is
   already there, it is simply not in the prompt. Closes the door for every seat that says so.
2. **The library deployer checks for a sibling collision** before deploying an existing tool: if
   another site in the same network already serves this component, refuse and file a
   `capability_gap` naming the sibling's URL (the brief's own instruction: link, don't build).
3. **`structure_floor_unmet` reads the vertical/brief** before counting tools as a floor — or
   drops `tool` from the universal set.
4. **`brief-fidelity-audit` dispatches, or at least BLOCKS, on a brand-new site** — a record-mode
   verdict that names a live violation should hold the dispatch loop for that site, not decorate
   the queue.

## 6. How to verify

Re-run the improvement loop over gamedesign.uk after the fix: `add_tool` count for the site must be
**0** while the brief still says no tools — and the disconfirming half: run it over
`gamesdesign.co.uk`, where tools ARE the seat, and `add_tool` must still fire. A zero on both is a
blinded suggester, not a fixed one.
