# NOTES — fleet-wide claim patterns (`bugs_open/104`)

Append-only, newest at the bottom. Evidence, commands, what the system actually
said, and every misstep.

---

## 2026-07-28 — session "bugsearch 6", opening

Picked `104` from `HANDOFF_2026-07-28_bugsearch2_session.md` §4, which named it
unowned and the natural follow-on to `106`.

**Ownership re-checked, because the handoff's check was hours old.**
`scripts/who-owns.py 104` returns **"OWNED or recently active"** with *no owning
workstream identified* — the only two commits are the filing
(`02da9491e`) and a triage sweep (`e2634eeb7`), both 07-27. So the verdict is the
tool being conservative about recency, not a competing thread. But
`platform/orchestration/datahelpers/claims.go` was committed **today at 13:19**
(`2e665591b`, oufe), so the file is live under another lane's hands even though
the bug is not.

`104` is also **oufe decision O11** (`oufe/DECISIONS_2026-07-26_oufe.md:231`),
where it sits under *"These are new since the register was written, and they are
the owner's because each changes behaviour beyond this site."* oufe lists `104`
under "Bugs filed here", i.e. filed for someone else, not owned. Its sibling O13
(`bugs_open/105`) has since been implemented (`606f485f7`, `b18dd564d`), so the
decision list is live and being worked through.

**Grounded the figures before quoting them** (`104` says to, and says it drifted
within hours of being written). Unchanged from the 07-27 triage: **7 of 15 armed,
8 at zero**, and the two named worst cases — vetcomparison.uk and idea.uk — are
still at zero.

**The bug file's line numbers have already drifted.** It cites
`ScanBannedClaims` at `claims.go:284-325` and `ParseEvidenceBase` at `:121-128`;
they are now at `:442` and `:268`. `claims.go` grew a long comment block about
`factKindAliases` in between (that block references *a different* bug file's
candidate 1 — 105's, not 104's; easy to misread).

### Reuse before building

Was about to write a Go harness to run the patterns against live pages. Did not
need to: **`cmd/claimscan` already exists** and runs the same shared engine as the
gate — and `sql_for_agents/226`'s own verify section names it. Found it by reading
226 to the end rather than by grep. `go build ./cmd/claimscan` succeeds against
the shared tree.

### Misstep 1 — the loop terminated after one site and looked like a result

The first fleet run printed exactly one line (`ai-agent-orchestration.com … 0`)
followed by `=== DONE ===`. `kubectl exec -i` **consumes the `while read` loop's
stdin**, so the site list was eaten after the first iteration. Exit 0, no error.
Fixed with `mapfile` + `</dev/null` on every exec. In the RUNBOOK as GOTCHA 1.

### Misstep 2 — I grepped for a string the tool never prints

Second run completed all 15 sites and reported **0 findings everywhere**. That was
a false all-clear: I counted `grep -c "banned_claim"`, but claimscan prints line
prefixes `BANNED` and `NUMBER`. `banned_claim` is the JSON `check` value and
appears nowhere in CLI output. The correct count is 7, not 0.

This is the [check answers the question you ENCODED] failure exactly: there was no
filter to notice, the query was simply about something else. What caught it was
the **positive control**, not review of the code.

### Misstep 3 — two sites returned blank, not zero

`gamesdesign.co.uk` and `robot-hands.com` printed an **empty** count. `file` says
both scan outputs are `Non-ISO extended-ASCII text`, and plain `grep -c` returns
empty with **no error** on them. `LC_ALL=C grep -ac` fixes it. Both of these
sites turned out to matter (see below), so this would have hidden the finding.

### The positive control, which is what actually found the bug

Nine synthetic components, six that must block and three that must pass:

```
BANNED ctl_block_1 "claim without a source does not appear" — completeness-of-exclusion
BANNED ctl_block_1 "does not appear here"                   — short form
BANNED ctl_block_2 "Every figure is verified"               — verification-of-everything
BANNED ctl_block_3 "You can rely on our"                    — invites reliance
BANNED ctl_block_4 "Our reporting is always accurate"        — self accuracy overclaim
BANNED ctl_block_5 "Our method is not a disclaimer"          — repudiates the caveat
BANNED ctl_block_6 "Guaranteed accurate"                     — accuracy guarantee
claimscan: 7 finding(s) across 9 component(s)   exit=1
```

