# HANDOFF — 2026-08-22. `083` is CLOSED, the backfill is delivered as **3 of 27 with the other 24 refused on measured grounds**, and `bugs_open/357` is filed for the nine that must never be backfilled

**Supersedes `HANDOFF_2026-08-21_continue_here.md`** (which superseded 08-20b). Read this from disk,
then `NOTES_required_fields_repair.md` from the bottom — the 2026-08-22 entry has the full working.

> **Written ~10:00Z. Deploy facts have a shelf life of hours — re-read, do not quote.**
> Chassis is **`70e7b4f9c`** (`sha256:b83dc450…`, pods up 08-22 08:36Z), the third roll since this
> lane's code shipped. Capability re-probed on the newest binary rather than inferred from ancestry
> (`rendered_html_transform` 8, `code_span_to_code_tag` 5, control 0): a commit AFTER mine could
> delete the code and still leave mine an ancestor. **The seven `literal_markdown` repairs still hold
> at the served bytes across two roll boundaries** — prose backticks 0, in-script literals 4/44/8.

---

## 0. STATE

| thread | state |
|---|---|
| **`bugs_open/083`** | **CLOSED 2026-08-22 → `bugs_closed/`** (owner instruction, day 5 of the 7-day soak). Verified at HEAD by slug. Carries three caveats: `479`'s reclaim arm has never fired (7/0, live+archive), 41 of 42 still-`detected` rows have **no handler at all** (that is `149`/`114`/`236`, not 083), and the two-strike re-route trap is invisible to 083's own instruments |
| **`bugs_open/277`** | clause 1 MET and proven; **the backfill is done to the limit the data permits** (§1). 17 of 27 stay parked, 9 of them permanently unless `357` is fixed. What remains is a **decision about re-rendering**, not a data-recovery gap |
| **`bugs_open/357`** | **FILED today** — a whole tool page stored in a slot claiming to be the shared `hero` component. 9 rows, 2 sites, one as recent as 08-08. Root cause deliberately NOT asserted; in the loop as `63d4d1a7-ffec-4570-866b-8a0a41e3c69d` |
| `530`/`531` | done — APPROVED r1 (`c00fbfd8`), advisories answered |
| `540` + CQ-029 | applied; council **`cd8e555d`** submitted, verdict pending at the time of writing |

---

## 1. THE BACKFILL — what the owner asked for, and what the data permitted

| outcome | rows |
|---|---|
| recovered, round-trip byte-identical → **written** | **3** |
| blocked: template drift confined to `<style>` | 7 |
| blocked: template drift in **markup** | 8 |
| refused: stored HTML is not that component's output at all → `357` | 9 |

**Read this before proposing a looser gate.** `datahelpers.ContentDataCanFillTemplate` returns true
when content_data holds **any one** of the template's top-level fields. Writing a single recovered
field therefore flips a component from "cannot regenerate" to "can regenerate", and the next
regeneration renders that one field and blanks the rest under `missingkey=zero`. **A partial backfill
is destructive, not incomplete** — the parked state was the protection. Byte-identity is exactly the
property that makes the flip safe.

**The 15 drifted rows can never meet that bar**: their HTML was rendered by template versions that no
longer exist (`component_versions`: 367 rows over 202 components, **zero** for any of the nine
components involved). Re-rendering them would change the page. **Whether to do that is a decision
about pages, not a data-recovery question** — see §3.

## 2. STILL TRUE, AND STILL OWED

- **Read the `cd8e555d` verdict** and act on it (the only council item outstanding).
- **`63d4d1a7`** — the `357` diagnosis. Read it before anyone asserts a cause in that file.
- **`copy_edit_proposed` exclusion in the promoter's `pre_query`** (owner decision D2, 2026-08-12) —
  still deliberately not done by a session: it changes which rows are dispatched.
- **Two disconfirmable predictions now running.** (a) `learn-index` should be born `detected`, not
  `unresolved`, at the ~08-25 rotation sweep, because both its strikes age out first — **do not
  hand-flip it**. (b) The daily `review-queue-revalidate-daily` pass should stop reporting *"carries
  no content_data"* for the three rows `540` wrote.

## 3. THE ONE DECISION LEFT IN THIS LANE

**Do the 15 drift-blocked components get re-rendered?** Recovering their data is impossible (the
templates that made their bytes are gone), so the only route to a rebuildable state is to let the
CURRENT template render them — which changes what those pages serve. That is a deliberate content
change on 4 sites and is the owner's call, not a session's. Everything needed to decide is in
`bugs_open/277` §8–§9.

If the answer is no, **277 can close**: routing delivered, clause 1 proven, the repairable subset
repaired, and the residual split into a filed defect (`357`) and a stated, costed decision.

## 4. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py` by slug for `277`, `357` ·
chassis stamp + `git merge-base --is-ancestor` for anything you think shipped ·
`SELECT status, count(*) FROM site_work_items WHERE item_type='required_fields_missing' GROUP BY 1;`
(30 parked at the time of writing; 3 of them should retract on the next revalidator pass) · then §2.
