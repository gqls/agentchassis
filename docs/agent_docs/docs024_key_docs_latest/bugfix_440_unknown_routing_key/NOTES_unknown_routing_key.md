# NOTES — bugfix 440 (append-only, newest at the bottom)

## 2026-09-02 — lane opened; evidence gathered; one wrong call caught

Spun out of 410's candidate 1 (owner decision). Evidence base assembled first-hand — census,
probe, migration reads — all `[MEASURED 2026-09-02]`, recorded in the bug file. Missteps:

- **Wrong call, caught before filing but after saying it in-session**: read "2 rows contain the
  warning string" as "warning fired twice in production". Both rows were the 404 lane's council
  runs QUOTING the string in their submission payload. Caught by reading one member row
  (`current_step = complete_revise/complete_approved`). Logged in WRONG_CALLS 2026-09-02. The
  corrected finding (zero production firings + prose minted via migrations that bypass the
  creator) became the load-bearing "many doors" argument — the error, chased, was worth more
  than the number.
- Side effect of the same read: learned 404's r4 is `complete_approved` — their design is
  through. Their session has not yet read/recorded the verdict; nothing of theirs touched here.
