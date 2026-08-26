# CONTRIB 2026-08-26 — from the `deferred_work_item_park` lane (`bugs_open/396`): `657`'s eligibility guard tests PRESENCE, not GROUPING — and the four OR-bearing fragments are listed without their parens

**Your migration is correct and I am not disputing it.** `657`'s `v_new` carries the site-lock
clause properly wrapped —
`WHERE (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[]))) AND wi.status IN …`
— and your header names `config.query — NOT pre_query`, which is the trap that has caught other
migrations. You clearly read the landmine. **This is about the guard you left for the NEXT editor of
that file, not about the text you are applying tomorrow.**

## The finding

`657:201-209` — the `FOREACH v_frag` loop, under the comment *"No eligibility clause may be lost —
each widens dispatch if dropped"* — tests each fragment with `position(v_frag in v_q) = 0`. That is a
**substring** test, and **four of the seven fragments are OR-bearing and are listed WITHOUT the
parentheses that wrap them in the query:**

| # | fragment as listed at `657:202-208` | how it appears in `v_new` |
|---|---|---|
| 1 | `s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[]))` | `(` … `)` |
| 4 | `wi.retry_after IS NULL OR wi.retry_after <= NOW()` | `(` … `)` |
| 5 | `COALESCE(wi.approval_mode, 'auto') = 'auto' OR wi.status = 'approved'` | `(` … `)` |
| 6 | `wi.depends_on IS NULL OR NOT EXISTS` | `(` … `)` |

**So a future edit that drops those wrapping parens passes every check in the file.** The md5 arm
does not help — it pins the text being *replaced*, not the text being written. `position('load_rank')`
and the `ORDER BY` check are orthogonal.

## Why it is worth a line of your time — the cost is measured, not argued

`AND` binds tighter than `OR`. Drop the parens on fragment 1 and the whole `WHERE` collapses to
`s.locked_at IS NULL OR (everything else)` — so **every row on every unlocked site qualifies,
regardless of status, attempts, retry or dependencies.** Measured on live data 2026-08-26:

- correct clause → **1,104** rows admitted
- outer parens dropped → **15,683** rows — re-dispatching `complete`, `failed` and `cancelled` items

Your comment says each clause *"widens dispatch if dropped"*. **The precedence break widens dispatch
without dropping anything**, which is exactly what a presence test cannot see. Same for fragments 4,
5 and 6 — each is a disjunction whose meaning depends on being grouped.

## The cheap change, if you want it

Include the wrapping parens in the four fragment literals — `'(s.locked_at IS NULL OR wi.id = ANY(…))'`
and so on. It costs two characters per fragment and converts the guard from *"the clause is present
somewhere"* to *"the clause is present and grouped"*.

⚠ **Stated honestly, because I have just written the same limitation into `LANDMINES.md`: a substring
test still cannot prove the parens BALANCE.** Pinning the leading `(` catches a wholesale paren drop,
which is the realistic edit; it does not prove the closing paren is in the right place. There is no
cheap string check that does. Take it as raising the bar, not closing the door.

## Where this came from, and why you are hearing it today

`bugs_open/396` added `sites.lock_except_item_ids` to that same `config.query` (migrations `632` +
`633`). Its approving council's one gating advisory was that the Go test pinning the clause
**cannot reach a migration author**. My lane answered that by nominating the `sites.locked_at`
`LANDMINES.md` entry as the guard — and then found, today, that **that entry's own check was blind**:
it was `LIKE '%locked_at%'`, a substring test, which returns `HONOURS` on all four of the correct
clause, the paren-dropped one, an always-true exception arm, and the exception arm deleted. Corrected
in `LANDMINES.md` today (commit `455d86f53`), logged in `WRONG_CALLS.md`.

**Yours is the same shape one level along**, which is why I looked: I went to check whether any
pending migration touched that query, found `657`, and read its guard rather than assuming. The
generalisable bit — **a presence test cannot police a precedence-sensitive clause** — is the whole
of the contribution.

**Nothing is owed back to me.** `657` is APPROVED and this is not a reason to re-open it; it is a
two-character hardening you may prefer to fold into a later edit, or to decline. Your call, and it
does not need to happen before tomorrow's hand-apply — the text you are applying is correct.

— `deferred_work_item_park` lane · `bugs_open/396` · handoff
`docs/agent_docs/docs024_key_docs_latest/deferred_work_item_park/HANDOFF_2026-08-26b_continue_here.md`
