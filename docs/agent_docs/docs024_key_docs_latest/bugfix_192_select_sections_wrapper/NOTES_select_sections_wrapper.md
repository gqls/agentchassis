# NOTES — `bugs_open/192`, the section_plan wrapper

Append-only, newest at the bottom.

---

## 2026-08-04 ~09:35 — picking the bug, and the two candidates I put down

Swept `bugs_open/` against live session transcripts rather than `who-owns.py` alone
(that script reads COMMITS, so a session mid-fix is invisible to it — the standing
lesson). 30-odd sessions are live on this tree; nearly every recent bug is claimed.

Two candidates I opened and deliberately put down, recorded because the reasons are
reusable:

- **`132`** (every B2 site serves a raw JSON blob instead of a 404) — genuinely
  unowned, still reproducible (I re-curled webdesign.co.uk, robot-hands.com,
  oufe.com today: all still raw JSON; relojistas.com still the nginx control). Put
  down because its fix is a **Cloudflare-side edit with no deploy path from this
  tree** — no `wrangler.toml`, and the repo's `scripts/cloudflare/worker.js` is
  provably NOT what is deployed (its miss branch returns plain "Not found", the edge
  returns JSON). I could have written a fix I could neither ship nor verify.
- **`114`** (imagery generated, deployed, never referenced) — unowned, but its
  load-bearing candidate is "drain or stop filing `image_landed`", and those rows
  are `item_type='needs_page'`, which is exactly the seam the `187` lane fixed and
  committed **today**. Taking it would have been competing, not contributing.

Took **`192`** instead: filed 30 minutes earlier by the `154` lane while
live-verifying `178`, explicitly *"not yet diagnosed"* and handed off. Confirmed
only the filing lane and mine had it in transcript.

## 2026-08-04 ~09:40 — filed the `090` run before asserting anything

Per the CLAUDE.md default (durable, cross-cutting claim ⇒ file first). Intake
`b45144fa-2051-4ff8-9318-e052a9c3a084`, run `aea3cc68-b274-4df4-a1c1-f60ba47bf09e`.
Checked the diagnosis queue first — no open `needs_diagnosis` touched this. Then
diagnosed first-hand in parallel rather than waiting the ~30 minutes.

## 2026-08-04 ~09:45 — the diagnosis, and the measurement that turned it

The bug file's own open question was whether the fallback fails for an **ordering**
reason (the step running before `input_data.section_plan` is merged). It is not
ordering. The step that decided it was enumerating the KEYS rather than reading the
path:

```sql
SELECT (SELECT string_agg(k, ',' ORDER BY k)
        FROM jsonb_object_keys(collected_data->'input_data'->'section_plan') k) AS keys,
       jsonb_array_length(coalesce(collected_data->'input_data'->'section_plan'->'sections_ready','[]')) AS direct,
       jsonb_array_length(coalesce(collected_data->'input_data'->'section_plan'->'section_plan'->'sections_ready','[]')) AS nested
FROM orchestration_states WHERE owner_agent_type='page-content-writer' ORDER BY updated_at DESC;
```

```
3b692317 | applied,reason,section_plan              | direct 0 | nested 7   <- FAILING
0883b1aa | applied,matched,section_plan             | direct 0 | nested 1   <- FAILING
df69efd6 | applied,matched,section_plan             | direct 0 | nested 1   <- FAILING
4edcdaeb | deferred_count,…,sections_ready          | direct 3 | nested 0   <- flat, pre-seed
```

Two shapes. The failing ones carry a **wrapper**; the data was never missing, it was
one level down. `sections_for_render` on all three is `{}`.

