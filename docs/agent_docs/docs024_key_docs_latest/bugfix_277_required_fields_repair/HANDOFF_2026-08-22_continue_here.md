# HANDOFF — 2026-08-22. **THE LANE IS CLOSED.** `083` and `277` are both in `bugs_closed/`; the successor is `bugs_open/357`

> **If you are picking this up cold: there is no work left in this lane.** Go to `bugs_open/357`
> (nine tool pages mislabelled as `hero` components — cause UNVERIFIABLE, precedent to copy named)
> or read `SUMMARY_2026-08-22_lane_closed.md` for the whole arc. Everything below is the closing state.

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
| **`bugs_open/277`** | **CLOSED 2026-08-22 → `bugs_closed/`** (owner ruling). Router live and classifying everything; both repair shapes proven at the served bytes; the `no_content_data` repair built, APPROVED (`cd8e555d` r1) and applied. **24 of 27 rows stay parked** — 15 blocked by vanished templates (a decision about re-rendering, ruled not-now), 9 in `357`. ⚠ the file's §9 says "17 of 27"; the arithmetic is 24, corrected visibly in the close |
| **`bugs_open/357`** | **FILED today — THE SUCCESSOR.** A whole tool page stored in a slot claiming to be the shared `hero` component. 9 rows, 2 sites, one as recent as 08-08. Diagnosis `63d4d1a7` returned **UNVERIFIABLE** ("scope-not-narrowing") — neither confirmed nor refuted, recorded as such. Two of its three gaps are closed in the file; the third (which writer) is narrowed to six writers + one labelled lead. **A live proven precedent exists** (`loancalculator_couk/decompose`) — read it before writing anything |
| `530`/`531` | done — APPROVED r1 (`c00fbfd8`), advisories answered |
| `540` + CQ-029 | applied; council **`cd8e555d`** **APPROVED r1** — all six objections answered by checking, and the guardian's was RIGHT (540 was already claimed by another lane; cosmetic, the ledger keys on filename). See NOTES §5 and the RUNBOOK's new section on this lane's migrations being absent from `schema_migrations` |

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

## 3. THE DECISION THAT CLOSED 277 — recorded, so it is not re-litigated

**"Do the 15 drift-blocked components get re-rendered?"** — the owner's answer, by closing 277 on
2026-08-22, is **not now**. Recovering their data is impossible (the templates that made their bytes
are gone), so the only route to a rebuildable state is letting the CURRENT template render them, which
changes what four sites serve. They serve correctly today and stay as they are. The costing is in
`bugs_closed/277` §8.4/§8.5 if anyone reopens it.

## 4. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py` by slug for `277`, `357` ·
chassis stamp + `git merge-base --is-ancestor` for anything you think shipped ·
`SELECT status, count(*) FROM site_work_items WHERE item_type='required_fields_missing' GROUP BY 1;`
(30 parked at close; 3 of them should retract on the next `review-queue-revalidate-daily` pass — a
disconfirmable prediction, ~16:01Z daily) · then §2. **If §2 is empty, this lane is finished; go to
`bugs_open/357`.**
