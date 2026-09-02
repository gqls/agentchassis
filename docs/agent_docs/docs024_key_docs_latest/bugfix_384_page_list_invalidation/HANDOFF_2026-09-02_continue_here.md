# HANDOFF — bugs_open/384 page-list invalidation · continue here

**Written 2026-09-02 ~18:45Z. SUPERSEDES `HANDOFF_2026-08-26_continue_here.md`**, which said the bug
was ready to close. It is not. Read the 08-26 file only for history, and note it now carries a
CORRECTED block at its head.

Cold-start read order: **this file** → `bugs_open/384_…md` (read the updates dated **2026-09-02**,
tail-first — they supersede the 08-26 CLOSE-OUT) → `RUNBOOK_page_list_invalidation.md` →
`NOTES_…` tail.

---

## 1. THE ONE-LINE STATE

**384 is OPEN and must not be closed.** The fix itself works — four of five demonstrations are
genuine. But **one generic page has been serving the defect since 2026-08-27** and the cause is now
traced to a mechanism that is **not this lane's** (`bugs_open/389`). There is one cheap, decisive
experiment outstanding, and it may already have resolved itself while you were away.

## 2. DO THIS FIRST — the experiment that is either confirmed or refuted by now

The two-strike anti-churn arm brands a new item when **2 or more** terminal siblings share its
`item_key` within a **rolling 7 days**. So the brand lifts when the count falls to 1 — i.e. 7 days
after the **second-newest** strike, not the newest.

**The arithmetic, per key `[MEASURED 2026-09-02]`:**
- Blog-listing key (`page_rerender_blog_4851f6fc…_section_data_resolved`) — 6 completes:
  08-26 14:42:46, 15:18:36, 15:28:52; 08-27 05:08:22, **22:37:25**; 08-28 09:39:12.
  Second-newest is **08-27 22:37:25 ⇒ brand lifts ~2026-09-03 22:37Z.**
- The `rerender-pages` keys (`needs_rerender`, `deactivated_component`) last completed
  **08-27 21:22–21:33 ⇒ ~2026-09-03 21:2xZ.**

**So the prediction is sharp: nothing before ~2026-09-03 21:30Z, and service should resume that
evening.** Verified still broken at 2026-09-02 18:5xZ (11 of 13, byte-identical to 08-31 — three
identical reads across three days).

### ⚠ A CHASSIS ROLL LANDED 2026-09-02 ~21:00Z AND IT CONFOUNDS THE NEGATIVE RESULT

`[MEASURED 2026-09-02 21:1xZ]` Fleet is mid-roll: `ebf27c60377f` (540 pods) → `0d2feee2ff61`
(61 pods, a strict descendant). **All four lane commits are ancestors of BOTH**, so the lane's own
behaviour is unaffected. **1,034 commits** in the roll.

**But `applyGrowthPostureDoor` (WDS-020) is NEW in it** — 0 occurrences in the previously verified
build `ef06af0e0afc`, 1 in `0d2feee2ff61`. It runs inside `writeWorkItem`, **the same seam
`insertWorkItem` wraps**, i.e. the seam that carries the two-strike arm — and it **parks items at
`status='deferred'` with `handler_agent=''`** (`growth_posture_door.go:88`).

**So a negative result no longer refutes anything on its own.** ~~If it is still 11 of 13 on 09-04
the chain in §4 is REFUTED~~ — **that is now WRONG as written**: a still-frozen site could be the
two-strike arm OR the new growth door, and the two are indistinguishable from the page alone.

**Discriminate at the item, not the page. This is the first query, before the curl:**
```sql
SELECT status, coalesce(nullif(handler_agent,''),'(EMPTY)') AS handler,
       left(summary,60) AS summary, spec ? 'growth_release_recipe' AS growth_door, created_at
  FROM site_work_items
 WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
   AND handler_agent IN ('rerender-pages','') AND created_at > '2026-09-03'
 ORDER BY created_at DESC LIMIT 10;
```
- `unresolved` + `[unresolved after N attempts]` ⇒ **two-strike arm still holding.** Chain intact;
  either the age-out has not landed or the arithmetic is wrong. Re-derive from the actual strike
  timestamps before concluding.
- `deferred` + `growth_release_recipe` in spec ⇒ **the NEW door, not this lane's chain.** The
  experiment is VOID, not refuted — say so, and note the roll date.
- `triaged`/`claimed`/`complete`, and `rerender-pages` runs for the site ⇒ **chain CONFIRMED.**
  Stamp the resume time into `bugs_open/389`'s evidence (the `dispatch_throughput` lane's request).

**Two more items in the roll touch this ground** — neither changes the arm, both worth knowing:
`8eca969cb` (315 reopen: a producer "stops filing guaranteed-skip rerenders", so filing VOLUME may
drop for reasons unrelated to branding) and `2a0bdb001` (035 P1: `rerenderFlatSections` extracted —
a MOVE; it is the function the `090` in §8 reported it could not see, so a re-run now seeds better).
**Verified unchanged across the roll:** the two-strike predicate
(`status IN ('complete','failed')`), the `[unresolved after %d attempts]` branding literal, and
`recurrenceExpected`. The arm itself did not move.

