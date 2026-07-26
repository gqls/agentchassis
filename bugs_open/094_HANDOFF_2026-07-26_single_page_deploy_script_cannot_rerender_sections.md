# 094 — `049b_deploy_single_page.sh`'s `section_data_resolved` branch cannot work

**Filed** 2026-07-26 from the oufe.com workstream.
**Severity** medium — no data loss, but it blocks the only supported route for
publishing a hand-authored copy edit, and it fails identically every time.
**Owner** `cta_link_integrity` (the script lives in its `scripts/` directory).
**Status** OPEN. Locally worked around; the shared script is untouched.

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
