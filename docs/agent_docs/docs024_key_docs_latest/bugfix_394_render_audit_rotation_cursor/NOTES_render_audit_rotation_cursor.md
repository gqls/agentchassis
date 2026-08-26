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

---

## 2026-08-26 — the two answers that most change the fix's shape

### 1. The retraction path is ALREADY scoped to what was measured — so a cursor is safe there

This was the risk I most expected to have to design around: `write_render_audit_findings`
RETRACTS standing findings, and a cursor makes every run look at a *different* subset. If
retraction inferred "resolved" from "not seen in this run", a cursor would silently close every
finding on every page the cursor happened to skip — fleet-wide, on the first run.

It does not. `write_render_audit_findings_action.go:219-222`, in the payload struct's own words:

> `PagesAudited` names the pages the adapter **SUCCESSFULLY MEASURED**, and is the **entire
> scope of what this action may retract**. Absent on an old-shape reply, which is why
> retraction is inert rather than wrong against an un-rolled adapter: an empty audited set
> retracts nothing.

and at :731, `if len(payload.Summary.PagesAudited) == 0 { … }` degrades to no retraction. The
adapter populates it per successfully-navigated URL
(`internal/adapters/browserrunner/render_audit_action.go:422`).

**So the most dangerous interaction in this change is already closed, by someone else, for a
different reason.** Recording it as a checked fact rather than an assumption, because the
opposite would have been a fleet-wide silent close and it is exactly the kind of thing a plan
asserts without opening the file.

⚠ The corollary is a constraint, not a freedom: **the cursor must never advance past a page the
adapter did not actually measure**, or coverage and retraction scope drift apart — the audit
would claim a page as "covered" that was never rendered. `len(PagesAudited) < Pages` exactly
when a navigation failed, and the adapter's own comment at :232-235 says so.

### 2. There is NO durable per-page audited list today — only counts

The bug's acceptance for candidate 1 is *"the union of audited pages reaches all 146, verified
by the audit's own durable page list, not the status"*. That list does not exist to verify
against:

```
grep -rn "PagesAudited|pages_audited" --include=*.go platform/ internal/ | grep -v _test
```
`[MEASURED 2026-08-26]` every persisted use is a **count**:
`request_render_audit_action.go:179` writes `"pages_audited": len(urls)` into the
`agent_error_log` context; `write_render_audit_findings_action.go:615` writes
`result["pages_audited"] = payload.Summary.Pages` and :654
`result["retraction_scope_pages"] = len(payload.Summary.PagesAudited)`. The **URL slice itself
is used and discarded**.

And the step's own result is not durable anyway — the file's header records that an awaiting
step's result never survives the park (RFC_012 addendum 2, owner-ruled option B), which is why
the truncation row is written to `agent_error_log` *before* the dispatch in the first place.

**So the acceptance arm the bug asks for is not runnable today.** Whatever the fix is, it has to
make coverage reconstructable — and the cheapest honest version is probably to record the cursor
window (`cursor_from` / `cursor_to` / `pages_audited`) on the durable row that is already being
written, so consecutive runs can be replayed into a union. An acceptance test nobody can run is
the same as no acceptance test, so this is part of the change, not a follow-up.

### 3. The second caller is MANUAL, which changes what a cursor means for it

`sql_for_agents/645_design_critique_agent.sql` states it plainly: **"NO cadence — manual trigger
only, per the recorded owner decision. Nothing emits work for this agent; nothing promotes into
it; the only path is a hand dispatch."** So `design-critique-agent`'s cap of 8 is a *sample for
a vision critique*, not a coverage sweep — its 25-site truncation footprint is real but it is
not accumulating a debt on a schedule the way the render-audit rotation is.

That file also records the co-dedup design: this agent is deliberately a **second producer** on
`contrast_failure` / `undeployed_asset` keys beside the rotation, and the shared key namespace
is "the point". Worth holding onto — it means the two callers' findings already interleave by
design, so a per-caller cursor must not be justified by "keeping their findings apart". If the
callers need separate cursors it is because their *coverage questions* differ, which is a
different and weaker argument, and the plan should say which one it is relying on.

### 4. The optional-key budget has headroom — checked, because it gates the design

`scripts/audit-optional-key-budget.sh --json` `[MEASURED 2026-08-26]`:

```json
{ "action": "request_render_audit", "optional_keys": 6, "over_budget": false,
  "optional": ["site_id_field","domain_field","max_pages","page_names","topic","capture_renders"],
  "consumers": 2, "agents": ["design-critique-agent","render-audit-agent"] }
```

Budget is **N = 10** (owner ruling 2026-08-14, register WFA-013), and **no action in the fleet
is over budget today** (0 of the audited set). So this action can take one or two new optional
keys without tripping the accumulated-surface review. Checked rather than assumed, because the
budget check is exactly the sort of thing a plan asserts in passing and gets wrong — and because
RFC_022's narrowing only exempts an opt-in field from *architecture* scope; it does not exempt it
from the count.
