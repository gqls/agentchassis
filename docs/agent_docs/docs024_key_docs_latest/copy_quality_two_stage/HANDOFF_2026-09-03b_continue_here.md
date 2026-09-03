# HANDOFF 2026-09-03b — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-09-03_continue_here.md`** (the
morning file; it carries six stacked correction blocks and is kept for the trail, not for
cold-start). First reads: this file → `OWNER_RULINGS_2026-08-25_six_decisions_on_the_copy_machinery.md`
(THE LEDGER) → `NOTES_two_stage_copy.md` 2026-09-03 entries (the full trail — long, and every
misstep is in it). The owner does NOT want this lane closed.

**Standing rule for whoever picks this up, earned four times today:** when the mechanism belongs to
another lane, **state the finding and ASK for the cause — do not supply one.** Four wrong calls in
one session, all the same shape (a field wired from a shared NAME · a fence's logic from its OUTPUT
· a model's ranking from its RESULTS · a domain's role from its PAGE COUNT); three caught by the
owning lane, one by luck, none by me. Full rows in `WRONG_CALLS.md` 2026-09-03. A sound observation
is the most dangerous carrier for an invented cause.

## The stack as deployed (verified at the artefact, 2026-09-03 ~16:30Z)

- **Chassis v1.0.1359** (pods 13:28Z). Carries the `gutted` change (below). Probe practice:
  capability literal with BOTH controls, never your own commit sha — the `build provenance`
  startup line was EMPTY on `--tail=3000` with pods four minutes old, so the ancestry route was
  unavailable; `kubectl exec <pod> -- grep -acq datahelpers.wordCount /proc/1/exe` with a
  must-be-present (`AcceptNegationRewrite`) and a must-be-absent (`zzNotARealSymbol_deadbeef`)
  control is what proved it.
- **Two owner rulings LIVE, both council-approved:**
  1. **Writer rule 18 = "say less or leave it out"** (migration `739`, applied 12:34Z, verified
     on the loaded row; corr `498080d9` REVISE→**APPROVED** r2). Replaces "ALWAYS better to be
     honest and general". `[MEASURED]` **33 of 60 live sites carry no `evidence_base`**, so this
     is the majority case. ⚠ It edits TWO rules — 19 restates 18 — and aborts unless both anchors
     hit exactly once. Rollback is in **`agent_definitions_backup`** (NOT `agent_definitions`;
     `snapshot_agent()` returns the SOURCE id, so a returned uuid is not evidence a backup exists).
     Binds **per-spawn**, not per-instant: an in-flight writer carries its own inline
     `config.workflow`.
  2. **Copy-gate `gutted` floor accepts his truncations** (`7cc16a5d0`, corr `b9b5fdf8`
     APPROVED; live in 1359). Was `len(to) < len(from)*2/5` — a proportion, backwards for a
     truncation repair (his own worked example scored 29.5% and was refused). Now
     `wordCount(to) < 5 || len(to) < len(from)/4`, constants in a measured gap (2 words/14.5% vs
     6 words/29.5%), mutation-proven.
- **The planner canary is ANSWERED for `build-site-planner`** — rendered the build standard at
  10:40:15Z and 14:15:16Z, `has_standard=t`, `unrendered=f`, read at the artefact. `678`
  (content-gap-planner) unobserved; `679` (visual-designer) **never will be — the row is dead**:
  zero `llm_call_log` rows all-history, zero live configs name it, the one Go hit is a storage
  env-var list. ⚠ Needle is the carrier-only form `BUILD STANDARD (applies to every site,
  regardless of inputs). Aim` — NOT `stands comparison with the strongest sites`, which also
  matches `domain-research-classifier`'s hard-coded copy. ⚠ `diagnose-agent` carries the form
  without being a consumer. ⚠ A "zero rows since T" here can turn non-zero LATER (rows arrive
  after the `created_at` they carry) — two false "still open" readings today.
- **BANNED_REGISTER v2** live end-to-end (unchanged from the morning file).

## Waiting on the OWNER (explained plainly in chat 2026-09-03 ~16:00Z; raise, don't re-derive)

1. **Carrier 675 dropped the build standard's SCOPE paragraph** while asserting "verbatim,
   byte-identical, ONE trim". The missing text is the counterweight — *"This standard governs
   QUALITY and FIT, not scope. Do not invent services, pages, features, or facts beyond what the
   evidence supports…"* — and the three opted-in rows are exactly the agents that decide what
   pages exist. **Window CLOSED**: build-site-planner has now consumed it twice. Can't restore
   verbatim (classifier-specific nouns: `confidence fields`, `adopted sites`); needs generalising,
   which is his wording. **Recommended: restore, generalised.**
2. **The whole 897-char block renders as ONE markdown H2** — `## {{.build_standard}}` was
   correct against the source (title · newline · body) but 675 replaced the line break with a full
   stop. Low severity. **Fold into (1)**: both change the carrier text, and the fix moves the canary
   needle in `NOTES`, this file and `scripts/fire-content-gap-planner.sh` in the SAME commit.
