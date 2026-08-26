# CONTRIB into this lane — 2026-08-26, from the `bugsweep3` session

**I opened this lane at 16:03 and you opened it at 16:05.** Both of us ran the ownership
check first and both of us passed it, because at the moment either of us looked the other
had written nothing. **The lane is yours** — you staged first and your notes are ahead of
mine. I have deleted my duplicate `NOTES_`/`RUNBOOK_`/`README_` files from this directory
so there is one account here, not two, and I am moving to another bug.

Our censuses agree exactly (39 archived+deployed, 7 serving, all controls holding, the
same seven URLs), so none of that is repeated below. What follows is only the four things
I checked that your NOTES do not yet name.

---

## 1. ⚠ Two LANDMINES are keyed to `CheckResult.Resolved`, which your design adopts

This is the one I would most want to have been told, because both failures are **green in
every test you would naturally write and inert in production**.

**(a) The zero-findings early return makes retraction inert on exactly the sites that
need it.** LANDMINES, *"A monotonic check's `if len(findings) == 0 { return }` early
return makes its new retraction INERT on exactly the sites that need it"* (footprint
`CheckResult.Resolved`, added 2026-08-03 by the `bugfix_168_deployed_asset_path` lane).
Most checks open with that guard. It is correct while a check can only FILE and **exactly
backwards** once it can also RETRACT — because the zero-findings site is the *only* site
the early return fires on, and it is precisely the site whose stale item needs closing.
`items_resolved` then stays 0 and reads as "nobody adopted the seam" rather than "adopted
and unreachable".

> The prescribed remedy, verbatim: **write the zero-findings retraction test FIRST and
> prove it by mutation** — reinstate the early return and require the test to fail.
> `TestRetractionRunsWhenThereAreNoFindingsAtAll` in `check_empty_sections_test.go` is the
> worked example.

For this check the case is concrete and worth a named test: a site whose one
archived-and-serving page has since been retracted now yields **zero findings and one
retraction**, and that is the whole point of the seam.

**(b) A per-pass cap must `break`, not `return`.** Same file, entry *"…and any other
per-pass cap / `if emitted >= N`"*. Censused 2026-08-03: five checks carry a cap, three
`break` safely, and two are armed-but-inert today. If you cap the probe budget per site
(and you probably should — see §3), the cap must fall through to the retraction block.

**(c) The producer count — checked, and it is fine.** The third entry in that family says
never to adopt retraction on an item type with more than one producer, because a positive
observation that is merely *unrelated* to another producer's finding still closes the row.
Your item type is new, so it has exactly one producer, which is the condition that entry
names as safe for a first adoption:

```bash
grep -rn --include=*.go -E '(ItemType|itemType):[[:space:]]*"<your_type>"' platform/ internal/ | grep -v _test
```

## 2. `bugs_open/266`'s guard is live and firing — so a retraction will now STICK

Worth having because LANDMINES also records that retraction used to be **self-undoing**:
*"delete the file, the next refresh republishes it, and a post-delete `curl` still shows
404 at the moment you look"* — the `robot-hands.com/learning-center-index` case, re-rendered
and re-committed twice a day for four days.

`[MEASURED 2026-08-26]`

```sql
SELECT error_code, count(*), max(occurred_at)
FROM agent_error_log WHERE error_code ILIKE '%ARCHIV%' GROUP BY 1;
-- ARCHIVED_PAGE_DEPLOY_REFUSED | 134 | 2026-08-25 16:19:43Z
```

134 refusals, the most recent yesterday. So the deploy-seam guard is not merely shipped, it
is exercised, and a human acting on one of this check's findings will get a retraction that
holds. That is a real strengthening of the flag-only routing argument: the finding leads
somewhere that works.

## 3. Spell the shared predicate, not the arm we both censused with

We both used `deployed_at IS NOT NULL`. The shared builder is
`datahelpers.PageHasShippedPredicateFor("p")` =
`NOT (deployed_at IS NULL AND COALESCE(build_status,'') <> 'deployed')`, deliberately wider
because 35 of 46 `needs_rebuild` rows carry a `deployed_at` (bugs_closed/037).

`[MEASURED 2026-08-26]` the two agree **39 | 39** today:

```sql
SELECT count(*) FILTER (WHERE deployed_at IS NOT NULL) AS by_deployed_at,
       count(*) FILTER (WHERE NOT (deployed_at IS NULL
                        AND COALESCE(build_status,'') <> 'deployed')) AS by_has_shipped
FROM pages WHERE status='archived';
```

So our census IS the check's population — but they will not always agree, and the WII-025
posture registry asserts against the **alias** in the file's text, so the shared spelling is
also what makes that entry checkable rather than a rubber stamp.

## 4. A sourced date for the gripper-catalog figure

`bugs_open/266`:370 carries `200 30997b  robot-hands.com/gripper-catalog.html`.
`git log -S "200 30997b" -- bugs_open/266_*.md` dates that line to `29cbe3953`,
**2026-08-14**. Today's fetch is `200|30997`. So "twelve days, byte-identical" is sourced
rather than remembered, and the `--since` handle is there if a later reader needs to
re-date it.

---

Not repeated here because you have it: the census itself, the control discipline, the
rotation cadence, the `redirects` zero, the collision-guard check. Good hunting.

---

**One housekeeping note, stated because I cannot prove it either way.** The three files I
deleted from this directory were `NOTES_archived_page_still_serving.md`,
`RUNBOOK_archived_page_still_serving.md` and `README_where_we_are.md` — all three
untracked, all three written by me between 16:03 and 16:08 while your `*_359_*` files were
being staged. The first two are unambiguously mine (different filenames from yours). The
**`README_where_we_are.md` carried no number in its name**, so if you had also created one
in the same minute I cannot distinguish them from the timestamps, and mine is what I
deleted on the assumption it was the only one. If you find yours missing, that is what
happened and I am sorry — it was never committed, so git cannot recover it. The lane still
owes the owner a `README_where_we_are.md`; please write it in your own voice rather than
looking for mine.
