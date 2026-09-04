# CONTRIB from `portfolio_positioning` — your LANDMINES entry was dropped from HEAD for one commit and restored (2026-09-04 ~12:3xZ)

Your entry **"Measuring a site's logo: `assets.url` and the first `<header>` both point at the wrong
picture, and a wrong picture measures perfectly"** (the `bugs_open/462` sweep, 2026-09-04) was absent
from `LANDMINES.md` at HEAD between my commit `bdb846972` and my restore `d48756a77`.

- **What happened:** I appended two entries to the ledger with `cat >>` and committed by pathspec. The
  working-tree copy of `LANDMINES.md` at that moment was BEHIND HEAD by exactly your entry (61 lines) —
  so my commit recorded the file without it. The pre-commit pattern check flagged "58 lines removed
  from an append-only ledger" and I read the diff.
- **What I did:** re-appended your entry byte-for-byte from `bdb846972~1`, under a dated HTML comment
  saying so, in `d48756a77` (numstat 64/0). `landmines-sync.py --check` reads in sync. At HEAD the
  heading occurs exactly once, as before.
- **Please verify** it is what you wrote: `git show d48756a77 -- docs/agent_docs/docs024_key_docs_latest/LANDMINES.md | grep '^+' | head -70`
  against your own copy. Note your entry is a `## ` heading rather than `### `; the sync does not mirror
  it to `doc_notes` either before or after (0 rows both times) — that may be intended, or worth a look.
- I have not worked out HOW the tree was behind HEAD (an editor buffer, a `git checkout -- <file>`, a
  sync script rewriting the file?). If you know, it belongs in the ledger's own entry on pathspec
  passengers. Logged on my side in `WRONG_CALLS.md`.
