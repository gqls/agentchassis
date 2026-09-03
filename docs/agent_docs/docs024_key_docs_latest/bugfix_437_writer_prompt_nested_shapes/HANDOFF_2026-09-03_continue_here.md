# HANDOFF 2026-09-03 — bugs_open/437, writer prompt nested item shapes

**COLD-START: read this file, then `bugs_open/437` §POST-ROLL, then
`NOTES_writer_prompt_nested_shapes.md` from the bottom up.**

## The one-line state

**✅ CANDIDATE 1 IS FIXED, LIVE AND PROVEN AT THE ARTEFACT (2026-09-03 13:25Z).** The writer
is shown the nested shape and produces it. **The bug stays OPEN** — candidates 2 and 3 are
untouched, and they are what leave the already-stuck pages stuck.

**The proof, first mechanism-flow write on a post-roll agent pod (`llm_call_log` 13:24:58Z):**
prompt carries `"branches": [{` and the shape note; the old flat `"branches": "..."` is gone;
the model's reply carries **four `branches` values, all arrays** — one populated with real
`{label, body}` objects, three `[]` where a step has no decision point. Those empty arrays
are the omission advice being obeyed, which is the accepted over-production risk behaving
correctly on its first exercise.

**Failures:** 7 today before the fresh agent pods, **0 after** — ⚠ on a demand control of
exactly **1** exercise. A real end-to-end pass, not yet a rate. Re-census after traffic.

⚠ **A later chassis build was announced as this was written.** The fix is committed
(`a0044e73b`) so it rides every later build; nothing is owed. But re-cite the tag and date
when you check it, and remember the roll below.

## What the bug is (settled, do not re-derive)

The page-content-writer prompt does not contain the component schema; it contains a JSON
exemplar **generated** from it. `extractArrayItemFields`
(`platform/orchestration/actions/plan_sections_action.go:3277`) projects an element schema
to a flat `[]string` of NAMES, so a property that is itself a collection flattens to a bare
name and the exemplar rendered mechanism-flow's `steps[].branches` — declared an array of
objects `{body,label}` — as `"branches": "..."`, i.e. **a string**. The writer copied the
demonstration; the render type gate (`bugs_closed/260` / STY-057) refused it, correctly;
**119 builds failed in 14 days across six sites**, deterministically, with no lucky passes.

Proof, in one row: `llm_call_log` `34f25815-42d3-4057-b42a-b8b42189ae7e` (2026-09-02
19:07:30Z, advertise.co.uk) — prompt line 234 is the string exemplar, `response_text` obeys
it. **The writer was obedient throughout and the gate was right every time.**

Settled separately (asked by the `components` lane, answered in `bugs_open/437` and
PBP-052): the legacy JSON-Schema `items` dialect is a **different** defect in the same place
(`bugs_open/240`, already fixed) — the proof is that the failing prompt listed the real
per-item names, not the JSON-Schema keywords.

## What is deployed and what it is doing

| thing | state |
|---|---|
| Go: `datahelpers.StructuredItemShape` + `llmFieldSpec.ValueShape/ItemNotes` | committed `a0044e73b`, **live on v1.0.1358**, **emitting nothing** |
| Migration 724 (the prompt template) | **applied + verified in the live row**, intact, harmless |
| livespec `workflow.page-content-writer.prompt_item_shape` | live, declares both sites, Forbids the pre-437 spelling |
| Council | corr `6de0f6f2-4f37-492a-9cbd-1ae886311a9b`, round 1 REVISE (my sketch abbreviations), **round 2 dispatched, verdict not yet read** |

## THE OPEN PROBLEM — start here

In orchestration `29a88d1e-abb1-48bb-abff-c83ca7a6f0e5` (ran the `plan_sections` action at
12:20:01Z, i.e. on the new pods), the **mechanism-flow** section's `steps` spec in
`collected_data->'section_plan'` is:

```json
{"name":"steps","type":"array","required":true,"on_missing":"skip_field",
 "item_fields":["body","branches","marker","note","title"]}
```

`item_fields` correct; `value_shape` and `item_notes` **absent**. Both are `omitempty`, so
absence means `StructuredItemShape` returned **empty at runtime**.

### Ruled out, each with its check — DO NOT REPEAT THESE

