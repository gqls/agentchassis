> **RESOLVED 2026-08-15** — the fix prescribed below was applied, five days
> after this was written and 15 days after the original revert. Migration
> `docs/agent_docs/sql_for_agents/411_model_directory_publisher_re_extend_to_company_protocol.sql`,
> commit `13e964287`. Verified with this doc's own discriminating query
> (entity counts 40/44/8, differ per kind) and at the HTTP artefact (all four
> URLs 200 with real content). Council submission `a3c418ea-4452-420d-b6e8-62ce78d5339e`
> (advisory, verdict pending at commit time). Full account:
> `NOTES_model_directory_pipeline.md` (2026-08-15 entry) and
> `README_where_we_are.md` (2026-08-15 entry). The find-sites kind-blindness
> defect noted below remains open, now owned by `portfolio_positioning`
> Phase B3c.

# FINDING 2026-08-10 — the four tracker feeds have never existed, and the 08-09 CONTRIB's "probably one dispatch" is WRONG

Written by `staged_component_build` at the owner's request ("which thread is handling the
tracker-feed 404s so we can get it fixed"). **Answer to that question first: nobody is.**
This lane's last owner commit is `758a90f43`, **2026-07-26** — 15 days ago. A transcript
sweep of all local sessions found no thread working it. Nobody is mid-fix; the work is
free to pick up. No status has been changed and nothing has been dispatched.

> **CORRECTION to `CONTRIB_2026-08-09_tracker_feeds_404_from_staged_component_build.md`
> (same lane that wrote it).** That file guessed: *"If the trackers' publish trigger simply
> never ran … the fix is probably one dispatch."* **That is wrong, and dispatching would
> have wasted a cycle and taught us nothing.** There is nothing to dispatch: the live
> publisher workflow has no tracker steps at all, so firing it a hundred times republishes
> only `model-directory.json`. Recorded here rather than silently corrected — the wrong
> guess is the useful part, because "the mechanism exists, it just hasn't run" is the
> shape this estate keeps mistaking for a scheduling problem.

## What is actually true

The Go half is **complete and kind-driven**. `directoryPublishProfiles`
(`platform/orchestration/actions/render_directory_action.go:98-126`) is a closed set of
three profiles, and the `company` / `protocol` entries name exactly the four 404ing files:

```go
"company":  {SnippetFile: "data/adoption-tracker.json",  FullFile: "data/adoption-tracker-full.json",  ...}
"protocol": {SnippetFile: "data/protocol-tracker.json",  FullFile: "data/protocol-tracker-full.json",  ...}
```

**What is missing is the DB config that asks for them.** The live
`model-directory-publisher` row has **three** steps — `render_model_directory_json`,
`commit_model_directory`, `complete`. Its only snapshot (v2) has three as well. No 7-step
config survives anywhere, and `grep -rn "render_adoption_json"` across the repo returns
only NOTES prose.

## How it got that way — from this lane's own NOTES

1. **2026-07-26 13:39** — the publisher was extended live to a 7-step chain
   (model→company→protocol), snapshot v2 taken. Applied by ad-hoc `UPDATE`; **the SQL was
   never committed** (commit `86602c3d1` carried only the components seed + migration 215).
2. **2026-07-26 14:24** — first run. All six steps reported success and **published the
   model register three times**, committing `data/model-directory.json` under the commit
   messages "Update adoption tracker" and "Update protocol tracker". Cause:
   `ExtractActionInputs` treated the string config `"kind": "company"` as a *reference*,
   never a literal, so it resolved to nothing and the Go default `"model"` won silently.
   (NOTES 773-809; `WRONG_CALLS.md:3494`.)
3. **Same hour** — fixed two ways: the live workflow was **reverted to the model-only
   chain from its own snapshot**, and `render_directory` was taught to read `kind` as a
   literal when the value is in the closed profile set (`bb99df77a`, now
   `render_directory_action.go:181-186`).
4. The NOTES' last word: *"Inert until an image roll; the chain gets re-extended then."*
   **The re-extension never happened.** The lane stopped that afternoon.

So the tracker feeds have never served. Even the one historical run wrote *model* files
under tracker commit messages — which is why the tracker paths 404 rather than serving
stale tracker data.

## The two things that make re-extending safe now, both checked

