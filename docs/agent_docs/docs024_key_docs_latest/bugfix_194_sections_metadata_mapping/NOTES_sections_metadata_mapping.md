# NOTES — `bugs_open/194`, `save_page_sections` and the caller-supplied metadata path

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 ~19:35Z — picking the bug, and how I established it was free

Concurrency check before claiming (CLAUDE.md: "check other threads are not already looking
at this bug"), in three layers because each is blind in a different way:

1. `scripts/who-owns.py 194` — **advisory and LAGGING**: it reads commits, so a session
   mid-fix with nothing committed is invisible to it. Said `likely OWNING workstream(s): (none
   identified)`.
2. **Live `.jsonl` transcripts** — the layer that sees uncommitted sessions. 12 sessions
   modified in the last 4 hours mention 194; I read the CONTEXT of each mention rather than
   counting them, because every session that runs `ls bugs_open/` gets a hit per file and a
   raw tally is mostly listing noise:
   - `444c951d` (the `087`/`193`/`195` lane) cites 194 as the case it *spun out* — 41 mentions
     of `save_page_sections`, but as the closing note of 087, not as work in flight;
   - `96678671` was *offered* 194 by its operator ("possibly bugs_open/194?") and picked
     `189` instead — it is mid-fix on `SiblingSignatures` with a lane dir already created;
   - `cf156d11`, `86c1f378` — 194 appears only inside a `bugs_open/` listing or a `who-owns`
     dump. `86c1f378` is a sibling bug-clearing thread still surveying (candidate list
     includes 194), which is why I claimed in a commit rather than after finishing the work.
3. **A claim commit before any research** (`1dc0ddcb7`), so the sibling survey sees it.

**The measurement that could have come out otherwise:** if the transcript grep had shown a
session with `sections_metadata_field` in its recent tool calls, I would have dropped 194.
It showed 2 mentions in `444c951d` — both inside the 087 close-out text.

## Validity re-check at claim time (the bug is 12 hours old, but the tree moves hourly)

Census re-run live, not read off the bug file. Six `save_page_sections` steps fleet-wide,
all with step key `save_sections`:

```sql
SELECT ad.type, s.key, s.value->'config'->>'sections_metadata_field', s.value->'config'->>'html_field'
FROM agent_definitions ad,
LATERAL jsonb_path_query(ad.default_config, '$.**.steps') AS steps,
LATERAL jsonb_each(steps) AS s(key,value)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' = 'save_page_sections';
```

| agent | `sections_metadata_field` | `html_field` |
|---|---|---|
| page-build-handler | `page_content.response.sections_metadata` | `validation_result.clean_html` |
| page-rebuild | `page_content.response.sections_metadata` (seed 310) | `assembled_page.html` |
| page-rerender | `rerender_sections.sections_metadata` | *(none)* |
| pageflow-builder | **absent** | `assembled_page.html` |
| site-work-orchestrator | **absent** | `assembled_page.html` |
| tool-recreation-handler | **absent** | `validation_result.clean_html` |

Still valid: three instances unfixed. `$.**.steps` matters — the step is nested inside a loop
`sub_workflow` in four of the six and a top-level `jsonb_each` misses them.

## The bug file's one [UNMEASURED] item, now MEASURED

`tool-recreation-handler`'s writer flow. Its step graph has **no writer call at all**:
`recreate_tool` (`execute_llm_prompt` → `tool_recreation`), `validate_tool`
(`validate_page_content` → `validation_result`), saving from `validation_result.clean_html`.
There is no `sections_metadata` anywhere on that path — it recreates a whole-page tool HTML
blob, not a section set. **So a NULL `content_data` there is not this defect**; copying the
key onto it, as the bug file warned, would have been an unmeasured claim. The other two
(`pageflow-builder`, `site-work-orchestrator`) both have `write_page_content` = `call_agent`
with `output_field: page_content` in the same loop sub-workflow as `save_sections`, so
`page_content.response.sections_metadata` resolves for both — the bug file's prediction holds.

## Dormancy — and why `orchestration_states` cannot answer it

`orchestration_states` retains ~1 day of `COMPLETED` rows (`min(created_at)` per status:
COMPLETED 2026-08-03, i.e. yesterday), so "0 runs in the window" there means "none since
yesterday", not "dormant". Sourced instead from `agent_run_stats`, which has no reaper and
spans 2026-07-26 → 2026-08-04 across 95 agent types:

| agent | run_count | last_ran_at |
|---|---|---|
| page-rerender | 2878 | 2026-08-04 |
| page-build-handler | 283 | 2026-08-04 |
| page-rebuild | 4 | 2026-08-04 |
| tool-recreation-handler | 3 | 2026-08-02 |
| **pageflow-builder** | **absent** | — |
| **site-work-orchestrator** | **absent** | — |

The disconfirming shape: if `agent_run_stats` only tracked leaf agents, an orchestrator's
absence would be an artefact of the table rather than a fact about the fleet. It does not —
`build-dispatch-loop` (2352), `content-feed-orchestrator` (161),
`med-price-scrape-orchestrator` (36) are all in it. So the absence is meaningful:
**the two callers a Go-side default would newly re-route are both dormant over 9 days.**
That is good for blast radius and bad for provability — a live run cannot prove them, so
whatever ships needs offline proof that can fail.