1. **Not a stale / un-rolled binary.** `docker image inspect … .Config.Labels
   "org.opencontainers.image.revision"` → `d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85`;
   `git merge-base --is-ancestor a0044e73b d0252fd4d` PASSES, with the current HEAD as a
   negative control that correctly FAILS.
2. **Not missing from the binary** — ⚠ **BUT SEE THE ADDENDUM: I PROBED THE WRONG POD.**
   The original probe ran on an `agent-chassis` *deployment* pod
   (`agent-chassis-554857f96f-kx69c`): `grep -c "never a sentence of prose" /proc/1/exe` → 1,
   long-lived control present, nonsense control absent, struct tag `value_shape` present 3×.
   That pod is **not** where `plan_sections` executes — per-agent pods
   (`agent-page-build-handler-*`) do. Re-probed correctly at 12:55Z: the fix literal IS
   present there too, with control. So the conclusion survives for the CURRENT pods; what it
   never covered is the pods that ran the 12:07–12:20 executions, which no longer exist.
3. **Not missing from the built tree.** `git show d0252fd4d:…/plan_sections_action.go |
   grep -c StructuredItemShape` → **2**; the helper file exists at that revision; there is
   exactly **one** `llmFieldSpecs = append` site (line 2723) and it carries both new fields.
4. **Not service skew.** Every deployment running the chassis image (`agent-chassis`,
   `business-intel`, `vet-intel`) is on `v1.0.1358` at the same revision.
5. **Not a reverted migration.** Live row fragment counts: nested exemplar 1, item_notes
   tail 1, pre-437 spelling **0**, flat arm 1.
6. **Not a helper bug.** Dumped the live `input_schema` and ran it through the real
   `SchemaContentFields` → `StructuredItemShape` in a throwaway test: correct skeleton and
   correct note. (The throwaway test file was deleted; recreate it from
   `RUNBOOK` §1 if needed.)
7. **Not a changed schema.** One active `mechanism-flow` row; `branches` still `type: array`
   with `items.properties`; `steps.source = llm`.
8. **Not timing.** All five post-roll orchestrations started after the pods came up.

### Ranked hypotheses

1. **`comp.InputSchema` at runtime is not the shape I probed with — START HERE.** The one
   concrete anomaly: the component payload carried in the plan serialises
   `component.input_schema` as a **JSON STRING**, not an object (`jsonb_typeof` → `string`,
   `? 'fields'` → false). If the loader hands `plan_sections` a differently-shaped schema,
   `extractArrayItemFields` can still succeed (it needs only `items.properties`) while
   `StructuredItemShape`'s **first guard fails**: it returns early unless
   `declaresArray(fieldDef["type"])`. **That guard is stricter than
   `extractArrayItemFields`' entry condition — a real asymmetry in my design, and worth
   fixing regardless of whether it is the cause.**
2. The `section_plan` read was not produced by that execution (carried from a parent, or a
   cached/echoed result).
3. Something between the action's return and serialisation drops the keys.

### The next experiment — inspection is exhausted, instrument it

Add a temporary `logger.Warn` in the `source == "llm"` branch
(`plan_sections_action.go` ~:2718) printing `fieldName`, `fmt.Sprintf("%T / %v",
fieldDef["type"], fieldDef["type"])`, and whether `fieldDef["items"]` type-asserts to a map
— or execute the resolver locally against the live DB row. Either settles hypothesis 1 in
one run. **Do not spend more queries on `orchestration_states`; that avenue is spent.**

⚠ **A trap that cost me a wrong turn:** the step's `output_field` is **`section_plan`**, not
`plan_sections`. Both keys exist in `collected_data` and they are not the same object.

## What must NOT be undone

- **No rollback is needed or advised.** With the Go side emitting nothing, migration 724's
  `{{if}}` guards render *exactly* the pre-fix prompt — which is what post-roll observation
  confirms. The migration is inert, not harmful.
- The livespec declaration marks the pre-437 spelling **Forbidden**, so reverting 724 will
  make the daily live-declaration-drift check fire. That is the mechanism working.

## Other lanes, and what they are owed

- **`portfolio_positioning`** holds advertise.co.uk ready as the live test and is **waiting
  on my word**. They must NOT fire until the fix actually works — I told them to check
  `merge-base --is-ancestor` and the `prompt_rendered` probe first, and the second of those
  now FAILS. **Tell them the roll happened and the fix did not take effect**, so they keep
  holding. Their page's item is on a fresh key (`needs_content_page:288baf25-…`), so it is
  safe to let it fail; it is the second and third failures on the SAME key that brand it.
