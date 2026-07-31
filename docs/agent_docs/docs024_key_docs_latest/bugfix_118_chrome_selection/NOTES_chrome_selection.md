# NOTES — bugfix 118, chrome component selection

Append-only, newest at the bottom.

---

## 2026-07-31 — picking the bug, and why 118

Swept `bugs_open/` against the live `.jsonl` transcripts of every session active
in the last five hours (`grep -oh 'bugs_open/[0-9]\{3\}' | sort | uniq -c`),
because `scripts/who-owns.py` reads COMMITS and is blind to a session mid-fix.
118 came back with 4 mentions fleet-wide and no owning workstream; the one
session with a real cluster of them (`15765e72`) turned out to be the
`analytics_gtm` lane, whose 12:39 bulk edit to seven chrome components is GTM
injection into templates — not selection. Last commit naming 118: 2026-07-27.

Checked validity before starting rather than after: `render_site_components_action.go:556`
still carries the bare `WHERE function = $1 ORDER BY name LIMIT 1`, and the live
library still has three deactivated footers sorting ahead of two active ones.

## 2026-07-31 — the measurement that changed the fix, three times

**First shape I had in mind was the bug file's candidate 1: add `AND is_active`.**
Three measurements killed it, in order.

1. **`ORDER BY name` under `AND is_active` picks `header-leopardess`** — an
   ACTIVE FORK of leopardessconsulting.co.uk's header. A fork carries its
   parent's `function`, so it sits in the generic pool. The bug file only
   considered the footer's tie-break (`footer-theme-chrome` vs `site-footer`)
   and the header's answer is worse than the disease. `GetComponentByFunction`'s
   own doc comment has said "forks should only be accessed by component_id"
   since it was written; two of the three call sites never honoured it. A doc
   comment enforces nothing.
2. **`is_active AND forked_from IS NULL` still picks `site-header`, which is
   `component_level='section'`** — a 6.6KB page-section component, not chrome.
   016b §9 recorded the principle ("`site-head` (`component_level=section`) is
   unreachable as chrome") long before anything encoded it in a predicate. With
   the level filter added there is exactly ONE eligible row per chrome function,
   so the tie-break question the bug file raised disappears instead of being
   answered by the alphabet twice.
3. **The fallback only fires on a slot with NO `site_components` row.** All 14
   real sites have all three rows. So the filed claim that candidate 1 "changes
   the rendered footer on every site" — the reason it had been parked for an
   owner call for four days — is wrong about the code path. It changes what an
   UNASSIGNED slot gets. Live blast radius today: `loancalculator.co.uk`,
   created 2026-07-30, zero chrome rows.

That third one is the useful correction, because it turns a blocked bug into a
shippable one. What genuinely needs an owner call is repointing the 11 sites
already pinned to `footer-4-column` — which candidate 1 never proposed to do.
**The bug file collapsed "fix the selection" and "repair the fleet" into one
decision, and only the second is fleet-visible.**

## 2026-07-31 — what I found that the bug file did not

`site_work_items` has been carrying `deactivated_component` items — *"Site
component footer points to deactivated component 'footer-4-column'"* — since
**2026-07-17**, two of them stamped `[unresolved after 2 attempts]`. So the
platform detects this state perfectly well. Their `HandlerAgent` is
`rerender-pages`, which re-renders **the component the row already points at**:
the deactivated one. The routed repair is structurally incapable of repairing the
finding, which is why they age out rather than close.

I did not fix that here. It is a different defect (a handler whose contract it
cannot satisfy) and fixing it means repointing assignments, which is the
fleet-visible half. Recorded in the bug file and in `LANDMINES.md`, because a
`complete` deactivated_component item reads exactly like a repaired slot.

## 2026-07-31 — missteps

- **I nearly shipped `ORDER BY name` on `GetComponentByFunction` as "just
  determinism".** It is not: the ordered query and the unordered one could have
  disagreed, and if they had, every page BUILD's chrome would have changed as a
  side effect of a tidy-up. I only found out by running both (RUNBOOK R3). They
  agree — for exactly two functions, which is the entire population that has a
  choice. **An `ORDER BY` added to an existing `LIMIT 1` is a behaviour change
  until you have measured that it is not.**
- **My commit message names three "same-file passengers" that were not in my
  commit.** I checked `git diff` on `LANDMINES.md` and `000_concept_index.md`,
  saw other sessions' uncommitted entries, and wrote a paragraph naming them as
  riding along. Between that check and the commit, the 137 lane committed both
  files (`f0a52f42b`) — carrying MY lines as THEIR passenger — so my pathspec
  matched two clean files and silently took neither. The paragraph is wrong in
  the record. Forward-only, so it stays; corrected here and in `WRONG_CALLS.md`.
  The cheap check I skipped: re-run `git status --porcelain <paths>` in the same
  command as the commit, not two tool calls earlier. On this tree a two-minute-old
  `git status` is a guess.
- **`sites` has no `deleted_at`.** I wrote `WHERE deleted_at IS NULL` from habit
  and got a hard error. Cheap and self-correcting, but it is the third time this
  repo's schema has punished assuming a soft-delete column; `\d sites` first.
- **A 0-byte `go.mod` is a full disk, not a broken repo.** `git archive HEAD |
  tar -x` into `/tmp` produced `go: error reading go.mod: missing module
  declaration` — a message that reads like the module is broken. `/tmp` is a 16G
  tmpfs at 94%. Extract somewhere else and `wc -c go.mod` immediately.

## 2026-07-31 — what shipped

`b052249d8` (+ `a77034379` gofmt). One predicate (`chromeEligibleSQL`), one
slot→function map (`ChromeSlotFunction`), one resolver (`ResolveChromeComponent`)
returning `(component, eligible, error)`; the two ASSIGNMENT call sites routed
through it; `GetComponentByFunction` given `ORDER BY name` and nothing else.
Registered as **CLC-013** in the same commit, per the ordering-exemption
condition that still stands. Council submission
`5bc232d6-590a-4476-a6b1-4fb6f61751c6`, submitted before the commit, trailer
`Council-Submitted:` (never `Council-Reviewed:` on an unread verdict).

Verified against a clean `git archive HEAD` tree because the working tree carried
another lane's mid-edit compile error — `cmd/agent-chassis` builds, package tests
green, and the two ordering tests proven non-vacuous by deleting both `ORDER BY`
clauses and watching them go red.

**Inert until a chassis image rolls.** The bug stays OPEN until then; a fix
committed but not live leaves the defect reproducible, which is the standing bar.
