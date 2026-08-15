# 034 FEATURE — extend the claims audit to cover `site_specs` prose

**Raised:** 2026-08-14 (evening) — **owner decision**, option (b) of the three put to him after
the leopardess donor-run fabrication (PLAN_2026-08-02 decision log, 2026-08-14): *extend the
claims audit to cover `site_specs` prose*, chosen over leaving one field permanently omitted on
one site, and over merging a known-unverifiable field with no detector anywhere.
**Filed:** 2026-08-15 by the owning lane, as the lane's next track.
**Status:** approved in principle, NOT designed, NOT built. Nothing below has an implementation.
**Owning lane:** `docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/`
(the lane also owns the offer analyser, BIZ-032, which is this feature's biggest consumer of
un-checked prose).

## 1. The gap, stated plainly

Every site carries a written premise — `site_specs` rows for `strategy` (including the four
Q-fields restored by B2), `audience`, `identity`, `content_direction` — and **nothing on this
estate checks whether the sentences in those records are true.** The claims machinery that
exists is aimed one layer down, at what pages SAY, not at what the premise RECORDS:

- `check_unverified_claims` scans **deployed** `page_components` / `site_components` HTML and
  stored `content_data`. Its only relationship to `site_specs` is as its comparison REGISTER
  (`evidence_base`, `banned_claims`) — never as its subject. Verified at
  `platform/orchestration/actions/discovery_checks/check_unverified_claims.go` (header, and the
  scan targets) on 2026-08-13, 2026-08-14, and re-verified at HEAD 2026-08-15.
- The rest of the `discovery_checks` package reads `site_specs` as CONFIG (`check_generic_theme`,
  `check_directory`) or as material for duplication comparison — grepped 2026-08-15, no check
  takes premise prose as its subject.
- The build-time gate (`validate_page_content` check 8) shares the same scan engine and the same
  blind spot: it checks page content against the register, and nothing checks the register's
  neighbours.

So a false sentence in a premise record is invisible for ever: no writer renders it onto a page
where the audit would find it (measured 08-12 — nothing reads the `strategy` aspect into page
content), and no audit reads the record itself.

**Verification basis, per the owner ruling of 2026-07-31:** this structural absence was not put
through the 090 diagnosis loop. The substitute, stated plainly: the claim is about the input
surface of one named function and one named package, read first-hand at file:line by two
sessions across three days (08-13, 08-14, 08-15), including a fresh grep of the package at HEAD
today; and the motivating case below was confirmed by a live measurement, not an inference.

## 2. Why the gap stopped being academic this week

1. **The leopardess fabrication (2026-08-13, NOTES).** A `domain-strategist` donor run wrote
   into `recurring_value`: *"The engineering insights blog publishes two technically deep
   articles per week on production concerns (agent failure modes, Kafka consumer group design,
   Postgres schema patterns)…"*. Checked against the live site: **6 blog posts in ~4 months, in
   bursts, on entirely different subjects.** Neither the frequency nor the topic list ever
   existed. This arrived in the ONE spec on the estate protected by a human claims ruling
   (2026-07-16, `hitl`), and it passed the regex screen built from that ruling — no banned term,
   no "department", no numerals. **Only reading it, then checking the most checkable sentence
   against the database, found it.**