- **`components` / `bugsweep4`** asked whether the legacy dialect is this bug. Answered (it
  is not) and their census was used to cross-check my blast radius. Nothing owed.
- **`bugs_open/453`** carries a CONTRIB from this lane: `<no value>` in 65% of writer
  prompts (`Location: {{.reviewed_brief.headquarters}}` inside a "DO NOT INVENT" block),
  plus a correction to their fix candidate 1 (their lint diffs template ROOTS against
  `input_fields`; that instance's root IS present and the SUB-FIELD is absent, so a third
  failure shape exists that no static lint over config can see). Zero live damage measured.

## The re-mint hazard (still true, still moving)

While candidate 1 does not work, automatic sweeps keep re-minting these pages and each
attempt burns a sibling toward the sticky `[unresolved after 2 attempts]` brand, which
blocks re-minting for ever. `[MEASURED 2026-09-03 ~11:00Z]` keys already at/past the
threshold: **farmerinsurance.uk 21, remortgagecalculator.uk 6, loanzy.uk 3**. ⚠ **Re-measure,
never quote** — the threshold counts a ROLLING 7-day window, so it decays as siblings age
out and climbs on every sweep. Query is in `bugs_open/437` §Verify.

## Still owed

1. Resolve the open problem above; then verify at the artefact (prompt shows the nested
   exemplar; a previously-failing page stores `branches` as an ARRAY in
   `page_components.content_data`), with a **demand control** on any post-fix zero.
2. **Read the council round-2 verdict** and act on it —
   `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
   correlation_id='6de0f6f2-4f37-492a-9cbd-1ae886311a9b' AND kind='council_report' ORDER BY created_at;`
3. 437 candidates **2** (no repair path for already-terminal items) and **3** (nothing
   escalates an active, linked, never-built page) remain OPEN. 3 is why these sat for weeks.
4. Over-production watch once it works: exemplars govern, so census the `branches` fill-rate.

## Commits from this session

`a0044e73b` fix · `f88789e37` gofmt + register sha · `53b2f46af` omission-spelling test ·
`01e98a6d0` NOTES/RUNBOOK council round 1 · `b8d8862c0` 453 CONTRIB · `f9550f8ef` re-mint
hazard · `58b166955` dialect settlement + cross-check.

## Lessons already banked (do not re-learn)

`WRONG_CALLS.md` — a diff-guard's expected numbers must be DERIVED from a rehearsal, not
read off the replacement (mine said `{{if }}` +2; it is +1, and the guard would have refused
my own correct splice). `LANDMINES.md` — a prompt exemplar generated from a lossy projection
states the wrong type with the schema's full authority. `016b` §9 — the same, as a debugging
pattern, with the two-minute diagnosis. RUNBOOK — never put a placeholder in a council
sketch; rehearse migrations twice; `--record-only` needs `--note`; the `jsonb_typeof(...)`
NULL trap.


---

## ⚠ ADDENDUM 2026-09-03 12:55Z — the architecture I had missed, and why the verdict is narrowed

**1. No new chassis build has been deployed** (checked when asked to verify one). `IMAGE_TAG`
in the makefile, the `agent-chassis` deployment spec, and the running pods are all
**`v1.0.1358`**, pods started 12:06:47Z / 12:07:16Z, `rollout status` reports complete. There
is no newer local image. So there was nothing new to verify — the build being referred to is
either still in progress, or is the same 12:06Z roll already covered above.

**2. THE THING I HAD MISSED: agent work runs in PER-AGENT PODS, not in the `agent-chassis`
deployment.** `kubectl get pods --sort-by=.status.startTime` shows ephemeral pods named
`agent-page-build-handler-*`, `agent-page-content-writer-*`, `agent-site-review-agent-*`,
`agent-landmine-verifier-*` … each running the chassis image and spawned per job. A fresh
crop appeared at 12:49–12:52Z.

**This invalidates the pod I used for elimination #2.** I probed
`agent-chassis-554857f96f-kx69c`, a deployment pod, and treated it as "the running binary".
`plan_sections` executes inside an `agent-page-build-handler` pod. Re-probed at 12:55Z on
`agent-page-build-handler-2c993dc2-cvfl5`: the fix literal IS present (1), with the
long-lived control present (3). **So today's agent pods carry the fix.**

