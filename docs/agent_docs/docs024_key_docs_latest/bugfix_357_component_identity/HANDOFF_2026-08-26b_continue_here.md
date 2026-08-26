# HANDOFF — `bugs_open/357`, component identity — 2026-08-26b (evening)

**Cold-start. Read this, then `HANDOFF_2026-08-26_continue_here.md` (superseded, but it holds
the precondition-4 failure evidence, the 578 procedure in §2 and the close-out bar in §3 —
both still current), then the bug files `bugs_open/357_…`, `bugs_open/408_…`,
`bugs_open/406_…`.**

---

## What changed since the morning handoff — three facts

1. **The credit blocker CLEARED ~09:00Z and held all day.** `llm_call_log` grouped by
   `success` with `output_tokens > 0` beside it: 200–300 successful calls/hr with output from
   09:00Z onward, re-read at ~21:00Z. (Keep the morning file's lesson: the table logs
   ATTEMPTS; never read row counts as recovery.)
2. **`v1.0.1345` rolled ~20:36Z** (pods `agent-chassis-5864bf97c5-*`, deployment spec and both
   pods agree on the tag). **It does NOT carry a 408 fix** — structurally, not by
   stamp-reading: no commit in this repo's history touches `multipage_actions.go` after
   `c4baa53e7` (328's round 2, pre-408), and the recursion is present at HEAD
   (`:1213`/`:1223`). Every buildable commit carries the defect, whichever one 1345 was cut
   from. **The cv1 canary and migration 578 remain forbidden.**
3. **A review pass traced the post-408-fix rebuild chain end to end and found the morning
   handoff's canary pass condition is VACUOUS once 408 is fixed** — and, more importantly,
   that **cv1 is structurally the wrong vehicle for precondition 4 altogether.** Both findings
   below, with the file:lines; full trail in `NOTES_component_identity.md` (2026-08-26 evening
   entry).

## Lane state (unchanged otherwise)

| | state |
|---|---|
| Phase 0 — provenance stamp | DONE, proven at volume |
| Phase 1 / F2 guard | DONE, proven with demand |
| Phase 2 — stop the mislabelling at birth | DONE, PROVEN IN PRODUCTION (2026-08-25 12:24Z), survived three rolls incl. 1345 |
| Phase 3 — repair the 22 | **NOT RUN — precondition 4 unmet, and see Finding 2: the planned test could never have met it** |
| The bug itself | **OPEN. `population = 22`, `adopted = 2`** (re-measured 2026-08-26; predicate one-liner below) |
| `bugs_open/408` (stack overflow) | **OPEN, UNFIXED at HEAD, blocks everything** |
| `bugs_open/406` (prune floor) | FILED, diagnosed, unfixed; shared-seam fix wants the council gate; does not block 357 |

---

## Finding 1 — the cv1 canary pass condition passes VACUOUSLY once 408 is fixed

Traced at HEAD (all as of 2026-08-26):

1. The writer skips the adopted page (0 sections compiled) → a **fixed** `extractFieldValue`
   returns `""` → `assemble_page` returns the skip shape (`multipage_actions.go:108-120`) — no
   crash any more;
2. `git_commit` skips via `checkUpstreamSkipped` (`git_deployer_actions.go:673`, keys on
   `assembled_page.skipped`);
3. `save_sections` RUNS — its early exit is keyed to the OWNED-page marker only, **deliberately**
   (`save_page_sections_action.go:71-90`; an ordinary content-failure skip must fall through,
   see the comment there) — and exits at `len(sections)==0` with `success:true, skipped:true`
   (`:344`/`:401`), BEFORE the DELETE-and-reinsert and BEFORE the Layer 2 carry-forward
   (`:555-600`);
4. the run reaches COMPLETED.

Result: `save_result` present, rows still 1, md5 unchanged, component still
`adopted-fragment` — **every clause of the morning handoff's pass condition holds, and the
conservation machinery never executed.** The row "survives" because nothing touched it.
**Do not read a green cv1 canary as precondition-4 evidence.** The REAL pass condition needs
`sections_saved > 0` — a save that actually processed incoming sections — which cv1 cannot
produce (Finding 2).

## Finding 2 — cv1 cannot test precondition 4; the 22 population pages are the right shape

What 578 depends on is the **armed identity-carry inside a real Layer 2 splice**: an incoming,
normally-generated section matching the adopted row's slot, and the stored bytes keeping their
own identity instead of the plan's (`save_page_sections_action.go:562-577` — *"without this …
the very next rebuild re-mints `hero` over an adopted row and the population renews itself"*).
That path runs only when the save carries >0 incoming sections for the slot.

- **cv1's adopted pages can never produce one**: their `pages.sections` plans name
  `adopted-fragment` itself (index planned=1, tool-example planned=1), so the writer always
  skips. [INFERRED at code level that the planner consumes `pages.sections`; measured at
  behaviour level — it planned exactly the plan's one section onto exactly the plan's
  component.]
- **All 22 population pages' plans name `hero`** (`plan_names_hero = t` on every row, measured
  2026-08-26). **16** are `rebuild_policy='generic'`, **6** `'owned'` (the gamesdesign
  `tool-*` six — 578's "six owned pages"). Post-578, a rebuild of a generic one generates a
  fresh hero per its plan, the save carries it, Layer 2 splices (match key is the SLOT, which
  578 preserves — it RAISEs if `slot_name` moved), and the armed carry must keep
  `adopted-fragment`. **That IS precondition 4.**

**Recommended vehicle — a one-row pilot of 578 itself** (a plan refinement to put to the
owner; 578 as written is all-22):

- Retype ONE row by 578's own procedure scoped to one page. Minimal candidate:
  **`mortgagecalculator.co.uk/tool-simple`** — planned=1, `generic`, single row.
- Rebuild that page. REAL pass condition: `save_result` present AND `sections_saved > 0` AND
  rows still 1 AND stored bytes preserved (md5) AND component STILL `adopted-fragment` (the
  carry held), not re-minted `hero`.
- **Then read the SERVED page, not just the DB row.** The chain commits the assembled
  (prose-hero) page to the sites repo BEFORE the save's preservation runs (LANDMINES'
  `assemble → git_commit → save` entry). This is a property of today's rebuild of ANY generic
  population page — not something 578 introduces — but the pilot exercises it, so the served
  artefact is part of the pass condition. The 6 owned pages are refused at assemble and not
  exposed.
