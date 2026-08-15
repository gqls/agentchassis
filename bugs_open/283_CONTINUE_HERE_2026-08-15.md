# 283 — CONTINUE HERE (2026-08-15). Council returned REVISE; read §3 before touching the code.

**Read `bugs_open/283_HANDOFF_…_element_ids_are_literal.md` first** — that is the case file
(diagnosis, evidence, the four options, the owner's ruling). This file is the working state:
what shipped, what the council said, and what to do next in order.

---

## 1. State in one paragraph

The owner ruled that **reuse should be a genuine property of the platform** — *"if we chose
to list all the calculators on one page we'd hope it would work"* — so candidates **A + C**
(namespace properly, and block the bad state) rather than the staged option B. Halves one and
two are **built, tested and committed**: a per-instance template value `{{.InstanceID}}` wired
on all three render paths, a three-class collision detector, and an opt-in guard that records
everywhere and refuses only when armed. **No template consumes any of it yet**, so it is inert
in production and nothing is at risk. The council gate returned **REVISE** with two objections
that are real and one that is a submission artefact. **The 22 calculator templates are
unconverted, so the defect is still live and 283 stays OPEN.**

## 2. What is committed

| commit | what |
|---|---|
| `03c1b0b90` | `{{.InstanceID}}` on all three render paths + `DetectInstanceCollisions` + register entry CLC-014 |
| `9372a82c3` | gofmt + the CLC-014 row in `000_concept_index.md` |
| `1e19aa6ab` | the guard: records unconditionally, refuses only when `enforce_instance_scope` is set |
| `06a74ba7d` | case file updated — owner ruling, three collision classes, the closed `[UNVERIFIED]` |
| *(this one)* | council REVISE recorded, the 5-vs-7 correction, this file |

All carry `Council-Submitted: 07635a2f-3605-4e67-9a6d-7636b07f16ca`. **Do NOT write a
`Council-Reviewed:` trailer** — the verdict is REVISE, not approved.

Tests: `go test ./platform/orchestration/actions/` green; 8 tests in
`component_instance_scope_test.go`, proven able to fail by mutating the detector to return an
empty report (3 of 4 detector tests then fail, each for a different reason).

> ⚠ **DEPLOY STATUS OF THESE COMMITS IS UNVERIFIED, and do not guess it.** The pod runs
> `aqls/agent-chassis:v1.0.1303`. The `build provenance` startup line has scrolled out of
> `--tail=6000`, and the binary probe does **not** work here: an older *real* commit sha is
> equally absent, so the binary carries only its own build stamp, not arbitrary shas. My first
> probe used all-zeros as the absent-control and it read **PRESENT** — binaries are full of
> zero bytes — which is a reminder that the control is the part that has to be chosen well.
> It does not matter yet (the code is inert), but it matters before anyone arms the guard.

## 3. ⚠ The council verdict: REVISE — objection by objection

`decided_by: gating objection from editquality`. **4 approve** (render_guardian,
debug_historian, constitution, mission), **5 object** (editquality, bug_historian,
reuse_agent, guardian, architecture), 7 abstained. Full report:
`SELECT body FROM diagnosis_artifacts WHERE correlation_id='07635a2f-…' AND kind='council_report';`

### 3.1 HIGH — "`InstanceTokenFromSlot` is undefined; this does not compile" (editquality, guardian)

**A SUBMISSION ARTEFACT, NOT A CODE DEFECT — but my fault, and do not dismiss it.** The
function exists in `component_instance_scope.go` and `go build ./...` is clean. I added it
*after* writing edit 1's sketch and never updated that sketch, so the plan the reviewers read
genuinely called an undefined symbol. The runbook says it outright: *reviewers judge the
sketch; it is the only view of your code they get.* **Fix on resubmit by including the helper
in the sketch.** Two seats independently flagged it and the guardian said it would raise to
veto if real — that is the gate working.

### 3.2 HIGH — "the slot fallback writes the SAME key with a WEAKER guarantee" (reuse_agent; architecture agrees)

**REAL, AND IT IS THE MOST IMPORTANT ONE.** `v3_site_actions.go` sets `InstanceID` from
`InstanceTokenFromSlot`, which cannot guarantee uniqueness on the 13 duplicate-slot pages,
while the other two paths set the same key from a provably-unique position. The seat's words:
this *"reproduces, under a new name, the exact ComponentID landmine this plan's own rationale
diagnoses: one field name, two different uniqueness guarantees, indistinguishable on any page
where only one instance is present."*

That is exactly right, and it is the shape I wrote a `LANDMINES.md` entry about in the same
session. Options: give that action a real index; or name the weak one differently so no
consumer can read `InstanceID` as "unique". **See §4 — `data_uuid` probably dissolves this.**

### 3.3 MEDIUM — "three call sites patched, the mechanism left generic" (bug_historian)

**REAL.** `RenderTemplate` still defaults a missing key to empty string. Any *other* action
that renders these templates and does not set `InstanceID` — the seat names
`create_tool_component_action.go`, `deploy_tool_action.go`, `rebuild_blog_listing_action.go`,
`section_editor_actions.go`, `create_report_page_action.go` — will render **the same empty
string for every instance** the moment a template references `{{.InstanceID}}`, silently
recreating the bug. This is 016b §9's *"one call site of a shared judgement gets the rigorous
fix; the sibling stays heuristic."* **Either enumerate and patch every render call site, or
put the value in at the shared `RenderTemplate` layer.** The latter is probably right.

### 3.4 MEDIUM — "you didn't check `page_components.data_uuid`" (prior_art_librarian)

**REAL, AND THE SEAT WAS RIGHT — see §4.**

### 3.5 MEDIUM — "a per-instance value already exists on that exact line" (reuse_agent, constitution)

`assemble_from_library.go:277` already computes `component_<function>_<idx>`. I added a second
parallel token beside it rather than exposing the existing one. Fair. Resolve by having the
one canonical primitive, not two.

### 3.6 LOW — "the detector is wired into nothing" (guardian, bug_historian)

Partly stale: the guard in `1e19aa6ab` wires it into `rerender_page_sections`. It is **not**
in CI or a build gate. The guardian also asks for an **end-to-end test that two instances on a
real page actually get two different `InstanceID`s** — only the standalone detector is tested
against synthetic HTML. **Worth doing; it is the test that would catch §3.3.**

### 3.7 The architecture seat: `ARCHITECTURE_SIGNAL: insufficient`

It did not block, and its reasoning is the sharpest thing in the report: the render pipeline
already has **three independent context-builders that disagree about what an "instance" is**,
`ComponentID` already means two different things across two of them, and this patch adds a
**third** concept with a different guarantee per path. *"The estate needs one canonical
per-instance identity, not three ad hoc derivations sharing a name."* Its recommendation:
**proceed, but file a follow-up to unify the three context-builders before any template or
build guard adopts `InstanceID` as load-bearing.** Treat that as the real design instruction.

## 4. `data_uuid` — the finding that probably reshapes the fix

The `prior_art_librarian` asked whether `page_components.data_uuid` already provides a
per-instance handle. **Measured after the verdict: 1,580 rows, 1,580 populated, 1,580
distinct.** It is a fully-populated, unique-per-row identifier that is *stable* (stored, not
recomputed) and needs no measurement to justify uniqueness — unlike `position`, whose
uniqueness rests on a fleet-wide count that could change.

**This likely dissolves §3.2 and §3.5 together:** every render path can load `data_uuid`, so
all three could derive one identity from one source and the weak slot fallback disappears.

Before adopting it, settle two things I did **not** check:
1. Is `data_uuid` loaded (or cheaply loadable) on all three paths? `storedSection`
   (`rerender_page_sections_action.go:103`) currently loads
   `componentID, slotName, contentData, renderedHTML, position` — **not** `data_uuid`, so this
   is a query change.
2. A UUID may begin with a digit, which is a valid HTML id but **not** a valid CSS identifier,
   so `querySelector("#3f…")` throws. Keep the letter prefix.

## 5. ⚠ A defect in the committed token, independent of the council

`InstanceToken` derives from `position`. **Measured: the LMC tool slot sits at position 0 on 7
pages and position 1 on the other 16.** So the same component renders **different ids on
different pages**, coupling every selector to section order. `oracle.py` addresses all **170**
of its checks by literal CSS id (`#loanAmount`, `#displayMonthly`) and would need per-page
knowledge rather than one prefix per tool, breaking silently on any reorder.

`data_uuid` does **not** fix this — it makes it worse (ids differ per page *and* are opaque).
The property the oracle wants is a token that is **the same on every page for a
single-instance component**. That argues for **component function + occurrence index within
the page**: `c-mortgages-repayment`, then `c-mortgages-repayment-2` for a second instance.
`rerender_page_sections` already loops the full section list, so the occurrence index is
computable there.

**These two pull in opposite directions and the choice is not yet made.** `data_uuid` gives
provable uniqueness and kills the multi-path inconsistency; function+occurrence gives selector
stability and keeps the oracle change to one prefix per tool. **Decide this before converting
any template** — it is the single most consequential open question in this bug.

## 6. Corrections recorded this session

1. **"7 live components depend on `{{.ComponentID}}`" is FALSE — it is 5.** The 7 counted
   components templating *any* id, which also catches `product-grid` and `category-listing`
   using their own domain fields. Corrected in place in the CLC-014 register entry; full
   incident in `WRONG_CALLS.md`. Caught by the `prior_art_librarian` seat objecting that the
   claim was unevidenced — it did not know the number was wrong, only that it was unchecked,
   and that was enough. The decision it supported (leave `ComponentID` alone) still stands.
2. **The case file's §2 understated the defect**: three collision classes, not one. Already
   folded into the case file at §8.2.

## 7. Do this next, in order

1. **Settle §5** (`data_uuid` vs function+occurrence) — everything downstream depends on it.
2. **Act on §3.3**: move the value into the shared `RenderTemplate` layer, or enumerate and
   patch every call site. This is the objection most likely to recreate the bug silently.
3. **Add the end-to-end test** the guardian asked for (§3.6): two instances on one real page
   render two different tokens. It is also the regression test for step 2.
4. **Resubmit to the council** with `RESUBMIT_CORR=07635a2f-3605-4e67-9a6d-7636b07f16ca` so the
   trail accumulates, with the helper included in the sketch (§3.1).
5. **Then convert the 22 templates** — namespace ids, scope lookups to the instance root, wrap
   scripts in an IIFE (16 declare at top level), replace `window.onload` (8 assign it).
   **Update `oracle.py`'s selectors in lockstep** and re-run the full sweep (baseline
   170/0/6). DB writes need the owner — this session's permission classifier refuses them.
6. **Only then** consider arming `enforce_instance_scope` anywhere, and verify the deploy at
   the artefact first (§2's warning).

## 8. Traps carried forward

- **13 active pages already emit duplicate ids today** (`generic-text-block` ×2–3 resolving
  its one id through the shared `ComponentID`). That is why the guard's default is OFF, and
  why converting templates is not the only way this detector goes non-empty.
- **Converting the templates ends the byte-identical property** the LMC lane verifies against.
  `b2_verify`'s verbatim-render check will need rebaselining — a deliberate decision, not
  something to discover mid-batch.
- **The concatenation the guard checks excludes chrome**, so a collision between a section and
  the header/footer is invisible to it. A clean result is not a clean page.
- **`{{.ComponentID}}` is NOT a per-instance value** on two of the three paths. Full trap in
  `LANDMINES.md`; it is the reason this whole seam exists.
