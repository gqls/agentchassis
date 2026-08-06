# HANDOFF — 071 fragment arm, written for a cold start

**State: code APPROVED and committed; INERT until the next chassis roll.** Nothing
is blocked. What remains is roll-time verification and three deliberately-deferred
pieces.

## What exists now

| thing | where | state |
|---|---|---|
| `dead_fragment_link` arm | `discovery_checks/check_phantom_internal_links_fragments.go` | committed `af2667453`, inert |
| wiring (second pass, `p.url` in the query, severity/routing) | `check_phantom_internal_links.go` | same commit |
| `SplitFragment` | `datahelpers/links.go` | same commit |
| `DocumentIDs` (extracted from `OrphanElementRefs`, which now runs on it) | `datahelpers/element_refs.go` | same commit |
| writer constraint (no invented `#` anchors) | `prepare_link_context_action.go` `buildLinkConstraintText` | same commit |
| claim-timeout exclusion | `sql_for_agents/322` + `220`'s declared list | **APPLIED AND RECORDED** 2026-08-06 10:20Z |
| register | LNK-031 new; **LNK-009 status corrected** | same commit |

Commits: `af2667453` (code+register+migration+docs) · `06d0e7695` (071
contribution, RUNBOOK, WRONG_CALLS ×2) · `72df8913e` (2 LANDMINES + sync) ·
`a13a60938` (council dispositions) · `eb35a0e13` (the looseness limit, in code) ·
`a865dc897` (owner log).

Council **APPROVED round 1**, `bbbb4132-4abe-4db1-a1ba-755377dab009`, 3 advisory
objections, none high. All dispositioned in NOTES §"2026-08-06 (evening)".
`af2667453` carries `Council-Submitted:`, so `098` credits it automatically —
**do not try to amend it** (forward-only).

## Do this after the next roll — in order

1. **Prove it shipped.** One exec, every replica, three strings:
   ```bash
   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
   kubectl -n ai-persona-system exec "$POD" -- sh -c "
     strings /app/agent-chassis | grep -c dead_fragment_link;         # >0 = shipped
     strings /app/agent-chassis | grep -c phantom_internal_link;      # POSITIVE control, 9 pre-roll
     strings /app/agent-chassis | grep -c zzz_no_such_string_control" # NEGATIVE control, 0
   ```
   **Pre-roll baseline measured on `v1.0.1257`, both replicas: `0 / 9 / 0`.** A
   roll is not evidence your fix shipped — the image may predate the commit.

2. **Induce a live finding.** The arm only speaks when a fragment misses, and
   today's estate has none, so nothing will appear on its own — **an empty
   `dead_fragment_link` queue after the roll is the EXPECTED result and proves
   nothing.** Plant one: add `<a href="#zzz-induced-control">x</a>` to a scratch
   page component on a low-traffic site, run `completeness-discovery-agent`
   against that site, assert exactly one `dead_fragment_link` item naming that
   page and href, then remove it and confirm the verifier closes the item.
   (Item type is **`dead_fragment_link`**, singular-style like its siblings —
   the check NAME `phantom_internal_links` returns 0 rows and reads like "never
   fired". That trap is now a LANDMINE.)

3. **Re-run the no-op case in the same window** (memory: check the no-op case,
   not only the damage case): a resolving fragment such as loancash's `#content`
   must file nothing.

4. **Re-run the fleet harness** post-roll and compare with today's 67/0:
   the dump query and the planted-control recipe are in
   `RUNBOOK_fragment_blindspot.md`.

## Open, and deliberately not mine to close

1. **The deploy gate still cannot judge fragments.** `validate_page_content`
   receives the writer's `page_html` **without chrome**, so gate-side fragment
   validation would false-positive on every chrome-satisfied anchor. Needs a
   chrome-aware id load at the gate. The `bug_historian` seat objected here at
   medium and it is a fair objection, not a solved problem.
2. **Nothing REPAIRS a dead fragment.** Unlinking a label-bearing anchor leaves
   the label as bare text (recorded landmine), so: detection first, volume
   second, repair third.
3. **No section component emits a stable `id`** — the capability half. Until it
   does, a fragment link can only be avoided, never made to work on purpose.
   Changes every page's rendered HTML fleet-wide ⇒ architecture round, not a bug
   patch.
4. **Three unaligned consumers now reason about link-target resolution** (the
   gate, this arm, `link_repair.go`). Both `bug_historian` and `architecture`
   named it; architecture still returned `point_fix` and noted `DocumentIDs` is
   positioned for (3). Whether that split is itself the thing to fix is an
   architecture-track question.

## Traps this lane paid for (all now in LANDMINES/WRONG_CALLS)

- **A check's NAME is not its `item_type`.** `phantom_internal_links` → 0 rows;
  `phantom_internal_link` → 119. The zero reads as "this check is inert", which
  is the premise that justifies building a second, separate check — i.e.
  `bugs_open/093`'s shape.
- **`RegisterVerifier` obliges two more edits** (220's declared list AND the live
  `scheduled_tasks.pre_query`), and the build names only one at a time.
- **A clean measurement that has never been shown to fail is not a measurement.**
  Bit three times today: the 0-findings harness (fixed by planting controls), the
  differential test (fixed by id-stripped variants — the first run agreed only
  about nil), and the register's row count.
- **`site_components` has no `component_type` column** — it is `slot_name`.
- One regex for `href="(/[^"]*#[^"]+)"` finds **5** of the estate's **66**
  fragment links; you need the bare-`#` shape too.

## Not this lane's, recorded so nobody re-finds it

`component_library.go:1136-1147` still holds the `primary_cta_url` →
`/contact.html` / `secondary_cta_url` → `/about.html` defaults map. `bugs_open/203`'s
08-05 fix (`880a405a6`) removed the `cta_url` **scalar** defaults only. That lane
is active; contribute there, do not fix it here.
