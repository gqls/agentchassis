# 283 — CONTINUE HERE (2026-08-16). Round 2 is **APPROVED and LIVE**; the next real work is converting templates.

**Read `bugs_open/283_HANDOFF_…_element_ids_are_literal.md` first** — that is the case file
(diagnosis, evidence, the owner's ruling, and **§9, which is this session's record**). This
file is the working state only.

Supersedes `283_CONTINUE_HERE_2026-08-15.md`. Workstream docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/`.

---

## 1. State in one paragraph

The owner ruled that reuse should be a genuine property of the platform, so candidates **A+C**
(namespace properly, and block the bad state). Round 1 shipped a per-instance token and a
collision detector; the council returned **REVISE**. Round 2 (commit `32d6e980a`) acts on every
objection: the token rule is now **component function + occurrence** (`c-mortgages-repayment`,
`c-mortgages-repayment-2`), there is **one** derivation instead of two, the two previously
**unbound** page-embedded render sites are bound, the shared render layer reports a template
that needs a token and was given none, and `pattern-check.py` refuses a new unbound render call
site. **Nothing consumes any of it yet — 0 of 243 active templates reference `{{.InstanceID}}`
— so it is inert in production. The 22 calculator templates are unconverted, so the defect is
still live and 283 stays OPEN.**

## 2. What is committed

| commit | what |
|---|---|
| `03c1b0b90` | `{{.InstanceID}}` on three render paths + `DetectInstanceCollisions` + register CLC-014 |
| `9372a82c3` | gofmt + the CLC-014 row in `000_concept_index.md` |
| `1e19aa6ab` | the guard: records unconditionally, refuses only when `enforce_instance_scope` is set |
| `06a74ba7d` | case file: owner ruling, three collision classes, the closed `[UNVERIFIED]` |
| `7042b4eaf` | council REVISE recorded, the 5-vs-7 correction |
| **`32d6e980a`** | **round 2 — see §3** |

All carry `Council-Submitted: 07635a2f-3605-4e67-9a6d-7636b07f16ca`. **Round 2 came back
`approved` at 2026-08-16 10:25 UTC** — "approved with 6 advisory objection(s) — none
high-severity". The `098` report resolves the correlation at report time, so those commits are
credited automatically; **no amend, and do not retro-fit a `Council-Reviewed:` trailer.** New
commits on this lane may carry `Council-Reviewed: 07635a2f-3605-4e67-9a6d-7636b07f16ca`.

**It is LIVE.** Chassis `v1.0.1304`, pods started 2026-08-16T10:41 UTC, digest-verified (case
file §10). **And still inert — 0 of 243 templates reference the token.** Approval is not the
same as the defect being fixed; the 22 calculator templates are still literal-id.

**All six advisory objections are worked through in case file §10** — two needed a code check
(both fine), three needed a measurement (all done), and two produced real artefacts: `RFC_032`
and a `LANDMINES.md` entry. §10.4 is the one to read: the general `missingkey=zero` exposure is
**not** closed by this work, and that is recorded as a known-unresolved root cause rather than a
risk bullet.

Tests: green on a clean `git archive HEAD` tree. **The working tree's test build is broken by
another session's staged-but-uncommitted `agent_definition_nullable_columns_test.go`, which
redeclares `stripLineComments`** — nothing to do with this lane, and do not "fix" their file.
Test like this:

```bash
SC=<scratch>/tree; rm -rf $SC && mkdir -p $SC && git archive HEAD | tar -x -C $SC
cp platform/orchestration/actions/<your changed files> $SC/platform/orchestration/actions/
cd $SC && go build ./... && go test ./platform/orchestration/actions/
```

## 3. What round 2 changed

- **One rule, one derivation.** `InstanceToken(function, occurrence)`; `InstanceCounter`
  walked in position order by the two paths that see the whole page. `InstanceTokenFromSlot`
  is **deleted** — it wrote the same key under a weaker guarantee, which was the council's
  strongest objection and was correct.
- **The single-section paths supply occurrence 0** to that same rule (`BindSingleSectionInstanceToken`).
  A possibly-wrong *input*, not a second guarantee; where it is wrong the instances collide and
  the detector reports it.
- **`section_editor_actions.go` (2 sites) bound nothing at all** — the real gap, on nobody's
  list including the council's. Now bound.
- **The shared layer reports.** `RenderTemplateReportingMissing` logs at Error when a template
  references `{{.InstanceID}}` and none is bound. It deliberately does **not** substitute.
- **`check_unscoped_component_render`** in `scripts/pattern-check.py` — 4 findings at HEAD,
  0 now. Allow-list entries are measured claims about the **slot** (chrome, `<head>`, an
  offline lint), not about the file.

## 4. Do this next, in order

1. ~~Read the round-2 verdict~~ **DONE — approved; §10 of the case file has every objection and
   what it turned into.**
2. **Convert the components — and it is 91 rows, not 22 templates.** ⚠ **SCOPED 2026-08-17; the
   owner now owes a decision on SHAPE. See `architecture_review/RFC_034` and case file §11.**
   Measured live: **91 component rows on 94 live pages across 22 domains**, 1,346 literal ids,
   886 `getElementById` calls. Convert by `content_components.id`, **never by `function`** —
   4 functions carry forks and a function-keyed sweep skips 9 rows.
   ⚠ **Two findings that decide the sequencing, both proven in tests, not argued:**
   **(a) converting the ids ALONE makes the page read clean and leaves it broken** — 0 duplicate
   ids, but both scripts still declare `runCalc` globally, so every button runs the last
   instance's logic (`TestIDOnlyConversion_readsCleanOnIDsAndIsStillBroken`). So ids and scripts
   convert in ONE step per component; a phased "ids now, scripts later" is worse than nothing.
   **(b) the IIFE route is FORCED** — `{{.InstanceID}}` renders as `c-mortgages-repayment`, which
   is not a valid JS identifier, so `function runCalc_{{.InstanceID}}()` is a syntax error
   (`TestInstanceToken_isNotAValidJSIdentifier`). That forces the 22 inline `on*=` handlers to be
   rewired, which is the part that is not safely mechanical.
   ⚠ Also 58 rows carry `<label for=>` and 33 reference an id from CSS — **neither throws** when
   the id underneath is renamed.
   ~~This is architecture-scope, and the council said so~~ — still true, and `RFC_034` is now that
   round. When it is decided: namespace ids with `{{.InstanceID}}-`, scope
   lookups to the instance root instead of `document.getElementById`, wrap each script in an
   IIFE (**16 declare at top level**), replace `window.onload` (**8 assign it**). DB writes
   need the owner — this session's permission classifier refuses them.
3. **Update `oracle.py`'s selectors in lockstep** and re-run the full sweep (baseline
   **170/0/6**). One prefix per tool is all it needs — that property is why the token rule is
   function-based.
4. **Rebaseline `b2_verify`'s verbatim-render check.** Converting ends the byte-identical
   property that lane verifies against. A deliberate decision, not something to meet mid-batch.
5. **Then, and only then**, consider arming `enforce_instance_scope` anywhere. ⚠ The
   `render_guardian` seat's words: arming it **before** the 13 known-colliding pages are fixed
   would itself be *"a high-severity fail-loud violation"*. Convert → re-measure → arm.
6. ~~Build the RFC_022 expiry trigger~~ **DONE — `instance-token-adoption-check`, a daily CronJob
   at 07:40 UTC, deployed and proven (first run 2026-08-16 15:29 UTC: adopters 0, control 5,
   active 243).** You do not need to remember the expiry; it will tell you.
   **What you DO owe it:** it is not in the makefile (the file had another session's uncommitted
   changes when it was built, and a pathspec commit would have taken them as a passenger), so add
   `deploy-instance-token-adoption-check` alongside its siblings when the makefile is quiet.
   **And retire the job once it trips** — a fired tripwire left failing daily gets muted.

## 5. Traps carried forward

- **13 active pages already emit duplicate ids today** (`generic-text-block` ×2–3 resolving its
  one id through the shared `ComponentID`). That is why the guard defaults OFF, and why the
  detector goes non-empty for reasons other than your conversion.
- **The concatenation the guard checks EXCLUDES chrome**, so a collision between a section and
  the header/footer is invisible to it. A clean result is not a clean page.
- **`{{.ComponentID}}` is NOT a per-instance value** on two of the three paths. Full trap in
  `LANDMINES.md`; it is the reason this seam exists, and it is deliberately **not** changed —
  re-pointing it would move the served ids of five live components, which is architecture-scope
  and is the follow-up the architecture seat asked for.
- **Verify a deploy at the artefact.** Case file §9.5 has the chain that works when the
  `build provenance` line has scrolled: pod `imageID` digest → local image `RepoDigests` →
  its `org.opencontainers.image.revision` label → `git merge-base --is-ancestor`. The digest
  match is the load-bearing step; without it a local tag is just a local tag.
- **Round 2's code is committed but not built or rolled.** Inert in production until the next
  chassis release.
