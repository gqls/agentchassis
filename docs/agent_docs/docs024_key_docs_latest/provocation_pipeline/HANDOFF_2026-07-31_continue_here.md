# HANDOFF (2026-07-31) — provocation pipeline: COLD START, read this first

**This is the cold-start doc for the `provocation_pipeline` lane.** Origin:
`gauntlet_dead_cta/HANDOFF_2026-07-30_B` (marked TAKEN, commit `d4bb5fd47`).

Read in this order: this file → `PLAN_2026-07-31_provocation_pipeline.md` →
`SUMMARY_2026-07-31_provocation_pipeline.md`. The RUNBOOK has every command with
its gotcha; NOTES has the missteps.

---

## 1. The one sentence that matters

**The site still does not rotate.** "Every day, one provocation" is a **false
claim live on vonc.com right now**. Phase 0 made rotation *possible* and did not
make it *happen*: the schedule's last entry is 26 Jul and nothing rebuilds on a
cadence, so tomorrow serves exactly what today serves.

Everything below reads like progress. That sentence is the job.

## 2. State, measured 2026-07-31 ~16:30Z

| thing | state |
|---|---|
| Phase 0 feed (schedule-driven, real `generated_at`, `today.slug`+`date`) | **LIVE**, published 15:03Z, `9cc860ada` |
| rotation invariants (`verify_rotation.py`) | **PASS**, 39 dates, 9/9 distinct |
| publish + rollback (`publish_feed.sh`) | **WORKING**, both directions dry-run tested |
| paired-mode prototype | **BUILT**, `4f1f669ae`, local only, never deployed |
| the site actually rotating | **NO** |
| a `scheduled_tasks` row for provocations | **NONE** (still 0 rows) |
| Grok generator / the gate | **NOT STARTED** |
| Gauntlet lane's seal work | **committed** (`bb9719d3c`, corrected by `f331dcf9d`); **feed published**; **renderers NOT live** |

