# 104 — `banned_claims` is per-site only, so 10 of 15 live sites are unarmed

**Filed** 2026-07-27 from the oufe.com workstream.
**Severity** medium-high — not a code defect. The engine works; it is pointed at
almost nothing. The two most exposed sites in the estate are among the unarmed.
**Class** coverage / platform configuration.
**Status** **CLOSED 2026-07-29 — fixed and LIVE on chassis v1.0.1196**, pod-verified
with markers this change created plus a positive and a negative control, and the gate
wiring covered by five tests that induce both directions. Nine patterns now apply to
all 15 sites, armed or not. See § "FIX BUILT" at the foot of this file, and the
workstream at `docs024_key_docs_latest/bugfix_104_fleetwide_claim_patterns/`.

> **Two things remain true and neither reopens this bug.** (1) The council round is
> **advisory and still open** — round 1 REVISE on a guardian *scope* objection (not a
> veto), round 2 submitted on corr `899ed92e-1bf7-4707-96d8-24f102aa14fa`. Closure
> follows the `/bugs_closed` bar (fixed AND live), which is independent of the verdict
> — the same way `bugs_closed/124` closed carrying a REJECTED verdict. **If the
> guardian escalates to veto and the owner sides with it, the design changes and this
> file reopens**; that is a live possibility, not a formality. (2) The strongest of the
> ten candidate patterns is deliberately **excluded** pending a code-level negation
> guard — that is a separate, unowned, architecture-scope piece of work, filed in
> § "What is deliberately NOT fixed" below, NOT a residual of this bug.

Not a new idea — see "This decision is already filed" below.

## Measurement

```sql
SELECT s.domain, jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'::jsonb)) AS bans
  FROM sites s
  LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='evidence_base' AND ss.is_current
 WHERE s.status NOT IN ('pool','archived') ORDER BY bans DESC NULLS LAST;
```

2026-07-27: **5 of 15** live sites carry a single banned-claim pattern —
leopardessconsulting (19), oufe (28 after mig 226), vonc (9), relojistas (9),
fundamentallyai (6). The other ten carry **zero**, including:

- **vetcomparison.uk** — the site that published fabricated prices for 3,124 named
  real vet practices and required a contemporaneous legal record;
- **idea.uk** — the only site taking real money (£29 reports);
- robot-hands.com, ai-agent-orchestration.com, gamesdesign.co.uk — which have
  `facts[]` and a `writer_block` but no patterns.

## Why this is the defect and not something else

`ScanBannedClaims` (`platform/orchestration/datahelpers/claims.go:284-325`) is a
bare case-insensitive regex over prose blocks. It has no numeric gating —
`businessClaimContextRe` and `isExcludedNumber` gate only `ScanUnregisteredNumbers`
(`claims.go:365,369`). It will catch **whatever patterns a site is given**, about
anyone, numeric or qualitative.

So the engine is general and the wiring is sound. Every reader is keyed on
`site_id` (`validate_page_content.go:967`, `check_unverified_claims.go:229`,
`refresh_evidence_base_action.go:213`), there is no `site_id IS NULL` global row,
no network-level inheritance, and no merge of a base set into per-site sets. **A
pattern learned on one site cannot reach any other, and every new site is born
unarmed.**

Concretely, the lesson this estate paid for twice — fabricated social proof,
invented figures, promises of our own infallibility — is written down five times
and enforced five times, on the five sites where somebody happened to write it.

## This decision is already filed, and its precondition has lapsed

`docs/agent_docs/docs024_key_docs_latest/claims_verification/SPEC_claims_verification.md:250-252`
poses exactly this question and names exactly this use case:

> "Should `banned_claims` be fleet-shareable (some patterns are universal:
> 'Awards Won', invented-client shapes) or strictly per-site? **Proposal:
> per-site only until two sites have evidence bases.**"

Restated in `PLAN_2026-07-16_claims_verification.md:138-139` under what is
deliberately not built. The constraint behind it is landmine 7
(`SPEC:220-223`): *"Do NOT block builds fleet-wide on a layer only one site has
data for"* — entirely correct when n=1.

**There are now 8 sites with evidence bases.** The precondition lapsed months ago
and nothing re-opened the question, because a deferral with a numeric trigger and
no watcher is indistinguishable from a decision.

## The precedent to copy is one directory away

`platform/orchestration/datahelpers/voicetells.go` already solves this exact
shape for the sibling engine:

- `globalTellPhrases()` (`:121-137`) — a hardcoded fleet-wide list;
- unioned with the per-site list at `:109`:
  `append(globalTellPhrases(), spec.VoiceGate.BannedPhrases...)`;
