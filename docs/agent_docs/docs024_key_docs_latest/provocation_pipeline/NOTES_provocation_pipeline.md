# NOTES — provocation pipeline

Append-only, newest at the bottom. Missteps are the point of this file.

---

## 2026-07-31 — session 1, picking up HANDOFF B

**Task:** `gauntlet_dead_cta/HANDOFF_2026-07-30_B_the_daily_provocation_is_not_daily.md`.

### Re-verified the handoff's facts before acting on them

All still true 2026-07-31: served file `generated_at: 2026-07-26T00:00:00Z`,
9,797 bytes, HTTP 200; `scheduled_tasks` matching `vonc|provocation` = **0 rows**;
`build_provocations.py` has no date arithmetic. Archive = 8 entries, 28 Jun – 5 Jul.
`today` keys are `eyebrow/headline/body/primary_cta/secondary_cta/stats` — **no
`slug`, no `date`**, confirming it cannot join the archive as-is.

### Finding 1 — the server reads `today` [VERIFIED]

`internal/tools-api/handlers/round.go:44`, `FetchProvocation(domain)`: fetches
`https://{domain}/data/provocations.json`, requires `today`, caches per-domain for
`provocTTL = 5 * time.Minute` (line 25), 503s if absent (line 106).

This matters more than it first looks. I had already half-formed the "obvious"
cheap fix — ship a pool inside the JSON and pick by date **in the client**, since
the pages fetch it client-side anyway. That is wrong, and would have produced a
subtle, intermittent, extremely annoying bug: the page shows provocation N while
the Gauntlet argues provocation M. **Recording it because it is the design a
reasonable person reaches for first and nothing in the handoff warns against it.**

### Finding 2 — the publish plumbing already exists [VERIFIED at the artefact]

`content-feed-refresh`: enabled, 21600s, last triggered 2026-07-30 13:53:54Z,
completed 14:02:29Z. Ends in `git_commit` (`registry.go:509`) with `files_field`
→ commits rendered JSON to the site repo for S3 deploy.

Proved it end-to-end rather than trusting the row — `latest-news.json` served by
`dartsonline.com` (13:56:56Z), `relojistas.com` (13:58:07Z), `webdesign.co.uk`
(14:01:03Z), all inside that window. `gaswholesalers.com` at 07:57:53Z is the
previous 6-hourly run, which is itself corroboration.

⇒ Handoff B costs option 2 ("a scheduled regeneration") as the middle option and
its landmine describes hand-publishing via `gh api --input -`. **The platform path
is an off-the-shelf action and the hand path should not be automated.**

### Finding 3 — prior art, half-built [VERIFIED]

`docs/social001_vonc_tiktok_social/PLAN_spark_provocation_pipeline.md`, dated
2026-06-25, designs this exact thing as a news-pipeline clone. Phases 1–2 shipped;
`agent_definitions` has no `provocation-generator` or `provocation-orchestrator`,
so Phases 3–4 never did.

Neither handoff B nor the four sibling handoffs mention this plan exists. I found
it by grepping `provocations.json` across the repo, which I only did because I
wanted the consumer list — the prior art was a side effect. **Grep for the
artefact, not just the concept: the concept had been renamed ("Spark") but the
filename had not.**

### MISSTEP 1 — I nearly filed a working page as broken

Handoff B says `/provocations/index.html` "paints neither today's provocation nor,
apparently, much else (1,293 chars of visible text)" and suggests it may be broken
independently. I set out to confirm that.

Rendered it with `/snap/bin/chromium --dump-dom --virtual-time-budget=9000`
(playwright is not installed). First pass, `grep -c` per entry title: all 8
present, plus `"Nothing filed yet"` (the empty state) **and** `"Nobody actually"`.

I read that as: the empty state is showing AND today's headline is leaking onto
the archive page. **Both readings were wrong.**

Printing 600 chars of context around each match showed:

- `Nothing filed yet` sits in `<p class="provocations-archive__empty" hidden="">`
  — present in the DOM, not painted. `--dump-dom` gives the DOM, which is a
  different question from what a visitor can read. This is the *same* landmine
  `HANDOFF_2026-07-30_C` records, and I walked into it anyway, one file away from
  where it is written down.
- `Nobody actually` was the **29 Jun archive entry**, "Nobody actually reads terms
  of service — and that's rational". Today's headline is "Nobody actually *wants a
  personalised internet*". A 15-character substring matched two different
  provocations. **A substring is not an identification.**
- The blank leading row I then suspected turned out to be
  `<a class="provocations-archive__item" data-archive-template="" hidden="">`,
  with a `[data-archive-template] { display: none; }` rule as well. Correct by
  design, twice over.