```bash
curl -s https://leopardessconsulting.co.uk/blog.html | grep -o 'src="[^"]*card-[^"]*"' | wc -l
# 11 of 13 = still broken (the state on 08-31 and 09-02, byte-identical both days)
# 13      = it repaired unaided ⇒ §4's chain is CONFIRMED
```
Then, whichever way it went:
```sql
-- did rerender-pages resume for the site?
SELECT count(*) FROM orchestration_states WHERE owner_agent_type='rerender-pages'
  AND collected_data::text LIKE '%4851f6fc-71cf-4160-a270-e03d6d3e0732%';
-- did the array get rewritten, and BY WHOM?
SELECT created_at, application_name, source FROM page_component_history
 WHERE page_id='05269427-0c77-4e7a-a69c-a5b7b2dd7ca1' ORDER BY created_at DESC LIMIT 5;
```
**Both outcomes are informative and neither costs anything.** Repaired ⇒ the chain in §4 holds and
the lane's remaining work is to stop it recurring. Not repaired ⇒ the chain is wrong and something
else holds the site; start again at §4's step 3.

## 3. What is TRUE and settled

- **The fix is live and it works.** All four lane commits (`7720dc76c`, `bafd4411c`, `72469c556`,
  `efc0db7bc`) are ancestors of every build seen so far — `ef06af0e0afc` (08-31), and after the
  2026-09-02 ~21:00Z roll both `ebf27c60377f` and `0d2feee2ff61`. **Re-read the stamp, never
  remember it** — this file has already been overtaken once, and §2 carries what that roll changed.
- **§4's demonstrations are genuine — 4 of 5.** finetuning's three (19:13:31, 19:14:21, 19:15:17)
  and vonc's `archetypes` (08-27 08:12:56) are real `save_page_sections_overwrite` writes. **Only
  the leopardess 15:30:33 one was another action's repair** (`action:rebuild_blog_listing`).
  I doubted all four on 09-02 and was wrong; the doubt is withdrawn.
- **The defect is BLOG-LISTING-specific.** A blog listing's array is written only by
  `rebuild_blog_listing`; `save_page_sections` has written `leopardess/blog` **once in seven weeks**
  (2026-07-12). The `page-rerender` workflow has no `rebuild_blog_listing` step. So on a blog page
  the seam's item renders from stored `content_data`, deploys a real sha, and writes nothing —
  while the same item on a tool-listing page repairs correctly. **That is why exactly one generic
  page fleet-wide is stuck and every aggregate reads healthy.**
