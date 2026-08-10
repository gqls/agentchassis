# HANDOFF — provocation pipeline, 2026-08-10

**Supersedes `HANDOFF_2026-08-05_continue_here.md`** for "what to do next". Its traps
list still applies and is not repeated in full here.

> **⚠ THREE SESSIONS HAVE WORKED THIS AREA THIS WEEK.** Twice, work was already
> half-built by another session before this one started — once it would have been a
> **compile failure on shared HEAD**, not a mere duplication. **Before you touch
> anything below: `git status` on the target package AND grep other sessions' live
> `~/.claude/projects/*/[0-9a-f]*.jsonl` for the symbol.** `git log` and
> `scripts/who-owns.py` read COMMITS and were blind to all of it.

---

## 1. State in one paragraph

**The site is telling the truth.** Since 04:41Z today it serves a provocation dated
today, ending 13 days of a false "Today's Provocation" heading. Six provocations are
approved and dated through **15 August**, after which the shelf is empty again.
Categories are **live and dormant**. The publish-time allegation check is **built,
council-approved, and NOT live** — it ships from the island VM, not the chassis.
Two RFC_020 items remain unbuilt.

## 2. Verified live [MEASURED 2026-08-10, at the artefact and the pod]

| thing | state | evidence |
|---|---|---|
| served feed | `today.slug=you-love-being-from-your-city`, **date "10 Aug"**, archive 8 | `curl` cache-busted; `generated_at 2026-08-10T04:41:15Z` |
| pool | 6 approved `llm` (10–15 Aug) + 8 approved `human` + 1 **retired** | `SELECT … FROM provocations` |
| categories (`feedFilename`, `shouldBootstrap`) | **LIVE, both replicas** | pod-grep = 1 each; **negative control** `"provocations for %s — refusing"` = **0**, its replacement = 1 |
| `injectRobotsNoindex` (bug 232, other session) | **LIVE, both replicas** (=2) | same exec |
| `provocations-pets.json` | **404** — correctly nothing, no second category | `curl` |
| namecheck publish gate | **committed, NOT live** — island VM deploy, not ours | `3ec99efb1`, council `73dc4e78` APPROVED |

**Two reconciliations worth keeping, because both looked like bugs and neither was:**

- **The owner re-dated the drafts.** I parked them 09–14 Aug; he approved on 09 Aug
  at 09:39 and re-ordered them to 10–15 Aug. Editorial control working as designed —
  do not "correct" the dates back.
- **Archive stayed 8, not 9.** `group-chats-replaced-friendship` was **retired** (it
  was the one entry with no full case). As today advanced, the 26 Jul entry moved
  *into* the archive (+1) and the retirement took one out (−1). ⚠ **They cancelled,
  which is also why the shrink guard did not fire** — a retirement on a day with no
  rotation WOULD have been refused (8→7). Worth knowing before retiring another.

## 3. What to do next — in order

### 3.1 CONTENT RUNS OUT 2026-08-15. Six days.

The only item with a deadline. Two routes:

- **The generator** (`provocation_generator_action.go`, another session's, council
  `bbbc9fca8`) is built and approved but **has never been run against a real model**.
  Its own lane flagged the calibration run as owed: `cmd/provocation-gate-calibrate`.
  **Read their docs, not this file, for its state.**
- **By hand.** The owner has now twice asked a session to draft them and then edited
  the result, so the working pattern is: session drafts → owner re-dates/approves.
  **Insert as `status='draft'`** (invisible to the feed by construction) and let him
  approve. Recipe: RUNBOOK § "Adding a provocation"; audience: **PLAN §11.1**.
- **⚠ Apply PLAN §11.2's test to anything new:** *does answering this well require
  naming a real person or business and saying something checkable about them?* If
  yes, reframe or drop. One draft ("Restaurant food has got worse") was pulled on
  exactly this ground.

### 3.2 RFC_020 §5.3 and §5.4 — owner has authorised both

He said "go ahead as you suggest" on 2026-08-10 covering all of §5.2/§5.3/§5.4. Only
§5.2 got built before this handoff.

> **§5.4 IS DONE AND LIVE — 2026-08-10, commit `438b18d65`.** Both surfaces, verified
> at the artefact. Card (`5da50747` `js_content`, md5 `dec3a0b8`): *"The judge rates
> how well the case was argued — not whether it is true."* Record page (`71a54cc2`,
> md5 `aaac7950`): the same plus *"No claim on this page has been checked for
> accuracy."* Build status for all four items is now in **RFC_020 §7**, which is the
> file to read rather than this one. Details and the two traps it paid for are in
> `NOTES_provocation_pipeline.md` under 2026-08-10.
>
> ⚠ **It does not discharge §5.2 or §5.3.** §5.4 is a supporting control by the RFC's
> own words: it makes the artefact honest about what it is, it does not stop a round
> naming somebody.

- **§5.3 — this is now the open one, and the more important of what remains:** a
  published report/takedown route plus a written process. This is what preserves the
  operator position given that posters are anonymous (`bugs_open/139`). Mostly not
  code — a contact surface and a documented procedure. It is the item whose absence
  is hardest to explain after the fact.

### 3.3 The island deploy for §5.2 — NOT OURS TO FIRE

