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
