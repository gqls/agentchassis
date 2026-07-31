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