- **The Go fix is live in the pod.** It added no string literal, so the binary was dated by
  descendant literals from the same tree — `directory_citation_unverified`,
  `carried_empty_template`, `deduped_open_item` (all added 2026-08-03) each grep 1 on both
  replicas at `v1.0.1280`, with `ZZZ_NEG_CONTROL` → 0. A binary carrying 08-03 literals was
  built from a tree containing the 07-26 fix. **The trap that caused the 07-26 defect is
  closed.**
- **There is real data behind all four files** — so they would not publish empty:

  | kind | active entities | with current found claims |
  |---|---|---|
  | company | 32 | 32 |
  | model | 27 | 27 |
  | protocol | 4 | 4 |

  Corroborated at the artefact: `/adoption-tracker.html` already server-renders 15 cards
  and `/protocol-tracker.html` renders 4 protocols from these same rows.

Also verified: the site opt-in is present and correct (`site_specs` `classification`,
`is_current`, ai-agent-orchestration.com carries `adoption_tracker` and `protocol_tracker`
alongside `model_directory`), both listing pages are `active`/`deployed` so the full files
would be produced as well as the snippets, and the scheduled task
`model-directory-publish` is healthy and on cadence (every 21600s; last completed
2026-08-10 14:49:53).

## The fix — a config UPDATE plus a force-trigger, not a dispatch. Config-only, no roll.

The full worked SQL (snapshot-first, because `agent_definitions` has a `(type,version)`
unique key that a bare copy collides on — the trap the 07-26 session hit) is in this
lane's own materials and is reproduced in the session record. Shape:

1. `INSERT` a v3 snapshot of the live publisher row.
2. `UPDATE` its `default_config` to the 7-step chain — three `render`/`git_commit` pairs,
   with `"kind": "company"` and `"kind": "protocol"` as **literals** in the step config
   (which is what `bb99df77a` made readable), and `error_step` on each commit so one
   kind's git failure cannot abort the other two.
3. `UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name='model-directory-publish';`
   — the platform's own force-trigger idiom, verbatim from
   `internal/core-manager/admin/pipeline_admin_handlers.go:213-224`. The scheduler ticks
   every 30s and `loadDueTasks` selects on `last_triggered_at IS NULL`, so it fires next tick.

**Verify by the FILES, never the statuses** — that is exactly the check that caught the
07-26 defect, and *identical output across two supposedly-different kinds is the whole
signal*:

```sql
SELECT collected_data->'adoption_render_result'->'files' ? 'data/adoption-tracker.json' AS adoption_ok,
       collected_data->'protocol_render_result'->'files' ? 'data/protocol-tracker.json' AS protocol_ok,
       collected_data->'adoption_render_result'->>'entity_count',
       collected_data->'protocol_render_result'->>'entity_count'
FROM orchestration_states WHERE workflow_plan::text LIKE '%render_adoption_json%'
ORDER BY created_at DESC LIMIT 1;
```

Expect **different** entity counts per kind (~32 vs 4). **If they match each other, the
literal-kind read is not live and you have reproduced the 07-26 defect.**

## Marked unverified, deliberately

- `[UNVERIFIED]` that a 3-render/3-commit chain finishes inside the trigger's
  `call_publisher` `timeout_seconds: 600`. The 07-26 run completed all six steps, so it did
  once; not timed since.
- `[UNVERIFIED]` that committing the JSON into the `sites` repo reaches the bucket within
  any particular window. Inferred from `model-directory.json` serving 200 by this same
  path — a strong inference, not a measurement.
- `[UNVERIFIED]` whether the reverted 7-step config carried `error_step` on its commit
  steps; the NOTES record only "7-step chain, each render+commit pair". Added above on
  judgement, not recovered.
- **A latent defect worth fixing in the same edit:** the trigger's find-sites query gates
  only on `content_features.model_directory` — it does **not** check `adoption_tracker` /
  `protocol_tracker` opt-in. Harmless today (exactly one site matches and it is opted into
  all three), but a second site opting into `model_directory` alone would silently receive
  tracker feeds.
- **Absence of tracker publish runs is bounded by retention, not proof of never:**
  `orchestration_states` holds 2,057 rows and has zero rows for 2026-07-26, so the one
  historical run has aged out. The decisive evidence is structural — the live workflow has
  no tracker step — not the run count.

— `staged_component_build`, 2026-08-10. Raised because the owner asked who owns it; the
answer was "nobody", so the evidence is written down where the work lives rather than held
in a session.