**Conclusion, which is a correction to handoff B:** the archive page is **not
broken**. All 8 entries paint; 7 linked, 1 deliberately non-openable (no
`detail_body`, the builder's documented Journey B.3). 1,293 chars is what a
working page of 8 short entries measures. The handoff's figure came from a probe
looking for **today's** headline, which correctly is not there — the number was
right and the inference from it was not.

**Cheap check that would have caught it first time:** print the match context
before drawing any conclusion from a count. Now written into
`RUNBOOK` §4 as gotchas 2 and 3.

### MISSTEP 2 — asked the owner a question containing a claim I had just refuted

I put a decision question to the owner whose second half was framed on
`/provocations/index.html` "already paints all 8 entries with their full case
text" — true — but I had drafted the same question minutes earlier while still
believing the page might be broken. The framing survived the correction because I
edited the facts and not the sentence built on them. The owner rejected the
question for clarification, which was the right call for other reasons, but
**a superseded premise had already made it into an outward-facing artefact.**
Re-read questions after any correction lands, not just the notes.

### Decisions taken by the owner this session

- **Archive rule settled**: "It can be archived when the new one is published" —
  i.e. at the next rotation, never during its own day. This also happens to close
  the leak path that immediate promotion would have opened on the archive page
  (`HANDOFF_2026-07-30_C`).
- **Direction: LLM generation behind a gate** checking (a) safe, (b) interesting,
  (c) current/relevant to audience, (d) other good-provocation properties.
- **Images later**, celebrities possible, legal route to be found.
- **New shape raised: the paired provocation** — private, team-based, organiser-set.

### On the gate — the distinction that shapes it [INFERRED, from reading the corpus]

A provocation is a *deliberately contestable* assertion, so the claims rail would
reject all 9 of ours. But the *body* carries ordinary factual claims that are
fully subject to it — the four-day-week entry asserts the pilots "measure
self-reported output", which is simply true or false. **Thesis exempt, supporting
facts not.** Marked INFERRED because I have derived it from reading the 9
provocations, not from any measurement; it should be tested by running a
prototype gate over the corpus and seeing whether it passes all 9.

`bugs_open/149` C1 row 2 is the same failure in the wild — "10,000 Monte Carlo
trials per query" for a tool that computes the binomial analytically and calls
`Math.random` nowhere. A factual claim smuggled into prose, deployed with nothing
objecting. `bugs_closed/043` is the earlier instance of the class.

### [UNMEASURED] claims I am carrying into the plan

- That the 9 existing provocations are a sufficient specification for the gate.
  Untested — the test is whether a prototype gate passes all 9 and rejects
  deliberately bad samples.
- That a human approval queue "costs a few minutes a day". Never timed.
- That paired-provocation engagement data would make criteria (b)/(c) measurable.
  Plausible and it is the owner's own reasoning, but nothing has been built to
  check it.

---

## 2026-07-31 — session 1 continued: Grok, categories, and the paired prototype

### Grok integration verified before repeating the owner's claim [VERIFIED]

Owner said we already query Grok for news and have a key. True:
`platform/orchestration/actions/feed_actions.go:733` (`resolveLLMNewsProvider`)
handles `xai`/`grok`, reads `XAI_API_KEY` then `GROK_API_KEY`, and targets
`https://api.x.ai/v1/responses` — the **Responses API with `web_search` and
`x_search` tools**, doc-recommended model `grok-4-1-fast`.
`platform/agentenv/provider_keys.go:53` carries the key into agent environments.

So Phase 3's *sourcing* half is largely reuse. **That is the second time this
workstream has costed something as "build" and found it built** (the first was the
scheduled publish path). Noting the pattern rather than the instance: on this
platform, check before costing.

### The categories/`round.go` interaction — found by re-reading my own landmine

Owner wants categories "sooner rather than later". Went to write that up as a
straightforward data-model note and realised it collides with the constraint I
filed as a landmine this morning: **`round.go` requires exactly one top-level
`today` key per domain.** Several simultaneous category dailies cannot be
expressed in that shape at all — it needs either one file per category or a
change to `today`, and `tools-api` must move in step or the page and the engine
disagree about what is being argued.

Worth recording *how* I caught it: not by analysis, but because I had written the
landmine down four hours earlier and it was still in working memory when I touched
an adjacent design. The same constraint would have been invisible to a thread that
picked up "add categories" cold. That is the argument for `LANDMINES.md` in one
example.

### MISSTEP 3 — my atomicity test was green against a broken implementation

Wrote the paired prototype with the seal enforced structurally (two view types,
`SealedView` has nowhere to hold a peer's position), plus 14 tests written to break
it. All green, `go vet` clean.

Then mutation-tested, per the standing rule that a passing test may be passing
because the rule is absent. Three mutations:

| mutation | caught |
|---|---|
| `SealedPeer` gains a `Position` field, populated | yes — 3 tests |
| non-responders receive the reveal | yes — 2 tests |
| `RevealedAt` stamped per read instead of per session | **NO** |

The third is the interesting one. `TestRevealIsAtomic` read all three
participants' views at the **same** `now`, so a per-read timestamp and a
session-wide one produce byte-identical output. **The test could not distinguish
the property it was named after.** It asserted the three stamps equalled *each
other* — which is trivially true when you hand all three the same clock.

Fixed by reading at three *different* times and asserting each stamp equals the
moment the last participant committed. Re-ran the mutation: now fails on all three
readers, with the message printing both the observed stamp and what a per-read
stamp would have returned.

**The cheap check that would have caught it at writing time:** ask of every
assertion "what input variation is this test's subject supposed to be invariant
*to*?" — and then vary it. An atomicity test that holds the clock constant is
testing nothing about time. Same family as
[[check-answers-the-question-you-encoded]]: the test answered the question I
encoded (are these three values equal) rather than the one I meant (did the reveal
happen once).

### Verified the prototype at the artefact, not the build

`go build` and `go test` prove nothing about the running thing. Drove the HTTP
flow with curl against the live server:

- 3-person session, Alice and Bob commit
- Carol's page: **0** occurrences of either position, and correctly "2 of 3 committed"
- organiser's page: **0** occurrences of either position
- Carol commits → all three pages show all three positions

### [UNMEASURED] carried forward

- The audience recommendation (r/changemyview / HN / tech-X) is reasoned from the
  shape of the existing 9 provocations and from where the exchange card would read
  as native. **No measurement of any kind.** It is a hypothesis about reposting
  behaviour and should be labelled as such until something is actually posted.
- "Grok supplies topics, not provocations" is a design position, not a finding —
  no Grok call has been made from this workstream yet.

---

## 2026-07-31 — Phase 0 built: the builder is schedule-driven

Owner agreed the four paired-mode decisions and ruled **no human approval for
now**. Recorded as taken; PLAN §10 states what an unreviewed publish path
obliges instead (fail closed to yesterday's provocation, gate errors count as
rejections, decisions logged, rollback written before first publish,
calibration against the 9 no longer optional).

### The new builder, and why it lives in this lane

`provocation_pipeline/builder/build_provocations.py`. The original stays where it
is in `gauntlet_dead_cta/p4_sources/` — it produced the live file and that is a
fact about history, so calling it "superseded" would be false until something
else has actually published. Also: that lane is actively working (they shipped
the share card this morning), and a new file in my own directory cannot collide
with theirs.

Selection is a **schedule**: each entry carries a `publish` date, `today` is the
latest entry whose date has arrived, and the archive is everything published
strictly before it. That implements the owner's rule exactly — an entry is
archived the moment a later one is published and never during its own day — as a
property of the data structure rather than a step someone has to remember.

### Verified behaviour-preserving before anything else [VERIFIED]

The strongest check available for a refactor: run it for 2026-07-31 and diff
against the **served** file. Result — identical, apart from the computed
`generated_at` and the two additions (`today.date`, `today.slug`) that are the
entire point.

That diff caught a real regression I had introduced and would otherwise have
shipped: I derived `arena.cards[0].desc` from the archive teaser, but the live
card carries a longer hand-written blurb. Deriving it would have silently
reworded live copy under cover of a "no behaviour change" refactor. Fixed with an
optional `card_desc` on the schedule entry, so the card is still derived from one
source of truth and the live string is preserved.

> **Note the near-miss in my own method.** My first comparison checked
> `card[i].title` and reported SAME for all six cards. The titles *were* the
> same; `desc` was not, and I had not looked at it. **A field-by-field check is
> only as good as the field list**, and I had chosen a list that happened to
> exclude the one I had changed. Caught by then diffing whole documents instead
> of chosen fields. Same family as this morning's entry — a check that answers
> the question you encoded rather than the one you meant.

### `verify_rotation.py` — assert the mechanism, not a day's output

39 dates across the schedule span plus 10 days past the end. Invariants: today is
never also archived; today always has slug and date; the arena's first card
matches today; the archive is exactly the earlier entries, newest first; the
archive grows by exactly one when today changes and not at all when it does not;
9 distinct provocations appear in schedule order; and a date before the schedule
starts must FAIL rather than invent one. All hold.

### MISSTEP 4 — my verifier passed a mutation of exactly the class I filed this morning

Mutation-tested the verifier by breaking the builder six ways. Four caught
immediately. Two did not behave:

- **`today.date` frozen to a literal — NOT CAUGHT.** The verifier asserted the
  field was *present*, never that it was *right*. And `today.date` is what gets
  carried into the archive on promotion, so a frozen value would date every
  archived entry identically while looking completely plausible.
  **This is the same defect I filed in `LANDMINES.md` this morning** — the
  hardcoded `generated_at` whose freshness check reads a literal. I wrote the
  landmine, then wrote a checker with the identical blind spot four hours later,
  in the same file it was about. Knowing a failure mode is not the same as
  checking for it. Fixed: assert `today.date` and every archived entry's date
  equal the scheduled date, with the short-date formatter **re-derived in the
  verifier** rather than imported — a verifier that formats dates with the code
  under test cannot detect a formatting change, only agree with it.
- **Slug deleted — "caught", but by crashing.** `KeyError` before the check that
  was supposed to report it. Non-zero exit, so a mutation test scores it as
  caught while the operator learns nothing. Fixed by ordering the presence check
  first and skipping the slug-dependent checks for that date.
  **Worth generalising: "the mutation test went red" is not the same as "the
  checker reported the defect".** Read the output, not the exit code.

All six mutations now produce a named, dated failure line.

### [UNMEASURED] / not done

- Nothing has been published. The new builder's output has never been served;
  everything above is local. Publishing is an outward-facing change and is the
  owner's call.
- The schedule still has 9 entries and the last publishes 26 Jul, so on today's
  date the output is *identical to the stale live file*. **Phase 0 makes rotation
  possible; it does not make the site rotate.** That needs new entries or the
  generator, plus the scheduled task.

---

## 2026-07-31 — Phase 0 PUBLISHED and live

Owner authorised publishing. Done, verified, no regression.

### Target proven before writing [VERIFIED]

`sites.github_repo` is **empty** for vonc.com, so the DB could not confirm the
deploy target and the "wrong repo succeeds silently" landmine was live. Proved it
instead: fetched the `gqls/sites` blob and compared with the served bytes —
both 9,797 bytes, md5 `b6d1e766…`, `cmp` identical. That repo path *is* the file
the site serves.

### MISSTEP 5 — my rollback was blocked by my own preflight

Wrote `publish_feed.sh` as both the publish and the rollback path, on the
reasoning that a revert only ever run in an emergency is a revert nobody knows
works. Then dry-ran the rollback against the pre-Phase-0 backup.

**It refused.** The preflight required `today.slug`, and the pre-Phase-0 live file
has no slug — that being exactly the defect Phase 0 fixes. **The escape hatch was
gated on the improvement it exists to undo.** Had that shipped, discovering it
would have happened at the worst possible moment.

Fixed by splitting the preflight into two tiers: **safety** checks run on every
path (fields the live loader reads, `today` present, no duplicate slugs, today not
also archived — failing these is an outage, since `round.go` 503s without
`today`), **quality** checks are skipped under `--rollback`. Verified all four
combinations: forward passes, rollback passes with quality skipped, rollback file
without the flag is still refused, and a feed with `today.headline` removed is
refused on **both** paths.

**Transferable:** a guard on a publish path must be tested against the *oldest*
artefact you might need to restore, not only the newest you intend to ship. And
this was found by *running* the rollback, not by reading it — the reasoning that
produced the bug would have kept producing it.

### Published [VERIFIED at the artefact]

`gqls/sites:vonc.com/data/provocations.json`, 9,797 → 9,869 bytes. Served and
matching **~45 s** after the PUT (the script polls until the served md5 equals
what it pushed, rather than trusting the PUT's 200).

Live feed now: `generated_at 2026-07-31T15:03:31Z` (computed, not a literal, for
the first time), `today.slug=nobody-wants-personalised-internet`,
`today.date=26 Jul`, archive 8.

### Regression check — rendered, not grepped

- **home**: headline still paints in `.pc-headline`, body still paints. Unchanged.
- **archive**: 8/8 entries, 7 linked + 1 deliberately non-openable, empty state
  `hidden`, and **0** occurrences of today's slug or body — the owner's archive
  rule holding on the live site.
- **gauntlet**: `gi-sealed` present, **0** occurrences of the headline or body.
  The seal is intact; publishing did not open a leak.

Safe because the live loader reads only six fields
(`eyebrow/headline/body/primary_cta/secondary_cta/stats` — confirmed by grepping
the **served** `assets/js/snippets.js`, not the repo copy) and all six are
byte-identical to before. `date` and `slug` are additive and read by nothing.

### What is STILL NOT TRUE, and must not be overstated

**The site does not rotate.** It now *can*: the builder is date-driven, the feed
has the fields archiving needs, and publishing is one verified command. But the
schedule's last entry is 26 Jul, and **nothing rebuilds or republishes on a
cadence** — so tomorrow the site will serve exactly what it serves today.

**"Every day, one provocation" is still a false claim on the live site.** Phase 0
removed the structural obstacles; it did not fix the defect this workstream
exists to fix. Two things are outstanding and both are required: provocations to
rotate *to* (Grok generator, or hand-written entries as a bridge) and a
`scheduled_tasks` row that rebuilds and republishes daily.

---

## 2026-07-31 (evening) — INBOUND from the gauntlet_dead_cta lane: your builder now enforces the seal

**I changed two of your files.** Owner ruling 2026-07-31 on
`gauntlet_dead_cta/HANDOFF_2026-07-30_C`: today's provocation must be readable in the
Gauntlet, after entry, and **nowhere else** — home and Arena show a PAST provocation in
full as the worked sample instead. Your builder is the only place that can enforce it,
so it went there rather than into a competing generator. Live and verified: 0 of 19
pages paint today's provocation (was 3), and `POST /round` still returns it.

**`build_provocations.py`**
- `as_today()` **untouched** — your engine contract is intact and now asserted.
- Two new top-level keys, deliberately SIBLINGS of `today` rather than fields on it,
  because `RoundHandler` persists the whole `today` object as the round's provocation
  and anything added there would land inside every stored round: `seal` (display copy)
  and `sample` (a past provocation in full, derived from the newest archived entry with
  a case, so it follows your schedule and needs no edit when rotation lands).
- `arena.cards[0]` is now the **sealed** card. This is the change to your design worth
  knowing about: your docstring point 5 makes card 0 derived from today, which is right
  for rotation and was **2 of the 3 leaking surfaces**, since both the home lobby grid
  and the Arena lobby render it. Sealing it there fixed both with no JS change.
- New `check_seal()` refuses to emit in BOTH directions, because they pull opposite
  ways: `today` MUST keep headline/body/slug/date, and nothing outside `today` may name
  today's provocation. Induced-fault tested both ways.

**`verify_rotation.py`** — your invariant "arena card 0 MUST match today" is
**INVERTED** (it encoded the pre-seal design) and extended with the engine contract and
the seal-copy-exists check. All 39 dates × 9 provocations still pass.

### The thing I owe you, because your own work nearly caught my mistake and I bypassed it

My first attempt enforced the seal by **deleting `today.headline`/`today.body`** from
the feed, on the premise that the Gauntlet page never fetches it (true — verified by
request interception). **Your `publish_feed.sh` safety preflight would have refused
it**, and so would `verify_rotation.py`. Neither fired, because I was about to publish
through my own path. What caught it was diffing the live file against my generator's
output first — the `today.date`/`today.slug` your Phase 0 added are fields my
(superseded) generator could not produce, and pulling that thread led here and to
`round.go`. **Your two-tier preflight and your rollback-is-the-same-script decision
both did their job on me from a distance; the split between safety and quality checks
is exactly why the seal change passed the right gate.**

Also: your README says of the 15:03 publish that "the home page still shows the
provocation … Nothing broke." That was accurate and correctly scoped to rotation — the
leak was pre-existing, not yours. Flagging only so the line is not later read as a
clean bill on the seal.

**Nothing of yours is blocked by this.** The two things your README names as still
outstanding — provocations to rotate to, and a scheduled rebuild — are untouched, and
the sample/seal keys need no maintenance when they land. One trap for the scheduled
job, on top of the timestamp one you already filed: **the leak check must derive its
probes from `today` at run time.** My sweep hardcoded the lobby card text and started
reporting a leak for finding the seal itself.

Instrument, if useful: `gauntlet_dead_cta/scripts/provocation_leak_sweep.py`
(renders every active page; exit 0 clean / 1 leak / 2 incomplete).

---

## 2026-07-31 (evening) — my own step A was not executable, and I only found out by pricing it

Picked the lane back up and started on HANDOFF step A: *"add a bridge of
hand-written provocations to `SCHEDULE`, then add a `scheduled_tasks` row that
rebuilds and republishes daily."* I wrote that sentence this afternoon. **It
cannot work**, and the reason is not subtle:

```
SELECT type FROM agent_definitions WHERE type ~* 'provoc|gauntlet'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- 0 rows

grep -rn "build_provocations" --include=*.go --include=*.yaml --include=*.sql . | grep -v '^./docs/'
-- nothing
```

`scheduled_tasks.target_agent_type` is NOT NULL. **A row needs an agent to
dispatch to, and there is none.** The schedule lives in a Python literal under
`docs/`, which the cluster cannot execute and `make build-*` does not ship. So
adding entries to `SCHEDULE` changes nothing live until a human runs
`publish_feed.sh` — a person, not a mechanism, which is the exact failure this
workstream exists to fix.

**How the error was made, because that is the reusable part.** Both halves of
step A were true individually — the entries *are* needed, and a `scheduled_tasks`
row *is* needed. What I never checked was whether anything connects them. I had
described the news pipeline correctly ("ends in `git_commit`, proven live") and
let that stand as evidence that *our* feed had a path, when what it actually
showed is that a *different* feed does. **A named precedent is not a wired one.**
The cheap check was one query against `agent_definitions`, and I ran it only
when I went to execute the step rather than when I wrote it.

Same shape as `a-helper-with-no-callers-is-not-a-refactor` — every component
present, nothing calling anything. Logged in `WRONG_CALLS.md`.

### What it actually takes (measured, not assumed)

Template found: `directory_export_action.go:113` `DirectoryExportJSONAction` —
query DB → marshal JSON → `sendExportFilesToGit(ctx, params, "sites", domain,
msg, files)`. That sender at line 478 is **already shared by the med and
directory exporters**, so a provocation exporter is a third consumer of proven
machinery. Its agent definition (`directory-json-exporter`) is two steps:
`export_json` → `complete`. Dependency order is pool-in-DB → Go action (council
gate) → agent def → scheduled row → content.

⚠ **Port the seal invariants, not just the rotation ones.** `check_seal()` now
refuses in both directions; a Go rebuild carrying only rotation reopens the leak
the gauntlet lane closed today. `verify_rotation.py` is the specification.

### The cheap design, and why the seal kills it

Publish the whole schedule once; let `round.go` and `snippets.js` each select by
today's date. No job, no action, no table — and rotation becomes a property of
the data that *cannot silently stop*, which is strictly better than a cadenced
job that can. I liked it a lot.

**It is foreclosed by the seal ruling.** Any future entry in a world-readable
file is tomorrow's provocation in the clear, and today's ruling is that even
*today's* is not readable until you step in. Withholding future entries is a
daily republish by another name.

Recording it because the interaction is the finding: **the seal is what makes the
daily job mandatory.** That is a real cost of the ruling, not an oversight here —
and it is exactly the kind of cross-lane consequence that is invisible from
inside either lane.

---

## 2026-07-31 (late) — the mechanism, built: pool table, Go action, agent, schedule

Built the four things the corrected step A named. State at the end of the session:

| piece | state |
|---|---|
| `provocations` pool table + the 9 live entries | migration **282 applied + recorded** |
| `render_provocation_feed` action + 14 tests | commit **572ae8dc6**, council `6612dc0b` |
| `provocation-feed-publisher` agent + `provocation-feed-refresh` schedule | migration **283 applied + recorded**, row **DISABLED** |
| the action in the running chassis | **NOT THERE** — 0 occurrences, control 3 |
| the site rotating | **still no** |

### Three design calls worth the words

**The pool stores no `published` flag.** Whether a provocation has been published
is a fact about its date and today's. Storing it would create a second source of
truth for one fact, and the two can disagree — which is a whole bug class avoided
by not having the column. `status` carries only the editorial state, which is not
derivable from anything.

**A duplicate publish date is refused by the DATABASE**, via a partial unique index
over approved dated rows, not by a check in the action. Two entries on one day
would make "the latest" ambiguous and hand the day's provocation to plan order.
Ranked by what closes the door: an index makes the bad state unrepresentable; a
validation makes it merely unlikely.

**The action SKIPS the commit when only `generated_at` would change.** This was the
most useful five minutes of the design. A daily job that always commits would
advance the one field people use to judge freshness while the site repeated
itself — the original bug wearing the fix as a disguise, and the exact landmine
filed this morning. Skipping means the file's git history is an honest record of
rotation, and a frozen timestamp on a no-rotation day becomes correct behaviour
rather than a symptom to chase.

### The verification that actually mattered

Not the 14 invariant tests — those check the Go side obeys the rules. The risk
here is different: this is a SECOND implementation of a feed that already has a
proven one serving the site, and two implementations drift while each stays
self-consistent.

So: dump the 9 real rows out of the pool, build the feed in Go, build it in Python
for the same date, compare. **Structurally identical.** Plus a companion test that
perturbs an input and requires the same comparison to FAIL — because a parity test
that cannot fail is just an expensive way to write `true`.

Migration 282's own verify block did the same job earlier and cheaper: it asserted
the DB selector picks `nobody-wants-personalised-internet`, which is the slug the
live feed serves. The port agreed with production before a line of Go ran.

### Missteps

**Walked into a documented landmine: backticks in `git commit -m` EXECUTE.** The
message for `572ae8dc6` said "reads the whole \`today\` object"; the shell ran
`today`, printed `today: command not found`, and committed "reads the whole
object". Forward-only means no amend, so the damage stands. It is in
`MEMORY.md` under shell traps and I hit it anyway — knowing a trap and pausing at
the moment you are stepping into it are different skills. Use single quotes for
the whole message, which is what I did for the next commit.

**Nearly reported a broken build as mine.** `go test` failed with five type errors
— in `nav_prune_floor_test.go`, which I have never touched. It looked like my
package-level additions had collided with something. They had not: another lane
committed a test at 22:33 whose implementation half is still uncommitted in the
shared tree. Checking `git status` on THEIR file (clean vs HEAD) and then testing
against `git archive HEAD` separated the two cleanly — HEAD is green, the tree is
not, and neither fact was about my code. Worth repeating: on this tree "the tests
fail" is not evidence about your own change until you have asked which files are
dirty and whose they are.

### The ordering, and why I did not roll the fleet

The action is inert until a chassis image carries it, so the scheduled row is
seeded disabled and the migration is safe to apply at any time. I did **not**
build and roll an image, for two reasons: a roll ships every other lane's
committed HEAD, which is not mine to decide; and my own council review was in
flight, which a roll would have killed. Any other session's roll will ship this
action anyway — that is the standing property of the shared tree, not a
workaround.

The remaining gap to the claim being true is now **content, not machinery**: the
newest pool entry is 26 Jul, so even with the job enabled the site would serve
that same provocation. Rotation needs entries dated forward from today.

---

## 2026-08-01 — council round 1 came back REVISE, and it was right

**Verdict:** `revise`, 12 reviewers, 5 abstained, decided by a **gating objection
from editquality (high)**, independently echoed by `bug_historian`, `guidelines`,
`guardian` and `prior_art_librarian`:

> registry.go only registers the action so it CAN be invoked by name; it does not
> attach it to any agent_definitions workflow step nor create/point a
> scheduled_tasks row at an agent. Nothing in the plan actually causes this action
> to run on a schedule.

`bug_historian` named the pattern exactly: *016b §9, "a renderer fix is inert
until something re-renders — and nothing schedules a re-render."* The objection
was **correct about the submitted plan**. I submitted before building the wiring,
then built it (migration 283) forty minutes later. Round 2 leads with it.

### The objection I had waved off, and should not have

`bug_historian`, medium:

> checkAgainstServed's shrink-guard … is disabled whenever the HTTPS fetch of the
> live artefact fails. The one guard built specifically to stop silent content
> loss is bypassed precisely during the infra flakiness most likely to co-occur
> with a bad deploy. Author flags this in risks but it is not resolved.

I had written the tolerance deliberately and justified it in the submission: the
content is fully determined by the pool, so publishing blind is still *correct*.
That reasoning is true and beside the point. The comparison is not there to
validate content — it supplies the shrink guard's **denominator**. Tolerating a
failed fetch switched off the only defence against provocations silently
disappearing. Now fails closed, with `allow_unverified_publish` as a stated
override. **Naming a risk in the submission is not the same as answering it**, and
a reviewer treating my own risks section as an unresolved objection is the system
working.

### Answered with evidence rather than argument

`prior_art_librarian` asked whether an existing feed agent could host the step
instead of a new agent type — a fair challenge to a new type. Checked:
`content-feed-trigger`'s selector requires
`classification.content_features.news_feed.recommended = true` **and** existing
`content_sources`. vonc.com has `news_recommended` NULL and zero content_sources,
so hosting it there means it never fires for the only site that needs it.

`guardian` flagged the outbound fetch as an unguarded SSRF surface — the domain
comes from step config and is interpolated into a URL. Correct. Now allow-listed
against the `sites` table with a shape check in front of it.

`reuse_agent` and `architecture` both asked for the Python builder to be demoted
**in this change, not as a follow-up**, so the second-implementation risk does not
quietly become permanent. Done: a NOT AUTHORITATIVE header naming its two
remaining jobs (parity oracle, manual/rollback fallback) and stating that its
SCHEDULE list is no longer the source of truth.

### The misstep, and it is a bad one

While answering `tooling_provenance`'s request for a `doc_notes` entry I concluded
that `LANDMINES.md` and the landmines scripts **did not exist in this repo**, told
the owner so, and wrote it into round 2's risk #6 — where it is now a false claim
in front of twelve reviewers.

They all exist. `LANDMINES.md` is 342KB and tracked. Every check I ran was a
relative path resolved against a cwd left behind by a `cd` five tool calls
earlier. Five checks — `ls`, `find .`, `git log --`, `git ls-tree -r HEAD`, `ls`
again — spanning filesystem, history and index, all agreeing, **all taking the
same hidden parameter**. Unanimity across different methods is not independence
when the methods share an input.

What caught it was not doubt about the finding: it was the council trigger script,
which I had run successfully an hour before, failing with "No such file or
directory". A file that demonstrably existed being suddenly absent is not a
plausible world.

Filed as a landmine (with the sync applied, so it is in `doc_notes` — which also
answers `tooling_provenance` properly) and in `WRONG_CALLS.md`. **The rule I want
to keep: an absence claim is a claim about a search, so state the search's scope.**
Writing "not present under `builder/`" would have shown me the error as I typed it.

---

## 2026-08-01 — council round 2: APPROVED, and what the six advisory objections were worth

**Verdict:** `approved` — 13 reviewers, 4 abstained, "approved with 6 advisory
objection(s) — none high-severity". Correlation
`6612dc0b-8e03-4039-a8c8-fe4fabaaddeb`. The round 1 gating objection (nothing
invokes the action) is answered by migration 283.

Advisory does not mean ignorable, so each was checked rather than filed. Two were
already answered by machinery I had not looked for, one was a false positive, and
three are genuine limitations now recorded as such.

### Answered by existing machinery — I should have checked before flagging a risk

**guardian (medium): a third concurrent writer into the same git repo, with no
serialisation described.** Fair on the submission, wrong about the platform. The
git adapter runs **2 replicas** and three exporters now write the same `sites`
repo, so the concern is real in principle — but `CommitToRepo`
(`internal/adapters/git/github_client.go`) already carries a **ref-race retry**
from `bugs_open/120` (owner ruling 2026-07-28), with **4 tests** pinning its
subtleties: the base head is re-read *inside* the retry loop (retrying on a stale
head loops on the same non-fast-forward for ever), and blobs are created *outside*
it because they are content-addressed. So a third writer is safe by an existing,
tested mechanism. **I had listed no concurrency risk at all — not because I had
checked, but because I had not thought of it.** The reviewer thought of it; the
platform had already solved it.

**tooling_provenance (medium): still no `doc_notes` entry.** True at submission
time and false by the time the verdict landed — the landmine went in and
`landmines-sync.py --apply` put it in `doc_notes` (verified by selecting the row).
This one is my own fault twice over: I told the council the file did not exist.

### A false positive, checked rather than assumed

**debug_historian (medium): the file matches the `sql_for_agents/*.sql` landmine
about a trailing `ROLLBACK` protecting nothing unless `BEGIN` is live.** That
landmine's own test is "if the first non-comment statement is not `BEGIN;`, there
is no transaction". Both migrations:

```
282: first non-comment = BEGIN;   last = COMMIT;   stray ROLLBACK/ABORT = 0
283: first non-comment = BEGIN;   last = COMMIT;   stray ROLLBACK/ABORT = 0
```

The objection fired on the **path pattern**, not the contents. Worth recording
because the reviewer was right to look — the pattern is a real trap, and the cost
of the check was one command.

### Genuine limitations, conceded

- **editquality (medium): the Python "demotion" is a docstring, not a functional
  change** — no enforcement that both sides move together. Correct, and I said as
  much in the risks. Documentation is not a mechanism. The honest position is that
  the drift risk is *managed by discipline*, which is weaker than the parity test
  makes it look, and the parity test only pins today's fixtures.
- **guidelines (medium): WRAPPER-ORCHESTRATOR and DECLARED CONTRACTS** —
  `processing_mode: task` doing an outbound fetch and a git commit inline, and no
  `input_contract` declaring `domain`/`repo_name`/`data_path`/`filename`. Measured
  before answering: **none of `directory-json-exporter`, `med-json-exporter` or
  `content-feed-orchestrator` declares an `input_contract`**, and both proven
  exporters use `processing_mode: task` doing exactly this work inline. So this is
  a fleet-wide convention gap, not a defect introduced here — and unilaterally
  diverging would make this the odd one out. Recorded, not "fixed".
- **bug_historian (medium): does a refusal surface, or is it a swallowed
  orchestration error?** Partially answered by the design: the action updates
  `scheduled_tasks.last_completed_at` **only on the success paths** (published or
  skipped-as-unchanged), so a refusal leaves it stale and staleness is the
  monitorable signal. That is a check nobody is running yet. Added to the RUNBOOK.
- **bug_historian (low) + editquality (low):** `allow_unverified_publish` can
  reopen the hole if set casually, and the row is `enabled=false` so the action is
  operationally still uninvoked. Both true, both deliberate, both documented.

### The pattern across both rounds

Round 1's gating objection and round 2's guardian objection are the same shape:
**a reviewer asking what happens at the seam between my change and the rest of the
platform** — what invokes it, what else writes where it writes. Both times I had
reasoned carefully about the inside of the action and not at all about its edges.
The parity test, the seal checks and the fail-closed guards are all *interior*
work. Nothing I wrote unprompted looked outward.

---

## 2026-08-02 — the mechanism went live, and the artefact showed what the tests could not

The owner rolled a chassis build. Everything below is verified against the live
system; the queries are in the RUNBOOK.

### 1. It shipped, and both replicas carry it

`v1.0.1229`, both pods started 18:39Z. Grepped the running binary, not git and
not the tag:

```
render_provocation_feed                    1   (added by this lane)
deploy_image_asset                         5   (positive control — pipeline works)
render_provocation_feed_NOT_A_REAL_ACTION  0   (negative control — grep discriminates)
```

**Honest note on the negative control.** The standing rule wants a string the
change REMOVED, expecting 0. This change was purely additive, so no such string
exists and I used a synthetic near-miss instead. That proves the grep
discriminates; it does NOT prove the image postdates the commit the way a
removed string would. The commit-provenance evidence is separate: all eight of
this lane's commits are ancestors of HEAD, and the action's symbol is present.

### 2. The first live run — predicted, then confirmed

Before enabling anything I ran the Python oracle for today and diffed it against
the served feed: **identical apart from `generated_at`**. So a real run today had
to end in the no-change skip, which made it the safest possible first firing —
full path exercised, site untouchable.

Enabled the schedule at 18:57. It fired within seconds (both timestamp columns
were NULL, so it was due immediately) and returned:

```json
{"today": "nobody-wants-personalised-internet", "committed": false,
 "reason": "no change since the served feed", "archive_entries": 8}
```

Selector, builder, fetch, comparison and skip: all live and correct.

### 3. Inducing the half that had never run

`committed: false` means the git writer was still unexercised, and an unexercised
writer is exactly what this platform gets bitten by. Since the content was
provably identical today, forcing a commit could only test the git path.
Temporarily set `force_commit` in the schedule's `input_data`, let it fire,
removed it again (the row is back to its seeded shape — verified).

It committed: **`a1bf37d55`**, "Update daily provocation —
nobody-wants-personalised-internet (26 Jul)".

### 4. What the artefact showed — and what every test had missed

The commit was **+119 / −119 lines**. Every line of the file changed to publish a
timestamp. Two byte-level differences, both of which parse identically:

| | oracle (Python) | action (Go) |
|---|---|---|
| markup | `<em>` literal | `backslash-u003c-em-backslash-u003e` (spelled in words — see §6) — 12 escapes |
| top-level key order | `generated_at, today, seal, …` | `archive, arena, generated_at, …` |

Cause of the first: `encoding/json` HTML-escapes `<`, `>`, `&` by default. Cause
of the second: Go marshals a map with its keys sorted.

**Why the tests were blind.** `TestGoFeedMatchesThePythonBuilderExactly`
unmarshals both sides before comparing, so an encoding difference is invisible to
it *by construction*. It passed the entire time the escaped form was shipping.
The sharpest part: I had already written, in this very package, that "a test that
formats dates with the code under test cannot detect a formatting change, it can
only agree with it" — and applied that reasoning to date formatting while leaving
the document encoding unexamined. **The insight was present and under-applied,
which is a different failure from not having it.**

### 5. What I fixed, and what I deliberately did not

`marshalFeedFile` with `SetEscapeHTML(false)` (`59f3c67dd`). Escaped markup
matters because the oracle is retained as the hand-edited manual fallback, and
the escaped form (`backslash-u003c-em-backslash-u003e`) sitting in a body of prose is a standing invitation to "correct"
it into something that then double-escapes.

**Key order left alone, deliberately.** The 119-line diff is a one-off cost of
changing writer, not a daily one — Go's ordering is deterministic, so tomorrow's
diff is small. Pinning the order means a second hand-maintained copy of every
object's field list living apart from the code that builds it, which is precisely
the drift surface this feed exists to remove. Recorded rather than fixed, with
the caveat: if the manual Python fallback is ever used, the next Go run rewrites
the file wholesale again. That ping-pong is the price of keeping two writers.

New test asserts on **bytes**, with a control proving the old encoder fails it.

### 6. Three missteps, all the same trap

- **The `\uXXXX` channel-decode trap fired three times in one session.** It ate
  the assertion needle in the new test (which asserted the presence of the very
  character it was meant to prove absent); it ate the before/after pair in
  `59f3c67dd`'s commit message, whose escaped-vs-literal pair was decoded into
  two identical sides, so it now says nothing;
  and it would have eaten the fix again if I had not built the needle by
  concatenation. It is already in my memory as a landmine and it still caught me
  three times, because I was writing *about* an escape sequence rather than
  looking for one. **The test's own control is what caught it** — the assertion
  and the control failed together, which is a contradiction a single assertion
  could never have surfaced.
- **The cwd trap fired again.** A `grep` returned "No such file or directory"
  because a `cd` into `builder/` five calls earlier was still in effect. Caught
  instantly this time, because it is a landmine I filed last session after it
  cost me a false claim in front of twelve council reviewers.
- Un-gofmt'd code reached a commit; the pre-commit check is advisory and reported
  it after the fact. Fixed forward in `3e78b0ba9`.

### 7. Where this leaves the lane

The mechanism is **live, enabled, and proven on both paths** — skip and commit.
Every part of the machine works.

The site is still making a false claim. `provocations.json` says "Today's
Provocation" over an entry dated 26 Jul, because the pool's newest row is 26 Jul
and the selector correctly serves the latest one that has arrived. **The
remaining gap is content and nothing else** — adding provocations is now
`INSERT`s, no code, no roll. That is the owner's editorial call.
