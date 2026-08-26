# NOTES — render audit rotation cursor (bugs_open/394)

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-26 ~16:20 — LANE CLAIMED, before any research

**This file exists first and is committed first, deliberately.** I lost `bugs_open/359`
an hour ago to a session that opened the same lane two minutes after me: we both ran
`who-owns.py`, we both ran the tree grep, and both instruments told both of us the bug was
unowned — because at the moment either of us looked, the other had written nothing. The
window is not a gap in the instruments; it is the interval they cannot see into. Full
account in `WRONG_CALLS.md`, 2026-08-26.

So: **claim in the commit log before doing the work**, not after. If you are reading this
and you also want 394, `git log --oneline -- <this directory>` will tell you who was first,
which is the thing neither of us could establish last time.

### Ownership as at 2026-08-26 16:20

| instrument | result |
|---|---|
| `scripts/who-owns.py 394` | `likely OWNING workstream(s): (none identified)` |
| commits touching the bug file | **one**, ever: `3cb6be421` (2026-08-25), the filing commit |
| lane directory named `bugfix_394*` | none existed before this one |
| cross-references | `bugfix_358_unread_finding_codes` cites it twice — the lane that FILED it, not one working it |

Its two siblings, filed in the same commit, ARE owned and I am not touching them:
`bugs_open/392` → `bugfix_392_link_context_unread`; `bugs_open/393` →
`bugfix_393_ungraded_completions`. 394 is the one of the three nobody took.

### What the bug asks for, in one line

`bugs_closed/242` made render-audit truncation **loud** — `pages_total`/`truncated` stamped
into the durable result plus an `agent_error_log` `RENDER_AUDIT_TRUNCATED` row. Nothing
reads it. Meanwhile the mitigation it shipped (raise the cap 25→60) has been outgrown:
webdesign.co.uk went 109 → 125 → 131 live pages in six days, and the writer's own message
says the unaudited tail is **the same pages every run**.

The owner commissioned a reader (ruling 2026-08-25, decision 4). The bug ranks three
candidates and candidate 1 — **persist a per-site cursor so the next run starts where the
cap cut off** — is the one that retires the signal's cause rather than reporting it. That
is the direction I intend to take, subject to what the evidence says next.

Next: re-validate the four `RENDER_AUDIT_TRUNCATED` rows against the live DB, and resolve
the `[UNEXPLAINED]` `5 of 26` row by reading the dispatching config rather than assuming
the cap is 60 everywhere.

---

## 2026-08-26 — the bug is VALID, and materially worse than filed

### The mechanism, from the code rather than from the message

`platform/orchestration/actions/request_render_audit_action.go` selects the site's live pages
with

```sql
ORDER BY COALESCE(nav_order, 999), name
```

and then, in the scan loop:

```go
total++
if len(urls) >= maxPages {
    continue // keep counting so the truncation is reportable
}
```

So it takes a deterministic **prefix** and counts the rest. The row it writes says *"the
unaudited tail is the SAME pages every run"* — that is not an inference by the author, it is
true by construction, and reading the loop is what establishes it.

### `[MEASURED 2026-08-26]` webdesign.co.uk: 146 live pages, 60 audited, 86 never

The bug quotes 60 of 131 on 2026-08-24. Two days later:

| live pages | audited | tail |
|---|---|---|
| **146** | 60 | **86** |

Fifteen pages added in two days, all of them into the tail.

### The finding that decides the fix: the tail is a whole CLASS of page, not a random 86

`nav_order` on webdesign.co.uk's active pages, `[MEASURED 2026-08-26]` — and it is never NULL
on this site, so `COALESCE(nav_order,999)` never fires here:

| nav_order | pages |
|---|---|
| 0, 10, 20, 30, 40, 90 | 1 each (6 nav pages) |
| **100** | **94** (the tools, then alphabetical by `name`) |
| **200** | **48** (all named `tool-*-guide`) |
| 201 | 1 |

A cap of 60 therefore covers the 6 nav pages plus the first **54 tools alphabetically**, and
cuts between:

```
rn 60  nav_order 100  tool-head-architect     <- last page ever audited
rn 61  nav_order 100  tool-html-minifier      <- first page never audited
```

**Every one of the 45 remaining `tool-*-guide` pages at `nav_order` 200 is structurally
invisible to the render audit and always has been.** No cap below 98 reaches them. That is the
argument against the bug's candidate 3 (raise the cap) stated as a measurement rather than as a
preference: a constant cannot chase a site that adds 15 pages in two days, and the specific
constant we would have to pick to reach the guides today is one the site would outgrow next
week.

### Two callers, and the cap is PER-DISPATCH, not per-agent

`[MEASURED 2026-08-26]` from `agent_definitions`, live rows only:

| agent | step | `max_pages` |
|---|---|---|
| `render-audit-agent` | `audit` | **60** |
| `design-critique-agent` | `audit` | **8** |

`design-critique-agent` was seeded yesterday (`sql_for_agents/645_design_critique_agent.sql`)
and **is already truncating from birth** — two rows today for leopardessconsulting.co.uk at
`8 of 37`, 14:22Z and 15:10Z.

### The bug's `[UNEXPLAINED]` "5 of 26" row — RESOLVED

The bug flagged the 2026-08-11 loancalculator row (`5 of 26`) as unexplained and warned against
assuming the cap is uniform. It was right to. The row's own context settles it:

```
2026-08-11 18:08:54Z | render-audit-agent | step audit
{"max_pages": 5, "pages_total": 26, "pages_audited": 5}
```

So: **the standing agent, running with a per-dispatch override of 5.** Not a different agent,
not a config regression in `agent_definitions`. The originating orchestration row has since
aged out of `orchestration_states` (rolling window), so where the override came from cannot now
be recovered — recorded as unrecoverable rather than left open.

**The conclusion that matters for the fix:** `max_pages` is a per-dispatch value, so any design
that reasons from "the cap is 60" is wrong.

### `[MEASURED 2026-08-26]` fleet exposure, at both live caps

25 sites have more than 8 live pages.

- **At cap 60** exactly ONE site truncates — webdesign.co.uk, tail 86.
- **At cap 8** — the design-critique caller — **25 sites truncate**, tails from 4
  (noted.co.uk) to 138 (webdesign.co.uk).

So this is two problems wearing one message: the render-audit caller has a **deep tail on one
site**, and the design-critique caller has a **shallow tail on the whole fleet**. A fix that
only serves the first leaves the second, which is now the larger population.

### The driver, for the record

`scheduled_tasks.site-render-audit-rotation` → `render-audit-agent`, `interval_seconds=3600`,
enabled; its `pre_query` picks ONE site whose `site_discovery_rotation.last_selected_at` for
that agent is older than **3 days**, stamps it, returns it. So a site is audited at most every
three days, and at that rate webdesign's 86-page tail would take a fortnight to cover even with
a perfect cursor — worth knowing before promising coverage in a given window.

⚠ `site_discovery_rotation` (`site_id`, `agent_type`, `last_selected_at`; PK on the first two)
is written **only from that SQL**, never from Go. A cursor written by the action would be the
first Go writer of that table — a fact any reuse-versus-new-table decision has to face rather
than assume away.
