# 452 — `page_names` is declared in `request_render_audit`'s input spec and read by nothing

**Filed 2026-09-03** by the `bugfix_394_render_audit_rotation_cursor` lane, on closing that bug.
Spun out rather than fixed inside it: 394 was about coverage, and this is a dead config key. Named
as follow-up 4 in that lane's PLAN and in `bugs_closed/394`'s close-out.

**Status: OPEN, unowned. Severity: LOW — nothing is broken today.** No agent configures the key, so
no promise is currently being broken. What it costs is a *false affordance* and a budget slot.

## The defect

`RequestRenderAuditInputSpec` (`platform/orchestration/actions/request_render_audit_action.go:61`)
declares:

```go
Optional: []string{"site_id_field", "domain_field", "max_pages", "page_names", "topic", "capture_renders", "rotate_coverage"},
```

`[MEASURED 2026-09-03]` `page_names` occurs **exactly once** in that file — the spec line above.
The action body never reads it:

```
grep -c "page_names" platform/orchestration/actions/request_render_audit_action.go   ->  1
```

So a step config that sets `page_names` is accepted, ignored, and audits the whole site anyway.

## Why it matters, stated at the strength the evidence supports

1. **It advertises a capability that does not exist.** The spec is what an author reads before
   writing a step. `bugs_closed/394` §6 wanted exactly this key for `design-critique-agent` — "give
   the vision critique a *curated* 8 pages rather than the nav-order prefix" — and the natural move
   is to set `page_names` and expect it to work. It would silently audit 8 arbitrary pages instead.
   That is the shape MEMORY files under *a dead config key looks like a live one*.
2. **It spends an optional-key budget slot.** `[MEASURED 2026-09-03]` the action declares **7**
   optional keys against the ruled budget of **N = 10** (WFA-013, owner ruling 2026-08-14). One of
   those seven is dead, so the real affordance is 6 and the budget is 10% tighter than it reads.
3. **It is NOT currently breaking anything.** `[MEASURED 2026-09-03]` neither live carrier sets it:
   ```sql
   SELECT type, s.value->'config'->>'page_names'
   FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
   WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
     AND s.value->>'action'='request_render_audit';
   -- design-critique-agent | (null)
   -- render-audit-agent    | (null)
   ```
   **Say this plainly rather than inflating it:** there is no live damage, and a fix is cheap
   precisely because nothing depends on the current behaviour.

## Fix candidates, ordered by what closes the door

1. **Delete the declaration.** One line. Makes the false affordance unrepresentable, frees a budget
   slot, and is correct unless someone wants the capability. ⚠ Check first that nothing outside this
   repo's Go — a seed, a doc, an agent's `input_contract` — advertises it, or deleting it turns a
   dead key into a broken promise somewhere else.
2. **Implement it**: when `page_names` is set, resolve those `pages.name` values to URLs and audit
   exactly those, bypassing the cap and the cursor. Only worth doing if the answer to
   `bugs_closed/394`'s open question is "curate the design critique's sample". ⚠ It would need an
   explicit interaction with `rotate_coverage` — a named page set and a rotating window are two
   different selection policies and one must win, stated rather than implied.
3. Leave it and document it. Weakest: the next author still reads a spec that lies.

> **⚠ THE PRODUCT QUESTION IS ALREADY ANSWERED — OWNER RULING 2026-09-03.** Candidates 1 and 2 are
> mutually exclusive, and the only live reason to IMPLEMENT the key was to curate
> `design-critique-agent`'s 8-page sample. The owner ruled *"leave it"* on that: the critic keeps
> its nav-order prefix and stays a taste instrument rather than a coverage one
> (`bugs_closed/394`, ruling A; the operative control is that agent's entry in
> `render_truncation_acks.json`).
>
> **So candidate 1 — delete the declaration — is the LIVE one, and candidate 2 is PARKED.** Do not
> re-derive this choice; if you think it should be reopened, the thing that would reopen it is
> ruling A's own stated precondition: `design-critique-agent` gaining a cadence.

## How to verify a fix

- For candidate 1: `grep -rn "page_names" --include='*.go' --include='*.sql' --include='*.json' .`
  returns nothing that expects the key to work, and
  `./scripts/audit-optional-key-budget.sh --json` shows `request_render_audit` at **6**.
  ⚠ The optional-key-budget CronJob carries a hand-kept literal
  (`deployments/kustomize/services/optional-key-budget-check/base/check.py`) — **move it in the same
  commit** and re-apply the overlay, or the cluster keeps the old count. The parity test
  `go test ./cmd/config-key-audit/ -run BudgetCron` catches the repo half; nothing catches the
  cluster half but applying it.
- For candidate 2: a step configured with two page names audits exactly those two, and the
  truncation row is either absent or states the named-set policy — not a coverage window it did not
  use.

## Filing basis

No `090` run and no substitution claimed, because **no mechanism or root cause is asserted**: this
is a one-line reading of one file plus two queries, all reproducible above. If a fixing thread wants
to widen it to "how should this estate audit declared-but-unread config keys as a class", that is a
`090` or an RFC, not this file.

## Related

`bugs_closed/394` (where it was found and named as follow-up 4) · register **WFA-013** (the optional-key
budget it spends a slot of) · MEMORY *a dead config key looks like a live one* ·
`bugs_open/231` (a static config value for a spec-defaulted field is dead — the same family, one
field over).
