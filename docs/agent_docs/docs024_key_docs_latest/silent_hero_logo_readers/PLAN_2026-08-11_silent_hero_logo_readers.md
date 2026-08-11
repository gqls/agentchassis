# PLAN — make the three silent hero/logo readers say something (commission item 2)

**Authority:** owner ruling 2026-08-10, item 2 of
`rfc012_await_findings/COMMISSION_2026-08-10_owner_rulings_five_pieces.md` — *"2. yes."*

**Why now:** item 5 is done, approved and live (`diagnosis_schema_visibility/`), so the
commission's order 5 → 2 → 1 → 3 puts this next. It is also more than "next" — item 5's
`090` re-run proved the *tool* works and the *data* is gone, so **item 2 is now item 1's
unblocker**, which the commission did not anticipate. Recorded in `bugs_open/236` §5b and in
that lane's handoff §3.

---

## 1. The three sites — LINE NUMBERS CORRECTED

The commission's line numbers had drifted by the time this lane opened. Re-located
2026-08-11 by `grep -rn "hero_deployed\|logo_deployed" --include=*.go platform/ internal/`:

| commission said | actually | what |
|---|---|---|
| `v3_site_actions.go:1020` | **`:1125`** | `hero_deployed` → `hero_url` |
| `v3_site_actions.go:1031` | **`:1136`** | `logo_deployed` → `logo_url` |
| `assemble_from_library.go:452` | **`:452`** | `extractLogoURLFromParams`, third rung of a ladder |

All three are `ok`-guarded map accesses with no `else`. A miss writes nothing anywhere.

The third is not the same shape as the first two, and the difference matters for the design:
it is the **last rung of a three-rung ladder** (`render_context.logo_url` → `site_record.logo_url`
→ `logo_deployed.image_url` → `""`). A miss there is only interesting when the whole ladder
failed *and* a deploy actually ran.

## 2. Two decisions that depart from the commission's letter — and why

### Decision 1 — a durable `agent_error_log` row, not only a `Warn`

The commission says *"a `Warn` naming the key it wanted, and what the map actually contained"*.
**A `Warn` alone cannot deliver item 2's stated purpose**, which `bugs_open/236` §5b states as
writing the evidence *"into `agent_error_log` … where nothing prunes it on an orchestration's
schedule"*. Three measured facts, none of them mine to re-derive:

1. **Pod logs do not keep it.** These readers run in `agent-chassis`, the busiest service on
   the estate. The item 5 lane measured its startup `build provenance` line **absent from
   `--tail=3000`** hours after the roll (2026-08-11). A `Warn` on a page build is gone by the
   time anyone asks.
2. **The orchestration row does not keep it either — the window is 4 hours.** The
   `bugfix_236_site_availability` lane measured `database-cleanup` (live row, hourly) deleting
   `AWAITING_RESPONSES` states after **4 hours**, and that is exactly the state in which
   `hero_deployed`/`logo_deployed` exist, *because they are the awaited responses*. The three
   observations of this bug (0 of 1,667 → 2 each → 0) are one 4-hour window opening and closing.
3. **This table is documented as the only sink that survives.**
   `agenterrors.go:20-24` and `log_action_error.go:14-18`, verbatim: *"THIS TABLE IS THE ONLY
   SINK THAT SURVIVES AN AWAITED STEP: the collected_data sibling key was refuted live."*