2. **B4 now grades every audit-due site against these records, sweep-driven.** The offer
   analyser was enrolled in the improvement loop on 2026-08-15 (migration 409, owner's call). A
   premise carrying an invented fact means the analysis judges a site against something that was
   never true — `bugs_open/161`'s shape (*a false fact causes a claim and then vouches for it*)
   one layer up. `site-review-agent` (B1) reads the same rows into every sweep's strategic
   review.
3. **13 premise records were refreshed on 2026-08-12 and none has ever been claim-checked.** The
   refresh's stability measurement (12 of 13 kept `primary_model`) says the strategist re-derives
   the same commercial answer; it says nothing about whether the newly written sentences are
   true — a correction recorded in NOTES 08-13 after the leopardess case proved the two come
   apart. A 3-site eyeball (dartsonline, gamesdesign, finetuning) found a vaguer, forward-looking
   register with no flatly checkable falsehood — **a sample read by eye, not a check.**

## 3. Design constraints already learned the hard way (all in NOTES 08-13/08-14)

- **A banned-term screen is NOT a claims check.** A banned list is a record of what was already
  caught; the next invention uses different words. The leopardess fabrication carried no banned
  term and no numerals, so both deterministic scans (`ScanBannedClaims`,
  `ScanUnregisteredNumbers`) would have passed it. This check has to READ the prose — an LLM
  pass, or a human one — which makes it unlike the two scans the existing audit runs.
- **Invented specificity is the signal.** A frequency, a count, a named topic list, in a field
  nobody supplied source material for. The working method: find the most checkable sentence and
  check it against the live tables (`pages` answered both halves of the leopardess case:
  cadence from `created_at`, topics from the titles).
- **Findings terminate at HUMAN review, and that is the drain, not a missing one.** The
  content-governance rule in the existing audit (auditors raise items, never rewrite; truth
  decisions are human) applies with MORE force to premise records — two of the 22 current
  strategy-adjacent rows are human-authored (`owner_direction`, `hitl`) and no machine should
  rewrite any premise on the strength of its own reading. The HITL-terminal item shape
  (`needs_human_review`, no handler agent — `check_unverified_claims`'s own precedent) is the
  route. The lane's "a detector ships with its drain or not at all" rule is satisfied by that
  shape; what it forbids is findings parked in `detected` with no terminal state
  (`bugs_open/115`).
- **Do not re-run a generator hoping for cleaner prose** — cherry-picking launders a
  fabrication into the record. The check exists to catch what WAS written, not to be a filter
  that quietly regenerates.

## 4. First population — genuinely disconfirmable

The 13 premises refreshed on 2026-08-12, never claim-checked. The disconfirming outcome — "all
13 clean" — is available and would itself be worth recording (it would say the leopardess case
was donor-run-specific, not refresh-general). The confirming outcome routes each finding to
human review. Either way the answer is new information; the 3-site sample cannot give it.

## 5. Open design questions (for the lane, then the owner where marked)

1. **Vehicle:** a new check in the `discovery_checks` array (registry `init()` + coverage tests
   + image roll, per the settled recipe — but this check needs an LLM call, which the
   deterministic check layer does not currently make), or a config-only auditor agent like B4
   (no image roll, LLM-native, but outside the checker layer's scheduling)? The lane leans
   config-only-agent for the same reasons B4 shipped that way; not decided.
2. **Scope of aspects:** the four Q-fields and `strategy` narrative first (they are what B4 and
   B1 grade against); whether `audience` / `identity` / `content_direction` prose follows.
3. **Cadence and trigger:** on premise write (catches the next donor-run fabrication at birth),
   or sweep-enrolled like B4 (catches drift, costs an LLM call per sweep), or both.
4. **Item shape:** `needs_human_review`-terminal per §3; `item_key` shape to be stated in the
   register entry at ship time (RFC_010 §1's condition) — likely `claims_spec:<aspect>:<site>`.
5. **(Owner, later) whether a finding may quarantine a field** — e.g. mark a premise field
   `unverified` so B4's degraded arm announces it (`inputs_missing`) until a human rules. That
   would connect this check to the analyser's existing honesty machinery with no rewrite. It is
   also new authority on a shared seam, so per the 2026-08-02 ruling it ships as an opt-in
   field, default OFF, if it ships at all.

## 6. Relates to

`features_open/030` §4 + §6 (the analyser's honesty ceiling this feature lifts) · BIZ-032
(register: "its inputs are unverified prose … until then this ceiling stands") ·
`bugs_open/161` (the shape: a register that ratifies the claim it caused) · `bugs_closed/262`
(claims revalidator certifies DB state while the served page drifts — the mirror-image gap, one
layer down) · PLAN_2026-08-02 decision log 2026-08-14 (the owner's ruling and the three options)
· NOTES 2026-08-13 (the fabrication, the screen that passed, the correction on refresh safety).