The producer is `load_current_section_content_action.go` (`178`'s fix, `08d0515f3`),
wired by seed `299` with `output_field: section_plan` — and `coordinator.go:1859-61`
(`storeActionResult`) stores an action's return value **wholesale** under
`output_field`. The action returns `{section_plan, applied, reason|matched}` on
**every** path including its eight "pass-throughs", so the flat plan is replaced by
the wrapper on every build in every mode.

**The part I nearly missed:** this kills BOTH of `select_sections`' fallback paths
from one cause. Path 2 directly; path 1 because `resolve_links`' input_mapping is
`"sections?": "input_data.section_plan.sections_ready"` — the resolver child is
handed no sections and returns `sections_ready: null`, which is the "present but
explicitly null" value the bug file reported as if it were an independent second
fault. I would have written that up as two problems. It is one.

## 2026-08-04 ~09:50 — MISSTEP AVOIDED, and the one I nearly shipped

I had drafted the seed to fix `resolve_links`' input_mapping too, pointing it at the
nested path so link resolution worked again immediately. **That would have been a
silent time-bomb.** `input_mapping` has no ordered fallback: the moment the Go fix
rolls and the plan is flat again, a nested-path mapping resolves to nothing, the
resolver silently gets no sections, and there is no error to notice — the exact
shape of defect I am here to fix. Dropped it. `select_sections`' shim is safe only
because its paths are tried **in order** and the flat path sits ahead of the shim, so
the shim retires itself on the roll. Recorded in the seed header as a deliberate
non-change.

## 2026-08-04 ~09:55 — the filing's onset is wrong, and it pointed away from the cause

`192` says "failing broadly since ~2026-08-03 21:00" and cites hourly FAILED spikes
of 11/14/12. Re-measured over the full retained window, splitting on the step name:

```
08-04 08:00 | select_sections_miss  | FAILED |  3
08-04 01:00 | iter_generate_content | FAILED |  1
08-03 23:00 | iter_generate_content | FAILED | 12
08-03 22:00 | iter_generate_content | FAILED | 14
08-03 21:00 | iter_generate_content | FAILED | 11   (+ 12 COMPLETED the same hour)
```

The overnight spike is `process_sections_loop_iter_N_generate_content` — a step only
reachable **after** `select_sections` has already succeeded. It is a different,
still-undiagnosed defect. This bug's mode (`current_step='process_sections_loop'`
exactly) starts at **08:20 on 08-04**, which is `agent_definitions.updated_at` for
both agents — i.e. when seed `299` was applied.

This is not pedantry: the wrong onset is what made "178 cannot be the cause" look
sound, because 08-03 21:00 predates the trigger. Logged in `WRONG_CALLS.md`, and
corrected in place at the head of the bug file.

## 2026-08-04 ~10:05 — the test that passed on the broken code

`load_current_section_content_action_test.go` asserted `result["applied"]`,
`result["reason"]`, `result["section_plan"]` — i.e. it encoded the wrapper as the
contract, and so **passed on the code that broke every page build in the fleet**.
Its own header, two lines above, says the step "must leave section_plan
byte-identical". The test and its comment disagreed and the comment was right.

Rewrote it to assert the contract instead of the observed shape, and added
`TestLoadCurrentSectionContent_NeverWraps` walking all four reachable return paths.

**Then proved it could fail**, because a passing test proves nothing on its own:
mutated the action back to the wrapper form, re-ran, confirmed **all five** cases
fail with the intended message, restored, and diffed to confirm the restore was
clean. (Short window on a shared tree — copy first, restore in the same command.)

```
--- FAIL: TestLoadCurrentSectionContent_NoMode_Passthrough
    pass-through must return section_plan unchanged.
    bugs_open/192 regression: the plan is nested under its own key …
--- FAIL: …ModeRecreate_Passthrough / …EditLive_AttachesMatchingSlot /
    …NoReadySections_SkipsQuery / …NeverWraps (all 4 sub-cases)
```

## 2026-08-04 ~10:15 — seed 308 applied, deliberately before its image

Inverts the house "image first, then seeds" rule, and the header says why: that rule
exists because a seed naming an **unregistered action** fails at runtime, and this
seed names no action. The third fallback path is plain data; `required` is a config
key the running binary provably ignores (`ExtractFieldsAction` reads only `fields`,
`field_map`, `defaults`). So it restores builds on the binary already deployed.

Applied clean: `UPDATE 1`, both `DO` verification blocks passed, `COMMIT`. Wrote the
verification as `DO`/`RAISE` rather than bare `SELECT`s — a `SELECT` below an
`UPDATE` cannot stop a `COMMIT` under `ON_ERROR_STOP`, which is a landmine already on
file.
