# PLAN — bugs_open/208: an owned page is committed over before the guard refuses it

**Status:** **SHIPPED IN CODE, `cb7b4d759`, inert until the chassis rolls.** Council submitted,
corr `5d1dcb10-7929-431e-b9e5-496992ce3229`. Registered PBP-036 in the same commit.
This file records the design **and its reasons**, including the options rejected — see the
amendments block below for what the review changed, which is the part worth reading.

> **AMENDED 2026-08-06, after the second model's review — three corrections, kept visible rather
> than edited away.**
>
> 1. **Layer 2 is THREE touch points, not one.** A skip at `assemble_page` alone does not survive
>    the iteration: `save_page_sections` has no upstream-skip check, so it reaches its own
>    ownership refusal and hard-errors, and `update_page_status` then stamps the page. Damage
>    prevented, run still lost. Both now honour the skip.
> 2. **My open question in (e) was answered the OPPOSITE way to my assumption, and it was the
>    dangerous direction.** I had assumed a skipped owned page would sit at `needs_rebuild`
>    (untidy, harmless). In fact `UpdatePageStatusAction` stamps it **`deployed`** — both existing
>    deploy guards pass for an owned page (it has components; it has no planned sections) — and the
>    same statement writes `built_from_plan_version`, which makes reconcile's `decideEmit` return
>    `skip_built` and **permanently suppresses the `owned_page_review` item that is this design's
>    own visibility channel.** The "harmless" option would have disabled the alarm.
> 3. **The visibility arm is UNHELD and shipped.** I had parked it citing the 114-junk-items
>    incident. That objection does not transfer: `bugs_open/204` filed `needs_new_component` per
>    unresolvable slot with **no stable key and an unbounded population**, whereas
>    `owned_page_review` has one deterministic key per page arbitrated by the `(site_id, item_key)`
>    partial unique index over open statuses. Different shape. Emitting.
>
> **And one correction in the other direction — the review's model of the third loop was wrong.**
> Its plan assumed `site-work-orchestrator` is fed `needs_page` items. Measured: **all 158
> `needs_page` rows fleet-wide route to `page-build-handler`**, which saves before it deploys and
> is safe. That agent consumes `handler_agent='page-content-writer'` items —
> `literal_markdown`/`placeholder_contact` — of which **11 of 14 targeted owned pages on
> 2026-08-04, all failed.** The door is real and observed; it is just a different door, and it
> changed which collected-data shape the guard must resolve.
**Bug:** `bugs_open/208_HANDOFF_2026-08-06_page_rebuild_commits_a_regenerated_owned_page_before_the_guard_refuses_it.md`

## The defect in one line

A `rebuild_policy='owned'` page (a tool/widget page, or a runtime-fill shell) enters a **generic**
page-composition loop, is recomposed by an LLM, **git-committed over the live tool**, and only then
refused by the ownership guard in `save_page_sections` — which protects the database row, not the
file the site serves.

## What the evidence forces, before any design choice

| # | Established | Where |
|---|---|---|
| 1 | Selection has no ownership clause, and an `include_all:true` branch with no status filter either | `get_pages_to_build_actions.go:88-165` |
| 2 | **Three** live agents run `assemble_page → deploy_page (git_commit) → save_sections`, not one | live `agent_definitions`, nested walk (RUNBOOK R1/R2) |
| 3 | `assemble_page` has **exactly those three consumers and no others**; the sanctioned owned-page paths (`page-rerender`, `section-editor`) do **not** use it | live `agent_definitions` |
| 4 | The sanctioned path `page-rerender` already runs **save before commit** | live `agent_definitions` |
| 5 | A refusal must not be an error: `continue_on_error` is unset on all four build loops, so a failing iteration **kills the whole run** | `loop_error_handler.go:66-80` + live config |
| 6 | The skip protocol already exists and `git_commit` already honours it | `multipage_actions.go:38-62`, `git_deployer_actions.go:576-588`, `output_field: assembled_page` on all three |
| 7 | Exposure: 14 owned pages at `needs_rebuild`/`planned` over 6 domains; 13 currently serve working tools; the 14th is a 404 with 0 components | `BASELINE_2026-08-06_owned_pages_served.txt` |
| 8 | Design intent named only two refusals (reconcile, save) because it assumed the only route in was `reconcile → needs_page` | `sql_for_agents/164_pages_rebuild_policy.sql` |

## The design

Two layers, because they answer two different questions. Neither is a substitute for the other.

### Layer 1 — an owned page never ENTERS a generic build loop (`get_pages_to_build`)

Add ownership exclusion to `queryPagesForBuild`, **in both branches** (the status-filtered one and
`include_all`), with an **opt-in config field whose unsafe value is the non-default**:

```
include_owned: false   (default — owned pages are excluded)
```

Why a field and not a documented contract: owner ruling 2026-08-02 §2 — *"when a seam's widest
branch is licensed by 'callers must all be X', make X a field with the unsafe default OFF … a
comment is not a control on a tree this many sessions share."* The widest branch here is "sweep
owned pages into a generic rebuild", so that is the branch that must be asked for by name.

