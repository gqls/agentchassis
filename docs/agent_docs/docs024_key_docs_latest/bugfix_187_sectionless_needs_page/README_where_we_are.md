# Where we are — bugfix 187 (needs_page items that can never be worked)

Plain-prose running log, append-only, newest at the bottom.

## 2026-08-03 evening — picked up, checked it's real, and it's worse than filed

The bug: five different parts of the platform raise "this page needs building"
work items, and 28 of them are sitting in the human-review queue where nobody
can ever finish them. When the ticket was filed there were 24; three more
appeared today, so the leak is live.

What I found tonight, checking every one of the 28 against the live database:

- About seventeen are asking for something impossible: "rebuild the sections
  of this page" where the page has never declared any sections (they're
  calculator/tool pages, which are complete as they are). Same disease the
  177 ticket just fixed for a sixth emitter — these were minted broken.
- Seven are the opposite: they were reasonable asks, and either the page has
  since been built by some other route (four of them — the item just never got
  told), or the page is buildable *right now* and the item is still parked
  (three of them). Nothing in the platform ever looks at this queue again once
  an item parks — a claim in the ticket that a "revalidator" would drain these
  turned out to be wrong when I read the code: that mechanism has never
  covered this item type.
- Five are doing exactly their job: pointing at pages that genuinely should
  have sections and don't. Those must NOT be silenced.
- One asks for a page that has since been archived — moot.

Direction (details in the PLAN once written): teach the emitters to check
"could anyone actually do this?" before raising the item — using one shared
piece of code rather than a third and fourth copy of the check the 177 fix
introduced — and register this item type with the existing queue-drain
mechanism so items whose ask has since been satisfied get closed with evidence
instead of parking for ever.

## 2026-08-04 just after midnight — fix written, tested, submitted for review, committed

The fix is in. One new piece of shared code answers the question "could the
page-builder actually do anything with this page?" and it is asked in three
places now: by the two parts of the platform that were raising impossible
requests (they now decline, visibly, instead), and by the queue-drain
mechanism, which can finally close a parked request once the page it asked
for genuinely exists — with the evidence recorded on the row.

Worth being honest about two things. First, the heavy lifting was done by the
opus model as asked, but it ran out of its weekly allowance a hair short of
the finish line, so the last stretch (fixing a test harness mistake it made,
and running the full verification) was done by this session directly. Second,
my first version of the clean-up script would have retired four rows it had
no right to retire — pages that are in the site plan and so could still be
built by the layout-borrowing fallback. The script's own safety check is what
caught it, which is exactly why the safety check exists.

Everything is committed and submitted to the review council; the verdict is
expected within the half hour. Once it lands the change gets built into a
fresh chassis image, proven on the running pods (both directions — a string
the change added must be there, a string it removed must be gone), the
clean-up script runs, and the ticket closes.
