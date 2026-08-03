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