Why exclusion is the right default, measured rather than argued: the only owned page this removes
from a live selection set is `vonc.com/blog/provocation`, which **404s and has zero
`page_components`** today — the page migration 164 deliberately parked. So the cost to the live
fleet is nil and the benefit is 13 working tools. (Evidence 7. This is the no-op case checked, not
just the damage case.)

Why it is not merely an optimisation: it is what stops the LLM spend and, given evidence 5, what
stops one owned page aborting an operator's whole batch.

### Layer 2 — an owned page cannot be COMMITTED even if it enters (`assemble_page`)

`AssemblePageAction` resolves the page and, if it is `owned`, returns the **existing** skip shape:

```go
return map[string]interface{}{"html": "", "skipped": true,
    "skip_reason": "page is rebuild_policy=owned …", …}, nil
```

Why this seam and not a wider one: evidence 3. `assemble_page` means precisely "generic
composition of freshly generated content, about to be committed" — its `next_step` is `deploy_page`
in all three consumers, and nothing else uses it. A guard in `git_commit` would be the widest net
and would **break the paths migration 164 says must stay ungated** ("page_rerender / assemble …
is how owned pages deploy"). So the wider net is not the safer choice here; it is the wrong one.

Why a skip and not an error: evidence 5 and 6. The skip is already understood by `git_commit`,
needs **no config change on any agent**, and lets the remaining pages rebuild.

Why it must exist even with Layer 1: Layer 1 cannot reach `site-work-orchestrator`, which selects
from work items, not from `get_pages_to_build`. Layer 2 also covers any future pipeline that
composes-then-commits — which is the recurrence this bug is an instance of (`TL-001`, and the
`bugs_closed/143` asset class: *"regenerates an artefact from a source AND upserts the row
describing it"*).

### Ranked as OPTIONAL, not part of the safety fix

3. **`save_page_sections` honours an ownership skip.** After Layer 2 skips the assembly, the save
   step still reaches its own owned refusal and hard-errors, so the batch still dies (evidence 5) —
   damage prevented, run lost. The narrow version: return the skip shape **only when the assembly
   for this item was skipped**, leaving the loud hard error intact for an owned page that arrives
   *with* content (old image, or a future path that bypasses `assemble_page`). That keeps the
   guard's teeth exactly where they are load-bearing. Listed separately because it is a third file
   and the fix is safe without it.
4. **Visibility.** An excluded page stays at `needs_rebuild` and is silently re-excluded for ever,
   so an operator's explicit request does nothing and says nothing. The framework already has the
   surface for this — `reconcile_site_plan_action.go:232-270` emits
   `item_type='owned_page_review'`, `status='needs_human_review'`, deduped on
   `item_key='owned_page_review:'+name` with `ON CONFLICT DO NOTHING`. Reuse it; do not invent an
   item type. **Held pending a judgement on spam risk** — a pre-fix build attempt on this estate
   once filed 114 junk items per site, so an emission on a hot selection path needs the dedup key
   to be proven, not assumed.

## Options rejected, and why (so nobody re-proposes them)

- **Guard inside `git_commit`.** Widest net, breaks `page-rerender`/`section-editor` — the only
  ways an owned page legitimately deploys. Evidence 3 + migration 164.
- **Reorder `save_sections` before `deploy_page`.** Changes commit/save ordering for *every* page
  on three pipelines, and `deploy_page`'s output (`page_deployed.commit_sha`) is consumed
  downstream by `update_page_status`. A config reordering also cannot be unit-tested and is
  invisible to `go build`. (Note it would move the generic loops onto `page-rerender`'s ordering —
  evidence 4 — so it is *directionally* right and still the wrong instrument.)
- **Operational stopgap: take the 14 pages out of `needs_rebuild`.** Leaves the trap armed for the
  next site and the next operator, and the bug is precisely that the queue state is trusted.

## Verification (never by rebuilding a real tool page)

1. Unit, sqlmock, following `save_sections_stored_slot_identity_test.go:324-327`. **Mutation-prove
   each guard**: a mock's own bookkeeping cannot assert a negative, so break the guard and require
   a named test to fail.
2. Query-level: run the fixed selection against a site with an owned page at `needs_rebuild` and
   assert the owned page is **absent by identity**, not by a count that dropped.
3. Live: dispatch on a **generic-only** `needs_rebuild` set, then re-run the baseline sweep — all
   13 bodies must be byte-identical (sha256 in `BASELINE_2026-08-06_owned_pages_served.txt`), and
   a generic page in the same run **must** have changed (negative control: a fix that excludes
   everything passes without one).

## Process obligations this change carries

- **Council gate** before/alongside the commit (`platform/` scope). Submit with
  `Council-Submitted: <corr>` if the verdict has not landed; never write `Council-Reviewed:` on a
  verdict not read.
- **Concept register in the SAME commit** (owner ruling 2026-07-29, condition 2 — the one surviving
  condition): the new `include_owned` field and the `assemble_page` refusal are a shared seam.
  Next free id is **PBP-036** in `page-build-pipeline.md` (clean file; `rebuild-cascade.md` and the
  index are dirty from another session — same-file passenger risk).
- **Tell the other consumers, don't just measure them** (owner ruling 2026-07-29 §3):
  `features_open/021`'s operator bulk rebuild went LIVE today and is owned by another workstream;
  after this change an owned page it explicitly names is refused rather than rebuilt. That is
  their guarantee changing.