- still gated by the per-site opt-in, so it only fires where the gate is enabled.

Two further precedents for fleet-wide patterns with **no** opt-in at all:
`checkMetaCommentary` (`validate_page_content.go:273,1003-1023`, severity blocker)
and `check_placeholder_contact.go:95-124`.

## Fix candidates, ordered by what closes the door

1. **Mirror `globalTellPhrases` into `ParseEvidenceBase`** (`claims.go:121-128`):
   a `globalBannedClaims()` list unioned with the per-site set. Universal patterns
   become default-on for every armed site, and a lesson learned anywhere protects
   everywhere. Small, precedented, one image roll.
   *Residual:* still gated by `eb != nil`, so the seven sites with no register at
   all remain uncovered.
2. **1, plus make the universal set apply even without a register** — i.e. treat
   the global list like `checkMetaCommentary`, which needs no opt-in. Closes the
   door completely, and is the only candidate that protects a brand-new site on
   its first build. Bigger blast radius: it changes what every site's build gate
   can block, so it wants a council round and a dry-run count of what it would
   have fired on across live pages before it is switched on.
3. **Write the patterns into each site by hand** (config, no code). Zero risk,
   immediate, and it is what mig 226 did for oufe — but it is ten more UPDATEs,
   it does not survive site number sixteen, and it is exactly the state that
   produced this bug.

Recommend 1 now and 2 behind a measured dry-run, because the whole point is that
the next site should not have to be armed by someone remembering.

## Suggested starting content for the universal set

Tested against real copy in `sql_for_agents/226_overclaim_patterns_oufe.sql` —
10/10 fabrication shapes blocked, 13/13 legitimate sentences passed. The
overclaimed-reliability family is a good first candidate for universality because
**no site should ever assert it**: completeness-of-exclusion, verification-of-
everything, invitations to rely, self-accuracy claims, infallibility claims.

The line those patterns draw generalises: **a site may describe what it does; it
may not claim what that guarantees.**

## How to verify a fix

Pick a site with **no** register today. Add a page whose copy asserts "every claim
on this site is verified", and build it. Under candidate 1 with a register seeded,
and under candidate 2 with none, the build must fail with a `claims` blocker.
Then confirm a legitimate process sentence ("we cite each figure and date it")
still builds — **induce both, because a checker that fires on everything is as
useless as one that fires on nothing.**

## Related

- `sql_for_agents/226` (arms oufe), `227` (corrects the false premise that no
  scanner could catch this class).
- `bugs_open/083_HANDOFF_2026-07-26_detected_findings_never_reach_a_handler.md` —
  the other half of the reach problem: even where a site *is* armed, the
  post-deploy sweep that would catch drift effectively never runs.
- `docs024_key_docs_latest/WRONG_CALLS.md` 2026-07-26 — how a wrong belief about
  this engine's reach nearly produced a redundant subsystem.

---

## Triage 2026-07-27, later the same day — the title figure has already moved: 8 unarmed, not 10

Verification sweep, not a fix. **Re-run the § Measurement query before quoting this file** —
it drifted within hours of being written, which is itself the point.

```
 oufe.com 28 · leopardessconsulting.co.uk 19 · ai-agent-orchestration.com 10 ·
 vonc.com 9 · relojistas.com 9 · fundamentallyai.com 6 · finetuning.uk 3
 gaswholesalers.com 0 · robot-hands.com 0 · gamesdesign.co.uk 0 · idea.uk 0 ·
 vetcomparison.uk 0 · system.internal 0 · dartsonline.com 0 · webdesign.co.uk 0
```

**7 of 15 armed** (was 5): ai-agent-orchestration.com gained 10 and finetuning.uk gained 3,
by hand, since this was filed. `site_specs` now holds **9** current `evidence_base` rows.

Nothing about the defect changed — and the drift is the argument for candidate 1, not
against it. Two sites were armed by somebody remembering, one at a time, which is precisely
the state § "Fix candidates" says produces the bug. **The two named worst cases are still at
zero: vetcomparison.uk and idea.uk** — the site that published fabricated prices for 3,124
named practices, and the only site taking real money.

### The filed decision, located and quoted verbatim (this file's § "already filed" is correct)

`claims_verification/SPEC_claims_verification.md` § 10 "Open questions for the owner", q2:

> "Should `banned_claims` be fleet-shareable (some patterns are universal: "Awards Won",
> invented-client shapes) or strictly per-site? Proposal: per-site only until two sites have
> evidence bases."

Restated in `PLAN_2026-07-16_claims_verification.md` under what is deliberately not built:

