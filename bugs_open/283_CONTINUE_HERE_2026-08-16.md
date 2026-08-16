# 283 — CONTINUE HERE (2026-08-16). Round 2 is submitted; the next real work is converting templates.

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

All carry `Council-Submitted: 07635a2f-3605-4e67-9a6d-7636b07f16ca`. **Do NOT write a
`Council-Reviewed:` trailer until you have read an approved verdict** — round 2's verdict was
not back when this was written.

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

1. **Read the round-2 verdict** — it is keyed on the same correlation:
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='07635a2f-3605-4e67-9a6d-7636b07f16ca' AND kind='council_report'
   ORDER BY created_at;
   ```
   Two reports will be there; the newest is round 2. The code is already on the shared branch,
   so a REVISE is acted on forward, not held.
2. **Convert the 22 calculator templates.** Namespace ids with `{{.InstanceID}}-`, scope
   lookups to the instance root instead of `document.getElementById`, wrap each script in an
   IIFE (**16 declare at top level**), replace `window.onload` (**8 assign it**). DB writes
   need the owner — this session's permission classifier refuses them.
3. **Update `oracle.py`'s selectors in lockstep** and re-run the full sweep (baseline
   **170/0/6**). One prefix per tool is all it needs — that property is why the token rule is
   function-based.
4. **Rebaseline `b2_verify`'s verbatim-render check.** Converting ends the byte-identical
   property that lane verifies against. A deliberate decision, not something to meet mid-batch.
5. **Then, and only then**, consider arming `enforce_instance_scope` anywhere.

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
