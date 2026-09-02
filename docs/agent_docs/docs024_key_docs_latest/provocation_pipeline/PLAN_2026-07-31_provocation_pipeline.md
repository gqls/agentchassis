# PLAN (2026-07-31) — make the daily provocation actually daily, and gate what it says

**Origin:** `gauntlet_dead_cta/HANDOFF_2026-07-30_B_the_daily_provocation_is_not_daily.md`.
Owner-raised 2026-07-30: *"the provocation didn't change today and has never
changed."* Measured true in substance — daily rotation was never built.

**Prior art, and it matters:** `docs/social001_vonc_tiktok_social/PLAN_spark_provocation_pipeline.md`
(2026-06-25) designed this already, as a clone of the news-feed pipeline. Its
Phases 1–2 SHIPPED (that is why `provocations.json` and the JS shells are live).
**Phases 3–4 never did** — measured 2026-07-30, `provocation-generator` and
`provocation-orchestrator` do not exist in `agent_definitions`. This plan is the
continuation of that one, not a competitor to it.

---

## 1. What is true, measured

All measured 2026-07-30/31; queries in `RUNBOOK_provocation_pipeline.md`.

| fact | evidence |
|---|---|
| the feed is a static file, four days stale when read | `https://vonc.com/data/provocations.json` HTTP 200, 9,797 bytes, `generated_at: 2026-07-26T00:00:00Z` |
| nothing regenerates it | 0 rows in `scheduled_tasks` matching `vonc\|provocation` |
| the builder has no rotation logic | `p4_sources/build_provocations.py`, 232 lines, every provocation a Python literal |
| it has changed 6 times, all by hand | `gh api repos/gqls/sites/commits?path=vonc.com/data/provocations.json` |
| the archive stopped 25 days ago | `archive.entries` = 8, dated 28 Jun – 5 Jul |
| today's provocation is not in the archive and cannot be | `today` has no `slug` and no `date` |

### 1a. Three findings the originating handoff did not have

**(i) The SERVER reads `today`, not just the browser.**
`internal/tools-api/handlers/round.go:44` — `FetchProvocation(domain)` fetches the
same `https://{domain}/data/provocations.json`, requires the `today` key, caches
it in-memory for 5 minutes (`provocTTL`), and returns 503 if it is missing.

> **This is the load-bearing constraint of the whole plan.** It kills the cheap
> fix that everyone reaches for first — putting a pool in the JSON and selecting
> by date *in the client*. Do that and the page displays provocation N while the
> Gauntlet argues provocation M, because the Go handler reads `today` and knows
> nothing about your selector. **Selection MUST happen at generation time and be
> written into `today`.** Any design where the browser chooses is wrong.

**(ii) The scheduled-publish plumbing already exists and is proven live.**
This is the good news, and it re-prices the options. The news pipeline does
exactly what this needs and is running now:

- `scheduled_tasks` row `content-feed-refresh` — enabled, 21,600s (6h), last
  triggered 2026-07-30 13:53:54Z, completed 14:02:29Z.
- it ends in `git_commit` (`platform/orchestration/actions/registry.go:509`),
  config `files_field` → commits rendered JSON into the site's git repo for S3
  deploy (`sql_for_agents/090_content_feed_orchestrator.sql`, step 7).
- **verified at the artefact, not the status:** `dartsonline.com`,
  `relojistas.com` and `webdesign.co.uk` are all serving `/data/latest-news.json`
  with `updated_at` of 13:56/13:58/14:01 on 2026-07-30 — inside that run's window.

⇒ Handoff B's landmine ("published via `gh api --input -`, payload on stdin
because argv blows `ARG_MAX`") describes the **hand** path. The **platform** path
is an off-the-shelf action. We should not hand-publish this on a cadence.

**(iii) CORRECTION — the archive page is NOT broken.**

> **CORRECTED 2026-07-31.** Handoff B said `/provocations/index.html` "paints
> neither today's provocation nor, apparently, much else (1,293 chars of visible
> text)" and told the next thread to check whether it is broken before designing
> the archive. **It is not broken.** Rendered in headless chromium
> (`--dump-dom --virtual-time-budget=9000`): all 8 archive entries paint with
> date, title and teaser; 7 are openable and the 8th is deliberately
> non-openable because no case was written for it (the builder's documented
> Journey B.3 behaviour); the empty state `<p class="provocations-archive__empty">`
> is correctly `hidden`; the blank leading row I first suspected of being a
> visible defect is a properly hidden `data-archive-template` with `hidden=""`
> *and* a `display:none` CSS rule. The 1,293-char figure is consistent with a
> page that is working — that measurement was probing for **today's** headline,
> which correctly is not on the archive page.
> **What caught it:** printing 600 characters of DOM context around each match
> instead of trusting `grep -c`. The same context print also caught a false
> positive of my own — "Nobody actually" matched the *29 Jun* entry ("Nobody
> actually reads terms of service"), not today's headline ("Nobody actually
> wants a personalised internet"). A substring is not an identification.
> **Consequence for this plan:** the archive needs no repair. The only archive
> defect is that today's provocation never joins it.

---

## 2. Decisions taken by the owner (2026-07-31)

1. **Archive rule: DECIDED.** *"It can be archived when the new one is
   published."* → a provocation joins the archive at the next rotation, never
   during its own day.
   **This is also the security-correct answer** and it was not obvious: the
   archive page paints full case text, so promoting today's provocation
   immediately would have opened a *third* leak path for the provocation the
   Gauntlet is built to hide (`HANDOFF_2026-07-30_C`). The chosen rule closes it
   by construction, and it unblocks C's option 3 (home shows yesterday's as the
   sample).
2. **Direction: LLM generation, behind a gate** — "an equivalent of the claims
   check", asserting the provocation is (a) safe, (b) interesting, (c) current or
   relevant to our intended/current audience, (d) has the other properties of a
   good provocation.
3. **Images: later, and legally constrained** — possibly celebrities, via
   sketches / watercolours / licence-free images; possibly restricting the field
   to celebrities for whom licence-free images exist. See §6.
4. **New product shape raised for consideration: the paired provocation.** See §5.
5. **Acknowledged by the owner, and it shapes the sequencing:** *"when we have
   live contestants we will have a better idea of what they like and don't
   like."* Criteria (b) and (c) are **not measurable today**. See §4.

---

## 3. Phase 0 — prerequisites that every option needs

None of these depend on where provocations come from. All are in
`p4_sources/build_provocations.py`. **Do these first; they are the difference
between a file that can rotate and one that cannot.**

1. **`generated_at` must be computed, not a literal.** It is currently the string
   `"2026-07-26T00:00:00Z"` at line 226.
   > **LANDMINE.** This defeats handoff B's own verification step 3
   > (`curl … | print(generated_at)`). Re-run the builder today and it still
   > emits `2026-07-26T00:00:00Z`. **A freshness check that reads a hardcoded
   > field will report a stale file as fresh for ever.** Fix this before writing
   > any verification that trusts it.
