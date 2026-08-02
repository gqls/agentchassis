# HANDOFF — staged_component_build, 2026-08-02 (fresh chat starts here)

**This supersedes `HANDOFF_2026-07-31c_continue_here.md`.** That file's Part A (author and
prove `teaser-reveal-panel`'s own criteria fence) is **DONE**. Only Part B remains, and it
is this file's whole subject. Read this file instead; go back to `c` only for Part A's own
detail, which also lives in `SUMMARY_2026-08-02_the_fence_is_proven_and_persisted.md` and
`NOTES_staged_component_build.md`.

---

## 1. Read these, in this order

| doc | why |
|---|---|
| `README_where_we_are.md`, last two entries (`2 August 2026`) | plain prose, fastest way in |
| `SUMMARY_2026-08-02_the_fence_is_proven_and_persisted.md` | current-state milestone read-out for the just-finished piece |
| `NOTES_staged_component_build.md`, last two entries (search `## 2026-08-02`) | the fence build AND the chassis-roll recheck — read before writing any code |
| `PLAN_2026-07-30_staged_component_build.md`, section **P2** (and P1's now-closed correction just above it) | P2 is this file's whole subject |
| `docs026_concept_register/register/documentation-system.md`, DOC-068 entry | the verify-later line this handoff is trying to close the second half of |
| `docs026_concept_register/register/tool-lifecycle.md`, TL-036 entry | what the mutation-proof harness can and cannot do generically — relevant if the S6 dispatch design needs its own mutation proof later |

Register: **DOC-068** (component travelling docs, both halves live), **CLC-012**
(`teaser-reveal-panel`), **TL-036** (offline fence trial + mutation prover, PLUS the
2026-08-02 correction that only the trial half is actually generic).

---

## 2. State — what is live and proven, what is not

**DONE and proven at the artefact, as of 2026-08-02:**

- **`teaser-reveal-panel` has a real, persisted, mutation-proven criteria fence.** 12
  checks, `try_fence.go` 15/15 against the live URL AND against the fence read back out of
  the DB (byte-identical). Mutation proof: 12/12 mutants caught, all 12 checks watched red,
  via a new bespoke harness (`prove_fence_can_fail_teaser_reveal_panel.go`) — the lane's
  existing `prove_fence_can_fail.go` turned out to be hardcoded to a *different* component
  (`tool-review-council-simulator`) despite reading as generic; do not assume it works for
  whatever you test next either — run it and watch it fail on stale mutants first, the same
  check this handoff itself required.
- **`doc_plans`/`doc_notes` hold real rows**: `subject_type='component', subject_key='teaser-reveal-panel'`,
  `is_current=true`, body 19,953 bytes, plus two backfilled/new NOTES rows. Not a probe,
  not deleted.
- **Old file-based `PLAN_teaser-reveal-panel.md`/`NOTES_teaser-reveal-panel.md`** are marked
  superseded with a pointer to the DB rows, per their own stated instruction. Left on disk
  as history, no longer updated.
- **Chassis rolled to `v1.0.1229`** (from `v1.0.1219`) for reasons entirely unrelated to
  this lane (`bugs_closed/165`, `168`, `179`, `097`, portfolio work). Re-checked, not
  assumed: the `doc_plans` row is unchanged, and a fresh pod-grep still shows the component
  vocabulary marker present. **Re-check the running image yourself before citing any figure
  from before this roll as current** — standard practice, restated because it is easy to
  skip when nothing in the diff looks relevant.

**NOT yet done — this is the whole of what remains in P2:**

`request_browser_run` (`platform/orchestration/actions/tool_acceptance_actions.go:87-152`)
resolves the page to test by reading a `function` string and looking it up **directly
against `pages.name`** (`SELECT url FROM pages WHERE site_id=$1 AND name IN ($2, 'tool-'||$2)`).
That encodes **one function ⇒ exactly one page**. `teaser-reveal-panel` is placed via
`page_components` on **5 distinct pages across 2 distinct sites** (re-verify — this moves;
query is in HANDOFF `c` §3 and RUNBOOK). Nothing in `request_browser_run` can express
"which of the 5" today. This has not changed since `c` was written, and nobody else has
touched this action (`git log -- platform/orchestration/actions/tool_acceptance_actions.go`
shows no new commits as of the chassis-roll recheck above).

---

## 3. THE NEXT ACTION — decide, then wire, the S6 dispatch for a multi-page component

**Two ways to close the gap, carried forward from `c`, still not decided:**

- **(a) Extend `request_browser_run`** with an opt-in `page_id_field` (+ implicit `site_id`)
  that, when present, bypasses the `pages.name` lookup and resolves instead via
  `page_components.component_id = <uuid> AND page_id = <given>`. Less code, one action's
  config surface; adds a branch to a function that has exactly one path today.
- **(b) A sibling action** (e.g. `request_component_browser_run`) that never touches the
  tool path. Zero blast radius on a working mechanism; near-duplicate plumbing (envelope,
  headers, profile handling — roughly lines 153-230 of the same file, which are already
  identity-agnostic and could be shared via a helper either way).

**Make this decision explicitly and write it into the PLAN (P2 section) before coding it**
— this handoff deliberately does not pick one, the same way `c` didn't, because it is a
real tradeoff (blast radius vs. duplication) and not a default.

**Council-gate note:** either option edits `platform/orchestration/actions/`, in-scope for
the advisory council gate (CLAUDE.md, "Platform seams"). Per the 2026-07-29 owner ruling, an
addition needs architecture review only if it changes what the shared mechanism
*guarantees* — an opt-in field or a new sibling action that nothing calls until a component
agent names it does not change `request_browser_run`'s existing guarantee for tools, so the
normal council gate (not an RFC) looks right. **Confirm against the actual diff once one
exists** — this paragraph is a read of the rule, not a ruling on code that doesn't exist
yet. Submit before or alongside the commit (`Council-Submitted: <corr>` if the verdict
hasn't landed yet); register the addition in the concept register in the SAME commit
(TL-036/DOC-068 area), per the "platform seams" section's condition (2) — not later.

**Dispatch shape to copy** — `tool-acceptance-agent`'s live workflow (re-verify against the
live `agent_definitions` row, it may have changed):
```
ensure_site_record → load_docs (load_doc_context, subject_type: tool)
                   → request_run (request_browser_run)
                   → judge (judge_acceptance_results)
                   → complete
```
For a first proof, you likely do **not** need a real `agent_definitions` row at all — the
Go-gate probe in this same lane (`scripts/PROBE_doc_subject_go_gate.sh`) and the fence build
both proved that an inline `config.workflow` in the dispatched message (`selectWorkflow`
Priority 1, `processor.go:922-928`) is enough, with nothing to seed and nothing to clean up
afterwards. Same trick applies here: set `load_docs`'s `subject_type: component`,
`subject_key_field` pointing at `teaser-reveal-panel`, and whichever `request_run` shape you
build pointing at the one `(site_id, page_id)` this handoff's own fence was proven against:
`page_id=ebc2c413-61e2-465e-b22b-9aab0167abc9`, `site_id=4851f6fc-71cf-4160-a270-e03d6d3e0732`
(`leopardessconsulting.co.uk/services.html` — re-verify this row is still current before
using it, per the standing landmine about placements changing).

**Gate:** a deliberately broken `teaser-reveal-panel` (or a deliberately wrong page_id)
makes S6 go red — run the negative control **in the same dispatch**, the same habit the
Go-gate probe and this handoff's own fence build both used, and for the same reason: a green
run and a run that skipped silently look identical unless something is watched to fail.
Read the result honestly per RUNBOOK §10 (a FAILED run still reports `status=COMPLETED`;
the real error is in `collected_data->'__step_error'`, not `error`).

**Once dispatched successfully, close DOC-068's remaining verify-later half** ("an S6 run
citing its fence") — it is currently the one explicitly open thing in that entry.

---

## 4. Open items behind this one, unchanged from `c`

3. **The authoring backlog: components/tools with no PLAN at all.** Honest, not a check
   failure. Do NOT wire any naming check into a component's birth path as a side effect of
   this work — out of scope.
4. **`features_open/028`** (rename orphaning) — filed, unowned.
5. **`has_visible_area` checks owed to every existing fence**, now that `bugs_closed/157` is
   long closed. Not this lane's to duplicate — check who owns it before touching it
   (`scripts/who-owns.py`).

---

## 5. Do NOT do these (unchanged from `c`, still correct)

- **Do not rebuild the eight-stage ladder.** Owner cut it (D8).
- **Do not fire an acceptance run at the arena tool.** `gauntlet_dead_cta` lane's decision.
- **Do not run `./scripts/migration/run-migrations.sh --apply`** blind.
- **Do not roll the chassis to ship anything.** Builds come from committed HEAD; your
  commit ships on anyone's next roll regardless.
- **Do not adopt `features_open/015`.** Accepted decomposition stands.
- **Do not assume `prove_fence_can_fail.go` (the original) works for whatever you test
  next.** It is hardcoded to one specific component's source strings. Run it, watch it
  report stale mutants, build a sibling — the same discipline this handoff's own Part A
  had to apply to itself.

---

## 6. The one thing to carry into this piece specifically

Twice now in this lane, a piece of machinery that read as generic — `request_browser_run`'s
page lookup, and separately `prove_fence_can_fail.go`'s mutation list — turned out to encode
an assumption specific to the first thing it was built for, and the only way either was
caught was by actually running it against the new case rather than trusting its name, its
doc comment, or the RUNBOOK's own phrasing. **Whatever you build for the S6 dispatch, run it
against a case it wasn't written for before calling it done** — a second component with a
different placement shape, if one exists, or at minimum the negative control this section
already asks for.
