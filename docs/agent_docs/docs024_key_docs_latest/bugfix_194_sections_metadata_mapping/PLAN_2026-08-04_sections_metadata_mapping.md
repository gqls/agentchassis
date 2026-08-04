# PLAN — `bugs_open/194`: the saver that depends on being told where its own input lives

**Opened:** 2026-08-04 · **Lane:** bug-clearing thread, session `da43ef00`
**Bug:** `bugs_open/194` · **Council:** `b6023fc1-ae70-4486-b752-d399e9b1afcc`

## The problem, stated as a mechanism

`SavePageSectionsAction` persists two representations of each page section:
`rendered_html` (what a visitor sees) and `content_data` (the only thing
`rerender_page_sections` can regenerate the section from). Which `collected_data` path
holds the structured half is supplied **per caller** by the config key
`sections_metadata_field`, introduced 2026-02-18 (`d7b07f4fa`) and never defaulted. A
caller that does not set it falls through to the regex HTML-parse path, whose sections
carry no `content_data`, so the INSERT writes SQL NULL — and the save reports success.

Four of the six live callers had never set it.

## Decisions, and the reasons

### D1 — fix BOTH halves, config first, and do not let the config half stand as the fix

The config half (seed 312) is live on apply and closes the bug as filed. The Go half is
inert until the next chassis roll and closes the *class*. Both, in that order, because:

- config alone leaves the next caller free to make the same mistake, and seed 312 was
  already the third and fourth hand-written copy of one path string (055/065, 034, 310);
- Go alone leaves the two dormant callers broken until a roll, and makes the live
  mechanism an implicit default that nobody's config names — worse to review, not better.

**The emergent property is what makes this safe:** after the seeds, five of six callers
NAME the field and the sixth has no `page_content` in scope at all, so the new default is
consulted by **zero live steps today**. The Go change's live blast radius is nil by
measurement, which is what makes it a normal council-gate change rather than an
architecture round.

### D2 — a single default, NOT an ordered probe

An earlier shape tried both known paths in order (`page_content…`, then
`rerender_sections…`). Rejected: `page-rerender` names its field explicitly and always
will, so probing its path from every save adds a second way to resolve for zero benefit —
and a save that finds a *different* run's metadata under a path nobody configured is a
worse failure than the NULL it was meant to prevent. One default, consulted only when the
caller has said nothing at all.

The default is `defaultSectionsMetadataField`, **referenced from
`validate_page_content_stats.go:75`, not copied**, so the gate and the save cannot drift
about where a page's structured content lives. Two hand-maintained copies of one path is
the exact drift class this bug is an instance of.

### D3 — the report is not a guard, and the refusal is opt-in with the unsafe default OFF

The bug's teeth are the silence: a save that strips `content_data` succeeds, serves, and
tells nobody. Three responses were available and the choice between them matters.

- **A durable record** (`CONTENT_DATA_REGRESSION`, severity `warning`) — chosen, always on.
- **A sixth unconditional refusal** — rejected. This function already carries five, and
  `LANDMINES` records the consequence: *"a green orchestration status no longer means the
  sections were written"*. Adding one to the fleet's highest-traffic save path
  (`page-rerender`, 2,878 runs in nine days) on the strength of a prediction rather than a
  measurement is how a guard gets deleted six weeks later.
- **An opt-in refusal** (`require_sections_metadata: true`) — chosen, **seeded on nobody**.
  RFC_010 (owner ruling 2026-08-02): new authority on a shared seam ships as a field with
  the unsafe default OFF, because "a comment is not a control on a tree this many sessions
  share". The record is what turns the later per-caller opt-in into a measurement.

**The honest cost of D3, named:** a mechanism nobody enables is one this estate has been
bitten by before. If the report runs quiet for a week and nobody opts in,
`require_sections_metadata` is dead weight and should be either seeded or deleted — it is
written into PBP-031's `verify-later` so the question comes back.

### D4 — `tool-recreation-handler` gets a declaration, not the key

The bug file flagged its response shape `[UNMEASURED]` and was right to. Measured: it has
no writer step at all (`recreate_tool` → `validate_tool` → save from
`validation_result.clean_html`), so it has no structured sections and its NULL is correct.
`rerender_page_sections_action.go:318` already agrees, exempting self-contained tool
sections from the missing-content escalation. It is the intended consumer of
`expects_no_sections_metadata` — the declaration exists so that "this caller has no
structured content" is a fact a reviewer of the CALLER can see, rather than a comment in
the callee.

### D5 — proof is offline, because the changed callers cannot be proven online

`pageflow-builder` and `site-work-orchestrator` are absent from `agent_run_stats` over its
whole 9-day span. A live run cannot prove them, so the proof is:

- seven unit tests over the resolution seam, with **discriminating** fixtures (the
  configured-vs-default test holds *different* arrays at the two paths — identical ones
  would pass whichever path won);
- **four mutations actually run** against the shipped code, each failing the test that
  names it, all green again on restore;
- the whole `actions` package green against `git archive HEAD` **plus these three files**,
  because the shared working tree's suite is red on another session's uncommitted
  `revalidate_review_queue` work.

`site-work-orchestrator` IS directly dispatchable
(`scripts/initial_messages/170_work_item_flow_build/075d_simple_maintain_trigger.sh`), so
one of the two can be proven live once the image rolls. That is the post-roll acceptance
run, and its disconfirming outcome is written down before it is run (see below).

## Verification, stated before it is run

**Acceptance (post-roll).** Dispatch `site-work-orchestrator` at a site whose target pages
are `rebuild_policy != 'owned'` (check the column first — 087's run was blocked by exactly
that guard), then:

```sql
SELECT slot_name, length(rendered_html), length(content_data::text), updated_at
FROM page_components WHERE page_id='<uuid>' ORDER BY position;
```

**Pass:** `content_data` non-NULL on every row at the new run's `updated_at`, AND the save
step's result carries `sections_source: 'metadata'`. The second half is not decoration —
`content_data` can also arrive via the interactive carry-forward, so a bare non-NULL check
is a **false pass**.
**Disconfirming:** still NULL, or `sections_source: 'html_parse'` — the writer's reply is
not reaching the save on that path and the mapping premise is wrong.

**No-regression (24h post-roll).**

```sql
SELECT agent_type, count(*) FROM agent_error_log
WHERE error_code='CONTENT_DATA_REGRESSION' GROUP BY 1;
```

**Pass:** zero rows for `page-build-handler` and `page-rerender`.
**Disconfirming:** any `page-rerender` row — the report's predicate is misconceived, and
the follow-up opt-in must not proceed until it is understood.

## What this does NOT fix (so nobody inherits an overclaim)

- The **161 of 1,201** already-NULL `page_components` rows. Repair is re-running the build,
  never restoring `page_component_history`: its `component_id` is NULLed by the FK's
  `ON DELETE SET NULL`, and pairing yesterday's `content_data` with today's `rendered_html`
  makes the next rerender reinstate the old page.
- The single-component writers outside this action (`bugs_open/136` names them).
- **Partial** `content_data` loss — deliberately outside the report's predicate, because a
  partial-loss test fires on legitimate compositions (a new plan that drops one section
  keeps content on the rest).