So: **both**. The `Warn` for immediate visibility and for the post-roll pod-grep (a log string is
a real literal — the commission's own note), and the durable row for the evidence item 1 needs.
This is not scope creep away from "observability only": it is the same observability, put where
it survives.

### Decision 2 — log only when there was DEMAND, never on plain absence

`BuildRenderContextAction` runs for every page build, and **most pages never deploy a hero or a
logo**. An `else` on the outer guard would file a row on every page of every site — noise that
buries the signal and costs rows fleet-wide.

**The demand signal is the presence of the container key.** `hero_deployed` exists in
`collected_data` only because a `deploy_hero_image` step ran. So:

| state | reading | action |
|---|---|---|
| key absent | no hero was wanted | **silent** — the no-op case, deliberately |
| key present, is a map, has usable `image_url` | working | silent (the existing `Info` already logs it) |
| key present, is a map, **no usable `image_url`** | **the 236 shape** | `Warn` + durable row |
| key present, **not a map** | anomalous shape | `Warn` + durable row |

This is the demand-control discipline applied at write time rather than at measurement time:
the row is filed only when something actually asked for an image and did not get one.

> ⚠ **STATED BLIND SPOT.** This is silent if a future mechanism removes the container key
> *entirely*, because then there is no local evidence a deploy ran at all. The observed shape
> (`bugs_open/236` §1) has the key **present** carrying
> `response`/`response_status`/`response_received_at`, and `coordinator.go:2719-2748` preserves
> then adds — so key-present is the shape to catch today. Detecting "a hero was wanted" without
> the key would require the reader to know the workflow declared a deploy step, which a reader
> cannot know locally. **Not built, and not measured whether it ever happens.** `[UNMEASURED]`

### Decision 3 — declared provenance inheritance

The finding belongs to the step that is executing (the reader), so these sites call the
**declared** opt-in `LogActionEntryInheritingProvenance` / `LogActionFindings` rather than
letting provenance be filled silently. That is the unsafe-default-OFF discipline
`log_action_error.go:34-46` implements per owner ruling 2026-08-02 (RFC_010 §2), and it means a
reviewer reading only the call site can see the decision.

### Decision 4 — scope routing: council gate, not architecture RFC

Checked against `PROCESS_architecture_review.md` as worded (owner ruling 2026-08-10 item 4,
*"it adds, changes or removes an exported symbol other packages depend on"*):

- every symbol touched is **unexported and package-local** (`extractLogoURLFromParams`,
  `buildRenderContextFromParams`, and the new helper);
- it **adds no shared vocabulary** — the one new `error_code` is a queryable literal in a table
  that already carries dozens, with no automated consumer;
- **nothing a shared mechanism guarantees changes** (owner ruling 2026-07-29 §1): it calls an
  existing door in its documented way and adds no authority to it.

Two signature changes (`buildRenderContextFromParams`, `extractLogoURLFromParams` gain `ctx`)
are package-local and mechanical. → **council gate.**

## 3. What is NOT in scope

- **Not widening the read.** Do not also accept `response.data.file_path` /
  `response.data.files[0]`. The commission is explicit, and `bugs_open/236` §4 candidate 2 says
  why: it encodes the merged shape at three call sites, which is the patch-one-site-by-hand
  pattern the RFC_012 census already found at `unified_extractor.go:200`. **That is item 1's
  design decision and it is reserved to the owner.**
- **Not fixing the loss.** Item 1's root cause is still open (`bugs_open/236` §5) and this lane
  makes no claim about it.
- **Not the `deploy_image_asset` writer.** See NOTES for a finding about its existing
  collected_data sibling-key workaround; it goes to `bugs_open/236` as a contribution, not into
  this change.

## 4. Verification

| step | how | state |
|---|---|---|
| unit tests incl. the no-op case | `go test ./platform/orchestration/actions/ -run ...` | pending |
| the guard is real, not a mock's bookkeeping | mutate the condition, watch the test fail | pending |
| council verdict | `097_TRIGGER`, read the objections | pending |
| committed | pathspec commit | pending |
| chassis image + roll | whole-fleet, owner runs `make release` | pending |
| pod-grep the new literal after the roll | a log string is a real literal, so this one greps | pending |
| **the real test: a durable row appears** | `agent_error_log` query in the RUNBOOK, after a build that deploys a hero/logo | pending |

**The last row is the only one that proves the point.** Everything above it proves the code
shipped; only a row filed at the moment of failure proves item 1 has its evidence. It needs a
site build that deploys a hero or logo, so it cannot be forced from this lane alone.