> "**No fleet-wide banned list.** Per-site until at least two sites have evidence bases
> (spec open question 2, unchanged)."

**The precondition is a number and the number is 9.** So this is not a technical decision
waiting on evidence — it is an **owner call on a deferral whose own trigger has fired**, and
nothing watches it. Nobody needs to re-derive the design; candidate 1 (mirror
`globalTellPhrases`) is precedented one directory away.

### One thing that changes how much candidate 1 buys today

The post-deploy half of this layer does not run at all — `check_unverified_claims` is
reachable only through `improvement-sweep`, disabled since 2026-05-02
(`bugs_open/083`; `claims_unverified` has **0** live rows and **1** ever, from 2026-07-17).
So arming a site fleet-wide arms its **build gate** and nothing else. That is still the
half that matters for new copy, and it should be said plainly when the value is estimated.

---

## Dry run 2026-07-28 — candidate 1 would break three live sites, and 4 of its 7 blocks would be punishing honest disclosure

Session "bugsearch 6". Not a fix: the blast-radius measurement this file asks for.
Workstream `docs024_key_docs_latest/bugfix_104_fleetwide_claim_patterns/`
(PLAN · RUNBOOK · NOTES · README_where_we_are). **The defect as filed is confirmed
and unchanged — 7 of 15 armed, vetcomparison.uk and idea.uk still at zero. What
changed is the recommendation.**

### Method

`cmd/claimscan` — which runs the **same shared engine** as the deploy gate and the
post-deploy audit, and which this file's § Related already names via `226` — over
the **stored `rendered_html` of all 15 live sites (908 components)**, using the 10
tested patterns extracted from `sql_for_agents/226`. Positive control first, both
directions: **6 of 6 overclaim shapes blocked, 3 of 3 legitimate sentences
passed.** Commands and five gotchas in the RUNBOOK.

### Result

**7 findings, 3 sites, 6 surfaces** — leopardessconsulting 1, robot-hands 4,
vonc 2; the other twelve sites clean. **Four of the seven are false positives, and
every one of them fires on a negated sentence:**

- robot-hands `index/features` — "Where manufacturer data has **not** been
  independently verified, that is stated explicitly."
- robot-hands `gripper-detail` — "When a figure **cannot** be independently
  verified, it is marked as unverified rather than carried forward…"
- vonc `about/platform-comparison` ×2 — "…are Spark's own assessment, **not**
  independently verified."

Severity is `blocker` (`validate_page_content.go:930`), so each is a failed page
build — for making exactly the hedged disclosure this layer exists to encourage.
A single pattern, `(fully|independently|externally|properly)
(verified|audited|fact.?checked)`, causes **6 of the 7 hits and all 4 false
positives**. The three true positives are real overclaims and would be correctly
caught ("Every component is verified against production.").

### CORRECTION to § "Fix candidates, ordered by what closes the door"

This file attaches the dry-run requirement to **candidate 2 only**, on the
reasoning that candidate 1 is contained by `eb != nil`. Measured, that containment
does not hold where it matters:

| site | facts[] | banned | `ParseEvidenceBase` |
|---|---|---|---|
| robot-hands.com | 5 | 0 | **non-nil** |
| gamesdesign.co.uk | 4 | 0 | **non-nil** |
| vonc.com | 4 | 9 | non-nil |
| leopardessconsulting.co.uk | 18 | 19 | non-nil |