- **The owned-page residual is unchanged**: 14 blanks across 3 `rebuild_policy='owned'` pages.
  Structural, correct, still needs its own seam (migration `486`'s `section_edit` route).

## 4. THE CHAIN — why `rebuild_blog_listing` has not run since 2026-08-27 21:34:20

`[MEASURED 2026-09-02]`
1. `rebuild_blog_listing` lives in exactly one live agent: **`rerender-pages`**.
2. `rerender-pages` is spawned by `build-dispatch-loop` (36 of 37 runs in 24h).
3. It ran **37 times across 17 sites in 24h — zero for leopardess.**
4. leopardess gets 18 `rerender-pages`-handled items in 7 days, but **15 are born `unresolved`**,
   every one carrying the `[unresolved after N attempts]` brand. Last completion **08-27 21:22–21:33**;
   last listing write **21:34:20**.
5. Born-terminal ⇒ never eligible ⇒ no `rerender-pages` ⇒ no `rebuild_blog_listing` ⇒ blank cards.

**The strikes were this lane's own seam succeeding** on 08-26/27 after card landings. `insertWorkItem`
counts `status IN ('complete','failed')` on a shared `item_key` — **successes count as strikes.**

⚠ **The brand is NOT unique to leopardess.** relojistas 36/43 branded, dartsonline 17/24,
gaswholesalers 13/22, boxingonline 3/10 — all still served, because some keys still get through.
leopardess has none left that do. **[INFERRED]**, and §2 is the test.

**A much better control, supplied by the `dispatch_throughput` lane 2026-09-02** — 21-day daily
rerender-family completions (live+archive). The discriminator is **post-outage recovery**, which my
snapshot could not see: 09-01/09-02 dartsonline **268/106**, gaswholesalers **52/139**, relojistas
**40/72**, leopardess **2/1**. leopardess froze at 08-28 (…89, 179, then 5, 4, 2, 1) and is
**uniquely unable to recover** while its comparators did. ⚠ **The weekly PERIODICITY is untestable
in this window** — the 08-28→31 trough is the fleet LLM outage and every site dips, so **do not
claim a sawtooth**. §2 is the clean test.
**If it resumes, stamp the resume time into `bugs_open/389`'s evidence** — the peer lane's request,
and the confirmation that mechanism needs.

## 5. WHAT IS SOMEONE ELSE'S — do not fix these here

- **The two-strike mechanism belongs to `bugs_open/389`** (owned by the `bugfix_308` lane;
  `who-owns.py` confirms). Its class 1 already names the chain. **Evidence contributed there
  2026-09-02**, including the property they did not have: the strikes and the parked item came from
  **two different producers sharing an `item_key` on purpose**, and all six strikes were successes.
  A fix candidate is noted there (`insertWorkItem` already exempts `item.recurrenceExpected`) and
  deliberately **not taken**.
- **The `detected` backlog: RESOLVED as designed — there is NO bug and nothing to take forward.**
  Corresponded with the `dispatch_throughput` lane 2026-09-02. Promotion is `detected-item-promoter`
  (900s tick, enabled, firing, **independent of site selection**); `sites.build_status` is **inert
  for dispatch**; they added a zero-eligible starvation census to their runbook (`155c36812`).
  ⚠ **I claimed the handler door "parks 100% of the detected population by construction" and that
  was WRONG** — verified at `scheduled_tasks.pre_query` (line 51,
  `AND COALESCE(wi.handler_agent,'') <> ''` inside `scored`): handler-less rows are excluded
  UPSTREAM and never reach any door. They are **flags** — records with no automated handler, whose
  permanent home IS `detected`. I had tested rows against my own paraphrase of the door instead of
  reading the query. **Do not re-open this.**

## 6. THIS LANE'S OWN SWEEP HAS NEVER RUN — and its watch item is vacuous

`check_page_list_stale` (migration `603`, live 2026-08-25) has filed **12 items in its lifetime,
live AND archive, all `unresolved`, `attempt_count=0`, zero ever claimed.** Same two-strike arm.
It detects correctly every time and every report is binned at birth.

⚠ **§8.1 of the old handoff asked for the escalation rate to be re-read against a 1-in-36 baseline,
recording "zero escalations". That is zero over an EMPTY denominator. Do not re-read it as written;
count the RUNS before quoting a RATE.**

## 7. TRAPS THIS SESSION PAID FOR — read before you measure

Four instrument failures in one day, three of the same shape (an instrument that could only return
the answer it gave). All are in `WRONG_CALLS.md`.

- **`page_components.updated_at` is NOT trigger-maintained.** I rested "the array was never
  rewritten" on it across three docs. Use **`page_component_history`**: its archive triggers fire on
  real change, and **`application_name` NAMES THE WRITER**.
- **`page_component_history.component_id` is the `page_components` ROW id, and `save_page_sections`
  DELETES AND RE-INSERTS that row.** So a join to live `page_components` drops **98.3% of history**
  (44,781 of 45,540) — and the survivors are exactly the in-place writers, which produced a
  confident, wrong culprit. **Key on `page_id`.**
- **A NOW-census of `triaged`/`approved` tells you nothing**: every site reads zero-eligible,
  including ones served today. Claimed rows leave the status and completed rows archive out. I
  built a "structurally invisible site" theory on it and had to retract it to a peer.
- **`unresolved` is in `workItemTerminalStatuses`.** A born-terminal backlog reads as OPEN
  in-flight work to the 090 coverage clause (`NOT IN ('complete','cancelled','rejected')`) — it
  refused my diagnosis run over the very rows the bug is about. `FORCE=1` was correct there; read
  the items first.
- **Seeding a `090`**: `SEED_SCOPE` with `file.go:Symbol` omits the callee holding the deciding
  branch — **seed whole files**. And the bundle does **not** auto-fetch workflow steps, so quote any
  load-bearing live `agent_definitions` value **in the symptom text**.

## 8. The `090` that was run

Intake `d4f745e6-3f79-42a8-8f71-bb611736912c`; **run correlation
`149ec925-ffb7-41eb-806a-1595b8ff2226`** (artifacts are written under the RUN one). 5 iterations,
verdict **`UNVERIFIABLE`**. **Not a waste** — it refused to confirm, and its "still needed" list
named the `updated_at` ambiguity that sent me to `page_component_history`, which is what cracked the
case. A re-run is possible with the §7 seeding fixed, but **do §2 first** — it may make the question
moot.

## 9. Where the knowledge lives

- **Bug + every 09-02 update:** `bugs_open/384_HANDOFF_2026-08-24_….md` (tail-first).
- **Contributed evidence:** `bugs_open/389_…` (CONTRIB dated 2026-09-02).
- **Traps:** `docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md` (two entries, 2026-09-02).
- **Commands:** `RUNBOOK_page_list_invalidation.md` — incl. the `page_component_history` recipe and
  the `090` seeding notes.
- **Owner prose:** `README_where_we_are.md` (three entries, 08-31 and 09-02).
- **Peers:** `dispatch_throughput` lane (live session `throughput`, corresponded 09-02);
  `bugfix_308` lane owns 389.
- **Code:** `rebuild_blog_listing_action.go`, `rerender_page_sections_action.go`
  (`planSection`, `carryStoredSection`, `rerenderFlatSections`),
  `load_work_item_actions.go:1985-2033` (the two-strike arm), `work_items_common.go:42`
  (terminal statuses), `queryresolve/queryresolve.go`, `discovery_checks/check_page_list_stale.go`.