3. **`679` opted a dead agent in.** Roll back, revive, or leave recorded. Recommended: roll back
   unless a design agent is meant to exist.
4. **Which built sites are for sale** (the one that BLOCKS work): the about-commercial-block
   renders on nothing until each site has a `site_specs.commercial` row, and today **exactly one**
   site fleet-wide has `for_sale_requested=true` (relojistas.com). He said built sites "should be
   in sedo now" — `[MEASURED 16:20Z]` **0 of 42 built sites are in draft8** (the latest Sedo sheet,
   15:54Z). They are HELD by the sedo lane pending a real price; the valuation lane's coverage is
   **588 of 2,945 (20%)** and they asked that provisional tiers not drive a floor (D2 makes a
   wrong price INVISIBLE — it shows only in which offers never arrive). **He can unblock it by
   naming floors himself**; he gave webdesign.uk as potentially seven figures.
5. `plain_words` boundary case; banned-WORDS repair arm; spec-fed class; xAI top-up; word to the
   offer lane on Decision D — all unchanged from the morning file.

## The about-page / Sedo seam (this lane owns the block's copy; sedo/valuation own the lists)

- `about-commercial-block` is live on **3** sites (relojistas.com `sobre-nosotros`,
  advertise.co.uk `about`, finetuning.uk `about`). Destination is **config** —
  `site_specs.commercial.marketplace_url`, `on_missing: skip_field` — neither "afternic" nor
  "sedo" is in the 3,699-char template. Owner ruled **"Yes, point to Sedo"** (settles D1's on-site
  CTA half). **Blocked**: relojistas has never been listed on Sedo (zero rows in every sheet), so
  there is no URL to point at; **no Sedo URL pattern is documented anywhere — do not compose one.**
  relojistas keeps its working GoDaddy/Afternic lander meanwhile.
