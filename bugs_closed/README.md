# bugs_closed/ — cases that are fixed AND live

Split out of `/bugs_open/` on 2026-07-19. `/bugs_open/` had accumulated closed
cases, so its listing no longer answered the question it exists to answer:
**what is biting production right now.**

## The bar for moving a case here

A case moves here only when its fix is **fixed AND live in production** —
deployed and verified against the running system, not merely written or
committed.

**A fix that is committed but inert until the next image roll STAYS in
`/bugs_open/`.** That is the whole point of the bar: between commit and image
roll, the defect is still reproducible in prod, and that is exactly when the
next thread to hit it needs to find the case file.

> **UPDATED 2026-07-20 19:00 BST:** `008` and `017` were named here as examples
> of that inert state. Both have since shipped and been verified live, and both
> are now in this directory. The *rule* stands unchanged — it is the examples
> that expired. If you are looking for a case currently in that state, check
> `/bugs_open/` rather than trusting a list here; a named example goes stale
> within days, which is exactly what happened to this one.

> **UPDATED 2026-07-20:** `012` (improver truncation) was named here as staying
> open *because migration 169 was unapplied*. That condition is gone — 169 and
> its correction 170 are applied, the guard shipped in **v1.0.1139**, and the
> whole chain (component untouched · refusal logged · item to
> `needs_human_review` · note written) was **driven and verified against
> production**, not inferred from config. `012` has moved here. `008`'s
> `stop_reason` code did ship in the same image, but its own case file owns that
> verification, so it stays open until that thread confirms it.

"Superseded by a later case" also qualifies — see `004`, superseded by `005`.

## Numbering is preserved, and continuous across BOTH directories

Moved files keep their original number and filename. Numbering is a single
sequence shared by the two directories — **there is no renumbering, ever.**

So a stale pointer resolves trivially: if `bugs_open/NNN` is not there, look in
`bugs_closed/NNN`. Many older references also use the directory's former name,
`aaa_fails_to_mend/` — same rule, same number.

This was chosen deliberately over rewriting every reference: roughly 40 files
across docs, Go comments and SQL point at these paths, several of them owned by
concurrently-running threads whose working trees must not be touched. A stable
number that resolves in one of two adjacent directories is cheaper and safer
than chasing pointers through other sessions' work.

## ⚠️ Duplicate numbers exist — check the slug, not just the number

The numbering was assigned by concurrent threads and **collided twice**:

| number | this directory | still in `/bugs_open/` |
|---|---|---|
| `016` | **both** here now: `ssh` ignores `$HOME` (passwd entry); **and** council revise prompts drop reviewer output (closed 2026-07-21) | — |
| `017` | — | static cutover orphans entry forms **and** unregistered action marked complete (two files, both `017`) |
| `018`, `027`, `028`, `029` | — | two files each in `/bugs_open/` (concurrent-thread collisions of 2026-07-18/19) |
| `040` | — | kafka dial timeouts fleet-wide **and** failed page build leaves page deployed (two files, both `040`, both 2026-07-20) |
| `044` | — | no capability inventory / dormant agents undetectable (19:22) **and** plan_sections defers an empty-schema component by name heuristic (19:57) — two files, both `044`, both 2026-07-20 |

A bare reference to `bugs_open/016` or `bugs_open/017` in older docs or code
comments is therefore **ambiguous** — resolve it by the slug or the described
mechanism, never by the number alone. Do not "fix" this by renumbering; the
numbers are cited in commit messages and Go comments.

## Contents

| # | case | closed because |
|---|---|---|
| 004 | Landing an image can silently blank an article body | superseded by `005`, which found the real root cause |
| 005 | Article-body blanking — root cause is LLM truncation (`max_tokens`) | fix deployed v1.0.1126; re-verified live 2026-07-19 (19/19 healthy, config survived a re-seed, repair fn present in the running pod) |
| 014 | VM-site artefacts silently deploy to the default `sites` repo | both causes fixed (v1.0.1126 + pin removal) |
| 008 | `GenerateText` never decoded `stop_reason` — truncations and refusals surfaced as successes or as parse faults | all 5 items live in the image rolled 2026-07-20 18:58 BST; **re-verified in the running pod** |
| 012 | The improver truncates and destroys the component it is repairing | guard + migrations 168/169/170 live in v1.0.1139; chain **driven** against production (not inferred from config); **re-verified in the running pod under v1.0.1140** |
| 013 | fix-implementer commits un-`gofmt`'d LLM output, so the gate burns the whole run | `formatGeneratedGo` at commit-prep; **re-verified in the running pod** 2026-07-20 |
| 014 | VM-site artefacts silently deploy to the default `sites` repo | both causes fixed (v1.0.1126 + pin removal) |
| 016 | `ssh` ignores `$HOME` and expands `~` from the passwd entry | fixed in the box scripts |
| 017 | An unregistered action fails as "requires a topic" — and the failure is stamped `complete` | registration + `handlerReportedFailure` + a registry parity test, live in v1.0.1139 |
| 031 | A stale register entry asserts a content-hash rerender skip that never existed, blocking correct plans | corrected in all 6 places, including the live council seat prompts |
| 032 | The completion verifier reads a DELETED component as a successful fix | conservative floor (error, not verdict) shipped; **re-verified in the running pod** 2026-07-20 |

Note `002` is also filed here as a routing document rather than a fixed defect —
its A–C legs are done and it points at the owners of the rest. Read it before
assuming everything it describes is closed.

Closure evidence lives inside each case file. `005`, `008`, `012`, `013` and
`032` were independently verified against the live system by the thread that
moved them (`008`/`013`/`032` by grepping the running pod's binary on
2026-07-20); the others rest on their filing thread's own verification.

**`032` closed with a known residual, deliberately.** Its stronger fix — treat
absence as *deletion* when the page still expects the component — was left to the
owning thread. A case can close on its safe floor while a better answer stays
open; say so in the file, as `032` does, rather than letting the floor read as
the finished shape.

## Still the rules

- **Grep BOTH directories before filing a new bug.** The point of keeping closed
  cases in the repo is that a recurrence is recognisable — several members of the
  truncation family (`005`/`008`/`012`) were found separately by four threads
  because nobody grepped first.
- **A case here can reopen.** If the mechanism recurs, move the file back to
  `/bugs_open/` rather than filing a new number, and say what recurred.
- §10 of `docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md`
  remains the index of record for both directories.
