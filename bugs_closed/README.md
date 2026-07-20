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
next thread to hit it needs to find the case file. `008` (stop_reason) and `017`
(unregistered action) are in that state and deliberately remain open.

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
| `016` | `ssh` ignores `$HOME`, uses the passwd entry | council revise prompts drop reviewer output |
| `017` | — | static cutover orphans entry forms **and** unregistered action marked complete (two files, both `017`) |

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
| 016 | `ssh` ignores `$HOME` and expands `~` from the passwd entry | fixed in the box scripts |

Closure evidence lives inside each case file. `005` is the only one
independently re-verified against the live system by the thread that moved it;
the other three rest on their filing thread's own verification.

## Still the rules

- **Grep BOTH directories before filing a new bug.** The point of keeping closed
  cases in the repo is that a recurrence is recognisable — several members of the
  truncation family (`005`/`008`/`012`) were found separately by four threads
  because nobody grepped first.
- **A case here can reopen.** If the mechanism recurs, move the file back to
  `/bugs_open/` rather than filing a new number, and say what recurred.
- §10 of `docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md`
  remains the index of record for both directories.