- The pilot does not structurally require the 408 fix (its plan generates real content), but
  any no-content outcome on the way still crashes the pod until 408 is fixed — **fix 408
  first regardless.**

## What to do, in order

1. **Fix `bugs_open/408`** — candidate 1 in the bug file (ordered candidate list
   [`original`, `stripped`, `with-response`], no self-call), plus demote the per-recursion
   `Warn`. `extractFieldValue` has exactly **1** caller as of 2026-08-26
   (`multipage_actions.go:106`), which already handles `""` — nothing else to change. Test
   MUST run under `go test -timeout` (the failure mode is non-termination; a plain assert
   hangs, not fails). Platform Go → council gate (before/alongside commit), then
   `make build-agent-chassis` (bump `IMAGE_TAG`), roll, and verify at the pod with the
   known-sha probe + controls.
2. **Optional, now cheap after the 408 fix:** re-run the cv1 canary on `index` capturing the
   page-content-writer child orchestration live (by correlation, before the reaper) — it
   converts the by-elimination cause (`adopted-fragment` binding → empty render) into an
   observed one at the cost of a clean no-op rebuild instead of a pod crash. It does NOT bear
   on precondition 4 (Finding 1).
3. **Put the one-row pilot to the owner, then run it** (Finding 2). If the carry holds at the
   DB row AND the served page is acceptable, precondition 4 is met on the honest evidence.
4. **Run 578 in full** — §2 of the morning handoff, by hand, never the runner. Verify at the
   served pages per §3; **357 closes at `population = 0` verified at the artefacts.**
5. **Separately:** `bugs_open/406`'s durable fix (one-entry plan for tool-recreation routing)
   through the council gate. Owner's standing instruction (2026-08-25) to continue to phase 3
   on clean verification still stands.

## Traps for the next session

- The morning handoff's canary vehicle scripts lived in a previous session's scratchpad
  (`scratchpad/canary_rebuild.sh`, `canary_before.txt`) — **session-specific paths, likely
  gone.** The pinned baseline is reproduced in the morning handoff. Rebuild dispatches must be
  receipt-asserting: `scripts/kafka-publish-lib.sh` (OPP-009), never the silent-drop
  `kubectl run -i | kcat -P` shape in `110_page_rebuild/072_page_rebuild`.
- `llm_call_log` logs attempts — always `GROUP BY success` with an `output_tokens > 0` filter.
- A row count of **2** on an adopted page after any rebuild is the carry-forward landmine
  firing (the re-append arm) — STOP, do not run 578.
- Nothing enforces the *"nothing should ever plan a section onto adopted-fragment"* sentence —
  it is prose in a seeded description (LANDMINES entry of 2026-08-26). Plans naming it are the
  mechanism behind Finding 2.
- Spawned agent pods are ephemeral (logs gone in minutes) and `-l app=agent-chassis` is the
  wrong pod set for them — capture live or use the DB.

## Keys

| what | value |
|---|---|
| cv1.co.uk | site `8c3e9118-2455-4f0d-b01a-5dcde13dcf99` · adoption corr `468cb727-…` |
| adopted-fragment component | `9d4b922b-a548-4ca2-987c-ecacc7904b1f` ("Adopted Fragment") |
| its version row | `3301ef65-4d83-4ea5-aa7c-65cb38e83653` |
| pilot candidate | `mortgagecalculator.co.uk` page `tool-simple` (planned=1, generic, single row) |
| 578 file | `docs/agent_docs/sql_for_agents/578_retype_mislabelled_tool_rows_HOLD.sql` (+ `_ROLLBACK`) |
| baseline (cv1, unchanged) | `index` md5 `26f484f2744ab3e9cd19e50f600a52b8` 17,595 B · `tool-example` md5 `291b88d876e182a32a4a538c514878d2` 20,076 B |

```bash
# lane state, one line (population + adopted)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -F'|' -c "
SELECT (SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.function='adopted-fragment') AS adopted,
       (SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.name='hero' AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0) AS population;"

# is the fleet actually able to call an LLM? (attempts are NOT successes)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -F'|' -c "
SELECT success, count(*), count(*) FILTER (WHERE output_tokens > 0) AS with_output
FROM llm_call_log WHERE created_at > now() - interval '30 minutes' GROUP BY 1;"

# the 22, with the two facts that matter for the pilot
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -F'|' -c "
SELECT s.domain, p.name, p.rebuild_policy, jsonb_array_length(p.sections) AS planned
FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE cc.name='hero' AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0
ORDER BY 1,2;"
```

## In one sentence

**Credit is back and the machinery is proven, but the fresh roll carries no 408 fix, the
planned canary could only ever pass vacuously, and the honest route to closing 357 is: fix
408, pilot 578 on one of the 22 (whose plans — unlike cv1's — actually exercise the
conservation carry), read the served page as well as the row, then run the full repair.**
