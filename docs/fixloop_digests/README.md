# Fix-loop digests

A plain, daily account of what the self-fixing loop has been doing — so you can
stay aware without going digging. This directory is the **committed-file
delivery** of that account.

- **`DIGEST_latest.md`** — the most recent digest. Read this.
- **`archive/DIGEST_YYYY-MM-DD.md`** — one file per day; the git history of this
  directory is your diary of the loop over time.

## How it works

1. Inside the platform, the `fixloop-digest` agent composes the digest —
   **deterministically, with no AI in the path** (facts gathered by SQL,
   rendered as plain text), so it can only report what it can point to. It
   writes to the `doc_notes` table.
2. A pod can't write to this repo, so `094_pull_digest_to_file.sh` reads the
   latest digest out of `doc_notes` and writes it here.
   - On demand: `./094_pull_digest_to_file.sh`
   - Daily/automatic: a local cron running `./094_pull_digest_to_file.sh --commit`
     (pair it with the scheduled digest agent).

## What each section means

- **Runs** — every fix-loop orchestration in the window: what agent, its status
  and terminal step, the build-gate verdict, and any pull request opened.
- **Decisions by correlation** — for each bug worked on: what was written
  (diagnosis bundles, plans, council reports) and the **latest council decision
  and its reason**.
- **Agent config changes** — every change to the platform's agents in the
  window, with the reason given. This is the ledger of *changes to the machine
  itself*: if anything anywhere starts changing the platform in a direction you
  didn't intend, it shows up here.

## Scope — what the digest does and does NOT watch

The digest reports on the **fix loop's own activity**. It is a rear-view mirror
of what the loop did, not a forward-looking scanner of the platform. It does
**not** crawl sites for problems (e.g. missing pages) — detecting those is the
job of the build workflow and the audit/checker agents, which escalate genuine
code-caused problems into the loop as `needs_diagnosis` items. See
`../agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/` for the loop
itself and the architecture notes.
