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
