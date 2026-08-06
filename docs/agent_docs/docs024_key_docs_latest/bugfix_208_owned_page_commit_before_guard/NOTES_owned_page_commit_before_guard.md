# NOTES — bugs_open/208, owned page committed before the guard refuses it

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-08-06 — session pickup and first-hand re-verification

Picked up `bugs_open/208`, filed hours earlier by the `bugfix_201_page_content_writer_dispatch`
lane in pre-flight for an owner-authorised rebuild of `ai-agent-orchestration.com`. That lane
did **not** take the fix on — its own transcript says "208 is handed off. Continuing with 201
symptom 2", so the file's "OPEN, unowned" is accurate.

**Ownership check, and why the advisory VERDICT was overridden.** `scripts/who-owns.py 208`
returns **OWNED or recently active** — but only because the *filing* commit (`aaf8779e2`,
today) touches the file. That is the tool's documented blind spot: it reads commits, so a
filing looks identical to a claim. Per the memory note that every ownership check is lagging,
I also grepped the live session transcripts under
`~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`, which is the only source that
sees an *uncommitted* session:

```
for f in $(ls -t *.jsonl | head -30); do
  hits=$(tail -c 400000 "$f" | grep -oE "208_HANDOFF|rebuild_policy|queryPagesForBuild|get_pages_to_build" | sort | uniq -c | tr '\n' ' ')
  [ -n "$hits" ] && echo "$f [$(date -r $f +%H:%M)] $hits"
done
```

One session (`fef871d4…`, active) was deep in the mechanism — the filer. Reading its last
outputs showed it had handed 208 off and moved to 201 symptom 2. Two other sessions had 2
incidental `rebuild_policy` hits each, no 208 involvement. **Taken on.**

### The bug is still valid, and the blast radius is bigger than filed

Every link re-read first-hand rather than taken from the handoff.

**1. Selection ignores ownership — confirmed.** `queryPagesForBuild`
(`platform/orchestration/actions/get_pages_to_build_actions.go:88-165`) filters on
`site_id + status='active' + COALESCE(build_status,'planned') IN (...)`. No `rebuild_policy`
clause. Note a second branch the handoff did not mention: **`include_all: true` drops the
status filter entirely**, so that branch would sweep every `deployed` owned page too.

**2. Live consumers — TWO, both `include_all:false`** (nested walk over `agent_definitions`,
because a top-level `jsonb_each` under-reports steps nested in a loop sub_workflow — the
landmine noted at `save_page_sections_action.go:180`):

| agent | step | build_statuses |
|---|---|---|
| `page-rebuild` | `get_pages_to_rebuild` | `["needs_rebuild"]` |
| `pageflow-builder` | `get_pages_to_build` | `["planned","needs_rebuild"]` |

So `pageflow-builder` is exposed too, and it additionally selects `planned`.

**3. The commit-before-guard order — confirmed live, and it is THREE agents, not one.**
`assemble_page → deploy_page (action `git_commit`) → save_sections (action
`save_page_sections`) → update_page_status` is the order in `page-rebuild`,
`pageflow-builder` **and** `site-work-orchestrator` (the third selects from work items via
`load_work_items`, not from `get_pages_to_build`).

**4. Live exposure census.** 14 active pages are `rebuild_policy='owned'` AND
`build_status IN ('needs_rebuild','planned')`, across **6 domains** — the handoff named 2 on
1 site:

```
agent-complexity-estimator, password-entropy      ai-agent-orchestration.com  needs_rebuild
ai-agent-roi-estimator                            finetuning.uk               needs_rebuild
tool-drop-rate-simulator, tool-ehp-calculator,
tool-jump-physics, tool-lanchester-sim,
tool-progression-architect, tool-ttk-calculator   gamesdesign.co.uk           needs_rebuild
password-entropy                                  leopardessconsulting.co.uk  needs_rebuild
provocation (blog-post)                           vonc.com                    planned
tool-archetype-taster-quiz, tool-arena-interface,
tool-gauntlet                                     vonc.com                    needs_rebuild
```

`tool-arena-interface` and `tool-gauntlet` are on **vonc.com** — the same site as the "vonc
arena clobber" that motivated the ownership marker in the first place. A further **189** owned
pages sit at `deployed`: not selected today, but selected by the `include_all:true` branch.

### The finding that decided the design

`assemble_page` has **exactly three live consumers — the same three agents above — and in all
three its `next_step` is `deploy_page`.** The *sanctioned* owned-page deploy paths do **not**
use it: `page-rerender` goes through `rerender_single_page`, `section-editor` through
`apply_section_edit`. Both of those still `git_commit`.

That is what rules out the tempting fix. A guard inside `git_commit` would be the widest net,
and it would **break the only paths by which owned pages legitimately deploy** — which
migration 164 says in terms: *"page_rerender / assemble (re-assembly of EXISTING
page_components) is deliberately NOT gated — it is how owned pages deploy."* `assemble_page`,
by contrast, means precisely "generic composition of freshly generated content, about to be
committed". It is the seam.

**Verified the skip protocol reaches the commit.** All three `assemble_page` steps declare
`output_field: "assembled_page"`, and `checkUpstreamSkipped` (`git_deployer_actions.go:576-588`)
reads `collectedData["assembled_page"].skipped` as its first branch. `AssemblePageAction`
already returns `{"html":"", "skipped":true, "skip_reason":…}` in two existing cases
(`multipage_actions.go:38-62`). So a refusal expressed as a skip needs **no change to
`git_commit` and no config change on any agent** — existing machinery, not new.

### Prior art that shapes the fix rather than duplicating it

- **Migration 164** (`docs/agent_docs/sql_for_agents/164_pages_rebuild_policy.sql`) — created
  the column as Experience Loop guard rail 1, mechanising TL-001. It names exactly two Go
  refusals (reconcile emits `owned_page_review`; `save_page_sections` hard-refuses) because it
  assumed the only route into a generic build was `reconcile → needs_page`. **A path that
  selects straight off `pages.build_status` was never in its model — that is the gap 208 is.**
- **`reconcile_site_plan_action.go:232-270`** — the framework's existing answer when an owned
  page reaches a generic builder: exclude it and emit a `site_work_items` row of
  `item_type='owned_page_review'`, `status='needs_human_review'`, deduped by
  `item_key='owned_page_review:'+name` with `ON CONFLICT DO NOTHING`.
- **`features_open/012`** (LIVE v1.0.1149, council-approved) — establishes the precedent that
  regenerating an existing page's composition requires **explicit named intent**
  (`recompose_pages`). An opt-in override on selection is the same philosophy one layer down,
  not a new invention.
- **`features_open/021`** (LIVE and proven **today**) — the operator bulk-rebuild entry point
  whose first real dispatch surfaced this. **Owned by another workstream**, so it is a
  consumer to be told, not just measured.

### Open question logged before it is answered

`[UNMEASURED]` Whether excluding an owned page at selection leaves it stuck at
`needs_rebuild` for ever and re-selected on every subsequent run — i.e. what
`update_page_status` does for a page the loop skipped. Being answered before the design is
fixed, not after.
