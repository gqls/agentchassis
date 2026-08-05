# HANDOFF — provocation pipeline, 2026-08-05

**Supersedes `HANDOFF_2026-08-02_continue_here.md`** for "what to do next". That
file's traps section is still worth reading; its state table is now stale in one
important way (it says the generative half is unbuilt — it is built and committed).

> **⚠ THIS LANE HAS HAD TWO SESSIONS IN IT TODAY, WORKING DIFFERENT HALVES.**
> Another session built the gate + generator; this one did categories. Neither
> blocked the other, but **check the working tree and other sessions' live
> transcripts before assuming anything here is unclaimed** — `git log` and
> `scripts/who-owns.py` both read COMMITS and were blind to work in progress for
> most of today. See §5.

---

## 1. State in one paragraph

The **rotation** mechanism is live and has been since 08-02. The **gate and
generator** (PLAN Phases 2–3) were built, tested and council-approved today by
another session — committed, **not yet live**. **Categories** (PLAN §9.2) were
ruled on by the owner today and the **publisher half is committed, migration
applied, also not yet live**. The site is *still* serving a provocation dated
**26 Jul** under "Today's Provocation", because the pool still holds nothing
newer and the generator that would fix that has not rolled. **Nothing is
half-applied and nothing is broken; three separate pieces are simply waiting on
the same fleet release.**

## 2. What is true right now [VERIFIED 2026-08-05]

| thing | state | how checked |
|---|---|---|
| `render_provocation_feed` live | **yes**, v1.0.1238 | earlier pod-grep, unchanged |
| `provocation-feed-refresh` schedule | **enabled**, 21600s, fired 16:24Z 08-04 | `scheduled_tasks` |
| pool | **9 rows, newest `publish_on` = 2026-07-26** | `SELECT … FROM provocations` |
| served feed | `today.slug=nobody-wants-personalised-internet`, date **"26 Jul"**, 10,798 B, 0 escapes, 3 literal `<em>` | `curl https://vonc.com/data/provocations.json` |
| gate + generator | **committed, NOT live** — `e3ac4e15d`, council APPROVED `bbbc9fca8` | `git log` |
| categories (publisher half) | **committed, NOT live** — `40746962a`, council **pending** `ccc32c3c` | `git log` |
| migration 320 (per-category index) | **APPLIED and live** | `pg_indexes` shows the two new indexes, old two gone |
| `tools-api` category support | **does not exist and is not ours** | RFC_013 §2.2, unruled |

## 3. What to do next — in order

### 3.1 ~~READ THE COUNCIL VERDICT~~ — **DONE. APPROVED round 1** (`ccc32c3c`)

Approved with 2 advisory objections, none high-severity; `architecture` signal
`point_fix`, explicitly because the mechanism change had already been through the
RFC. Three objections were answerable by query and were answered (NOTES has the
detail): **exactly one consumer** of this action fleet-wide (one agent, one
scheduled row, vonc.com), the **ledger did record 320** for this file, and the
missing `BEGIN...COMMIT` was an artefact of the submission *sketch* — the real
migration has both.

**One objection is left open and it gates §3.3, so read it before seeding a
category:** `bug_historian` points out that vonc.com is Cloudflare-fronted, where a
refusal can be indistinguishable from origin behaviour by status code alone. The
bootstrap branch assumes a 404 means "the artefact does not exist". That is
UNEXERCISED LIVE. **Before the first category is seeded, confirm by hand that the
intended path 404s at the ORIGIN, not merely at the edge.** It is harmless until
then, because nothing reaches the branch.

### 3.2 The stale site is STILL the live defect, and it now has a built fix

Ten days stale. Two routes, and the first is the owner's call:

- **Content:** you supply provocation text, a session `INSERT`s it. Minutes. They
  publish as the owner's opinions under his name, so **a session must not invent
  them.**
- **The generator:** built and approved but needs (a) a fleet release and (b) the
  live-model calibration run the other session flagged as owed — `cmd/provocation-gate-calibrate`.
  Its own README/NOTES are the authority; **do not re-derive its state from here.**

### 3.3 Categories: what is left, and none of it is ours alone

The publisher can write `provocations-<category>.json` today. **Nothing reads it.**
Completing categories needs `tools-api` to learn which category a round argues,
which is RFC_013 **§2.2, unruled**, and the `gauntlet_dead_cta` lane's code. Also
still open: **§2.3** (should `gauntlet_rounds` record its category — cheap now,
unrecoverable later, because rounds publish to permanent public URLs) and **§2.4**
(should the contract become a shared Go type). They have been told: an `INCOMING`
block sits in their cold-start `HANDOFF_2026-07-31_continue_here.md`.