**Candidate 1 arms 9 of 15 sites, including both sites carrying false positives.**
The residual as written ("still gated by `eb != nil`, so the seven sites with no
register remain uncovered") is literally true but reads as though the gate keeps
these sites out. **Both candidates needed the count.** Also: § Measurement's query
conflates "no row" with "row, empty array" — both render `0`, and the distinction
is what decides candidate 1's reach. Split it: **6 no row, 2 row-with-facts-only,
7 with patterns.**

### Why this was invisible to `226`'s testing, and the design point underneath

`226` tested 10 fabrication shapes blocked and 13 legitimate sentences passed —
careful work. **The pass-list contained no sentence that negates one of its own
patterns**, and on a single site that never came up.

The structural version: `ScanBannedClaims` has **no** false-positive apparatus,
deliberately, and the reason is written beside it (`claims.go:439-441`) — *"Every
match is a KNOWN falsehood for this site (each pattern was audited out by a
human)"*. Its sibling `ScanUnregisteredNumbers` has an elaborate one and says why
— *"a scanner that always reports something is one people stop reading"*.
**Making the patterns fleet-wide removes the premise that justified the absence
while keeping blocker severity.** The human per-site audit *was* the false-positive
apparatus.

### What a fix now has to include

- **No negation-guard prior art exists in the estate.** `voicetells.go:212` looks
  like one and is not — it *checks for* "defines by negation" as a style tell.
- **Go's RE2 has no lookbehind**, so this cannot be fixed in the pattern string.
  A guard must be code, in the shape of `isExcludedNumber`.
- Pattern 7 of the tested set contains the literal alternative **`oufe`** — not
  universalisable verbatim.
- § "How to verify a fix" needs a third induced case: **"where a figure has not
  been independently verified, that is stated" must still build.**

**Nothing is biting today**: every armed site scores **0** against its own live
register. This is a latent trap in the proposal, not a live outage.

### Status

Still **OPEN**, still unowned as a fix, and still the owner's call — it is oufe
decision **O11** (`oufe/DECISIONS_2026-07-26_oufe.md:231`), routed to the owner
because it changes behaviour beyond one site. Now costed rather than cold:
candidate 1 is **not** "small, precedented, one roll" until the set is
negation-safe. The two live options are:

- **(a) ship a narrowed set** with the one offending pattern dropped. **Measured,
  not assumed:** the remaining 9 patterns fire **exactly once** fleet-wide —
  leopardessconsulting `for-engineering-teams/features`, *"Every component is
  verified against production."*, a true positive. So option (a) costs **one page
  build on one site, correctly blocked**, and zero false positives across the
  other 907 components. It needs a human to confirm that one sentence is an
  overclaim, and the copy fixed, before or with the roll.
- **(b) add the negation guard as code**, which is a shared-scanner change and
  therefore architecture-scope under `CLAUDE.md` § "Platform seams and the
  ordering exemption" — council round, and registered in the concept register in
  the commit that ships it. This is the only option that lets the dropped pattern
  (the strongest of the ten — it caught the shape oufe actually shipped) come back.

> **CORRECTED, same session:** the sentence originally here claimed the narrowed
> set "measurably fires on nothing fleet-wide". It was written before it was run,
> and it is wrong — it fires once. Caught by running it. The distinction matters
> because option (a) is not free: it lands a blocker on a live page.

---

## FIX BUILT 2026-07-28 — owner ruled O11, nine patterns are fleet-wide. STAYS OPEN: committed, not yet live.

Owner ruling on oufe decision **O11**, taken with the dry-run numbers above in
front of it: **option "narrowed set, all 15 sites"** — i.e. candidate 2, minus the
negation-prone pattern. Second ruling, on the single true positive:
leopardessconsulting's *"Every component is verified against production."* is
**acceptable** — a claim about a site's own delivered work is not an accuracy
overclaim — so the `every … is verified` pattern was narrowed rather than the copy
changed.

**Commits** `93003e6e0` (the fix + register CLM-015), `7eeb28417` (fixture
correction). **Council** submitted, advisory, `SUBMISSION_CORR
899ed92e-1bf7-4707-96d8-24f102aa14fa` — no verdict at time of writing.

### What shipped

`platform/orchestration/datahelpers/claims_global.go` — `globalBannedClaims()` (9
patterns) and `ScanAllBannedClaims(blocks, eb)`, which scans the fleet-wide set
plus the site's own, dedupes a pattern present in both, and is **nil-safe**: a
site with no `evidence_base` row is scanned. Wired at **both** enforcement
surfaces (`validate_page_content` check 8, `check_unverified_claims`
`scanComponentClaims`), because they are documented to agree by one
implementation. `cmd/claimscan` includes the set by default with `-no-global` to
isolate a candidate set.

Two constraints ruled out the obvious mirror of `globalTellPhrases()`:

1. **Not unioned into `EvidenceBase` at parse time.** `EvidenceBase` is marshalled
   **back** to `site_specs` by `refresh_evidence_base_action.go` and
   `evidence_citations.go`, so seeding `eb.BannedClaims` would persist the
   fleet-wide set into every site's stored register through write paths that never
   intended to touch it — the trap `claims.go` already documents for
   `EvidenceFact.Kind`. The set is held outside any parsed register, joined at
   scan time, and `globalEvidence` is unexported so it cannot reach a writer.
2. **`ParseEvidenceBase`'s nil contract is unchanged.** Only the banned half goes
   fleet-wide; the numeric scan stays strictly opt-in, because its false-positive
   rate is why it is never a blocker.

### Verified before commit, against the corpus rather than in the abstract

- Shipped set over the stored `rendered_html` of **all 15 live sites / 908
  components with no register supplied** — **0 findings**. Nothing on the estate
  becomes unbuildable.
- Positive control, same run, still with no register: **6 of 6** overclaim shapes
  blocked.
- The four previously-false-positive live sentences: **all pass**, and they are
  committed as regression fixtures (verbatim from `rendered_html` — retyping them
  from claimscan's elided snippets produced two quotes no site had published).
- 8 unit tests; `datahelpers`, `actions`, `discovery_checks` suites green at HEAD.

### Why this is still OPEN

**Go changes are inert until an image is rebuilt and rolled.** The defect is
reproducible until it ships, which is the `/bugs_open` bar. To close it:

1. Wait for (or read) the council verdict on `899ed92e`.
2. After the next chassis roll, pod-grep a symbol the change CREATED — not one it
   merely uses: `strings /app/agent-chassis | grep -c "completeness-of-exclusion"`
   should be ≥1, with a positive control.
3. Induce **both** directions on a site with no register (§ "How to verify a fix",
   plus the third case this session added): "every claim on this site is verified"
   must fail with a `claims` blocker; "we cite each figure and date it" and "where
   a figure has not been independently verified, that is stated" must still build.

### What is deliberately NOT fixed, and must not be quietly re-added

`(fully|independently|externally|properly) (verified|audited|fact.?checked)` is
**excluded from the set**. It is the strongest of the ten — it catches the shape
oufe actually shipped live — and it caused 4 of 7 dry-run findings, all on negated
sentences. It needs a **code-level negation guard**: RE2 has no lookbehind, and
there is no negation-guard prior art in the estate. That is a separate,
architecture-scope change and nobody owns it. The regression fixtures mean an
attempt to re-add the pattern fails the test suite rather than a production build.

Also latent: pattern 2, `(does not|doesn't|do not|don't) appear here`, is itself a
negative construction — "prices do not appear here because they change daily"
would match. Zero hits fleet-wide today; flagged in the code beside it and the
most likely source of the next false positive.

---

## 2026-07-29 — two council rounds AFTER closure produced two real improvements. Neither reopened the bug; both are shipped.

Recorded here because the closure above is easy to read as "done", and the useful part
happened afterwards. Round 1 REVISE (guardian, scope), round 2 REVISE (**gating seat
was `debug_historian` at HIGH — not the guardian, whose objections de-escalated to
medium/low and dropped the architecture escalation**), round 3 submitted. Advisory
throughout; none of it gated the roll, and the code was already live.

**What the council found that I had not:**

1. **My measurement was scoped by a variable the gate does not read.** Both my
   population queries filtered on `sites.status` — round 1's reviewer used
   `status='live'` (a value that does not exist here) and got `0`; I corrected it to
   `status NOT IN ('pool','archived')` and got 908. `debug_historian` pointed out I had
   *"replaced it with a different status-based exclusion rather than dropping status as
   a scoping variable entirely"* — and the gate never reads site status at all. There
   were **17 pool sites against my 15**, so the unmeasured slice was larger than the
   measured one. Re-measured unfiltered and grouped: **908 components / 14 sites, all
   `deployed`, and no other row** — pool and system sites hold zero stored components,
   so the filter excluded nothing and 908 was the whole enforcement surface. **The
   answer was favourable and the objection was still correct**; I inherited that filter
   from this file's own § Measurement query, where it is right for a different question.
2. **There was no way to withdraw the set without a build.** Guardian: *"shipping
   without a kill switch is a containment gap independent of how good the
   measurement is."* Now `check_claims_fleet_wide`, a `validate_page_content` config
   key **defaulting TRUE** — DB config is live immediately, so a bad pattern is
   withdrawn fleet-wide in seconds instead of commit + build + roll. Off restores the
   pre-104 scan exactly, and does **not** disarm a site's own audited patterns.
   Gate-only: `DiscoveryCheckContext` has no config map, so a sweep toggle would mean
   a new field on a shared context, and the sweep is unreachable anyway.
3. **A malformed fleet-wide pattern would have degraded silently.**
   `globalEvidence()` inherits `ParseEvidenceBase`'s fallback — an uncompilable regex
   becomes a literal substring — which is right for a site's hand-written config and
   wrong for our own code, because it has no logger and no error path, so a typo
   becomes a near-inert pattern that still looks armed. The fallback stays (panicking
   at init over a regex is worse); two tests now move that failure to CI, including
   the inverse case where an empty pattern compiles and matches **every** block.

Commits: `a1428c908` (lever + guards), `804d021b2` (docs). Full trail in the
workstream NOTES; council correlation `899ed92e-1bf7-4707-96d8-24f102aa14fa`.
