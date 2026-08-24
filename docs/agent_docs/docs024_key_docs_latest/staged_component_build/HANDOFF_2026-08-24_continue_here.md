# HANDOFF — 2026-08-24, fresh chat starts here: **the lane's deliverable is DONE — live, approved, gate-passed, and proven to catch its class.** What remains is not this workstream.

**Supersedes `HANDOFF_2026-08-22_continue_here.md`** and, transitively, the 08-21 file.
**DOES NOT supersede `HANDOFF_2026-08-20_continue_here.md`** — that file holds the gate's terms,
baseline, interim reads and the **CLOSE-OUT** (§2.11). Read it for the evidence trail; do not
re-derive it.

---

## 1. THE ANSWER TO "CAN WE CLOSE THIS LANE?" — **YES**, and here is what that rests on

RFC_029 §9 D2 Phase 2 — *make the resolver never guess* — is **built, approved, live and
gate-passed**.

| # | what | state |
|---|---|---|
| **1** | **THE FLIP** — conflicting whole-tree search → refusal | 🟢 LIVE since `v1.0.1323` (08-22 08:37Z), re-verified on `1326`/`1328`/`1332`. Council APPROVED r3 (`26186633`). **Gate PASSED** |
| **2** | Read-side tolerance retirement | 🟢 LIVE, same rolls. APPROVED r1 (`e05ea6f9`) |
| **3** | `bugs_open/330` resolver half (migration 516) | 🟢 PROVEN both directions (absence + presence, owner-approved positive control) |
| **4** | Standing form of 537's guard | 🟢 LIVE as `WFA-022` (parallel session); `bugs_open/334` CLOSED |
| **5** | The gate's own finding — improvement-loop's bare `page_id` | 🟢 FIXED: migration **571**, council `5ae2147d` APPROVED, **applied + verified live 2026-08-24** |

**The gate close-out is in `HANDOFF_2026-08-20_continue_here.md` §2.11.** Headline: **0
`1-resolve-and-warn` across three rolls and 3,957 orchestrations ⇒ no regression**, and 5 real
refusals. ⚠ **The close-out states first what the zero CANNOT prove** — the conflict table was
already near-silent before the roll, so a clean 48 h was expected either way. What carries the
behaviour claim is the 13 tests that fail on a one-line revert plus the binary probe; what makes
the flip *worth it* is the real defect it caught on day one (below).

---

## 2. THE ONE FINDING THAT JUSTIFIES THE WHOLE WORKSTREAM

`improvement-loop` asked for a bare **`page_id`**. The tree held **80+ candidate paths that were
elements of ONE findings array** — `…discovery_result.findings[0..64].page_id`, plus nested
`members[N]`/`components[N]`, doubled by a step/alias pair. **Pre-flip that resolved to
`findings[0].page_id`: one finding's page silently attached to work about a different finding.**
Same shape on `component_id`. Both runs COMPLETED after the refusal — absence beat the wrong value,
as the owner ruled 2026-08-15.

**No conflict-WARN count could ever have proven this safe, because the old behaviour WAS the
substitution.** That is the argument for step 5, and it is now evidenced rather than reasoned.

**The owner asked whether this is an identifier problem — measured, and NO** (2026-08-23):
839 pages / 839 distinct ids; **0** ids on more than one site; `UNIQUE(site_id, name)` already
exists; the 34 ambiguous ids were **all one site's**. Names repeat across sites (`index` 28,
`about` 20) so a NAME needs a site — an ID never does. Every candidate was valid and correctly
scoped: **the REQUEST was underspecified, not the identity**, and no identifier design can answer
"give me page_id" when the tree holds 34. Migration 571 fixes it where the ambiguity is — in what
the step declares it takes.

---

## 3. WHAT LEAVES THIS LANE — named, with owners