2. **`today` needs `slug` and `date`.** It has neither (keys are `eyebrow`,
   `headline`, `body`, `primary_cta`, `secondary_cta`, `stats`). Without them it
   cannot be promoted into `archive.entries`, whose shape is
   `{date, slug, title, teaser, detail_body?, url?}`. Note the shapes differ in
   more than that — `today` has `headline`/`body`, entries have `title`/`teaser`/
   `detail_body`. **Define the mapping explicitly; do not assume it is a rename.**
3. **Promotion on publish.** Previous `today` → head of `archive.entries`, per the
   owner's rule in §2.1.
4. **`arena.cards[0]` is a hand-copied duplicate of today's headline**
   (lines 206–216: the literal `"Nobody actually wants a personalised internet"`
   without the `<em>`). Derive it from `today` or it goes stale the first time
   anything rotates, and the site contradicts itself on the same page.
   *(This is a small live instance of `HANDOFF_2026-07-30_D`'s "vonc says the
   same things twice" — same class, different cause.)*
5. **Selection happens at generation time**, per §1a(i). Non-negotiable.

**Definition of done for Phase 0:** running the builder twice with two injected
dates produces two different `today` objects, a correct `generated_at`, and a
correctly promoted archive — verified by running it, not by reading it.

---

## 4. Phases 1–3 — rotation, then the gate, then generation

Deliberately in this order. Each phase is independently useful and shippable, and
each earlier one de-risks the next. This mirrors the 2026-06-25 plan's own
sequencing instinct ("prove the JS layer works before building the pipeline"),
which was right and is the reason Phases 1–2 of it survive today.

### Phase 1 — rotation, driven by a pool

The mechanism, with no LLM anywhere near it. Proves *daily* is real.

- pool of provocations + deterministic selection by UTC date
- a `scheduled_tasks` row, daily
- publish via the existing `git_commit` action (§1a(ii)), not by hand
- **9 written provocations exist today** (8 archive + 1 current) = 9 days of
  runway. State that plainly rather than discovering it on day 10.

**Verify the mechanism, not the outcome** — handoff B is right that rotation
cannot be checked in a day, and that a hardcoded provocation looks identical to
a rotated one on any single day. Same code, two injected dates, two results.

### Phase 2 — the provocation gate

The piece the owner asked for. **It is genuinely not the claims check**, and the
distinction is the whole design:

> **A provocation is a deliberately contestable assertion. The claims rail
> ("nothing on vonc.com claims a number that is not true by construction") would
> reject every good provocation we have. But the BODY of a provocation smuggles
> in ordinary factual claims that are fully subject to the rail** — e.g. the
> four-day-week entry asserts the pilots "measure self-reported output", which is
> either true or false and has nothing to do with the thesis.
>
> **So the gate must split the artefact:** the *thesis* is exempt by design; every
> *supporting factual assertion* is not. A single blanket "is this true?" rejects
> everything; a single blanket "it's opinion" lets falsehoods through. This is
> exactly the failure in `bugs_open/149` C1, whose second row is a factual claim
> smuggled into prose ("10,000 Monte Carlo trials per query" for a tool that
> computes analytically and calls `Math.random` nowhere) — and in
> `bugs_closed/043`, generated copy inventing quantitative claims.

**Derive the criteria from the corpus we already have, not from first
principles.** The 9 existing provocations share a measurable shape:

- a contestable claim stated flatly as fact, short, no hedging
- **genuinely two-sided** — every `detail_body` makes the case and then makes the
  counter-case ("The counter is…", "Against that:", "The rebuttal is…")
- **not tribal-political** — none touch party politics or the standard culture-war
  set. This is already a safety property of the corpus and should be made an
  explicit rule rather than left as an accident of taste
- arguable from ordinary experience; no specialist knowledge needed to disagree

That corpus is the specification. Writing the gate against it is cheap and
testable: it must pass all 9 and reject deliberately bad samples.

**Who judges what, and why the split:**

| criterion | judged by | why |
|---|---|---|
| (a) safe | automated | rule-shaped, and the corpus already defines the boundary |
| (d) good-provocation form | automated | two-sidedness, length, flat assertion — all checkable |
| body's factual claims | automated, existing claims rail | this is what the rail is *for* |
| (b) interesting | **human, for now** | we have no signal |
| (c) current / relevant to the audience | **human, for now** | we have no audience data |

**Recommendation, and it is the one judgement I made on the owner's behalf:
publish nothing unreviewed until (b) and (c) are measurable.** The owner's own
sentence is the reason — *"when we have live contestants we will have a better
idea of what they like and don't like."* An automated judge of "interesting" with
no contestant data is unfalsifiable: it will emit a confident score that nothing
can check, which is the precise shape of every entry in `WRONG_CALLS.md`. A human
approval queue costs a few minutes a day, removes the `149` C1 risk structurally
rather than gating it, and **retires itself** — once paired provocations produce
real engagement data (§5), (b) and (c) become measurable and move to the
automated column. Flagging this as a decision to overturn if the owner would
rather accept the risk for speed.

### Phase 3 — LLM generation feeding the gate

Only once the gate exists and the rotation is proven.

- `provocation-generator` agent — Phase 3 of the 2026-06-25 plan, never built
- reuse `feed-ingester` + `content_sources` for **currency**, which is criterion
  (c)'s objective half: what is actually being argued about this week. That half
  *is* measurable now, unlike the audience half
- output goes into the approval queue, never straight to `today`

---

## 5. The paired provocation — analysis

*Owner's idea, raised 2026-07-31: an organiser picks a set of people, sets their
own provocation, the team reply that day or over several days until all have
committed, the organiser chooses the timing rules, and results are distributed to
the team privately rather than published.*

**This is the strongest idea in the thread, and I think it is sequenced too
late.** Seven observations:

1. **It is the Gauntlet inverted, and it is the only version where the sealed
   reveal is load-bearing.** In the public product the seal is a nicety — and
   `HANDOFF_2026-07-30_C` shows it already leaks from the home page without much
   being lost. In a team setting the seal *is* the product: if Alice sees Bob's
   position before committing hers, the exercise produces anchoring and
   groupthink, which is the exact thing a team-deliberation tool exists to
   prevent. **This argues for resolving C in favour of keeping the seal**, rather
   than C's option 2 (retire it), because a feature that looks decorative today
   becomes essential here.

2. **"Until they've all committed" is a sealed-bid auction, not a form.** It
   needs, explicitly: a deadline; a rule for what happens when someone never
   responds (reveal anyway? reveal at quorum? organiser forces it?); and an
   **atomic** reveal — everyone's positions become visible at the same instant.
   If they dribble out, the early committers are penalised for being prompt and
   the last person to answer has read everyone else's. The owner's "choices given
   to the person setting it up" is the right instinct; these are the choices.

3. **It requires identity, and this system has none.** The public Gauntlet keys
   rounds on a hash of the client IP — and per `bugs_open/139` that hash is a
   *constant*: `sha256("172.18.0.1")` in 83 of 83 rows, so it has never
   distinguished anybody. There is no user table, no accounts, no invitations.
   **Named participants are a bigger prerequisite than the provocation pipeline
   itself**, and it should be planned as its own thing (magic-link invitations are
   probably the cheapest honest route — no passwords, no account management).

4. **It is the answer to the owner's own blocker.** He wrote *"when we have live
   contestants we will have a better idea of what they like and don't like"* in
   the same message. **Paired provocations are how you get live contestants** —
   named, repeat participants who arrive because someone they know invited them.
   That data is precisely what makes criteria (b) and (c) measurable and lets the
   gate's human column retire. So the dependency runs *backwards* from how it
   looks: the paired mode unblocks the automated gate, not the other way round.

5. **It has a buyer, which the public daily does not.** A team lead or facilitator
   has a budget and a recurring need; an anonymous visitor has neither. And it is
   a distribution mechanism in itself — every paired provocation brings N people
   to the site on a trusted invitation, which is a far better test than posting a
   provocation into a feed and hoping (the owner's 2026-07-29 distribution
   experiment).

6. **Privacy stops being a preference and becomes a correctness property.** Named
   individuals' opinions on contested topics, held on our infrastructure and
   distributed to their colleagues. A leak here is a real harm to a real person —
   workplace consequences for a recorded opinion. The public product carries none
   of this exposure because it is anonymous. **Access control must be genuine, not
   URL obscurity**, and that should be stated before anyone builds a prototype
   with unguessable links and calls it private.

7. **An AI verdict on a named colleague, circulated to their team, is socially
   different from an anonymous one.** "The AI judged Bob's argument weakest" sent
   to Bob's team is a performance review nobody agreed to. Worth designing around
   deliberately — e.g. share the *positions* with the group but keep each
   *verdict* private to its author, or drop ranking entirely in paired mode. Cheap
   to decide now, expensive to retrofit after the first team falls out over it.

**Suggested sequencing:** it does not block Phases 0–2 and should not delay them —
they are small and the site is making a false claim today. But it should be
planned *in parallel*, not after Phase 3, because Phase 3's hardest criteria
depend on data only this mode produces.

---

## 6. Images — the note to keep, for later

The owner's instinct (sketches / watercolours / licence-free) is right about
copyright and **incomplete about the other half**, so recording it now:

- **Copyright and personality rights are separate problems.** A watercolour of a
  photograph avoids infringing the *photographer's* copyright. It does nothing
  about the depicted person's **publicity / personality rights**, which attach to
  the identity, not the image. An original painting of a living celebrity used to
  promote a commercial service can still be actionable.
- **UK-specific and relevant, since this is a UK operation:** there is no general
  right of publicity in the UK; the nearest route is **passing off / false
  endorsement**, which turns on whether a reasonable viewer would infer the person
  endorses the product. That is a more permissive regime than the US, but it is
  not nothing, and it is exactly the inference a celebrity beside our branding
  invites. *(Not legal advice — this needs a real opinion before shipping.)*
- **"Only feature celebrities for whom we have licence-free images" inverts the
  priority.** It selects the day's topic by image availability rather than by what
  is worth arguing about. Given criterion (c) is *relevance*, that is a direct
  conflict with the gate we are about to build.
- **Cheapest safe route: illustrate the idea, not the person.** The provocations
  in the corpus are about ideas (privacy, mentorship, four-day weeks); none needs
  a face. Keep celebrity likenesses for a deliberate, separately-advised decision.

---

## 7. Open questions

1. **Human approval queue, or automated gate only?** §4 recommends human-for-now
   with a defined retirement condition. Owner to confirm or overturn.
2. **Where does the pool live** — Python literals in the builder (matches today,
   zero new machinery) or a DB table (queryable, needed eventually by the approval
   queue and by paired mode)? Phase 1 can start with literals; Phase 2's queue
   probably forces the table. Deciding early avoids one migration.
3. **What is the audience?** Criterion (c) is unanswerable as written until this
   is stated. Even a one-line answer materially changes the gate.
4. **Paired provocation: prototype or spec first?** §5.3 says identity is the real
   prerequisite; worth knowing whether the owner wants a throwaway prototype to
   feel the shape, or a proper design.

---

## 8. Coordination

`tools-api` and the vonc island belong to the **gauntlet_dead_cta** lane
(`MEMORY_workstreams.md`). Nothing in Phases 0–3 changes `tools-api` — `round.go`
appears here only as a *reader* whose contract we must not break. **Paired mode
would change it**, so that must be agreed with that lane, not implemented past
them.

---

## 9. OWNER DIRECTION 2026-07-31 (second round)

Four things, all answering §7's open questions or adding to them.

### 9.1 Grok is the topic source — and it is already wired [VERIFIED]

*Owner: "ask a different AI — maybe Grok for interesting current topical questions
as it has access to twitter and readership stats, we have an api key for it as we
already query Grok for news."*

Correct on every count, and cheaper than it sounds:

- `platform/orchestration/actions/feed_actions.go:733` — `resolveLLMNewsProvider`
  already supports `xai` / `grok`, reading `XAI_API_KEY` then falling back to
  `GROK_API_KEY`.
- It targets the **xAI Responses API** (`https://api.x.ai/v1/responses`) with the
  **`web_search` AND `x_search` tools** — that is precisely the X readership signal
  the owner is describing. Documented recommended model: `grok-4-1-fast`.
- `platform/agentenv/provider_keys.go:53` carries `GROK_API_KEY` into agent
  environments, so an agent can reach it without new plumbing.

⇒ **Phase 3's sourcing half is mostly reuse**, the same way Phase 1's publishing
half turned out to be (§1a(ii)). This is now the second time this workstream has
found the expensive-looking piece already built. *Check before costing.*

**But keep the roles straight, because this is where it could go wrong:**

> **Grok supplies TOPICS — what is being argued about right now, and how loudly.
> It does not supply provocations.** Turning a live topic into a provocation
> (a flat, two-sided, non-tribal claim in our house voice) is a separate step, and
> the gate sits after *that*, not after Grok.
>
> **And the gate must judge the OUTPUT, never the topic.** X is an adversarial
> content source — the most-discussed thing on any given day is frequently a
> pile-on, a hoax, or someone's harassment campaign. "It is trending" is evidence
> of volume and nothing else. The owner's "filter it against slop and danger" is
> exactly right and it belongs at the end of the chain, applied to the finished
> provocation, where the claims-rail split of §4 Phase 2 also applies.

### 9.2 Categories — make it a field NOW, while it is free

*Owner: "I want it to be hugely popular so I think we will need to have different
categories sooner rather than later, ranging from current political opinions to
pets etc each with a different target audience."*

Agreed, and the recommendation is to **add `category` as a first-class field from
day one even while only one category exists.** It is nearly free now and expensive
to retrofit, for two reasons that are not obvious:

1. **A category is not a tag — it carries its own gate configuration.** "Current
   political opinions" and "pets" cannot share a safety threshold, an audience
   definition, or a two-sidedness standard. If the gate is built assuming one
   global config, per-category thresholds mean rewriting it. If `category` is
   present from the start, the gate reads its config *from* the category and the
   politics one is a row, not a refactor.
2. **⚠ Categories BREAK the current server contract, and this is the §1a(i)
   landmine biting a second time.** `round.go` fetches the feed and requires a
   single top-level `today` key — one provocation per domain, full stop. Several
   simultaneous category dailies cannot be expressed in that shape. So per-category
   rotation needs *either* one feed file per category *or* a shape change to
   `today`, **and in both cases `tools-api` must change in step** so the Gauntlet
   knows which category's provocation a round is arguing. That is the
   gauntlet_dead_cta lane's code (§8). **Do not design multi-category rotation
   without agreeing this with them first** — it is not a vonc-side-only change,
   and discovering it late means a live mismatch between the page and the engine.

### 9.3 The first audience — recommendation

*Owner: "For now keep it narrow to an audience that would be likely to repost
their results on busy sites on topics that would encourage them to do that — we
could think about this."*

This finally answers §7 question 3, which the gate's criterion (c) could not be
written without. My recommendation, for the first and only category:

> **People who argue online recreationally in venues that are already busy** —
> concretely, the `r/changemyview` / Hacker News / tech-X axis. Topics: work,
> technology, AI, attention, institutions. The nine provocations we already have
> sit squarely in this space, which is a useful sign rather than a coincidence.

Why this one rather than something with more reach:

- **The share artefact already exists and is native to these venues.** The
  gauntlet lane shipped the exchange card on 2026-07-31 — challenge, defence and
  ruling on one image. "I argued this and here is what the AI ruled" is a
  *post format* that r/changemyview and HN already reward. In a sports or politics
  venue the same card reads as bait.
- **It is the safest possible calibration set for the gate.** These topics are
  contestable without being tribal, so the slop-and-danger filter gets its easiest
  job while we are still learning what it should reject. Starting on political
  opinion would mean tuning the hardest category with an untested gate — and
  `bugs_open/149` C1 is what an untested gate costs.
- **It carries paired mode's buyers with it.** Tech team leads are in this
  audience. The people most likely to repost a verdict are also the people most
  likely to run a paired provocation with their team (§5.5).

**The honest weakness:** this audience is small and sceptical, and "hugely popular"
is not where it leads on its own. It is a *calibration* audience — the one where
the gate can be proven cheaply — not the destination. The path to popular runs
through the categories in 9.2, and the argument for doing this one first is that
it is the only one where being wrong is cheap.

### 9.4 Paired mode — prototyped

*Owner: "Please prototype the paired mode."* **Done:**
`prototype/` (nested Go module, in-memory, no LLM), with its own README.

- the seal is enforced by the **type system**, not by a check: a pre-reveal
  response is a `SealedView` whose only description of a peer is
  `{Name, Committed}`, so a handler cannot leak a position by forgetting a
  condition
- three reveal rules — all-committed, quorum, deadline — plus an organiser
  force-reveal, which is the owner's "choices given to the person setting it up"
- **the organiser cannot read positions either**, and **non-responders do not
  receive the reveal** (stay silent under a deadline and you do not get to read
  the room — otherwise silence is the optimal play, which is the seal's own
  failure mode re-entering by the back door)
- 14 tests written to *break* the seal, **mutation-tested**: two mutations were
  caught, and a third — stamping the reveal time per read rather than per
  session — **was not**, because the atomicity test held the clock constant across
  all three readers. Fixed, and re-verified that the mutation now fails. Recorded
  because it is the exact "a quiet test passes when the rule is gone" trap
- driven end to end against the running server: with two of three committed,
  the third participant's page and the organiser's page each contain **zero**
  occurrences of either position; after the third commits, all three pages show
  all three

**What it deliberately fakes, so nobody mistakes it for a design:** nothing is
persisted; a token in a URL is the only access control (unguessable ≠
authenticated); there is no AI verdict; no email, no CSRF, no audit log. Full list
in `prototype/README.md`.

### 9.5 Consequent changes to §7's open questions

- §7.3 (*what is the audience?*) — **answered**, 9.3, pending owner confirmation.
- §7.2 (*where does the pool live?*) — **now leans strongly to a table.** Categories
  (9.2) plus a human approval queue (§4) plus Grok-sourced drafts (9.1) are three
  independent reasons the Python-literal pool stops being adequate. Recommend
  starting with the table rather than migrating to it.
- §7.4 (*prototype or spec first?*) — **answered by the owner: prototype.** Done.
- §7.1 (*human approval or automated gate only?*) — **ANSWERED 2026-07-31: no human
  approval, "we can try without for now."** The owner read the §4 recommendation
  and the 9.1 stakes and took the speed. That is his call and it is now the
  design; §10 states what it obliges us to build instead.

---

## 10. Building for an UNREVIEWED publish path (owner decision, 2026-07-31)

The recommendation in §4 was a human approving (b) and (c) until contestant data
exists. **The owner has decided against it, and the decision is recorded as taken,
not as pending.** What follows is not a re-argument — it is the set of properties
the pipeline now has to have *because* nobody is in the loop.

**The load-bearing change: the gate stops being a filter and becomes the only
control.** With an approver, a gate miss costs a moment of someone's attention.
Without one, a gate miss is a false statement on a live homepage — which is
`bugs_open/149` C1, witnessed on 2026-07-29 on another site. So:

### 10.1 Fail CLOSED, and the degraded mode is "yesterday's provocation stays"

If generation produces nothing that clears the gate, **publish nothing**. The feed
keeps the current provocation and the site is unchanged.

That degraded state is, precisely, the bug this workstream exists to fix — but
there is a real difference between it happening silently for a month and it
happening deliberately for a day with an alert attached. **A stale provocation is
a broken promise; a false one is a broken product.** Prefer the first.

**This is already how the news pipeline behaves and it is proven live** — step 6 of
`090_content_feed_orchestrator.sql` is an `evaluate_condition` on
`news_render_result.item_count`, routing `0 → complete` so `commit_news` never
runs on an empty render. Reuse that shape rather than inventing one.

### 10.2 A gate that ERRORS must count as a rejection, never as a pass

The failure to design against is not the gate judging wrongly — it is the gate not
running at all (timeout, API error, malformed response) and the pipeline reading
"no objection returned" as "no objection exists".

> **This exact bug has already been shipped on this platform.** A `!= nil` guard
> turned *unknown* into *no rule*, so an unpublished product range scored `Match`
> on a live page (fixed 2026-07-29, chassis v1.0.1196). **Absence of a verdict is
> not a favourable verdict.** Encode the default as reject and test the timeout
> path explicitly, because it is the path that will not be exercised by accident.

### 10.3 Somebody has to be able to see what it decided

An approval queue has a reviewer looking at every candidate as a side effect. Take
the reviewer away and **nothing observes the gate unless we build the observation.**
Otherwise the first evidence of a broken gate is a complaint about the live site.

Minimum: every gate decision persisted with its verdict, its reasons, and the
candidate it judged — including the rejections, which are the interesting half.
A gate that has rejected 100% or 0% for a week is broken in one of two directions
and both are invisible without the log.

### 10.4 A rollback that exists BEFORE the first automated publish

One command that restores the previous provocation, written and *tested* while
nothing is going wrong. Not "we will write one if we need it" — the moment it is
needed is the worst moment to write it. The feed is a single file in a git repo,
so this should be cheap; cheap and absent is the worst combination.

### 10.5 One publish per day is itself a safety property

At most one rotation per day means a broken generator can produce at most one bad
day before anyone notices, rather than a flood. Do not add a "catch-up" mode that
publishes several at once to fill a gap.

### 10.6 Calibration is no longer optional

§4 proposed testing the gate against the nine existing provocations. With a human
backstop that was good practice; without one **it is the only evidence the gate
works at all.** It must pass all 9 and reject a set of deliberately bad samples
(a bare insult, a factual claim dressed as opinion, a one-sided political take, a
piece of trending slop) *before* it is wired to anything that publishes.

### 10.7 What this does NOT change

The §4 split still holds and matters more: **the thesis is exempt from the claims
rail, the body's factual assertions are not.** Removing the human removes the
judge of "interesting" and "relevant"; it does not make the truth check optional.
Criteria (b) and (c) remain unmeasurable until there are contestants (§5.4) — they
will now be judged by a model with no data behind it, so treat their verdicts as
the weakest part of the gate and keep the confidence of that judgement out of the
publish decision where it is separable from safety and form.

---

## 11. OWNER RULING 2026-08-09 — the audience, and a risk brief that reorders §5

### 11.1 §7.3 is ANSWERED — everyday culture orthodoxies

Open since 2026-07-31, recorded in §9.5 as *"answered, 9.3, pending owner
confirmation"*. It is now confirmed, and **not in favour of §9.3's recommendation.**

> **RULED: the live slot carries everyday culture orthodoxies** — food, music and
> film canon, cities, generational habits.

§9.3's recommendation (the r/changemyview–HN axis: work, technology, AI, attention,
institutions) is **superseded as the target audience**, though its reasoning is kept
above rather than deleted — it was right about the *calibration* value and it said
so itself: *"this audience is small and sceptical, and 'hugely popular' is not where
it leads on its own. It is a calibration audience … not the destination."*

