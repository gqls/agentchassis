# 094 — `049b_deploy_single_page.sh`'s `section_data_resolved` branch cannot work (CLOSED — fixed, LIVE on v1.0.1182, and driven end to end)

**Filed** 2026-07-26 from the oufe.com workstream.
**Severity** medium — no data loss, but it blocks the only supported route for
publishing a hand-authored copy edit, and it fails identically every time.
**Owner** `cta_link_integrity` (the script lives in its `scripts/` directory).
**Status** **CLOSED 2026-07-28 — fixed, council-APPROVED, LIVE on v1.0.1182, and the
branch has now actually been run.** The shared script needs no change.

> ## ✅ CLOSED — the previously-impossible branch executed
>
> **Live:** chassis v1.0.1182 (pod up 2026-07-28T09:55:02Z). Pod-grep of strings this
> change created: `"rerender_page_sections: page resolved"` → 1,
> `"need page_name or page_id"` → 1.
>
> **End-to-end, on the branch this file says cannot run.** Fired
> `049b_deploy_single_page.sh <page_id> <site_id> gaswholesalers.com section_data_resolved`
> against `tool-gas-unit-converter` (chosen because all its sections have non-NULL
> `content_data` — this file's own warning). Result:
>
> ```
> orchestration          complete | COMPLETED
> steps recorded         rerender_sections, check_rerender_mode
> rerender_sections      { "page_name": "tool-gas-unit-converter",   <- RESOLVED FROM page_id
>                          "skipped": false, "section_count": 1, "escalated": true }
> ```
>
> **`rerender_sections` appears at all**, which is the whole point: before this fix the
> step died at the input gate with `missing required fields: [page_name]` and never ran.
> And `page_name` in its own output is the name **derived from the page_id the envelope
> supplied** — the fix, observed rather than asserted.
>
> **Nothing was damaged**, checked rather than assumed: the component's `updated_at` is
> still `17:52:24` from the previous day's re-render, **0** work items were filed in the
> window, and the page serves 200 at 21,340 bytes with its meta description intact.
>
> ### One loose thread, deliberately not chased here
>
> The step returned **`escalated: true` while filing no work item and mutating nothing**.
> For this page that is plausible — its single section is the tool widget, whose
> `content_data` carries no writable prose keys — but "reported an escalation that left no
> trace" is the shape of `bugs_open/091`, and I have **not** established whether the
> escalation is genuinely a no-op here or a report of something that did not happen.
> `[UNVERIFIED]`. Worth a look by whoever owns `rerender_page_sections`; it is not a
> regression from this fix, which is why it did not hold the close.

> ## STATUS 2026-07-28 (bugs thread) — candidate 1, not 2 or 3
>
> **The action now accepts EITHER `page_name` OR `page_id` and derives the other**, so the
> envelope this file describes works as published and every future caller is fixed at
> once. `page_name` moved Required → Optional; `page_id` added Optional.
>
> **Both lookups stay scoped to `target_site_id`.** `page_id` is globally unique, so an
> unscoped lookup would have turned this into a way to re-render another site's page
> through an envelope naming this one. `TestRerenderPageSections_PageIDCannotReachPastTheSite`
> pins it.
>
> **The script needs no edit** — candidate 2 becomes unnecessary once the action resolves.
>
> **Council:** APPROVED (`0a31be23-d7f6-41c5-be2d-02373a472fae`, 10 reviewers, 0 unreadable,
> 4 advisory objections, **6 seats abstained** — a small panel).
>
> ### Two council questions answered by grep, not by argument
>
> - **Does any caller outside `agent_definitions` rely on the old refusal?** **No.** Every
>   `.sh`/`.py`/`.sql`/`.go`/`.json`/`.md` in the repo was searched. The one shell hit
>   (`bundle_minilobby_trim2.sh`) *names* the action while assembling a docs bundle and
>   never invokes it; everything else is a comment or a cross-reference.
> - **Was there an existing id-or-name page resolver to reuse?** **No.** Two `name → id`
>   helpers already exist (`saveSectionsLookupPageID`, `lookupPageID` — themselves a
>   duplication worth collapsing), and nothing resolves the other direction.
>
> ### The objection that was right, and what it bought
>
> `page_id` is mapped by **no** step config, so it arrives through
> `ExtractActionInputs` **Strategy 2**, whose own comment warns it *"uses aggressive
> recursive search that can find stale values"*. Before this change a missing `page_name`
> failed loudly at the input-spec gate; now a stale `page_id` could resolve a **different
> page of the same site**. Site scoping does not help there.
>
> The action now logs which key resolved the page and what it resolved to. **That does not
> prevent a wrong resolution — it makes it attributable rather than invisible.** If you are
> debugging an unexpected re-render, `resolved_by` is the field to read.
>
> ### Owed
>
> - A roll, then a pod-grep for a string this change created:
>   `strings /app/agent-chassis | grep -c "rerender_page_sections: page resolved"`.
> - **The end-to-end run has NOT happened**: nobody has driven
>   `049b_deploy_single_page.sh … section_data_resolved` through since the fix. Before
>   doing so, heed this file's own warning and check no section has NULL `content_data`,
>   or the page escalates to the content writer and authored copy is regenerated.
> - **Bigger, not fixed here:** Strategy 2's recursive search is a generic exposure for any
>   action with Optional fields and no explicit `input_fields`. That is shared input
>   machinery and should not ride along on a page-lookup fix.

## Symptom

```
step rerender_sections failed: failed to execute action rerender_page_sections:
input extraction failed: missing required fields: [page_name]
```
Reproduced twice, on two different pages of oufe.com, 2026-07-26.

## Cause

`docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049b_deploy_single_page.sh`
publishes:

```json
{"action":"orchestrate","config":{"agent_type":"page-rerender"},
 "input_data":{"page_id":"…","site_id":"…","domain":"…","spec":{"reason":"section_data_resolved"}}}
```

`rerender_page_sections` declares:

```go
// platform/orchestration/actions/rerender_page_sections_action.go:80
Required: []string{"target_site_id", "page_name"},
```

Nothing between the envelope and the action derives `page_name` from `page_id`,
so the required-field check fails before the action does any work.

## Why it survived this long

The script's **default** branch (no reason argument) is assemble-only and never
calls `rerender_page_sections` — it stitches stored `rendered_html`. That branch
works, and it is the one most callers use.

The `section_data_resolved` branch is only needed when you have edited
`content_data` directly and want the page re-rendered from it. That is exactly
the situation where the failure is most expensive, because the operator has just
authored copy and has every reason to believe the publish step is routine.

The script's own header documents the branch in detail, including its gotchas,
which makes the gap easy to miss: the documentation is right about what the
branch is *for* and silent on the fact that it cannot run.

## Fix candidates, ordered by what closes the door

1. **Resolve `page_name` inside the action** from `page_id` when it is absent —
   the action already has the DB handle and reads the page row. This makes the
   bad envelope unrepresentable rather than merely documented, and fixes every
   existing caller at once, including any not yet discovered.
2. Add `page_name` (and `target_site_id`) to the script's envelope. Smaller, but
   only fixes this one caller and leaves the next one to rediscover the trap.
3. Fail earlier and louder: have the script look the page up and refuse before
   publishing. Weakest — it moves the error, it does not remove it.

A working envelope, for reference, is in
`docs/agent_docs/docs024_key_docs_latest/oufe/TRIGGER_rerender_page.sh`, which
also refuses up front if any section has NULL `content_data` (that would escalate
the page to the content writer and silently regenerate authored copy).

## How to verify a fix

Edit one `content_data` field on a deployed page, run the script with
`section_data_resolved`, and confirm the orchestration reaches
`COMPLETED | complete` — **and** that the edited string appears in the live page.
The orchestration completing is not sufficient evidence on its own; see 095.