**Do not seed a second category expecting to see anything happen.** The runbook's
new section says why.

## 4. What was decided today, and the two calls worth challenging

### 4.1 One file per category (OWNER RULING, RFC_013 §7)

Not a style preference — it turns on a measured property of the reader.
`FetchProvocation` validates `today` for **presence only** (three checks,
`round.go:73-78`), returns it **unparsed**, and `position.go:67`/`defend.go:67`
paste the raw bytes **into the AI prompt**. So a keyed-map shape change would pass
every check and be *served*: the model argues against a blob, 200, no error
anywhere. One file per category fails loudly instead — a 404 the engine already
turns into a 503. `go list -deps` proves the two sides share no type describing the
feed, so nothing static will ever catch a divergence either.

### 4.2 A 404 permits a first publish ONLY when the built archive is empty

The one place this change makes a failure path *more* permissive. Without it a new
category could never publish at all. Narrowed so the guard stays armed wherever it
has something to protect: the live general feed has 8 archive entries, so a
spurious 404 there still refuses. Implemented as a named, tested `shouldBootstrap`
rather than inline, precisely because an inlined fail-open path is reachable only
through a real 404 and therefore never exercised.
**UNEXERCISED IN PRODUCTION** — unit-tested only, and stated as such in the
submission.

### 4.3 A filename contradicting its category is refused, not overridden

Because migration 283's live row passes `filename` explicitly, so copying it is the
obvious way to add a category — and would publish one category's provocation over
another's file, served as everybody's daily, undetected.

## 5. Traps this lane paid for TODAY (the earlier file's list still applies too)

- **`git log` cannot tell you whether work is in progress.** Both today's sessions
  were invisible to it and to `who-owns.py` for hours. What worked: `git status` /
  `ls -la` on the package, then grepping other sessions' live `.jsonl` transcripts
  for the symbol. Two `GateProvocationAction`s in one package would have been a
  **compile failure on shared HEAD**, not a conflict to sort out later.
- **A test that re-states the condition it is testing cannot fail.** The bootstrap
  table was first written inline and would have passed against any edit to the
  code. Fixed by extracting `shouldBootstrap` and adding a can-fail control that
  asserts the two variants genuinely disagree.
- **My landmine was swept into another lane's pathspec commit** (`6f154d9b1`)
  before I committed it. Nothing lost, forward-only holds — but check HEAD before
  re-adding something you think is yours and uncommitted.
- **`LANDMINES.md` is appended to by many sessions at once.** Append with `>>`;
  a whole-file `Write` would silently drop a concurrent append. One landed between
  my two edits to it.
- **The council trigger wants `.plan` as an OBJECT.** An older submission in the
  tree has it as an array; copying that shape costs a refused dispatch.
- **The migration runner's `--apply` takes EVERY pending file.** Scope it:
  `MIGRATIONS_DIR=<dir with only your file> ./scripts/migration/run-migrations.sh --apply`.

## 6. Where everything lives

- Action: `platform/orchestration/actions/provocation_feed_action.go`
  (categories: `feedFilename`, `validProvocationCategory`, `shouldBootstrap`)
- Gate/generator (**other session's — read their docs, not this file**):
  `provocation_gate_action.go`, `provocation_generator_action.go`,
  `builder/rollback_provocation.sh`
- Migrations: `282` pool, `283` publisher+schedule, **`320` per-category index**
- RFC: `architecture_review/RFC_013_per_category_provocations_and_a_contract_no_compiler_can_see.md` (RATIFIED on §2.1; §2.2/2.3/2.4 open)
- Runbook: `RUNBOOK_provocation_pipeline.md` § "Adding a CATEGORY"
- Register: VONC-011 (updated today, with the presence-not-shape landmine), VONC-002, VONC-003
- Councils: `6612dc0b` (original action, APPROVED), `bbbc9fca8` (gate, APPROVED),
  **`ccc32c3c` (categories, PENDING — §3.1)**
- Landmine: `LANDMINES.md#the-gauntlet-engine-validates-the-provocation-feed-s-today-key-for-presence-only`
  (verifier returned NEEDS_HUMAN_REVIEW; it confirmed the reader-side half and
  could not see the writer-side symbols because the code index is frozen at 07-28
  — `bugs_open/108`, not a doubt about the claim)