**What decided it,** because the mechanism is reusable: the owner read seven drafts
written to §9.3 and called them boring. The diagnosis is not taste — **work topics
are the weakest possible on identity investment.** The product asks someone to argue
on a clock and be judged, which only appeals when having a take on the topic is part
of who they are. Almost nobody's identity is "person with views on meetings", so
there is nothing at stake in defending one. Everyday culture orthodoxies score high
on identity *and* on appetite for the fight — which pets, the owner's own earlier
example, does not: affection is high but nobody wants an AI ruling their dog opinion
wrong.

### 11.2 The risk brief — LOW APPETITE, KEEP THE VIRALITY

Raised the same day:

> *"someone can say something insulting about someone and then it gets validated by
> the site and posted over the internet. I may be liable … I still want the viral
> effect somehow but my risk appetite is low."*

Routed to **`architecture_review/RFC_020`** (the publish path is the
`gauntlet_dead_cta` lane's code) plus **`bugs_open/232`** for the one gap true
today. **The finding that matters to THIS lane:** publishing is an amplifier, not a
threshold — showing text about a named third party to one person is already
publication — so the exposure does not start at the publish button.

**Consequence for provocation selection, and it is now a standing constraint on
this lane:** a provocation about a *category* invites category argument; one that
implies a villain invites naming one. **This is a free lever and it is ours.** A
drafted provocation ("Restaurant food has got worse") was pulled on exactly this
ground before going live — it invites naming a specific business and asserting
something factual about it, which is the actionable shape. Apply the test to every
future provocation, human- or generator-written:

> **Does answering this well require naming a real person or business, and saying
> something checkable about them?** If yes, reframe it or drop it.

This also belongs in the generator's gate criteria whenever that is next touched —
it is a *safety* criterion (§4's column (a)), not a taste one, and it is currently
absent from both the gate and this PLAN's §4 table.

### 11.3 What this does NOT change

§10 stands entire — no human approval of publishes, fail closed, gate errors count
as rejections. §5's paired-mode analysis stands and gains force: §5.6 already said
privacy stops being a preference and becomes a correctness property once real people
are named, which is the same finding arriving from the other direction.

---

## 12. OWNER RULING 2026-08-09 (second, same day) — the safety layer moves OFF the judge

### 12.1 The ruling

Raised by the nine-round calibration on v1.0.1267
(`HANDOFF_2026-08-08b_continue_here.md` §4): on round 3 of 9, `cal-bad-insult` —
repetitive abuse, no argument — was **approved**. The judge ran, its own advisory note
called it *"pure repeated insult with no actual argument or fact-checkable content"*, and
it returned `safe: true` anyway. The entire safety decision is one arm,
`provocation_gate_action.go:469` `if !j.Safe`, with nothing binding the boolean to the
judge's own note. Four candidates were put up, ordered by what makes the bad state
unrepresentable.

> **RULED: candidate 1 — a DETERMINISTIC pre-judge abuse check**, in the shape of the
> existing `tribal_political` form-layer rule (`judge_ran=false`, never varied across
> nine rounds). Anything the form layer kills never reaches the judge's discretion.

Rejected by implication, and worth recording so they are not re-proposed:
best-of-N on the safety field (mitigates, never closes — and multiplies cost per
candidate), and relying on human approval before publish, **which §10 has already ruled
out** and which the owner did not reopen.

### 12.2 Why this ruling is load-bearing rather than tidy

§10 stands: **no human approval of publishes.** So after the generator and gate are
wired, the gate is the *only* thing between an LLM-written provocation and a live page
on a site that promises one daily. A 1-in-9 leak on the abuse criterion is not a quality
issue in that configuration; it is the whole exposure. This is also why §4a's inherited
"three green runs" bar was retired — six consecutive clean rounds followed the failure,
so the bar it set would have certified this gate ~70% of the time.

### 12.3 STANDING CONSTRAINT — §11.2's third-party test lands in the SAME change

§11.2 ruled the named-third-party test a **safety** criterion and said it *"belongs in
the generator's gate criteria whenever that is next touched"*. **This ruling is that
moment** — the gate's safety layer is being opened now, and adding one safety criterion
while leaving the other in prose would be the same omission twice.

> Does answering this well require naming a real person or business, and saying
> something checkable about them? If yes, reframe it or drop it.

**Open sub-question for the implementing thread, and it is a real one:** that test is a
*judgement*, not a keyword — so a purely deterministic form-layer rule cannot express
it, and putting it on the judge gives it the same stochastic exposure this ruling exists
to remove. The honest options are a deterministic proxy (proper nouns / brand-shaped
tokens as a **flag**, not a verdict), a judge criterion accepted as best-effort with the
weakness stated, or both in series. **Do not let it silently become judge-only** — write
down which was chosen and why.

### 12.4 What is NOT settled by this ruling

- **Which way the deterministic check should err.** A false negative publishes abuse; a
  false positive silently starves an already-exhausted pool. The gate is fail-closed by
  §10 and a pool can be topped up, so the drafting recommendation is to err toward
  rejection — but the cost is real and the starvation is **silent**, so it needs saying
  out loud rather than assuming.
- **The wording/severity of the pattern set itself**, which is where a form-layer rule
  either earns its keep or generates false positives nobody notices.
- **Sequencing against the six ungated drafts** (§12.5).

### 12.5 State this ruling lands into, measured 2026-08-09

The pool moved the same morning, under §11.1's audience ruling:

| | state |
|---|---|
| approved | **9**, all `category='general'`, newest `publish_on` **2026-07-26** |
| draft | **6**, all `source='llm'`, created 09:39:40, `publish_on` 08-09..08-14, **all ungated** |
| the previous 7 drafts | **gone** — the §9.3-audience set the owner called boring |
| wired and running | only `provocation-feed-refresh` → `provocation-feed-publisher`, every 6h |

**So the stale site has a precise cause, and it is not a bug.** `selectForDate`
(`provocation_feed_action.go:276`) takes every approved row with `publish_on <= today`
and picks the latest; `loadProvocations` returns `status='approved'` only. Nothing has
been gated into `approved` since 26 July, so 26 July is correctly the newest thing there
is to serve. **The missing link is the gate, which is the unwired step** — generate (ran
manually today) → **gate (not wired)** → approved → publish (wired, running).

Which means the six drafts are what un-sticks the site, and gating them is precisely
what should NOT happen until this ruling ships. That sequencing is a decision, not an
implementation detail.

---

## 13. OWNER RULINGS 2026-08-09 (third round) — the nine open decisions, answered

Put to the owner as an outline of everything still needing a decision after §12.
Answers recorded verbatim in substance, with the reasoning where he gave it.

| # | question | RULING |
|---|---|---|
| 1 | which way should the deterministic abuse check err? | **Err toward REJECTION** |
| 2 | the §11.2 third-party test — deterministic proxy, judge-side, or both? | **Best-effort (judge-side) is OK for now** |
| 3 | the six ungated drafts — gate now or hold? | **HOLD** until the abuse check ships |
| 4 | pool runs dry — top up, and be told? | **Both: generate more AND notify me** |
| 5 | cadence / buffer | **6 days is right — it keeps recency.** Shorten later as we learn, to get more current questions |
| 6 | `group-chats-replaced-friendship` (empty body) | **RETIRE the row** |
| 7 | `bugs_open/223` verifier fix | explanation requested before deciding |
| 8 | `/blog/provocation.html` | **fix as you see fit** (delegated) |
| 9 | RFC_013 category collision | **"fix category general now"** — AMBIGUOUS, see §13.3 |

### 13.1 What ruling 1 commits us to, stated so it is not a surprise later

Erring toward rejection on a fail-closed gate with **no human approval of publishes**
(§10) means the failure mode we have chosen is **a silently starving pool**, not a
published insult. That is the right trade and it has a cost: the pool is *already*
exhausted at the approved tier (newest `publish_on` 2026-07-26), so the check must ship
alongside ruling 4's dry-pool handling or the first thing it does is deepen an outage
nobody is told about. **Rulings 1 and 4 are one change, not two.**

### 13.2 Ruling 2 is explicitly PROVISIONAL, and the weakness must stay visible

"Best effort for now" accepts that the named-third-party test sits on the judge and
therefore inherits precisely the stochastic exposure §12 removed from the abuse
criterion. That is a deliberate, informed trade — not an oversight — and the reason it
is acceptable is ruling 1: a judge that is unsure now errs toward rejection. **It must
be written into the gate's own comment block as provisional**, so the next thread does
not read it as a settled design. Revisit when a deterministic proxy (proper-noun /
brand-shaped token flagging) is cheap to add.

### 13.3 Ruling 9 is AMBIGUOUS and has NOT been actioned — do not guess it

*"fix category general now"* admits two readings that are materially different work:

- **(A) engineering:** make `nextPublishDates` and RFC_013's per-category index
  category-aware **now, while every row is still `general`** — i.e. while the collision
  is dormant and therefore cheap and safe to fix. Nothing user-visible changes.
- **(B) product:** stop using `general` — introduce real categories now (food, music,
  film, cities, generational habits, per §11.1's audience ruling), which touches the
  generator's prompt, the feed's per-category selection and the seed data.

Asked rather than assumed. **Whichever is chosen, the standing constraint from §5/§7.4
holds: both halves must become category-aware in the SAME change, or a category is
silently never scheduled.** Reading (B) forces (A) as a prerequisite; reading (A) does
not require (B).

---

## 14. CORRECTION 2026-08-09 — §10's premise was REVERSED by the owner while §§12-13 were being written, and my reasoning cited it

**§§12.2, 13.1 and §11.3 all rest on "§10 stands: no human approval of publishes."**
A concurrent session put the same §4 finding to the owner in parallel and recorded his
answer in `HANDOFF_2026-08-08b_continue_here.md`'s banner:

> **"A human CAN approve them."** This **reverses the 31 July no-human-approval ruling
> (PLAN §10)**, which was the load-bearing premise of that whole section.

So the sentence I leaned on twice — *"the gate is the only thing between an LLM-written
provocation and a live page"* — **is no longer the policy.** Recorded here rather than
edited away at the point of use, because how the reasoning was built matters.

**What actually changes, and what does not:**

- **Ruling 1 (err toward rejection) STANDS.** The owner ruled it directly. Its
  *justification* softens — a human backstop makes a leak less catastrophic — but the
  trade he chose is unaffected, and §13.1's real point (rulings 1 and 4 are one change,
  because a silently starving pool is the failure mode we chose) is untouched.
- **My option 4 was NOT excluded after all.** I told the owner human approval was "off
  the table by §10". That was true when I wrote it and false within the hour. The
  correction matters because a live option was presented as dead.

### 14.1 ⚠ THE REVERSAL IS POLICY, NOT CODE — and today the difference is load-bearing

**There is no human-approval step implemented anywhere in the publish path.**
`provocation-feed-publisher`'s workflow is one action, `render_provocation_feed`, and
`loadProvocations` selects on `status='approved'` alone. Nothing consults a human, and
no column records that anyone read the text.

So **"a human CAN approve" is currently a statement about intent, not about behaviour.**
Until an approval step exists, the code still behaves exactly as §10 described, and any
reasoning that treats the human backstop as *in place* is wrong in the direction that
matters. **Whoever implements the reversal should add the step before quoting it as a
control** — this is the estate's standing `a-doc-comment-is-not-an-enforcement-mechanism`
shape, arriving as a policy/implementation gap rather than a comment.

### 14.2 The divergence the owner should see, because two sessions were answered differently

On the SAME §4 finding, within hours:

| put by | answer received |
|---|---|
| this session | **candidate 1 — a deterministic pre-judge abuse check** (§12) |
| the concurrent session | **"we can ask a model different ways, several times"** — candidate 3, sharpened to *varied framings*, not repetition |
| this session | the six drafts: **HOLD** until the abuse check ships (§13, ruling 3) |
| the concurrent session | the six drafts **go through the gate** — done; they are now approved and dated |

**The two safety answers COMPOSE and are not in conflict** — the concurrent session said
so itself, and it is right: a deterministic floor removes stochasticity, varied-framing
sampling reduces what survives it, and they sit in series. The sensible reading of both
rulings together is **build candidate 1 AND candidate 3's varied-framing sampler**, which
is strictly what each ruling asked for.

**The drafts answer IS a genuine conflict** and it resolved itself in practice: they were
gated and approved before this session acted on "hold". Not re-litigated here — see §14.3
for the only part that is time-sensitive.

### 14.3 What "hold" was protecting against is now scheduled

Measured 2026-08-09 16:5x: the six LLM provocations are `approved` with `publish_on`
**2026-08-10 … 2026-08-15**. `selectForDate` picks the latest approved row with
`publish_on <= today`, and `provocation-feed-refresh` is **enabled on a 6-hour tick**
(last completed 16:31:45).

- **today** the feed still serves `nobody-wants-personalised-internet` (2026-07-26)
- **tomorrow** it serves `you-love-being-from-your-city` — the first LLM-written
  provocation ever published on this site — **automatically, at the first tick after
  midnight**, with no abuse check built and no approval step in code.

The six were read this session and are on-brief for §11.1's audience; none names a real
person or business, so §11.2's test passes on inspection. **The exposure is procedural,
not textual**: the first automatic LLM publish happens before either safety ruling ships,
and it was never explicitly decided — it is the arithmetic of a date column and a cron.
Flagged to the owner rather than acted on unilaterally, because the dates are another
session's deliberate work and the gating had his sanction.

---

# PLAN 15 — OWNER INSTRUCTION 2026-09-02: daily, and no permission step

*"I'd like to make the challenges change every day and not be restrained by needing my
permission."*

## §15.1 — This REVERSES §10's reversal. It is the third position, not the second.

The trail, so nobody reads the current state as settled:

| date | position | recorded in |
|---|---|---|
| 2026-07-31 | no human approval — "decided, do not re-open" | PLAN §10 |
| 2026-08-09 | human approval REQUIRED; stamp column added | PLAN §14 correction, migration 320/321 |
| **2026-09-02** | **no permission step** | this section, commit `326370d6c` |

**Whoever changes it next is the fourth.** The column, the history and every verdict are
retained; only three query predicates moved, and each site carries the instruction to
restore it in **both** queries or neither — because the defect 320/321 were built to fix
was precisely a comment claiming a predicate the query did not have.

## §15.2 — The three sizing answers (owner, same day)

1. **Quality bar** — *"Rail fatal, and over-generate to absorb it."* The deterministic
   readability rail rejects rather than records, and the generator writes more than it
   needs so rejections do not thin the shelf.
2. **Shelf depth** — **14 days**, up from the §13 ruling of 6.
3. **If it runs dry** — *"create a new set and carry on."* So: **no alerting step.** The
   refill is demand-driven and self-healing; an empty shelf is a trigger, not an incident.

## §15.3 — Why the depth ruling could be reversed without contradicting itself

§13 chose 6 days so that *"one bad batch can never fill a long stretch nobody is
watching"*. **That reasoning was correct and its premise has been withdrawn.** It assumed
the owner was reading each batch; a shallow shelf was a deliberate trade of staleness risk
against unreviewed-content risk, and it was the right trade *while he was the check*.

Unattended, the trade inverts. The staleness risk is no longer hypothetical — it is the
measured failure that prompted this instruction (11 days). So the deeper shelf is not a
loosening of §13; it is §13's own logic applied to the new premise.

## §15.4 — Over-generation is delivered by CADENCE, not batch size

The owner asked to over-generate. The obvious implementation — raise `count` — is the
wrong one here. **4 candidates at an 8000-token budget is the only combination this lane
has actually proven** (2026-08-10). A larger batch risks a truncated completion, and this
pipeline has already been bitten by `stop_reason=max_tokens`, which **presents as
success**.

So the refill runs **twice daily with a cap of 4**, and the over-generation factor
(deficit × 2) lives in the pre-query. Same throughput, no unproven parameter.

## §15.5 — The generator must be demand-driven, or the pool grows without bound

The site consumes one provocation a day; the scheduler dates up to 14. An unconditional
periodic generator therefore pushes publish dates months out and spends model credits for
ever. The `pre_query` returns **no rows** once 14 days of inventory exist, and
`runPreQuery` treats no rows as *skip this task* — so a full shelf costs one `SELECT`.

**The verify block RUNS that pre-query rather than checking it exists**, because both
failure directions are silent: a typo returning nothing means the refill never fires
again, and one returning a row unconditionally means unbounded growth and spend.

## §15.6 — What replaces the human, enumerated

"The owner said so" is a reason to make the change, not a reason it is safe. What actually
carries the safety now:

- **`nextPublishDates` starts at tomorrow, never today.** The six-hour draft-to-live path
  that motivated the 08-09 stamp is structurally unavailable; there is always ≥1 day to
  retire a row. **This is the load-bearing one.**
- **The readability rail, now fatal** — arithmetic, so it cannot drift.
- **The deterministic abuse layer and the RFC_020 §5.2 third-party-harm publish refusal**,
  neither of which ever depended on the stamp.
- **Approved rows are never re-gated**, so the fatal rail cannot retract anything live.

**What is genuinely given up, stated rather than argued away:** nobody reads the text
before it is served. The judge is documented-stochastic, so it is a bonus on top of the
arithmetic floor, not the floor. A row that clears the rail and is wrong in a way no
arithmetic can see **will be served for a day**.

## §15.7 — §10.6 and the rail are mutually unsatisfiable, and that is not a bug

§10.6's rule is *"the corpus IS the specification"*. The owner changed the specification on
2026-08-11. All 28 then-approved entries fail the rail; an entry written before a standard
cannot be expected to meet it.

The resolution is **not** to refresh the calibration corpus — its size is pinned at nine on
purpose, and regenerating it would weaken the guard that stops a calibration silently
shrinking. Instead the acceptance test carries a **narrow, pinned exemption**: a rejection
is tolerated only if **every** fatal rule is `hard_to_read`, and the count is pinned at
exactly 8, in both directions.

## §15.8 — Open, and owed

- **The §10.6 LIVE calibration has NOT been re-run**, and two of its four bad-set fixtures
  changed. Nobody should cite that calibration as current until it is run.
- **Councils `c08d263a` (Go) and `fb31e95e` (config) are submitted, not read.**
- **685 is committed and NOT applied.** It needs: fleet roll → one attended generator run
  → apply. Its guard enforces that order mechanically.
- **Agent types keep the `-manual` suffix** while scheduled. Carried in `description`
  rather than renamed, because renaming `agent_definitions.type` risks in-flight dispatch
  and breaks 321/371's own verify queries.

## §15.9 — Council `c08d263a` (Go half): APPROVED, and what its objections were owed

**APPROVED**, 3 advisory objections, none high-severity. Health signals read the *gate's*
way, not the 4-seat council's: `unreadable: 0`, `gated_by_truncation: false`, and
`reviewers 9 + abstained 8 = 17` — every seat accounted for, the 8 filter-skipped by the
relevance gate. (`abstained: 0` is the **wrong** check on the 16/17-seat gate; it would
just mean every seat fired.)

Two medium objections were sound and are **discharged with evidence, not argument**:

**1. `debug_historian` — the load-bearing safety figure was three days stale.** Correct,
and the seat is right that this lane has a WRONG_CALLS history of exactly that. The
"0 unstamped approved rows" count came from 08-31 and was cited in a 09-02 submission.
**Re-run fresh 2026-09-02:**

| domain | status | total | unstamped |
|---|---|---|---|
| `vonc.com` | approved | **23** | **0** |
| `calibration.vonc.com` | approved | 8 | 8 |

Claim holds, now on a same-day measurement. **The lesson is not "the number was fine" —
it is that a load-bearing count must be re-read at submission time, because the one time
it has moved is the time nobody re-read it.**

**2. `architecture` — `railOnly` pinned to a literal 8 coupled the test to fixture SIZE.**
Correct: a legitimate regeneration of `realProvocations` would fail it for a reason
unrelated to the rule. **Fixed** — the expectation is now derived
(`len(realProvocations) - emptyBodied`), so it states the property (*every body-carrying
entry is pre-rail prose and must fail the rail*) rather than a count. Re-proven by
mutation: loosening the thresholds still collapses it to 0 and fires.

**`guardian` — blast radius asserted by naming convention only.** Measured instead:

```
provocation-feed-publisher   | render_provocation_feed
provocation-gate-calibration | gate_provocation
provocation-generator-manual | gate_provocation, generate_provocations
provocation-scheduler-manual | schedule_provocations
```

**Four agents, all in this lane. No other consumer of any touched action.**

**`reuse_agent` — was the existing voice/prose machinery considered?** Yes, and it is not
a fit, for three reasons rather than by assertion: `check_voice_tells` operates on
**rendered page components**, not generation candidates; its contract is explicitly
*"never an unreviewed auto-rewrite"* — it files HITL items, which is the opposite of a
synchronous reject in a pipeline with no human; and **nothing has ever closed a
`voice_tells` item**, so routing provocations at it would feed a queue with no drain.

### §15.9a — ACCEPTED DEBT, recorded because the seat asked for it to be

The `architecture` seat declined to block but put this on record, and it is right:

> *"a policy TOGGLE being carried as source-code surgery rather than as a single
> config-driven gate … Given three reversals already, a fourth is not hypothetical."*

**Accepted as debt, with a stated trigger: if this policy flips a FOURTH time, the flip
itself is not the work — converting approval into a single config-driven gate is.** Doing
it now would build a switch on the day the owner said he does not want the position it
switches to, which is how mechanisms rot unexercised (the same argument the owner used on
2026-07-29 to refuse default-OFF switches for seams). But the seat's cost analysis stands:
every reversal currently costs a review-and-deploy cycle plus the risk of unpairing two
predicates in two files.

**`guardian`'s sequencing concern** — that the fatal rail ships ahead of its dispatch
config and therefore stands idle — is answered by the config half already existing:
`685_HOLD` is committed (`83ab0b455`) with a mechanical ordering guard, not "tracked" or
promised. The interim state (rail live, nothing generating) is inert rather than harmful:
with no generator running, there is nothing for it to reject, and the site continues
serving what it already served. **Image before config is also the house rule**, and 685's
guard enforces that direction rather than trusting sequencing.

**`debug_historian`'s second point is OWED and not done:** after the roll, verify at the
**pod** that the deployed binary actually rejects a known-bad fixture — a unit test is not
a deployed binary. Added to `RUNBOOK §16f` step 2.

## §15.10 — OWNER RULINGS 2026-09-02 (evening): the rail is the floor, and the switch is not built yet

Both open decisions from §15.8/§15.9a were put to the owner with the live state measured
first. Both answered in one line: *"arithmetic rail is enough for now, we don't need a
switch yet."*

**1. The §10.6 LIVE calibration will NOT be re-run. The arithmetic rail is accepted as the
floor.** ⚠ **This does NOT make the calibration current, and it must not be read that way.**
It retires the re-run as *owed work*; it says nothing about the *figure*. The second half of
§15.8's first bullet stands verbatim — two of the four bad-set fixtures were rewritten to
clear the rail, so **nobody may cite that calibration as current**. What the ruling changes
is only that re-running it is off the critical path. If you are about to lean on a
calibrated *judge* threshold rather than on the rail's arithmetic, this ruling does not
cover you: run it first.

**2. The approval-gate config switch is NOT to be built on this flip.** §15.9a's trigger
stands exactly as recorded — a **fourth** flip makes converting approval into a
config-driven gate the work. The owner is confirming the `architecture` seat's debt, not
discharging it. The two paired predicates in two files remain the mechanism; each site
still carries the instruction to change both or neither.

**What these rulings do NOT touch, and what is therefore the lane's only live gap: nobody
has read the prose.** `[MEASURED 2026-09-02]` Eight rows were written by the generator on
2026-09-02 (four dated 09-05 → 09-08, four undated), and all eight have
`human_approved_at IS NULL`. The two rows dated 09-03 and 09-04 pre-date the change and DO
carry the stamp. **So the first piece served that no human has read is 2026-09-05, not
09-03** — the handoff's "first unattended day is 09-03" is true of the *mechanism* and is
not the deadline for reading. The rail measures sentence and word length; it cannot see a
riddle, and the owner's 08-11 rejection was of the pool's *plainest* entry.