`namecheck` is inert until `tools-api` is rebuilt and shipped to the island VM
(docker compose + scp, RUNBOOK `gauntlet_dead_cta` §5). **The council's `guardian`
seat explicitly asked that this be scheduled by the owning lane, not fired
opportunistically by us.** Their cold-start carries an INCOMING saying so. Verify
after: `curl -sI '…/round/<REAL published slug>' -H 'Origin: https://vonc.com' | grep -i x-robots`
— a 404 returns no header either and reads identically to a fix that is not there.

### 3.4 Categories are half a feature — and the other half is not ours

The publisher can write `provocations-<category>.json` today; **nothing reads it.**
`tools-api`'s `FetchProvocation` takes a domain only and always fetches
`provocations.json`. That is **RFC_020's sibling, RFC_013 §2.2, still unruled**, plus
§2.3 (should a round record its category — cheap now, unrecoverable later, because
rounds publish to permanent URLs) and §2.4 (should the contract become a shared Go
type). **Do not seed a second category expecting anything to happen.**

⚠ And when it is built: `provocStore` (`round.go:25-29`) is keyed by **domain alone**,
5-minute TTL. A per-category feed needs a category dimension in that key or categories
serve each other's provocations for up to five minutes.

## 4. Owed follow-ups that are recorded but not done

1. **Extract `datahelpers.NegationGuard` into a leaf package.** `namecheck` now
   carries a **second copy of the algorithm** (own cue vocabulary, per that guard's
   own "two vocabularies, one algorithm" doctrine — `bugs_open/222`). Not imported
   because `datahelpers` drags goquery, cascadia and five tdewolff minify packages
   into a service that parses no HTML — **measured**, 12+ heavy transitive deps for a
   three-field struct. **A THIRD copy should be the extraction, not another paste.**
2. **Measure namecheck's false-positive rate from real traffic** via
   `logPublishRefusal`'s signals. The allow-set is my own composition and the
   `architecture` seat said so: green tests are not evidence of correct tuning. A term
   firing constantly is too broad; one never firing is dead weight.
3. **Confirm how many rounds are published and whether any are strangers'.** Still
   unknown here — `gauntlet_rounds` is on the **island VM**, not `clients_db`. Last
   record: 3, all the lane's own harness rounds, 2026-07-31. **If still true there is
   no live third-party exposure**, which sets RFC_020's urgency.
4. **`bugs_open/232` stays OPEN** until the page is re-rendered after the roll. Its
   own §"Not done" lists the steps, including that **there are TWO page-head producers
   and only one honours `pages.noindex`** — a rebuild through the other path silently
   drops the tag.

## 5. Traps paid for since the last handoff

- **A test that re-states its own condition cannot fail.** Two of mine did. Both are
  now paired with can-fail mutation controls.
- **Tests failing for the "wrong" reason exposed three real detection holes:** the
  allegation list had `stole/stolen/steals` but not **`steal`**, `plagiarised` but not
  **`plagiarise`**, `embezzled` but not **`embezzle`**. Inflection lists rot silently.
- **A council objection can be right about the wrong thing.** `reuse_agent` said I had
  reimplemented `ScanBannedClaims`; that part was wrong (it needs `site_specs`
  tools-api cannot reach) but the *search it demanded* surfaced a real bug — no
  negation handling at all, so a **defence** of a named person was refused.
- **My own recommendation was superseded by a better one on bug 232.** I proposed an
  out-of-repo Cloudflare rule; another session shipped a `pages.noindex` opt-in field,
  which is the same idea in the form owner ruling 2026-08-02 §2 requires. Read a
  co-worked bug file's *ordering* carefully — my correction landed after their fix and
  reads as though it supersedes it. It does not; there is a reconciliation note.
- **`LANDMINES.md` is appended by many sessions at once** — append with `>>`, never a
  whole-file `Write`. Mine was also swept into another lane's pathspec commit
  (`6f154d9b1`); nothing lost, but check HEAD before re-adding what you think is yours.
- **The council trigger wants `.plan` as an OBJECT.** An older submission in the tree
  has it as an array; copying that shape costs a refused dispatch.

## 6. Where everything lives

- Feed action + categories: `platform/orchestration/actions/provocation_feed_action.go`
- Allegation gate: `internal/tools-api/namecheck/` + `handlers/publish.go` + `handlers/ailog.go`
- Gate/generator (**other session's**): `provocation_gate_action.go`, `provocation_generator_action.go`, `builder/rollback_provocation.sh`
- Migrations: `282` pool, `283` publisher, `320` per-category index, `352` `pages.noindex` (other session's)
- RFCs: **`RFC_013`** (categories — RATIFIED on §2.1; §2.2/2.3/2.4 OPEN) · **`RFC_020`** (third-party harm — §5.1 done by another session, §5.2 built-not-live, **§5.3/§5.4 NOT BUILT**)
- Bugs: **`bugs_open/232`** (noindex — fixed, open until re-rendered)
- Councils: `6612dc0b` action · `bbbc9fca8` gate · `ccc32c3c` categories APPROVED · **`73dc4e78` namecheck APPROVED**
- Register: VONC-011 (categories + the presence-not-shape landmine), SEO-003 (other session's)
- Owner rulings: PLAN **§11.1** audience, **§11.2** risk brief + the naming test; RFC_013 §7; RFC_020 §7 (unfilled)