**6 of 6 block cases fired; 3 of 3 legitimate sentences passed** ("We cite each
figure and date it", "The statute is the authoritative text", "We check our
sources before we publish"). The engine works and the harness is sound.

### The finding: 7 real hits, and 4 of them are false positives

10 patterns from `226` × **908 components across all 15 live sites**:

```
leopardessconsulting.co.uk  97 comps  1
robot-hands.com             90 comps  4
vonc.com                    55 comps  2
(all twelve others)                   0
TOTAL 7, on 6 distinct page+slot surfaces
```

Every one, read in full:

| site / surface | matched | verdict |
|---|---|---|
| leopardess `for-engineering-teams/features` | "Every component is verified against production." | **true positive** — verification-of-everything |
| robot-hands `how-it-works` | "…where available, independently verified test data." | borderline — hedged, but does assert independent verification |
| robot-hands `gripper-catalog` | "…pulled from manufacturer datasheets and independently verified." | **true positive** |
| robot-hands `index/features` | "Where manufacturer data has **not** been independently verified, that is stated explicitly." | **FALSE POSITIVE** |
| robot-hands `gripper-detail` | "When a figure **cannot** be independently verified, it is marked as unverified rather than carried forward…" | **FALSE POSITIVE** |
| vonc `about/platform-comparison` ×2 | "…are Spark's own assessment, **not** independently verified." | **FALSE POSITIVE** ×2 |

**All four false positives fire on a negated sentence** — the honest disclosure
this layer exists to encourage. Severity is `blocker`
(`validate_page_content.go:930`), so each one fails a whole page build. One
pattern, `(fully|independently|externally|properly) (verified|audited|fact.?checked)`,
accounts for **6 of the 7 hits and all 4 false positives**.

### Two checks that make the consequence certain rather than likely

1. **Candidate 1 does reach those sites.** It is gated on `ParseEvidenceBase`
   returning non-nil, which needs `facts[]` **or** patterns:
   `robot-hands facts=5 banned=0 → non-nil`, `gamesdesign facts=4 banned=0 →
   non-nil`, `vonc facts=4 banned=9`, `leopardess facts=18 banned=19`. So
   candidate 1 arms **9 of 15** sites, including both sites carrying false
   positives. `104`'s residual note ("still gated by `eb != nil`") is literally
   true but reads as though the gate keeps these sites out. It does not.
2. **There is no negation-guard prior art**, and it cannot be done in regex here.
   I grepped for negation handling and got a hit at `voicetells.go:212` — which on
   reading is a *check for* "defines by negation ('not X, but Y')" as a style
   tell, not a guard. **Semantic coincidence; I nearly cited it as precedent.**
   Go's RE2 has no lookbehind, so a guard must be code, in the shape of
   `isExcludedNumber`.

### Nothing is biting today

Each site scanned against **its own** live register: **0 findings, every site.**
6 sites have no `evidence_base` row; 2 have a row with 0 patterns but non-empty
`facts[]`; 7 have patterns. So this is a latent trap in the *proposal*, not a live
outage — no page is currently unbuildable.

### Misstep 4 — I suppressed stderr and it became a data claim

The first self-scan table reported **vonc: "no register"**. vonc has a current
`evidence_base` row with 9 patterns and 4,651 characters. The per-site fetch had
failed transiently and `2>/dev/null` hid it, so a `kubectl` flake was rendered as
a fact about the estate — on the single site whose register mattered most to the
finding. Caught only because it contradicted the § Measurement query I had run ten
minutes earlier. Now retried 3× with `FETCH_FAIL` printed distinctly from
`no-row`. In the RUNBOOK as GOTCHA 2.

### The design point underneath all of this

`ScanBannedClaims` has **no** false-positive apparatus, deliberately and with the
reason written down (`claims.go:439-441`): *"Every match is a KNOWN falsehood for
this site (each pattern was audited out by a human) — callers treat findings as
blockers."* Its sibling `ScanUnregisteredNumbers` has an elaborate one, and says
why: *"Noise is not harmless in a checker: a scanner that always reports something
is one people stop reading."*

**Fleet-wide patterns remove the premise that justified the absence, and keep
blocker severity.** Nobody had audited the oufe set against the other fourteen
sites' copy; this is the first time it was done, and it found four false
positives in the first ten patterns. `226`'s own test was "10 fabrication shapes
blocked, 13 legitimate sentences passed" — thorough, and it still missed this,
because **the pass-list contained no sentence that negates one of its own
patterns.**

### Side observations, not chased

- **vonc has two `page_components` rows on the same page+slot**
  (`about`/`platform-comparison`), both length 7,113 — ids `8847777f…` and
  `4a3d50e8…`. That is why its 2 findings are one sentence counted twice. Not
  investigated; may be a duplicate-component defect worth its own look.
- Pattern 7 of the tested set contains the literal alternative **`oufe`**, so the
  set is not universalisable verbatim regardless of the negation question.
- `104`'s § Measurement query conflates "no row" with "row, empty array" — both
  render as `0`. The distinction is load-bearing for candidate 1's reach.

---

## 2026-07-28, later — the narrowed set, measured rather than assumed

I wrote "a narrowed set measurably fires on nothing fleet-wide" into `bugs_open/104`
**before running it**. Then ran it: dropping the single offending pattern
`(fully|independently|externally|properly) (verified|audited|fact.?checked)` leaves
9 patterns which fire **exactly once** across all 908 components:

```
leopardessconsulting.co.uk  for-engineering-teams/features
  "Every component is verified"  — verification-of-everything: a claim about outcomes, not process
  …Every component is verified against production. The architecture is built on tools your…
TOTAL with narrowed 9-pattern set: 1
```

That is a **true positive**, so option (a) is still viable — but it is not free:
it lands a `blocker` on a live leopardess page until either the copy is fixed or a
human rules the sentence acceptable. Corrected in `104` with a visible correction
block. **"Fires on nothing" was an inference stated in the same voice as a
finding, and it took one command to falsify.**

Note what the 9-pattern result also says: with the negation-prone pattern removed,
**zero false positives across 907 other components.** So the false-positive class
is concentrated in one pattern, not diffuse — which is what makes option (a) a real
choice rather than a retreat.

## Housekeeping — one of my "new" gotchas was already written down

`016b` §9 already has *"A command that reads stdin truncates the `while read` loop
that calls it (2026-07-18)"* at line 1110 — precisely misstep 1. I hit a known
pattern, so it went in the RUNBOOK (where the command lives) and **not** into §9
again. The new §9 entry is the one thing here that is genuinely transferable and
not already recorded: *"A checker's missing false-positive apparatus may BE its
per-site human audit."*

---

## 2026-07-28, evening — the owner ruled, the fix is built, and I broke HEAD on the way

### The two rulings

O11: **narrowed set, all 15 sites** (candidate 2 minus the negation-prone pattern).
And on the single true positive: leopardess's *"Every component is verified against
production."* is **acceptable**, so the `every … is verified` pattern was narrowed
to claim/content nouns instead of the copy being changed.

Worth noting what made the door-closing option affordable: **the dry run was
register-blind all along.** I had scanned all 15 sites regardless of whether each
would be armed, which is candidate 2's shape, not candidate 1's. So candidate 2's
blast radius was already measured — same single finding — and the containment
candidate 1 offered turned out to buy nothing while leaving vetcomparison.uk and
idea.uk unprotected. I did not plan that; I noticed it when writing the options up.

### What the design constraint turned out to be

The obvious implementation is the one `104` recommends: mirror `globalTellPhrases()`
by unioning the global set into `ParseEvidenceBase`. **That would have been a data
corruption bug.** `EvidenceBase` is marshalled BACK to `site_specs` by
`refresh_evidence_base_action.go` and `evidence_citations.go`, so seeding
`eb.BannedClaims` at parse time persists the fleet-wide set into every site's
stored register through write paths that never intended to touch it. `claims.go`
documents that exact trap two hundred lines above, for `EvidenceFact.Kind`, and I
only found it because I went to read how `BannedClaim` was structured. **The
precedent named in the bug file was the wrong precedent, and the reason is in the
same file as the code.**

So the set is held outside any parsed register (`claims_global.go`), joined at scan
time by `ScanAllBannedClaims`, and `globalEvidence` is unexported so it cannot
reach a writer. `ParseEvidenceBase`'s nil contract is untouched.

### Verification, and the number that matters

Shipped set, no register supplied, all 15 sites / 908 components: **0 BANNED**.
Positive control same run: **6 of 6** blocked. The four negated sentences: pass.
`0 findings` is the claim I would most want challenged, so it is worth being
explicit that it was produced by the built binary (`go build ./cmd/claimscan`
after the change), not by the candidate JSON I started with.

### Misstep 5 — my commit shipped another session's half-finished dependency, and HEAD did not compile

`93003e6e0` named `validate_page_content.go` and `check_unverified_claims.go`
because I had edited them. Both also carried the bugs_open/102 `ClaimSurface`
plumbing from another session, **uncommitted** — so my commit shipped the two
CONSUMERS of a type whose DEFINITION, in `claims.go`, was still in the working
tree. HEAD stopped compiling:

```
check_unverified_claims.go:295: undefined: datahelpers.ClaimSurface
check_unverified_claims.go:414: too many arguments in call to eb.ScanUnregisteredNumbers
```

`make build-<service>` builds from committed HEAD, so any session starting an
image build in that window would have got a failure attributed to my commit.

**What caught it:** the diff's insertion/deletion counts looked too large for the
edits I remembered, and a hunk header read `func scanComponentClaims(html,
slotName string, eb *datahelpers.EvidenceBase)` — without the `surface` parameter
that was in my tree. **A hunk whose context is code you did not write is the
tell.**

**What fixed it:** forward-only. I prepared the missing half as a labelled
`sweep:` commit; by the time it ran, the owning session had committed it
themselves (`3ddb4ed2d`, ~4 minutes after mine). HEAD compiles and all three
suites pass there now — verified in a clean `git archive HEAD` export, not in the
tree. Logged in `WRONG_CALLS.md`, because the commit message asserted "suites
green" and that was true of my tree and false of the commit.

The pathspec discipline did work on everything it could: five other sessions'
modified files stayed out. It cannot help when two sessions edit one file, and
this is what that costs.

### Misstep 6 — two of my regression fixtures were sentences no site had published

claimscan elides its snippets with ellipses. I retyped two fixtures from that
output instead of from the source, and both were wrong: vonc's real sentence
begins *"Competitor characterisations reflect general platform mechanics…"*, not
*"These reflect platform mechanics…"*; the robot-hands catalogue one is a long
sentence about six actuation technologies, not the short paraphrase I wrote.

The tests passed either way — they assert a negated sentence is *not* flagged, and
a paraphrase negates too. **That is exactly why it was worth fixing**: a fixture
whose comment says "real copy from a live site" must be real copy, or the next
person cites a quote that never existed. Caught while checking `grounded_in`
fidelity for the council submission, by decoding the component base64 and grepping
the actual sentence. Fixed in `7eeb28417` and in the submission.

### Council

Submitted, advisory: `SUBMISSION_CORR 899ed92e-1bf7-4707-96d8-24f102aa14fa`. Queue
showed two other councils mid-flight (`review_architecture` ×2 at depth 440–450),
so ~30 minutes is the expectation, not 2. No verdict yet at time of writing.

---

## 2026-07-29 — the fix is LIVE, the council said REVISE, and one of its own checks was wrong

### Live, and verified with discriminating markers

Another session rolled the fleet to **v1.0.1196** (pods started 2026-07-28T22:37Z /
22:38Z, both on that tag). Pod-grep with strings this change created:
`completeness-of-exclusion` → **3** (three patterns share that reason),
`verification-of-everything` → **1**, positive control `banned_claim` → 2, negative
control → 0. So the fleet-wide set is enforcing on every site's build now.

**The standing fleet invariant was checked in the same pod** and passes:
`unknown execution-context field` → **1** (migration 258 needs chassis ≥1191,
`bugs_closed/124`). Recording it because it is owed after *every* roll and this was
a roll nobody from that workstream witnessed.

### Council round 1: REVISE — and NOT a harness artefact

`decision=revise`, `reviewers=12`, `abstained=5`, **`unreadable=0`** — so it is
substance, not the ~11% one-seat-unparseable-JSON failure. Gating objection from
**guardian**, verdict `object` (not veto), severity **high** on edit 3.

Its reasoning is worth keeping: it explicitly declined to invoke the
foundational-plumbing preference, and said what tipped it from veto to object was
"the unusually thorough containment already done — dry-run against all 908 live
components with a concrete zero-finding result, regression fixtures for every false
positive found, explicit exclusion of the one unsafe pattern pending a negation
guard, and the change being inert until the next chassis build." **The last of those
four is no longer true**, which is why round 2 discloses it first rather than
letting a reviewer discover it.

Its `missing` was a fair, factual question: is O11 a council-reviewed decision or an
out-of-band owner ruling being rubber-stamped? And it warned it would escalate to
veto if the precedent check came back empty.

### One of the council's own read-only checks returned a WRONG 0, and it was the load-bearing one

The check *"checks the claimed 908-component dry-run population against actual live
rendered_html row count, since the plan's zero-false-positive claim rests entirely
on that measurement"* returned:

```
count
0
```

That is a **wrong-filter artefact, not a refutation.** `sites.status` has no value
`'live'` in this estate — the vocabulary is `deployed` (14), `pool` (17), `system`
(1), which the council's own *sibling* check printed three paragraphs earlier. Re-run
with the real predicate the population is **908**, exactly as claimed. This is the
trap already in 016b §9 as *"sites.status: vocabulary, the legacy 'active' value, and
a wrong blast-radius filter"*.

**Worth sitting with: a seat reading `count 0` would reasonably have concluded the
measurement was fabricated.** The reviewers' checks are machine-generated and are not
themselves reviewed, so a wrong query in the evidence pack is indistinguishable from
a wrong claim in the plan — and it lands on the author to notice. That is a property
of the gate, not of this submission.

### The caller census the guardian asked for

Full-repo, non-test: `validate_page_content.go:980`, `check_unverified_claims.go:405`,
`cmd/claimscan/main.go:132,134`, plus `claims_global.go:196,203` inside the join
itself. **Three production call sites, all three already in the plan.** No other
`eb == nil` guard anywhere gates a banned-claim scan — the rest are the stat lane,
the kind accessors, `ScanBannedClaims`' own nil-receiver guard, and
`ScanUnregisteredNumbers` (untouched, still opt-in). There is no fourth branch.

Round 2 submitted on the same correlation with all of the above, the sketch fix
editquality correctly demanded (`GlobalBannedClaimCount` was called in edit 5's
sketch and never shown in edit 1's), and an explicit statement that if the scope
concern survives the facts, it goes to the owner rather than to round 3.

### Misstep 7 — my first gate test could not have failed in the direction that mattered

I wrote the gate-wiring test asserting on the returned `issues` list, treating a
returned error as a test failure. On any blocker the action returns **`(nil, error)`** —
the error *is* how the build fails. So my first version **failed the test on the one
outcome 104 wants**, and had I "fixed" it by loosening the assertion instead of
reading the contract, the test would have passed on both a working and a broken
gate: the `issues` map is nil on the blocker path, so `len(claimsIssues(nil)) == 0`
would have looked like "no findings" either way.

What caught it was the failure message itself — `content validation failed: 1
blockers, 0 errors` — which is the gate working. Same family as the `psql -t -A`
guard that could not fail, from the bugsearch-2 session. **A test whose harness
discards the success signal is worse than no test.** The discrimination is now
explicit: identical harness, one input errors, four inputs do not.

---

## 2026-07-29, round 2's verdict — REVISE again, and the gating seat was RIGHT about my measurement

Round 2: **revise**, 13 reviewers (**not 12 — the panel grew mid-trail**), 4 abstained,
0 unreadable. 6 approve / 7 object. **Gating objection: `debug_historian`, severity
HIGH — not the guardian.**

Guardian round 2 is worth reading as a de-escalation: still `object`, but medium/low,
and it explicitly **dropped** the architecture escalation — *"Blast radius is small and
fully enumerated (3 production call sites, all in-plan, census looks complete) rather
than 'many packages,' and the population-count correction (908) checks out."* The
council's own re-check now independently returns **908**.

### The gating objection, and why it was right

`debug_historian`: my *corrected* population query filtered
`s.status NOT IN ('pool','archived')` — and **nothing in the build-gate path filters
on `sites.status`**. So if pool-status sites' pages get built, the enforcement surface
is larger than measured and the zero-findings conclusion has a gap exactly the size of
what I excluded. Its sharpest line: *"Round 2 correctly caught round 1's status='live'
artifact but replaced it with a different status-based exclusion rather than dropping
status as a scoping variable entirely."*

**That is the [[narrow-filter-defines-the-conclusion]] landmine, and I walked into it
while congratulating myself for catching someone else's version of it.** I inherited
the filter from `104`'s own § Measurement query, where it is correct for *"which live
sites are armed"* — and never re-derived it for the different question *"what will the
gate fire on"*. 17 pool sites vs 15 measured: the excluded slice was **larger** than
the measured one.

Measured properly — status dropped entirely, grouped so the excluded slice is visible
rather than assumed:

```
 deployed | 908 components | 14 sites      <- and NO other row
```

Confirmed first that the gate never reads `sites.status`: the only status predicate in
`validate_page_content.go` is `WHERE site_id = $1 AND status NOT IN ('deleted','archived')`
on the **pages** table inside the link-index query — different table, different column.

**So the excluded slice is empty and 908 was the whole surface.** The answer was
favourable, which is the only reason it is safe to say the measurement was complete —
and I would not have known that without being asked. A filter inherited from another
document is still your filter.

### Two things built rather than argued

- **`check_claims_fleet_wide`, default TRUE** — the kill switch guardian asked for.
  Mirrors `repair_internal_links` for the same reason: DB config is live immediately,
  so a bad pattern is withdrawn fleet-wide in seconds instead of commit+build+roll. Off
  = the pre-104 scan exactly. A separate test proves withdrawing it does **not** disarm
  a site's own audited patterns. Gate-only, because `DiscoveryCheckContext` has **no
  config map at all** — a sweep toggle would mean a new field on a shared context, i.e.
  the very kind of seam this review is cautious about.
- **Two pattern-validity tests** — `bug_historian` caught that `globalEvidence()`'s
  regex fallback silently turns a malformed pattern into a near-inert literal, with no
  log and no test able to tell "compiled" from "degraded". The fallback stays (panicking
  at init over a regex is worse than one over-narrow pattern); the tests move the failure
  from a live build to CI, and the second one catches the inverse — an empty pattern
  compiles happily and matches **every** block.

Both are genuine improvements I would not have made unprompted. Recording that plainly:
the two most useful outcomes of this council trail so far are a defect in my measurement
method and a defect in my error handling, neither of which any amount of re-reading my
own diff would have surfaced.

Round 3 submitted on the same correlation.

---

## 2026-07-29 — round 3: APPROVED, and what the trail was actually worth

`decision=approved`, 12 reviewers, 5 abstained, **0 unreadable**, "approved with 4
advisory objection(s) — none high-severity". Committed the trailer
`Council-Reviewed: 899ed92e-1bf7-4707-96d8-24f102aa14fa` on `5db9f9081`.

**Verified the approval was mine before trusting it**, because a later approval can
attach to another lane's correlation: three `council_report` rows on my correlation
in order (revise → revise → approved), and the approved body references
`check_claims_fleet_wide` and the round-3 test names — symbols that did not exist
before round 3, so it reviewed my plan and not an earlier one.

### One advisory was worth code, and it is the sharpest thing anyone said in three rounds

**compliance (medium):** the kill switch I added in round 2 *at the guardian's
request* is a **silent** disarm path. Flipped false it restores the pre-104 unarmed
state — *"including for vetcomparison.uk or idea.uk, the two sites this whole fix
exists to protect"* — and a site with no register loses its only banned-claim
protection that way, with nothing anywhere saying so.

That is two seats pulling in apparently opposite directions and both being right:
guardian wanted a lever because an unwithdrawable blocker is a containment gap;
compliance noticed the lever I built could quietly undo the entire fix. **They only
conflict if pulling the lever is quiet.** The gate now logs at Warn when the set is
disabled, naming the site and whether it has a register at all. Same principle as
everything else in this workstream: a disabled checker and a working one must not
look identical from outside.

**guardian (low):** pattern 1 is itself a negation construction — the same shape that
caused all four false positives. Correct, and it is not one pattern but **three**: the
whole completeness-of-exclusion family reads negatively. The code comment now covers
the family rather than one member.

**prior_art_librarian (low ×2):** flagged both my "mirrors `repair_internal_links`"
precedent claim and my "no negation-guard prior art" absence claim as author-asserted.
Fair on the epistemics — the first I read first-hand and it holds; the second is the
class of claim that goes stale, and it is already recorded as such.

**architecture + guardian (medium): a fleet-wide blocker-severity mechanism across
three packages wants a filed RFC, and the substance of one exists without the
artefact.** Left as filed deliberately: that is an owner call on venue, not something
to self-approve, and `RFC_001` (from `bugs_closed/003`) is the precedent for how one
gets ratified. Recorded here rather than acted on.

### What the three rounds were actually worth, stated plainly

Approval was the least valuable output. The trail produced, in order: a defect in my
**measurement method** (status filtering the gate never applies), a missing
**containment lever**, a **silent failure mode** in the fallback I had copied without
re-deriving its justification, and a **silent failure mode in the lever itself**.
Not one of those would have been found by re-reading my own diff, and three of the
four are the same shape — *a thing that looks identical whether it is working or
not*, which is the failure this whole workstream started out being about.
