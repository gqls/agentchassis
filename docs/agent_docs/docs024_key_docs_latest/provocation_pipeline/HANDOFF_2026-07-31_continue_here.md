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

## 1b. LATEST STATE — 2026-07-31 ~22:15Z. Read this before the table below.

The mechanism the rest of this document says is missing **now exists**. What is
still missing is content and one image roll.

| piece | state |
|---|---|
| `provocations` pool table + the 9 live entries | **LIVE** — migration 282, applied + recorded |
| `render_provocation_feed` action, 14 tests | **COMMITTED** `572ae8dc6`; council `6612dc0b` |
| parity with the Python builder, from the 9 real rows | **PROVEN IDENTICAL** (+ a test proving the comparison can fail) |
| `provocation-feed-publisher` agent + `provocation-feed-refresh` schedule | **SEEDED** — migration 283, row deliberately `enabled=false` |
| the action in the running chassis | **NOT THERE** — `grep -c render_provocation_feed` = **0**, control `render_news_section` = 3 |
| **the site rotating** | **STILL NO** |

**Two things stand between here and the claim being true, in this order:**

1. **A chassis image carrying the action.** Not built here on purpose: a roll ships
   every other lane's committed HEAD, which is not one thread's call, and it would
   have killed the in-flight council run. Any other session's roll ships it anyway.
   Then verify ON THE POD (RUNBOOK, with the positive control) and flip
   `enabled = true`.
2. **Provocations dated forward from today.** The newest pool entry is 26 Jul, so
   even with the job running the site serves that same provocation. **This is now
   the whole remaining gap, and it is content, not machinery.**

Read `SUMMARY_2026-07-31b_provocation_pipeline.md` for the prose version, and
NOTES for the three design calls (no stored `published` flag; duplicate dates
refused by an index, not a check; the commit is SKIPPED when only `generated_at`
would move).

## 2. State, measured 2026-07-31 ~16:30Z — SUPERSEDED by §1b, kept for the trail

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

> **RESOLVED 2026-07-31 ~19:25Z — the transient state closed, as predicted, and
> the `[UNVERIFIED]` above is now answered.** The gauntlet lane delivered the
> renderers (`4f6af298e`). Re-measured by rendering `https://vonc.com/`:
>
> | | |
> |---|---|
> | served `snippets.js` | **25,532 bytes** (was 22,475), **11** refs to `seal`, **14** to `sample` |
> | today's headline painted on home | **NO — 0 matches.** The leak is closed |
> | today's body painted on home | **NO — 0 matches** |
> | lobby card | "Sealed until you step in" |
> | the `sample` block | **RENDERS** — `<span class="pc-eyebrow">A past provocation · 5 Jul</span>` |
>
> That eyebrow is the discriminator the earlier check lacked: an ordinary archive
> card cannot produce it. **`[UNVERIFIED]` → verified, by finding a string only
> the new renderer emits** rather than one both could.
>
> **The publish hazard in §3 is therefore CLEARED.** Feed and renderers now agree.
> The lane also left an inbound note in `NOTES_provocation_pipeline.md`
> (`eb700f293`) — read it; it documents the seal design and one trap for our
> scheduled job (**a leak check must derive its probes from `today` at run time**;
> theirs hardcoded the card text and began reporting the seal itself as a leak).

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

### ⚠ PUBLISHING IS NOW A HAZARD, not a one-liner — **CLEARED ~19:25Z, kept for the reasoning**

> **CLEARED 2026-07-31 ~19:25Z.** The gauntlet lane's renderers are live
> (`4f6af298e`) and the served `snippets.js` now understands `seal`/`sample` — so
> the specific incoherence below no longer exists. Evidence in §2's second
> correction. **The rule outlived the hazard**: the builder is still co-owned, and
> `publish_feed.sh` still publishes whatever the builder currently emits, so
> "check what the tree contains before publishing" stands. Read on for why.

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

**A. Make the claim true.**

> **CORRECTED 2026-07-31 ~19:30Z — the original step A could not have worked, and
> a thread following it would have found out only after writing the content.**
> It read: *"Add a bridge of hand-written provocations to `SCHEDULE` … then add a
> `scheduled_tasks` row that rebuilds and republishes daily."* Those are not two
> halves of one task. **A `scheduled_tasks` row has nothing to dispatch to.**
> Measured:
>
> - `SELECT type FROM agent_definitions WHERE type ~* 'provoc|gauntlet'` → **0 rows.**
> - `grep -rn "build_provocations" --include=*.go --include=*.yaml --include=*.sql`
>   outside `docs/` → **nothing.** The only cluster-side consumer of the feed is
>   `round.go`, which *reads* it.
> - The schedule lives in a **Python file under `docs/`**. Nothing in the cluster
>   can execute it, and `make build-*` does not ship it.
>
> So adding entries to `SCHEDULE` changes **nothing about the live site** until a
> human runs `publish_feed.sh` by hand — which is a person, not a mechanism, and
> is the same failure the workstream exists to fix.

> **BUILT 2026-07-31 ~22:15Z — steps 1–4 below are DONE.** See §1b. What remains
> of step A is the image roll, the `enabled = true` flip, and step 5: content.
> The list is kept because the dependency order is the reusable part.

**What daily rotation actually requires**, in dependency order:

1. **The pool moves to the DB.** A Python literal is unreachable from the cluster.
   (New table, or a `site_specs` aspect — undecided, see below.)
2. **A Go action** that selects today by publish date, builds the feed, and
   commits it. **Template: `directory_export_action.go:113`** — query DB → marshal
   JSON → `sendExportFilesToGit(...)`. That sender (line 478) is **already shared
   by two exporters**, so this is a third consumer of proven machinery, not new
   plumbing. → **platform code, council gate.**
3. **An agent definition**, one step + `complete_workflow` — copy
   `directory-json-exporter`, which is exactly this shape.
4. **A `scheduled_tasks` row** (`target_agent_type` = the new agent). Only now
   does this row have a meaning.
5. **Content** — the bridge entries. Cheapest, and *last*, because until 1–4 exist
   it is unpublishable prose.

⚠ **The Go action must carry the seal invariants across, not just the rotation
ones.** The Python builder now enforces both (`check_seal()` refuses in both
directions). A Go rebuild that ports only rotation **silently reopens the leak the
gauntlet lane closed today.** Port `verify_rotation.py`'s invariants as Go tests;
they are the specification, and they cost nothing to translate.

Verification is `today.slug` changing across two days, **not** the page looking
fine and **not** `generated_at`.

### Why the cheap design is foreclosed — do not propose it again

The obvious way to avoid all of the above: publish the **whole schedule** once and
let both readers select by today's date. No job, no action, no table, and rotation
becomes a property of the data that cannot silently stop.

**It is ruled out by the seal.** Publishing N days ahead puts every future
provocation in a world-readable file, and the owner's 2026-07-31 ruling is that
today's is not readable until you step into the round. Tomorrow's certainly is not.
Withholding the future entries collapses the design straight back into a daily
republish.

Worth recording because the two lanes' rulings interact: **the seal is what makes
the daily job mandatory.** It is a real cost of that ruling, not an oversight in
this plan.

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
