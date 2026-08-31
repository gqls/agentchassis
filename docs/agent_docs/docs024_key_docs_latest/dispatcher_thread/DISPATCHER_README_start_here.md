# DISPATCHER THREAD — start here (created 2026-08-31, owner request)

**What this is.** The owner's standing coordination thread: one Claude session, opened
FIRST (or kept always open), that watches for work addressed to *threads* and routes
it — because the cluster cannot push into a workstation session, and tonight's worked
case proved the gap (three lanes had owner-critique CONTRIBs filed and no live session
reading them; routing happened only because one session was watching a chat thread).

**The owner's stated intent, verbatim (2026-08-31):** "Maybe I can have one thread
always open (or that I open first every time) that monitors what threads need to be
opened? They might be requested from the task database or from another open thread on
this laptop."

## How to run it (for the session that cold-starts here)

You are the dispatcher. Run a self-paced loop (`/loop` with no interval; 1200–1800s
idle ticks are right — critiques are minutes-latency work, not seconds). Each tick:

1. **Poll the task database** for unrouted signals:
   `./poll_open_critiques.sh` (this dir) — prints open `owner_critique` items
   (the REQUEST CHANGES button files these; core-manager commit `f2b288b72`,
   inert until its roll) that carry no `routed_at`, plus any work item whose spec
   names `"consumer": "thread-dispatcher"`.
2. **Route each one**: read the critique + site; decide which lanes it touches
   (`MEMORY_workstreams.md` maps lanes→threads; `scripts/who-owns.py <slug>` for
   bugs; your own judgment for quality themes — copy → copy_quality_two_stage,
   design → editorial_design_uplift / visual-designer, imagery →
   inline_guide_imagery, delivery → site_delivery_and_editor). Then `ListAgents`:
   - lane has a LIVE session → `SendMessage` it the critique verbatim + the item id;
   - lane has NO live session → add it to the tick's **"threads to open"** report.
3. **Stamp the routing** so the next tick doesn't re-route (and so a second
   dispatcher never double-routes): the poll script prints the exact UPDATE —
   it writes `result.routed_at/routed_to`, status stays `triaged` until the
   receiving lane completes the underlying work.
4. **Report to the owner** ONLY when something happened: what was routed where,
   and the "threads to open" list — the names he should start, each with its
   cold-start doc path so the new session lands running. Quiet tick = noop.
5. **Receive requests from other threads**: any session can `SendMessage` this
   session by its name to ask for routing or for a thread to be opened —
   treat those exactly like DB rows (route or queue for the owner).

## Rules the dispatcher inherits

- **Route, don't work.** The dispatcher never fixes, never edits sites, never
  answers critiques itself — a dispatcher that starts working stops dispatching,
  and its judgment about lanes is only trusted because it has no stake.
- **A dead lane's work item is CORRECT state, not a failure** — it waits, durable,
  and the "threads to open" report is the remedy. Never mark an item routed
  because it *should* be picked up; only after a live session acknowledged or the
  owner was told to open one.
- **Completion is the receiving lane's job.** The dispatcher stamps `routed_*`;
  whoever does the work completes the item (and the origin review item stays
  gating delivery until the owner APPROVEs — that gate is untouchable from here).
- Re-read this file from disk each cold start — it is co-edited state.

## Across machines / teams (the owner's "work across teams" thought)

The task database is the bus that crosses machines: every laptop's sessions can
poll the same `site_work_items`, so a second team member runs their OWN dispatcher
session bridging the shared queue to *their* local threads. Nothing new to build —
the routing stamp (`routed_by` carries the dispatcher's session name) is what stops
two dispatchers double-routing one item; first stamp wins, and a dispatcher skips
rows already stamped. Cross-machine *live* messaging (cloud sessions / Remote
Control) can layer on later; the queue-plus-stamp design doesn't depend on it.

## Current state (2026-08-31)

- REQUEST CHANGES button + `owner_critique` item type: committed `f2b288b72`,
  council trail `9f1cb042`, **inert until the next core-manager roll** — until
  then critiques arrive the manual way (chat) and this thread routes those too.
- Worked example of the whole flow, manual version: the boxingonline critique of
  2026-08-31 — see `../site_delivery_and_editor/OWNER_REVIEW_2026-08-31_…` and the
  webdesign NOTES entries of the same date.
