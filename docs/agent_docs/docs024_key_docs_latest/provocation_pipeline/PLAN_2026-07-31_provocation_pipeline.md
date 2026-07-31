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
- §7.1 (*human approval or automated gate only?*) — **still open**, and 9.1 raises
  the stakes: an adversarial upstream source makes an unreviewed publish path
  riskier than it was when the pool was hand-written.