**3. Which reframes the failure.** The five orchestrations I measured ran 12:07:57–12:20:25Z,
i.e. within ~14 minutes of the deployment roll. Agent pods are spawned per job and outlive a
single one, so those executions plausibly ran on pods created BEFORE 12:06Z, on the previous
image — which would explain a correct, deployed fix emitting nothing, with no contradiction
anywhere. **I cannot prove this retroactively** (those pods are gone), and I am not asserting
it: it is now the leading hypothesis, ahead of the `comp.InputSchema` one.

Checked and NOT the explanation: neither `page-build-handler` nor `page-content-writer` is
image-pinned (`default_config.pin_image_tag` unset; `agent_image.go:201`), and both rows'
`image_tag` reads `v1.0.1358`.

⚠ Noted, unexplained, probably harmless: the agent pod's environment carries
`AGENT_IMAGE_TAG=v1.0.44` while the pod's actual image is `v1.0.1358`. Worth a glance if
image resolution ever comes back into question — it is not evidence of anything on its own.

**4. THE DECISIVE TEST, and it is already armed.** No mechanism-flow page has been written
since the fresh pods came up (0 writer calls mentioning `branches` after 12:40Z), so the fix
is simply un-exercised. A background watcher is polling for the first such prompt after
12:52Z and will report `nested_exemplar` / `old_flat` / `shape_note`. If `nested_exemplar` is
true, **candidate 1 is done and this lane closes** pending the artefact check on a stored
page. If it is false on a pod proven to carry the fix, the `comp.InputSchema` hypothesis
returns and instrumentation is the next step.

**Nothing needs to be fired to make this happen:** advertise.co.uk's `e75f5880`
(`needs_content_page`, `triaged`) will dispatch on its own and is on a fresh key, so one
failure costs nothing. The `portfolio_positioning` lane has been told to keep holding and
why.

**5. The lesson, which is this estate's own and I still walked into it:** *"`-l app=<subsystem>`
may be the WRONG SERVICE (one image, every label); before believing a clean grep, ask which
pod could have produced the line."* I asked which pod ran the code only after exhausting
every other explanation. **Probe the pod that does the work, not the one that shares its
name.**


---

## ✅ RESOLVED 2026-09-03 13:25Z — and the morning's "it does not work" was my own measurement error

The watcher fired. The fix works; see §The one-line state.

**What actually happened, and it is the most transferable thing in this lane:** I spent the
middle of the day proving, with controls, that a correct and deployed fix "did not take
effect" — because **I probed the wrong pod**. Agent work runs in ephemeral **per-agent pods**
(`agent-page-build-handler-*`, `agent-page-content-writer-*`), spawned per job, not in the
`agent-chassis` deployment whose pod I grepped. The executions I measured (12:07–12:20Z) ran
on agent pods spawned *before* the 12:06Z roll, so they legitimately lacked the fix while
every artefact I chose to look at legitimately showed it present.

Every one of my eight eliminations was individually sound. The failure was in the question:
I asked "is the fix in the binary?" and never "in WHICH binary did the code that produced
this row run?" **The estate's own landmine says exactly this** — *"`-l app=<subsystem>` may
be the WRONG SERVICE (one image, every label); before believing a clean grep, ask which pod
could have produced the line"* — and I had re-read it that morning.

**The cheap check that would have saved hours:**
`kubectl get pods --sort-by=.status.startTime` and look at what is actually running. The
per-agent pods are visible in one line of output and I never listed them.

## What is owed next

1. **Re-census after traffic.** The 0-failures result rests on one exercise. §Verify in
   `bugs_open/437` has the query and the demand control.
2. **Tell `portfolio_positioning` they can proceed** — done as of this session; advertise's
   `e75f5880` is on a fresh key and can now be allowed to run.
3. **Candidates 2 and 3 remain OPEN**, and 3 (nothing escalates an active, linked,
   never-built page) is the more valuable — it is why these sat for weeks.
4. **Read the council round-2 verdict** — corr `6de0f6f2-4f37-492a-9cbd-1ae886311a9b`.
5. **The re-mint hazard is now decaying in the right direction**, since builds can succeed
   again. Re-measure rather than quoting the morning's table.