- **`leopardessconsulting.co.uk` is PERMANENTLY excluded** — owner verbatim *"no
  leopardessconsulting need not be listed"*. The sedo lane holds it in a durable owner-withdrawal
  file. ⚠ **CORRECTED 2026-09-03 ~17:10Z: it is the owner's OWN consultancy, not a paying
  client's site.** D4's worked example (*"a paying client's site (leopardess)"*) is wrong on the
  facts — found by the valuation lane from the live site's own copy and confirmed by the owner to
  them. I relayed D4's gloss to three lanes as fact; D4 is now corrected at source. The exclusion
  stands on his word, not on a relationship-breach argument that cannot apply to his own site. Also `copyonline.co.uk` withdrawn (owner's own,
  possibly his wife's; either outcome was acceptable to him).
- **Both webdesign domains are in scope for sale and "will be the same endpoint one day."**
  ⚠ **`webdesign.uk` (18 pages) is the SHOPFRONT** — `CLAUDE.md:716` — and the one he valued at
  seven figures. I told two audiences it was `webdesign.co.uk` (155 pages); that was wrong and is
  corrected. The consolidation plan + both-for-sale is a coupling (sell one, break the merge) —
  raised to the sedo lane, commercially theirs.
- `commercial.tier` (1|2|3, resolver-owned) is **NOT** the valuation lane's A–E tier — different
  vocabulary, no mapping, theirs wired to nothing. Do not identify them.
- **66 of 2,945 domains are fenced** (48 live sites, 18 owner-withdrawn families); **6 are domains
  he no longer owns and were ALREADY listed at Afternic at $10k–$50k** — the afternic lane has
  them queued for removal. Their finding, attributed, not re-measured by me.

## Handed across the seam this session (not this lane's to finish)

- **Offer-analyser ordering prompt** (`747_…_HOLD.sql`, corr `aeaf9f88` — r1 approved on stale
  wording, **r2 APPROVED 16:18Z** on the current wording; still `_HOLD`, unapplied, a component
  fence run owed first). Carries TWO owner rulings: price into the exemplar, and the absolute
  *"never a description of us or of our inventory"* softened to *deprioritise — rank last by
  default, one clause, prefer the reader-benefit form*. **The mechanism, after three corrections
  (theirs), final form:** `from_field` is open, so the four named fields set a PRIOR not a wall;
  the binding constraint was the ABSOLUTE — five money_flow points exist and none states a price,
  so **zero of seven price questions were answered by a price point**; the model's refusals were
  correct throughout. **Owed post-apply, pre-registered:** re-run answer-rate-by-own-field for
  `money_flow` (baseline **2/7** vs 24/24, 29/30 elsewhere). Climbs toward the others ⇒ done.
  Plateaus ⇒ nothing asks a point to engage the NAMED alternative ("free AI") and a second clause
  is needed. **Do not fold that into 747.** ⚠ The fleet stretch rate (14 of 91 = 15.4%) is
  unexplained and NOT closed by "the asymmetry produces true unanswereds".
- **`bugs_open/420`** (SLUG-resolve) is with the `420 425` lane. They corrected my Scope note
  (`brief-negation-check` does NOT share the walker — struck through, `87780485a`), found
  `TestWalkerSkipsNonProse` vacuous (fixture values are single tokens), and found `name` carries
  two OPPOSITE contracts by producer — identity where a `url` sibling exists (898/898 no-space),
  display where none (752/752 prose-shaped), zero crossover. **No downstream guard stops a rewrite
  DROPPING a name** (`invented_name` only; figures are symmetric, names are not) — they are building
  `dropped_name` on the `dropped_figure`/`protectFrom` pattern as a control in
  `AcceptNegationRewrite`. The nightly `brief-negation-check` (`40 7 * * *`) pair is a **CONTROL**
  predicting NO movement; baseline 09-03 07:41Z = **11 of 39**; record N-of-M.
- **The empty-writer-context finding** (41% of 6,931 writer calls carry an empty `Company:`;
  `build_render_context` never merges `site_specs` while `identity` holds the name): diagnosis
  corr `fbe2be91` returned **UNVERIFIABLE (iteration cap)** — neither confirmed nor refuted.
  First-hand chain stands as evidence, not verdict. **Priority DROPPED** by the owner's reframing:
  feeding a name to a writer with no facts produces CONFIDENT vacuity. Worth fixing for evidenced
  sites; not the vacuity cause.

## The next real engineering (sized, NOT started, needs his word on shape)

**"Declined" is not distinguishable from "lost".** The licence to say nothing exists at page
level (planner: "too thin to describe"), one section type (his 2026-08-25 ruling), and field level
(optional → `""`). Rule 18 now lets the writer say LESS inside a planned section — but it cannot
hand a section back EMPTY without tripping the completeness/shrink/component floors, which exist
to catch content that was LOST. Prompt wording cannot fix this; a declined section has to be a
representable state. Two shapes he has not chosen between: teach the writer to decline (and make
the floors read the reason), or stop the planner planning the about page on no-evidence sites —
**except he wants an about page on every site** (it will carry the Sedo link), so it is the first.
Measured context: thin sites already skew to tools/articles (**16.4%** content pages vs 20.5% on
evidenced sites); the residual is `about` on 11 of 15 no-evidence sites.

## Standing bugs this lane owns or watches

- `bugs_open/422` — repair-vs-shrink-floor (unchanged). New evidence for it: technical-details
  rebuild lost **33.5%** of words and the gate removed **36** of the 612 — the WRITER shrinks,
  not the repair; the homepage rebuild grew +13.2%. So shrink is per-page writer behaviour, not a
  property of rebuilds.
- `bugs_open/456` (the SLUG `writer_emitted_a_malformed_closing_tag` — the number is doubled) —
  filed by the finetuning lane from this lane's after-pass: `<strong>…</strom>` reached the served
  page; nothing validates HTML between writer and bucket. Owner chose no re-run.
- 443 read-rule (theirs) unchanged.

## Peer state + channels (all exchanges CLOSED clean unless noted)

finetuning [060349] · sedo [9822ef] · domain valuation [22cc6c] · offer analyser [fb5808 /
351737 sock] · 420 425 [1618280 sock] · afternic [17bc44] (never messaged). A peer cannot grant
escalation. HEAD at write time `810184336` (another lane's), builds. Two migrations of the offer
lane's are `_HOLD` — an unscoped `run-migrations.sh --apply` by any session would take a
non-HOLD pending file, which is why they were renamed.

## This session's landmine harvest (all in LANDMINES/WRONG_CALLS with the checks)

- **A measured finding is the most dangerous carrier for an invented cause** — four times.
- **A before/after of an artefact cannot measure a step that REGENERATES the artefact** — read
  the component's own report (`copy_gate_N` carries before/after/rejected) instead.
- **A check that shares state with the thing it checks confirms intention, never outcome** —
  a verify must re-read the LIVE row (a peer's verify caught their own half-finished edit only
  because it did).
- **"I called the backup function" is not evidence a backup exists** — read the backup table.
- **A `content_data::text ILIKE` census counts MENTIONS, not settings** — and the extra rows are
  ALTERNATIVES that look safer than the real target (LANDMINES, new).
- **Two rates over one corpus are not comparable until you know what each counts** (reuse events
  vs questions touching a reused point; served HTML vs stored `content_data`; bytes vs chars —
  9,590 chars / 9,642 bytes / 9,643 with psql's newline).
- **Same-correlation resubmission is what keeps a `Council-Submitted:` trailer honest when the
  file moves under an in-flight round** — `098` resolves to the LATEST verdict; a fresh
  correlation would leave the stale approval credited with no mismatch flag.
- **An APPROVED verdict on a diverged file is the dangerous one, not the safe one.**
- **A `_HOLD` rename is a mechanism; a header banner is a hope.**
