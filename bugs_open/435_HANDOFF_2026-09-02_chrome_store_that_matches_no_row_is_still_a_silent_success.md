# 435 — a chrome store that matches NO ROW still reports success: the last silent branch in the render path, left open on a measurement

**Filed 2026-09-02** by the `bugfix_423_chrome_utf8` lane, at the council's asking
(`dc62975f` round 2, `bug_historian` medium). **Not a fix deferred out of laziness — a fix
deferred on a number**, and the number is the reason this file exists rather than a comment.

## The branch

`platform/orchestration/actions/render_site_components_action.go`, the `RowsAffected() == 0`
arm (`~:1411` as of 2026-09-02, after 423's edits):

```go
logger.Error("Failed to store rendered component: no row matched", ...)
return false, false, degraded, nil     // ← nil: the caller never hears
```

`bugs_open/423` fixed the four sibling dispositions in this function and left this one.
After 423 the file has five `degraded, nil` returns: `:1446` is the success return, `:1407`
is a lock refusal (correct — a lock is not a failure and files its own item), `:1222` is
`bugs_open/342`'s refusal (which reports), `:1241` is an empty-render Warn. **This is the
last one that is a genuine swallow**: the render succeeded, nothing was stored, and the
step reports success.

The lock arm above it already returns for a locked row, so reaching here means the
`site_components` row for this site+slot **does not exist** (the writable predicate is
purely lock-based — `datahelpers.AgentWritableSQLFor`).

## Why 423 did NOT simply route it into `chrome_render_failed`

Because the blast radius was measured first, and it inverted the obvious move.

`[MEASURED 2026-09-02]` **57** sites exist; only **34** have any `site_components` row at
all; **23** are missing at least one of the three slots.

```sql
SELECT count(*) FROM (
  SELECT s.id FROM sites s LEFT JOIN site_components sc ON sc.site_id = s.id
   GROUP BY s.id HAVING count(DISTINCT sc.slot_name) < 3) t;   -- 23
```

So filing a `needs_human_review` item here would mint **up to ~69** findings about sites
that have simply **never been built**. That is not a detector, it is a queue flood, and it
would bury the real instances the same way silence does.

## The actual fix, stated so nobody re-derives it

**Distinguish "this site has no chrome yet" from "this slot's row vanished."** They are the
same zero rows and completely different events. Candidates, ordered by what closes the door:

1. **Ask before writing.** Select the row (or its absence) before the UPDATE and branch on
   it: absent-and-never-built is a normal greenfield state; absent-after-having-existed is
   a defect. `trg_site_component_archive` (`sql_for_agents/344`) already archives outgoing
   bytes on every differing overwrite, so **a slot with archive rows and no live row is
   exactly the vanished case** — that is the discriminator, and it needs no new state.
2. Make the row's existence a precondition of the render loop, so a missing row is handled
   once at the top rather than as a store outcome per slot.
3. Report unconditionally and filter consumers — rejected here: it moves the flood
   downstream rather than removing it.

## How to verify a fix

The census above must stay the control: after any fix, a run over the **23** chrome-less
sites must produce **zero** items, and a deliberately deleted row on a built site must
produce **exactly one**. A fix that cannot tell those two apart has not fixed this.

## Relations

- `bugs_open/423` — the parent; the four siblings fixed, this one measured and deferred.
  Register **STY-059** carries the same measurement as a landmine ("do not tidy it without
  re-running that census").
- `bugs_closed/054` (chrome unresolved-field escalation and consumer) — the estate's
  precedent that a named log is observability, not escalation. This branch is still on the
  wrong side of that contract.
- `bugs_open/034` (validation errors dropped with no durable record) — named by the council
  as the open bug whose shape this is. Worth checking whether one fix serves both.
