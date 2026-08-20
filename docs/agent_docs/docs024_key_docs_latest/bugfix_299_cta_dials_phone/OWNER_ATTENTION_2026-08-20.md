# Points for the owner — `bugfix_299_cta_dials_phone` lane, 2026-08-20

Written at the owner's request so these can be picked up rather than living in a chat log.
**Nothing here blocks anything.** The bug is closed and live. These are the things a person
should know or decide, ordered by what actually matters.

---

## 1. NEEDS A DECISION — one only

**Should the fleet get a roll to pick up the capability-table prune?**

The capability registry (`service_binary_capabilities`) went live and immediately revealed a
leak in its own design: the chassis binary also runs as **short-lived one-off worker pods**, not
just the handful of long-lived services, and each one recorded itself and then vanished, leaving
its rows behind for ever.

- **Fixed in code** — the table now clears its own dead entries. Committed.
- **Not yet running** — Go changes need a rebuild, and the live binary (commit `447f3a8a8`,
  built 10:33) predates the fix (committed 15:08). The image is still `v1.0.1319`.
- **Meanwhile it regrows.** Currently modest: **3–5 pod starts an hour** in a quiet period,
  roughly 1,200 rows an hour. It was 15× that during this morning's rerender wave.
- **Clearing it is one line**, in the lane RUNBOOK, and takes a second.

**Recommendation: no special action.** Let it ride on your next ordinary release. If a week
passes with no release, run the one-liner. This is housekeeping, not risk — nothing reads the
table yet, so an oversized one costs disk and nothing else.

> ⚠ **I originally told you ~160 MB/day.** That was wrong by about 17× — I extrapolated from a
> burst window that happened to contain a fleet rerender wave, part of which was my own. The
> honest figure is a **range** that depends on fleet activity, and I have corrected it in the
> RFC, the register and `WRONG_CALLS.md`. Flagging it because urgency is what you would have
> acted on, and a wrong urgency is worse than a wrong detail.

---

## 2. WORTH KNOWING — no action needed

### (a) Three rerenders reported success while publishing the old page

Getting the fixed button onto the live site took three attempts. The first rebuilt nothing. The
second rebuilt everything correctly and then **threw the work away** — literally
`"success": true, "sections_saved": 0`. Both produced a real deployment, a real commit and a
`COMPLETED` status. Only the page itself disagreed.

The cause is that a rerender needs two fields most callers do not know about; omit either and it
silently does nothing useful. Written up in `LANDMINES.md` with the working command in the
RUNBOOK. **Why it matters to you:** this is the estate's most-repeated failure shape — a green
status over an unchanged artefact — and it is now documented rather than folklore.

### (b) A reviewer was right and my measurement was wrong

The council's guardian seat objected that the new capability writer would open a database
connection "on every pod of a ~41-replica fleet". I measured **5** live pods and was one step
from recording the objection as inapplicable. The real quantity was never the pod *count* — it
was the pod *start rate*, which is dozens an hour because of those short-lived workers.

What caught me was the tool I had just built, answering a question I had not asked. Logged in
`WRONG_CALLS.md`. **Why it matters to you:** the review process paid for itself here, and the
near-miss (retiring a correct objection with a confident wrong number) was worth more than the
defect it prevented.

### (c) Your phone button was a leftover the system would have defended for ever

As you suspected, it was never intentional — a week ago that section genuinely said "Prefer to
talk it through first? Call…". The copy was rewritten four times; the link never moved. The
uncomfortable part: **our fix would have protected it permanently**, because the whole point is
to teach the system that a phone link is deliberate and must be kept — which is what stops your
genuine "call us" buttons being destroyed. The system cannot tell deliberate from inherited.
That distinction needs a person, which is why your answer was the thing that closed it.

### (d) Two more of the same class are queued on your site, deliberately not fixed

`faq/hero` ("See what you get for it" → a phone) and `how-it-works/call-to-action` ("Still
deciding? The FAQ page covers the full terms…" → a phone). Both now filed by the new detector
and sitting in review. **These tel: links are believed genuine**, so the fix is the *copy*, not
the link — and that is exactly what the new destination stamp feeds the writer. They are the
first real test of it, so they are more useful left alone than hand-corrected.

---

## 3. FOR THE RECORD — what this lane produced beyond the bug

- **`bugs_open/312`** — the framework defect underneath: every page build computed the right
  links and threw them away. Before the fix, **0 of 33** computed answers survived; after,
  **4 of 4**. Its two tripwire candidates (a loud fallback, a lockstep test) remain unbuilt and
  are earned — that seam has now failed silently in both directions twice.
- **`RFC_040` / `BLD-023`** — the estate's standard way of answering "is my fix live?" turns out
  to be frequently *impossible*: the version line scrolls out of the log within hours and the
  binary carries one stamp rather than its history, so checking for your own commit returns
  "absent" for a binary that has it. Two separate threads had already been burned. There is now
  a table that answers it in one query — and it settled this very question for me an hour ago.
- **Three `LANDMINES.md` entries**, one closed bug, one open, three `WRONG_CALLS.md` entries
  (all three mine).

---

## 4. CAN THE LANE CLOSE?

**Yes — with one pointer left behind.** `bugs_open/299` is closed, fixed and live. RFC_040's
authorised scope is implemented and council-approved. The only live thread is item 1 above
(a roll picks up the prune), and it belongs to whoever runs the next release, not to this lane.

`bugs_open/312` stays open on its own merits and has its own file. It does not need this lane.