1. **`site-work-orchestrator`'s two unfixed pairs** (`result`, 11 agents under one bare key;
   `commit_sha`, 537's class on a second agent). Both want an explicit declaration. ⚠ **NOT a copy
   of 537 or 571** — the consuming steps are **dynamically generated**
   (`fix_items_loop_iter_N_call_handler`), so neither migration's static step-config method reaches
   them. Solving that addressing problem is the first real task for whoever takes it.
2. **`bugs_open/353`** — OWNED BY THIS LANE, fixed and backfilled, **still open**: see §4.
3. **`bugs_open/330` candidate 2** — the 269-pair / 75-agent remainder. The gate has now handed it
   a live, traced instance (§2), which is exactly why it was deferred until after the gate.
4. **The tool-birth refusal rate** — `tool_birth_instance_scope_refused` on 6 of 9 births
   2026-08-22 vs 2 on 08-21. **Cleared of the roll** (prover byte-identical across it);
   generation-side. Unfiled — file it if the rate holds.

---

## 4. `bugs_open/353` — THE ONE THING WITH A PENDING OWNER DECISION

Fixed (opt-in field, unsafe default OFF; decision extracted and mutation-proved) and **backfilled:
74 cross-link items restored across 34 tools / 19 sites**, driven through the real emitter so the
item shape cannot fork. Council round 1 REVISE — **both objections were right and are answered in
the code** (a literal `false` had made the decision's `pageLive` arm dead; and the
`replace_existing` reroute's exclusion was unstated) — **round 2 in flight, corr `642ecc3c`**.

**⚠ THE BACKFILL IS DONE AT THE DATA LAYER AND UNFINISHED AT THE SERVING LAYER.** 61 of 74 items
are `complete` and the links ARE in `page_components.rendered_html` — **but 51 of 51 pages carrying
a completed rewrite deployed BEFORE their rewrite landed (a ~7-second gap, repeated 51 times: an
ordering property, not a race), so visitors do not see them.** Verified by `curl` with a control.

**THE DECISION: redeploy those 51 pages?** It is a real fleet action across ~19 live sites, the
same class as the backfill, and it is deliberately **not** being fired without the owner's word.
⚠ When it is run, remember a re-render carries **every** improvement since each page last
rendered — do not size it by this change. Whether the pages redeploy on their own is **unproven in
both directions** (none has in ~17 h); do not write "they will pick it up naturally" without
measuring it.

---

## 5. TRAPS EARNED (the ones that cost something)

1. **A control that differs from the treatment in TWO variables is not a control — and its failure
   mode is a PASS.** `replace_existing` reroutes `create_tool_component` into a different function
   that returns before the emitter; it invalidated a peer's "free" control, then masked 353 for 19
   days, then drew a council objection. Split on the arm before comparing.
2. **A complete work item is not a repaired artefact** — 61 completions, 0 pages serving the link.
   The check that separates "writer failed" from "page is stale" is stored-HTML vs `deployed_at`.
3. **A residual is discharged by what is ABSENT SINCE, not by a falling count** — retention makes
   rows vanish and that proves nothing.
4. **A bare bug number routes to the wrong owner** — 029 is ambiguous; resolve by slug.
5. **Pinning a guard's INPUTS does not pin the guard.** 353's branch sat where no unit test could
   reach it while its inputs' tests passed throughout. Extract the decision so it can be CALLED.
6. **A comment-only sketch is refused by the council trigger** — put the real change in `sketch`,
   observations in `rationale`.

---

## 6. SESSION-START CHECKLIST

1. `kubectl -n ai-persona-system get pods -l app=agent-chassis` — one tag, one replicaset? If a new
   build has rolled and you intend to trust resolver rows, re-verify with **both controls**
   (§2.11's method): flip literal present · a known literal present · a synthetic literal absent.
2. **The gate is CLOSED — do not re-run a window.** Its result is recorded in
   `HANDOFF_2026-08-20_continue_here.md` §2.11.
3. Read the `642ecc3c` verdict (353 round 2) and act on it.
4. Council/ownership: `5ae2147d` APPROVED + applied; `26186633`, `e05ea6f9` APPROVED + live;
   `642ecc3c` pending. Nothing else owed.