> **CORRECTED 2026-07-31 ~16:45Z, within minutes of writing the table above.**
> Two rows were already stale by the time this file was committed, which is the
> point of the warning in §3 rather than an exception to it.
>
> - The seal changes were **uncommitted WIP** when I measured at ~16:30 and were
>   **committed as `f331dcf9d`** before my own commit landed. `git status` on the
>   builder is now clean. *("Your session-start `git status` is a snapshot; it
>   goes stale within minutes" — measured at eight.)*
> - The seal feed **has since been published**: live `generated_at` is now
>   `2026-07-31T15:27:46Z`, and the feed **does** carry `seal` and `sample`, with
>   `arena.cards[0].title = "Sealed until you step in"`.
>
> **And the hazard in §3 has already materialised — it is the LIVE state.**
> Measured by rendering `https://vonc.com/` at ~16:45Z:
>
> | | |
> |---|---|
> | today's headline painted on home | **YES — the HANDOFF C leak is still open** |
> | today's body painted on home | **YES** |
> | lobby card | says **"Sealed until you step in"** |
>
> So the home page currently prints today's provocation in full *and* announces
> that it is sealed, on the same screen. Cause: the feed shipped ahead of the
> renderers — the served `assets/js/snippets.js` is unchanged (22,475 bytes, zero
> references to `sample` or `seal`) and still reads `today.headline`.
>
> **This is very likely TRANSIENT — the gauntlet lane is mid-delivery** (three
> commits in the preceding hour) and the renderer step is the next one in their
> own sequence. **Re-measure before acting on it; do not "fix" it and do not
> re-notify them.** If it is still true after their delivery lands, that is a real
> defect and belongs in their lane, not ours.
>
> `[UNVERIFIED]` — I did *not* establish whether the `sample` block renders. The
> string I matched ("AI will never be funny on purpose") is also an ordinary
> archive card in `arena.cards`, so its presence is not evidence either way.

## 3. ⚠ READ BEFORE YOU TOUCH ANYTHING — the builder is CO-OWNED and dirty

`builder/build_provocations.py` and `builder/verify_rotation.py` are edited by
**two lanes**. Our lane owns **rotation**; the gauntlet lane owns **the seal**.
Same file, two concerns.

> **CORRECTED ~16:45Z:** an earlier draft of this section said those two files
> held *uncommitted* changes. They did when measured; they were committed as
> `f331dcf9d` minutes later and the tree is now clean. **The instruction below is
> unchanged and still applies — assume that file is dirty with someone else's
> work until you have just looked**, because on this tree it was true twice in
> one afternoon.

- **Re-run `git status` immediately before committing.** Never `git add -A`.
  Commit with an explicit pathspec naming only the files your task touched.
- **Do not revert seal-related changes.** They are a correction to a real
  near-miss (§4) and they pass all our rotation invariants — verified.
- Coordinate rather than compete:
  `gauntlet_dead_cta/HANDOFF_2026-07-31_continue_here.md` is their cold start.

### ⚠ PUBLISHING IS NOW A HAZARD, not a one-liner

`./publish_feed.sh` is one command and it publishes **whatever the builder
currently emits** — which now includes the seal work. **The renderers for it are
not live.** Verified: the served `assets/js/snippets.js` still reads
`today.headline` and contains **zero** references to `sample` or `seal`.

Publish today and you get a half-applied site: the lobby card says *"Sealed until
you step in"* while the page above it still prints today's provocation in full.
Not broken, but incoherent and visible.

**So: agree delivery order with the gauntlet lane before publishing anything.**
If you need to publish a rotation change urgently and their renderers are not
ready, build from a checkout of the committed builder rather than the dirty tree.

## 4. The near-miss worth knowing about (it is our landmine, fired)

The seal's first implementation (`bb9719d3c`) **removed `today.headline` and
`today.body` from the feed** to stop the leak — on the reasoning that the Gauntlet
page does not fetch the feed. The page does not; **the ENGINE does.**
`internal/tools-api/handlers/round.go` `FetchProvocation` takes the whole `today`
object server-side and persists it as the round's provocation. Stripping those
keys would have served **every round an empty question.**

The in-flight correction restores them and enforces the seal as a *renderer-level*
invariant with a structural guard in the builder. This is precisely the landmine
filed that morning ("the feed is read by the SERVER too"), and it fired in under a
day, in the same file, against a different session.

**Standing rule for this feed: sealing is a DISPLAY concern. Never seal by
emptying `today`.**

## 5. Landmines (all in `LANDMINES.md`, footprinted)

1. **`round.go` reads `today` server-side** (5-min cache). Selection must happen
   at generation time. No client-side date selectors. See §4.
2. **`generated_at` — half fixed, and the remaining half is the dangerous half.**
   Our builder computes it; **the OLD builder at
   `gauntlet_dead_cta/p4_sources/build_provocations.py` still hardcodes it and is
   still runnable**, so running that one silently reverts the timestamp while
   appearing to succeed.
3. **A real `generated_at` proves REBUILD, not ROTATION.** Once a daily job
   exists the timestamp advances every day whether or not the provocation changed
   — the original bug wearing the fix as a disguise. **For "did it rotate", diff
   `today.slug`. Never `generated_at`.**
4. **`sites.github_repo` is EMPTY for vonc.com.** The DB cannot tell you the
   deploy target. It is `gqls/sites:vonc.com/data/provocations.json` — prove it by
   `cmp`-ing the blob against the served bytes before writing (RUNBOOK §7).
5. **`--dump-dom` is the DOM, not what is painted.** `hidden` attributes and
   `display:none` leave text fully present. Print match context; a bare `grep -c`
   answers a different question.

## 6. Owner decisions already taken — do not re-open these

- **Archive rule:** a provocation is archived when the next is published, never
  during its own day. (Implemented as a property of the schedule.)
- **Source:** Grok, for topics. Already wired — `feed_actions.go`
  `resolveLLMNewsProvider` supports `xai`/`grok` against the Responses API with
  `web_search` **and `x_search`**, model `grok-4-1-fast`.
- **No human approval.** The owner read the recommendation and chose speed. It is
  decided; do not re-litigate it. What it *obliges* is PLAN §10 and that is not
  optional.
- **Categories** sooner rather than later, politics→pets, each with its own
  audience.
- **First audience:** the r/changemyview / HN / tech-X axis — a *calibration*
  audience, not the destination. `[UNMEASURED]`: a hypothesis about reposting
  behaviour with nothing posted yet.
- **Paired mode:** the four design decisions are agreed (organiser cannot read
  positions; non-responders do not receive the reveal; commit is final; reveal is
  atomic).

## 7. What to do next, in order

**A. Make the claim true.** Add a bridge of hand-written provocations to
`SCHEDULE` (author BOTH shapes — `headline`/`body` and `title`/`teaser`/
`detail_body`; the eight historical entries lack the today-shape and fall back),
then add a `scheduled_tasks` row that rebuilds and republishes daily. Reuse the
news pipeline's shape — it ends in the `git_commit` action and is proven live.
**Sequence with the gauntlet lane (§3).**

Verification is `today.slug` changing across two days, **not** the page looking
fine and **not** `generated_at`.

**B. The gate, then the generator.** In that order — the gate must exist and be
calibrated before anything generated can publish, because there is no human
backstop. PLAN §4 (thesis exempt from the claims rail, body's factual assertions
not) and PLAN §10 (fail closed; gate errors are rejections; log rejections;
tested rollback; one publish per day; calibrate against the 9) are the spec.
This is platform code → council gate.

**C. Categories.** Cheap now, expensive later — but note they **break the
engine's one-`today`-per-site contract**, so that is a conversation with the
gauntlet lane before it is a task (PLAN §9.2).

**D. Paired mode.** Prototype is built and waiting. Its real prerequisite is
identity, which the platform does not have at all — the public game keys rounds
on a hash that `bugs_open/139` measured as a constant across all 83 rows.

## 8. Quick commands

```bash
cd docs/agent_docs/docs024_key_docs_latest/provocation_pipeline/builder
python3 build_provocations.py --date 2026-08-02 --check   # what would today be?
python3 verify_rotation.py                                # invariants, 39 dates
./publish_feed.sh --dry-run /tmp/feed.json                # never skip the dry run — see §3
cd ../prototype && go test ./... && go run .              # paired mode, :8099
```

```sql
-- is anything pacing it yet? (0 rows = still no rotation)
SELECT name, enabled, interval_seconds, last_triggered_at
FROM scheduled_tasks WHERE name ~* 'provocation';
```

```bash
# did it ACTUALLY rotate? the only honest check
curl -s https://vonc.com/data/provocations.json \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['today']['slug'])"
```
