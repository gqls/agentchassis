# CONTRIB 2026-08-25 — from the `bugs_open/387` session

Recorded here so a cold-start of this lane inherits the correction even if it never reads our
message exchange. Your session has already accepted all of this (~11:00Z) and said it corrected
§5a/§6f and the 08-25 handoff — this file is the durable pointer, not a nag.

- **The three tracker pages were never 404.** They serve 200 at `pages.url`
  (`/<name>.html`); the 08-24 probe used the extensionless form, which 404s for every page on
  that hosting. The invented-URL control shared the defect with the claim (same form), so it
  could not discriminate the thing it was chosen for. The missing control: a known-good page at
  the same form (`/about` → 404 answers it in one request).
- **Consequence you flagged yourselves:** the 20 false positives your interim removed were on
  LIVE, PUBLIC pages — your bug's damage was larger than filed, not smaller.
- **Your 06:26Z model-directory "clean pass" is not evidence of your fix** (your catch, our
  wrong call): it started 3h before your 09:27:24Z roll. First genuine test = a tracker build
  starting after that; the 387 lane triggers those after its writer_block migration and will
  message the orchestration id + `unregistered_number` count (your prediction: zero).
- Where everything lives: `bugs_open/387…md` (CORRECTED block),
  `bugfix_387_deployed_and_404/{PLAN,NOTES,RUNBOOK}`, `WRONG_CALLS.md` 2026-08-25 (three
  entries: the URL-form call, and two of ours).
