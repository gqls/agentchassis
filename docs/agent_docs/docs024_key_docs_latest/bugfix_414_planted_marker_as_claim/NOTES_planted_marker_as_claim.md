# NOTES — bugs_open/414 (append-only, newest at the bottom)

## 2026-08-27 — session 1

### (a) Ownership: filed and handed on, not being worked

`scripts/who-owns.py 414` said OWNED by `portfolio_positioning` (142 commits/14d). Reading further
is what changed the answer: the bug is a **"Late addition (same evening)"** in that lane's handoff,
whose §1 first task is sitemaps, and the lane's session was idle. It filed, recorded the debt, and
moved on. So: resumed here, contributing INTO the bug file rather than forking an account.

### (b) The bug was still live, and the file's own status line was wrong

Re-measured rather than inherited: 2 + 1 served (control phrase "PRA handbook" = 0), 3 components
carrying it in **both** `content_data` and `rendered_html`, none locked.

Then the finding that changed the shape of the work. Running the file's §Population census over
**every** aspect instead of the one that had been edited: lendzy's **current `strategy`** row
(`96eaff0b`, `domain-strategist`, 2026-08-12) still read *"The acceptance marker 'checked against the
FCA handbook, rule by rule' should appear in the site's written copy…"*. The 08-26 fix had stripped
`content_direction` only. **"Regeneration can no longer re-plant the phrase" was false for ten
days.** `strategy` is read by `build-site-planner` and `webdesign-agent` (measured over live
`agent_definitions`), so the instruction was live in a surface an agent reads.

### (c) My own misstep, in the same query — `_` is a SQL wildcard

I first reported that the **key** `acceptance_marker` had propagated, from a column I had computed
as `data::text ILIKE '%acceptance_marker%'`. That predicate matches the PROSE "acceptance marker"
too. Escaped, the key form exists in exactly one superseded row. What caught it: grepping the pretty
JSON for `acceptance_marker` and getting nothing while the SQL said true — **two instruments
disagreeing, and I had believed the one I wrote on purpose.** Logged in `WRONG_CALLS.md`. It mattered
beyond tidiness: "the key propagated" would have pointed the fix at a key-strip, which cannot touch
prose. The truth — *an agent paraphrased the instruction into another aspect* — is what makes a
text-scanning detector the answer.

### (d) The history says the spec was re-planting it every time

`page_component_history`: **14** archived rows carry the phrase — `about` ×3 slots on 08-11
(including a `differentiators` slot that no longer exists) and the guide's `article-body` across
**4 versions, 08-15 → 08-24**. So the guide re-emitted the phrase on every regeneration while the
spec mandated it. Two consequences: the population was never really 3, and a framework rewrite
against a *clean* spec is evidenced (not assumed) to produce copy without it — the homepage instance
vanished on its own when index regenerated on 08-24.

### (e) The audit item was worse than the bug file said

`052d01b0`'s `current_value` is a **fourth** instance of the claim ("Our guides are checked against
the FCA handbook, rule by rule…"), attributed to `index` — a page that no longer carries it. Its
`suggestion` asked for a "How we verify our guides" methodology section, and its `acceptance_test`
was satisfiable innocuously (name any CONC rule), so a handler could pass the test **while** building
the methodology. Rejected under a guard that aborts unless the row is still `needs_human_review`.

### (f) Two things I nearly built that the measurements killed

1. **A generic claims scan over `site_specs`.** Measured: **21** hits over 522 current rows,
   effectively all false. Fifteen are the estate's OWN honesty instructions ("Never invent a person,
   company, scheme…") matching the never-invents pattern — and the negation guard cannot save them,
   because the match *starts* at "never" and the guard only looks backwards. Worse, `evidence_base`
   rows store each site's `banned_claims` **as data**, quoting the sentences they forbid: a generic
   spec scan convicts every site's own immune system, daily, for ever. `brief-negation-check`'s own
   header records this census as tried and **withdrawn within hours** on 2026-08-19 — I was about to
   re-run it.
2. **A content guard on `WriteSiteSpecAction`.** The plant arrived as a **manual** row
   (`source='manual'`, `created_by='cqls'`) and never passed through that action. A guard there
   covers the agent door only, and the agent door was hop TWO.

### (g) The calibration that decided the whole design

The three served sentences are not one shape but two, and the estate already had a family for each:

| component | shape | caught by |
|---|---|---|
| `content-block-about` "**Everything** on this site **is checked**…" | completeness | new indefinite-subject pattern |
| `article-body` "**Every figure** and every rule reference … **is checked**…" | completeness | window 30 → 60 |
| `hero-about` "…, **checked against the FCA handbook, rule by rule**" | diligence | practice-family P6 |

Over 2,405 live components: the existing completeness pattern fires **0** times fleet-wide at window
30 (it is inert today), **1** at 60, and the new entry **1** — each on exactly the sentence it was
written for. For P6, each half alone is unshippable: idiom alone **22**, verb+`against`+rulebook
alone **13** (including lendzy's own *correct* imperative "Check your loan against the FCA rules");
both together **3**, the planted ones, 0 of the other 2,402. **The conjunction is the design, not the
words.**

### (h) The skeptic pass earned its cost, and I was wrong twice

Ran an adversarial review of my own proposal before writing code. It killed the generic spec scan
(f1), and it caught a **stale justification**: I had argued for `brief-negation-check` over a
discovery check *because* discovery checks ride the improvement sweep, dead 2026-05-02 → 08-25. That
is false — discovery checks ride `site-discovery-rotation-*`, and `unverified_claims` has filed 17
items in 7 days. Same decision, honest reason (shape, not cadence). Logged in `WRONG_CALLS.md`,
because a reviewer told the wrong reason cannot re-derive the right one.

### (i) A test that could never fail, found by writing one that expected PRESENCE

`validate_page_content_fleetwide_claims_test.go`'s helper read
`out["issues"].([]ValidationIssue)`; the action returns `[]map[string]string`. So the helper always
returned nil and every false-positive assertion in that file — green since `bugs_open/104` — was
asserting that an empty list contained nothing. The blocker tests were unaffected (they assert on the
error, which is the real mechanism), which is exactly why nobody noticed. **A test expecting absence
cannot detect an instrument that reports absence.** Fixed the helper; all its tests still pass, so
nothing was hiding behind it.

### (j) The ordering I got wrong in the first plan, and the measurement that inverted it

I had written "pattern first, it protects against the Retry". The skeptic traced the enforcement
points: with the pattern live and the phrase still in `content_data`, a rerender regenerates HTML
carrying it, the **persistence floor refuses the save**, and the OLD `rendered_html` keeps serving.
The item lands `unresolved` and nothing a visitor sees changes — a stranded repair that reads like a
working gate. Repair first; the Go is inert until a roll anyway.

### (k) Mutation proofs run, not assumed

Six guards, each turned red by a deliberate break and green again on restore: the window widening,
the indefinite-subject pattern, P6's rulebook list, P6's order-B alternation, the attestation
exemption, the `evidence_base` exclusion, and the bare-aspect surface fallback. The strip's SQL guard
was induced too (corrupted the expected tail; it aborted with nothing changed) — a guard I had not
watched fire would have been a comment.

### (l) The dry run, and the export that lied twice

`claimscan` over the complete corpus: **3 BANNED** (2 mine, both lendzy; 1 pre-existing on
webdesign.co.uk) and **12 PRACTICE** (3 mine, all lendzy; 9 pre-existing). Zero false positives.
Getting the corpus out took three attempts: a single stream dropped **302 of 2,585** rows with an
"unexpected EOF" *after* printing plausible output; a per-domain loop hit the documented
`kubectl exec -i` stdin-eater and processed one domain. What works is writing inside the pod and
`kubectl cp`-ing the file out — with the row count agreeing three ways. **The row-count control is
the only reason I did not ship on an 88% corpus that scanned clean.**

### (m) State at the end of session 1

`fc588e445` committed (9 files, no passengers); council `f4c144ad` submitted and running
(`review_editquality` at 09:1xZ); HEAD verified — the only failing test in `datahelpers` is
**pre-existing**, from `component_hierarchy_walk.go:397` at commit `bc8167100` (another lane, 08-26)
hand-spelling the tombstone predicate. Spec sources clean fleet-wide (0 rows, any aspect, any site).
Audit item rejected. The two repair items are working through a **fleet-wide `resolve_links` /
`spawn_content_writer` flake** — the `about` item lost attempt 1 to `CHILD_ORCHESTRATION_FAILED` and
is retrying; several other lanes' page builds are failing the same way in the same window, so it is
not the payload.

## 2026-08-27 — session 1, afternoon

### (n) The about page landed while I was not looking, and the deploy "failure" was not one

Both repair items had failed again by 13:11 — and the errors were different, which was the tell.
`about` said *"step deploy_page failed: … timed out after 3 retries"*, and `deploy_page` runs AFTER
`save_sections`. So the save had succeeded: all three about components were rewritten at 10:19:03Z
with the phrase gone from both columns, and the served page reads 0 occurrences. **I nearly missed a
completed repair because the work item said "failed".** The item's status is a statement about the
orchestration, not about the artefact; reading it as the latter is the trap this file has warned about
in three other forms.

The framework's copy is good and grounded in the brief ("every figure we quote comes with the named
rule it's from and a pointer to check it for yourself"). I checked it against my OWN new patterns
before believing it, because the gate is inert until the roll and nothing else had looked: claimscan
over the whole repaired site reports about CLEAN and still convicts the guide in the same run. That
one-pass discrimination is the best control this lane has produced.

`edit_live` did rewrite the untouched `differentiators` slot, as the plan predicted (2,245 → 2,293 —
grew, so nothing lost). The before-snapshot is the only reason that is a checked fact.

### (o) I was wrong about the claim reaper, twice, and the second time was inside my own correction

Sequence, kept in full because the shape repeats:

1. I told the 413-owning lane a dropped-spawn claim is **"unreapable by construction"**, sourced "from
   the reaper's own text", with a falsifiable prediction attached, and wrote it into a work item's
   `error` field.
2. **False.** `claimed-item-timeout` (enabled, 120 s) covers it. I found out because the mechanism
   fired on my own item and returned the exact error string I had told another lane could not exist.
3. **How:** I ran the RIGHT query and piped it through `head -14`. Three of **twelve** matching rows
   were visible; `claimed-item-timeout` sorts first. The answer was in my own result set. I had read
   four reapers carefully and promoted "these four do not" into "nothing does".
4. I retracted to the lane and wrote a correction into `LANDMINES.md` — **and the correction was also
   wrong twice.** It claimed the false statement had been in LANDMINES (it had not: all three
   `unreapable` hits there are other sessions' entries about the `updated_at` trigger), and it said
   the whole CTE chain keys on 15 minutes.
5. The 413 lane read the `reset` stage's own WHERE: **40 minutes**, not 15 — the 15-minute key gates
   only the two auto-complete stages, which need completion evidence a dropped spawn can never have.
   So my item untouched at 34 minutes was **correct behaviour**, my hand-release preempted the
   mechanism by ~5 minutes, and the "scheduler gap" I had floated needed no explaining. Re-verified
   here by extracting the clause directly rather than taking it on trust.
6. Corrected the LANDMINES note again — which removed 17 lines from an append-only ledger and drew the
   `shared-ledger-not-appended` advisory. Checked the removed lines individually rather than by
   regex: **all 17 were my own block from an hour earlier**, no other session's text touched.

**What I take from it, and it is not "be careful".** Every one of today's five logged errors was a
claim I did not need to make, produced in passing to support a point that was not about it — a
baseline, a count, a rate, an absence, a threshold. The measurement-shaped work of this session (the
corpus calibrations, the mutation proofs, the dry run) was checked to a standard I am happy with. The
prose *around* it was not, and prose is what travels to other lanes. The cheap discipline that would
have caught four of five is the same one: **say N out loud before interpreting, and never conclude an
absence from output you truncated.**

### (p) The remaining blocker is not the queue any more

Since 11:30Z there is a fleet-wide, account-level LLM outage — measured first-hand in `llm_call_log`:
11:00Z **36 of 132** calls failed, 12:00Z **61 of 61**, 13:00Z **23 of 23**, every failure a
"usage limit" error. The guide's rewrite is LLM-bearing, so it cannot complete until that clears; the
about page got through at 10:19, before it began. The item is queued and correct. **Do not read the
next failure as a payload fault, and do not spend its last attempt while the outage stands.**

## 2026-09-02 — the follow-on work, and a correction from another lane

### (q) The owner's question moved the target, and measuring is what moved it

Asked "what can I do about the poisoned register hole, and shouldn't compliance be strong for
finance/legal/insurance?" — the instinct was right and the measurement pointed elsewhere. **319 facts
across 17 sites**: ~192 citation-backed (re-fetched and quote-checked **daily** by
`evidence-refresher`), 30 SQL-re-run daily, and **61 attested-only with nothing to re-check — 50 of
those on our OWN sites** (webdesign.uk 25, webdesign.co.uk 15, finetuning.uk 10). And **no live agent
writes the register at all**: 11 of 18 rows by the scheduler, 7 by human/session hands. The vector is
a person, which is exactly how 414's marker got in.

The real finance exposure is the inverse: **5 of 9 finance/insurance sites have NO register**, so
`ScanUnregisteredNumbers` never arms (it is opt-in on register presence), and 2 more have zero facts.

### (r) …and arming it naively would have made things worse — measured before recommending

Armed locally against those five (474 components, nothing written): **5 findings, all false.** Two
regulatory — loancash's flagged digit was the **`5` of `CONC 5A`**, and lendzy's was **`0.8% per day
under CONC 5A`**, a regulatory figure quoted beside its rule, *which that site's brief requires*. The
scan was convicting a site for doing the right thing. Three third-party survey figures in a news
listing (RFC_053's component-grain question — left there).

Fixed the regulatory half (`fad209b92`, council `1dd3d298` APPROVED): 5 → 3, fleet-wide unchanged,
live in chassis `v1.0.1354` by four-arm probe.

### (s) I nearly shipped a vacuous test — twice, in opposite directions, in one change

**Must-catch fixtures first:** "Our team of 37 advisers…", "We ran 12 conc tests…" — neither contains
a `businessClaimContextRe` noun, so the scan never reaches them. The test failed while the fix worked.

**Then the must-pass ones:** single sentences that **passed before the fix existed**, for the same
reason. And the mutation proof "passed" with both exclusions disabled — which I nearly read as *the
fix is unnecessary*. It was the fixtures being vacuous. Rebuilt from the **extracted live component
text** (the real window is wider than a sentence; "This site is independent" is the noun that makes
the scan look at all), they now give 2 findings with the exclusions off and 0 with them on.

**Third instance in this lane of the same thing**: a fixture composed from my model of the data
exercises my model, not the data. Logged before; logged again here because knowing it did not stop me.

### (t) Three council seats caught the same unverified assertion, and were right to

All three asked whether the documented forced `(?i)` on banned-claims reaches my new patterns — which
would make "case-sensitive" compile-time true and runtime false, taking the SUP/MAR/DISP narrowing
with it. **It does not**, verified three ways (the forced `(?i)` lives only in the four banned-claim
compilers; nothing case-folds the scanned text before the exclusion tests; and the existing must-catch
fixture proves it behaviourally). Written at the claim in `ad4824e73`, with
`TestRegulatoryCitationPatternsAreCaseSensitive` as the tripwire for a future refactor.

### (u) A correction FROM the copy_quality_two_stage lane, about my handoff

Their commit `51e05a374` notes "the Sunday handoff's '0 of 36' was the SPEC-CLAIMS half, two N-of-Ms
conflated". Checked: the report carries **two** `N of M sites` lines and they legitimately differ —
`11 of 34` for the negation half (writer-visible surface) and `0 of 39` for mine (union of every live
agent's surface). **My 08-31 handoff said "read N of M" without saying which**, and they hit it.

Corrected in place in that handoff (`580d1351d`), original struck rather than removed. It is the same
shape as the `scanned_fields` rule I logged on 08-31: **a reading rule is only as good as its ability
to be performed by someone who does not already know the answer — and "which number" is part of the
rule.** Two instances in three days, from one lane's instructions.

---

## 2026-09-03 — the detector shipped; the zero it produced is uninformative BY CONSTRUCTION

**Deploy verified at the artefact, not at the commit.** New replicaset `75b987cbd7`, pods ~17 min
old at 09:21 UTC. The `build provenance` startup line had already scrolled out of `--tail=400`
(expected on a busy service; **an empty result there means "not in range", not "unstamped"**), so I
probed the capability rather than the commit — which is the better question anyway:

| probe on `/proc/1/exe` | result | meaning |
|---|---|---|
| `invalid_banned_claim_pattern` (target) | **6** | the detector IS in the running binary |
| `zzz_not_a_real_symbol_qx7` (must be absent) | **0**, exit 1 | grep can return zero — not a blind pass |
| `stale_evidence` (must be present) | **6** | grep can return non-zero — not a blind match |

`evidence-freshness` then **ran at 09:10:23 and completed**, under pods that started ~09:04, so the
sweep executed on the new binary. `site_work_items` with `item_type='invalid_banned_claim_pattern'`:
**0**.

**⚠ And that zero cannot be read as "the fleet is clean", even though the fleet IS clean.** Both
result fields carry `omitempty`:

```go
InvalidBannedClaimPatterns        []invalidBannedClaimPattern `json:"invalid_banned_claim_patterns,omitempty"`   // :216
InvalidBannedClaimWorkItemsCreated int                        `json:"invalid_banned_claim_work_items_created,omitempty"` // :221
```

So a clean result **serialises to nothing**. Measured: of **23** evidence runs since 09:00,
**0** mention the field — and that figure is *identical* whether the code ran and found nothing or
never executed at all. There is no log line on the clean path either (only a `Warn` on write
failure), so nothing in the system distinguishes the two states.

This is the `a-post-fix-zero-needs-a-demand-control` family, but sharper than usual: the blindness
is **mechanical and citable at a line number** rather than a judgement about coverage. The clean
case is *designed* to leave no trace, which is reasonable for log volume and fatal for verification.

**What would actually prove it** — plant a deliberately broken pattern on a scratch site, confirm
the next pass files an item, remove it. Handed to the `claims-verification` lane with the
`omitempty` reasoning attached; **not run from this lane**, because they own the code and the
council round.

**Misstep worth recording in the same breath:** my first instinct was to grep the chassis logs for
evidence of the run. There is nothing to grep — I only established that by reading the function
rather than by grepping and finding nothing, which would have looked like the same answer and meant
something different. *An absence in a log is only evidence once you have read the code that would
have written the line.*

---

## (q) 2026-09-03, 09:43–09:52 UTC — the handoff went stale in 15 minutes, in three of its rows

Picked the lane back up from `HANDOFF_2026-09-03_continue_here.md`, written at 09:34 UTC. Its §0
says the lane is closeable and its §3 says one verification is outstanding and owed by another lane.
Both were true when written. Three rows were not true fifteen minutes later, and one of them I got
wrong on the first reading.

**1. The demand control was already planted — by its owner, four minutes after we unblocked it.**
`buytoletcalculator.uk` (`dc7a8ebf-…`, `sites.status='test'`, 0 pages) had a current `evidence_base`
created **09:34:48 UTC** by `created_by='claims_verification_probe'`: one `banned_claims` entry,
pattern `guaranteed(` (unterminated group), `reason` naming it a probe to revert. Our landmine
`71b85fcc2` — "a scratch site search by DOMAIN finds nothing; non-production is `sites.status`" —
was committed ~09:30 UTC. So the relay worked, and it worked in **four minutes**. Worth knowing that
the landmine channel moves that fast when the receiving lane is live.

At 09:49 UTC the assertion had **not** yet succeeded: `invalid_banned_claim_pattern` items **0**
fleet-wide, no `orchestration_states` row since 09:25 naming that site. Planted, not dispatched.

**2. The follow-up log line is NOT deployed, and I measured it rather than repeating the handoff.**
§3 said `996b40542`'s always-fired Info line is "Go, so inert until the next roll" — an inference.
`[MEASURED 09:47 UTC]` probing `/proc/1/exe` on **both** replicas of `75b987cbd7`:
`invalid_banned_claim_pattern` **6/6**, `patterns_checked` **0/0 (exit 1)**, control `stale_evidence`
**6/6**, control `zzz_not_a_real_symbol_qx7` **0/0 (exit 1)**. Pods started 08:57:46 / 08:58:07 UTC;
`996b40542` committed 09:29:46 UTC. Arithmetic and artefact agree.

The consequence is a live trap on exactly one branch, which is why it was worth relaying rather than
just recording: **if the dispatched pass files nothing, grepping for `patterns_checked` returns a
silence that is the un-deployed line, not a non-executing check.** Sent as
`claims_verification/CONTRIB_2026-09-03_from_414_your_demand_control_is_planted_and_the_log_line_you_will_reach_for_is_not_deployed.md`
(a NEW file, deliberately — that lane was committing into `RFC_060` every few minutes and a
same-file append would have been a passenger risk for no gain).

**3. MISSTEP — I read a current row and inferred a cause the history refutes.** §4 listed
`farmerinsurance.uk` as "7 facts but **0 `banned_claims`** — the 707 residue, flagged to lendzy
relay". The live row today shows **7 facts and 5 patterns**, written **09:11:23 UTC by
`evidence-refresher`** — one minute into the 09:10 daily sweep. I read that as *the daily sweep
closed the gap on its own*, which would have been a genuinely interesting claim about the refresher
minting patterns, and I nearly wrote it down.

The spec history says otherwise:

| created_at (UTC) | created_by | facts | banned |
|---|---|---|---|
| 09-02 15:11:19 | `evidence-researcher` | 3 | 0 |
| 09-02 15:27:19 | migration 698 (loanzy lane) | 7 | 0 |
| **09-02 18:34:47** | **migration 713 (loanzy lane)** | 7 | **5** |
| 09-03 09:11:23 | `evidence-refresher` | 7 | 5 | ← carried forward, minted nothing

The **lendzy relay closed it the previous evening**; the refresher merely copied it into a new
current row and stamped its own name and today's clock on it. **The lesson is narrow and reusable:
`created_by` on the current row names the last WRITER, not the AUTHOR of the value it carries — and
a refresher that rewrites a whole spec launders every field's provenance into its own name.** The
check that catches it is one query: read the aspect's full history, not its current row. Compare
`a-report-is-not-a-measurement` and `seed-sql-is-history-live-row-is-fact`; this is the inverse of
the second — here the LIVE row is the misleading one and the history is the fact.

**4. `loancash.co.uk` re-confirmed as the one genuine gap**, and it is unchanged: **no**
`evidence_base` row at all — not an empty one — beside 14 other current specs, on a `deployed`
finance site serving **30 pages**. RFC_060 Q1 makes a register required there. Still unowned.

## (r) 2026-09-03, 09:57 UTC — "the one genuine gap" was the one gap this lane was LOOKING at

Chasing §4's `loancash.co.uk` row produced a bigger number than the row implies, and one structural
finding that is worth more than either.

**Four of RFC_060 §1d's five register-less finance sites are done.** `lendzy.co.uk` now carries 8
facts and 5 `banned_claims` (its table row still says "not yet applied … 0 (pending)", true on 09-02
afternoon); `farmerinsurance.uk` 7 facts and 5 patterns. Only `loancash.co.uk` remains, so RFC_060
§3c's track 1 — *"populate the five register-less finance sites"* — is now **one site**. Re-measured
into the RFC at §1d (`ce71ab6bd`), because that instruction is what a session picking up track 1
reads, and it would have sent them at four sites that no longer need it.

**The population is 13, not 1.** `[MEASURED 2026-09-03]` **13 of 39 `deployed` sites hold no current
`evidence_base`**. Q1's requirement is finance-scoped so most are not violations today — but
`vetcomparison.uk` is on the list, and Q5's ruling was decided *on the owner's own fact* that vet and
legal are next. **The sector presets Q5 approved will land on a site with no register to apply them
to.**

**And nothing can raise the absence.** `resolveEvidenceSites` (`:290`) builds the daily sweep's
target list as `SELECT site_id FROM site_specs WHERE aspect='evidence_base' AND is_current` — the
sites that **have** a register. **The target set is defined by the presence of the very thing whose
absence is the defect.** So a register-less site is invisible to the freshness sweep, the fact
checks, and this lane's own new `invalid_banned_claim_pattern` detector alike — silently and for
ever. Q1 requires registers; no reader enforces or reports the requirement.

That is the same family as everything else this lane found, one turn further out. §3's `omitempty`
blindness hides a *clean result*. This hides an *absent subject*. In both cases the instrument
reports nothing and nothing is what a healthy system also reports — but here it is not a logging
choice, it is the population filter, and no amount of running the check more often would ever reach
these sites. **A detector whose target list is drawn from the population it is checking cannot find
what is missing from that population.**

Recorded in RFC_060 rather than built. It is that RFC's build and the `claims-verification` lane's
seam; this lane measured it and handed it over, which is where §2b's lesson landed.

---

## §(s) 2026-09-03 evening — the D2 restoration APPLIED and verified at the bytes, and D1 (vetcomparison) BUILT

All times UTC. The session clock is BST, one hour ahead — and I got that wrong twice, see §(s6).

### (s1) 743's round-2 verdict was REVISE, and the gating objection was FALSE OF THE FILE

Round 2 landed 17:00:28 — **REVISE**, `decided_by: gating objection from editquality`. Three seats
(editquality gating, guidelines, guardian) raised the same HIGH objection: the INSERT's last value is
`'auto'`, so `approval_mode='manual'` — the round's headline safety fix — "is not present in the SQL
as written".

**It is present.** Lines 157/164/171 of the file each end `'manual'`, and the verify block asserts
`w.approval_mode = 'manual' AND w.source = 'manual'` on all three rows, so the migration could not
have applied carrying `'auto'`. What ended `'auto'` was the **submission sketch**, a revision-1
leftover. Two more seats spent objections on the same staleness (a fourth page described as unhandled
that the file already covered; a "2 items" verify against a summary saying three).

I checked every other objection against the artefact rather than dismissing them as a set:

| objection | seat | checked | result |
|---|---|---|---|
| `approval_mode` is `'auto'` | editquality, guidelines, guardian | read the file + the verify block | **false of the file** — sketch only |
| `edit_live` may be inert; the landmine says `recreate` | editquality, prior_art_librarian | live `agent_definitions` for `page-build-handler` **and** the Go | `has_step=true has_edit_live=true`; `load_current_section_content_action.go:139` `if inputs.Get("mode") != editLiveMode { return passthrough("not_edit_live") }`, const at :98 — **edit_live is correct** |
| the fourth page is left to run | bug_historian | read the file | already covered — 743 rev 2 files **three** items |
| FSMA-2000-S19/S23 may not be in the register | prior_art_librarian | queried the 738 register | **both present** |
| did 745 file rows that could double-file? | reuse_agent | queried `created_by LIKE '%migration 74%'` | **0 rows** — 745 never applied |

And one check nobody asked for, because this lane's own landmine demands it (*paste, never retype*):
the three `restore_verbatim` blocks are **genuinely verbatim** from `page_component_history`, proven
by extracting each block programmatically and asserting `text in history`, with a deliberately-wrong
control returning False.

### (s2) Applied, and canaried — one page first, because 739 fired four at once

Applied 18:49. All guards passed (`block present on 12 page(s) - convention confirmed`;
`page-build-handler carries load_current_section_content AND the edit_live literal`), 3 items filed.

`approval_mode='manual'` means the dispatcher will not touch them: `load_work_item_actions.go:802`
— `AND (COALESCE(wi.approval_mode,'auto')='auto' OR wi.status='approved')`. So a manual item sits at
`triaged` until someone sets `approved`, and **the 48h reaper would then have flipped it to
`unresolved`, which reads as processed**. Left alone, the restoration would silently never happen.

I released **one page first** — `check-your-lender-is-authorised`, the simplest spec — because 739's
damage was four simultaneous items. Claimed 18:52:58, complete 18:55:43.

### (s3) The sentence-identity diff, which is the check 739 taught us to run

`[MEASURED at the component HTML, pre-image from `page_component_history`'s delete row]`

| page | pre | post | kept identical | removed | **orphaned** | added |
|---|---|---|---|---|---|---|
| check-your-lender | 35 | 40 | **35/35 (100%)** | 0 | **0** | 5 |
| loan-sharks | 20 | 30 | 15/20 | 5 | **0** | 15 |
| price-cap | 23 | 27 | 22/23 | 1 | **0** | 5 |

Compare 739: **36 of 37 sentences replaced** on the price-cap page. `edit_live` works, and the two
HIGH objections doubting it were answered by the artefact.

> **⚠ AND THE ACCEPTANCE TEST I WROTE WAS THE WRONG TEST.** It says *"ADDITIONS ONLY, no existing
> sentence removed or reworded"*. On loan-sharks, 5 sentences WERE reworded — and every one is a
> splice seam where the requested new material was inserted: *"lends money **for profit** without
> authorisation"*, *"the Illegal Money Lending Team**, which** investigates"*, *"Free**, independent**
> debt charities"*. Similarity 0.89–0.98, and **0 orphaned** — nothing lost. A restoration that must
> insert clauses into existing sentences CANNOT satisfy "no sentence reworded", so the literal test
> would have failed a correct repair. **The right measure is ORPHANED sentences (a removed sentence
> with no close survivor), not removed ones.** Fixing the test, not the site.

Served bytes, and 739's corrections intact: check-your-lender 62,823→67,856 · loan-sharks
61,579→62,848 · price-cap 61,601→61,897; the disclaimer markers all 0→1 and the site-wide count is
back to **15 pages** (12 after the damage, 14 before it — jargon-buster gained one).

> **CORRECTED, minutes after I wrote it:** I first read the price-cap page as having lost 739's
> correction, because `grep -c cumulativ` returned 0. The correction is fully present and better
> stated than the brief asked — *"the rules cap what that fee adds up to **across the whole
> agreement**"*, *"£15 is the ceiling whether one payment is missed or several"*, *"cannot lawfully
> be charged £45"*. **My needle was wrong, not the page.** A single-word grep for a concept the page
> may express in other words is not a check.

### (s4) D1 — vetcomparison's register, and the design decision that is the interesting part

Migrations **759** (register), **761** (posture record), **763** (council fix). 21 facts, 6
banned_claims, `citation_code_presets:["veterinary"]` — the preset already existed in `claims.go`
(`{"RCVS","VMD"}`), so it is reused, not invented.

**Three live errors found, two of them NEW.** Five lanes have now run this method; five have found
errors in their own site's live copy.

1. **The CMA final report is dated November 2024 on two guides. It is 24 March 2026.** NEW. The CMA's
   own case-page timetable says so, and its consultation page says *"In March 2026 we published the
   final report"* — both verified through the production matcher. November 2024 is the Inquiry
   Chair's BVA Congress speech, listed on the same page, which is almost certainly the origin.
2. **The £21 / £12.50 prescription caps are served as settled on seven pages** and are bracketed
   placeholders. Known to the vetcomparison lane since 08-24, unfixed. Confirmed here by reading the
   PDF rather than inheriting the claim: *"'Initial Primary Prescription Fee Cap' means [£21
   inclusive of VAT. This will be adjusted for inflation … before the Order is made]."*
3. **"36 service categories" is 36 SERVICES in 5 CATEGORIES.** NEW. Draft Schedule 1's own column
   heading is *"Service, product, treatment or procedure (36 total)"*; the five category counts
   12+6+6+9+3 sum to 36. The number is right, the noun is not.

**No copy touched** — the 695/699/738 precedent and the content freeze.

**THE PDF PROBLEM, which is the transferable finding.** The CMA's primary sources are PDFs and the
daily citation re-checker extracts visible text from HTML. `[MEASURED]` through `cmd/fcaquotecheck`:
`HTTP 200 raw=392144 visible=296699` and **every quote false**, including `"Compliance Date"` — and
**the absent control false too**, so at a PDF the probe discriminates nothing. A `source.citation`
there would read as `citation_lost` drift every day for ever. So 8 facts carry `source.attested_by`
+ `source_document` + `no_citation_because`. Verified at the code, not assumed: the re-fetch arm is
gated on `if _, has := src["citation"]; has` (:576), and `numberSupported` reads `Value`,
`ContextTerms`, `Tolerance`, `IsSeries()` and **never `Source`** — so an uncited fact arms the scan
exactly as fully. **Fourth signature for RUNBOOK §8g**, and unlike the three known ones the host is
fine and the document is right; it is the EXTRACTOR that cannot read the format, so a host-acceptance
check passes it. Now a LANDMINE.

### (s5) What the register actually arms — including where it arms nothing

`banned_claims`: all 6 compile in the exact consumer form, all 6 **fire on their own positive
control** through the production scanner, all 6 give **0 hits over the 23 served pages**, and the
negation guard suppressed **0** — so the zero is the site's, not the guard's. Re-run against the
**stored** row after apply (§8f's post-apply half): patterns survive the escaping layer intact, no
double-escape. The inherited finance set was **not** adopted; that lane warned it false-positives
here in reverse.

> **⚠ I NEARLY RECORDED A VACUOUS ZERO.** With the register loaded, `ScanUnregisteredNumbers`
> returned 0 on all seven non-editorial pages. A **demand control** — the same register with `facts`
> emptied — returns **the same 0**. The scan is high-precision (`businessClaimContextRe`) and the
> £21/£12.50/36 sentences live on `guide`/`blog-post`/`tool` pages that `editorialPageTypes` gates
> off. Replaced with a disconfirmable test: the control flags *"The threshold is 15 first opinion
> practice sites"*, the full register **supports** it, and *"We have 4,000 clients"* stays flagged in
> **both**. So: the facts do real work and do not blanket-support — but **the numeric half is armed
> and currently unexercised**, and the migration says so rather than implying coverage.

### (s6) Four things I got wrong, beyond those already corrected above

1. **The sketch/file divergence is a class, not an incident.** The council reviews the SUBMISSION.
   This round failed safe (REVISE on a non-defect); the same staleness pointed the other way draws an
   **APPROVAL for a file that is wrong**, and nothing downstream compares the two. 759 and 761 have
   **machine-extracted** sketches with a self-check asserting the file's prefix and suffix match.
2. **My first mutation of 759's guard PASSED — and the mutation was inert**, not the guard: it edited
   a different clause of the same field while the needle survived earlier in the sentence. This
   lane's own rule caught it. A mutation must assert it changed the thing the guard reads (`assert
   n==1`).
3. **And minutes later the same discipline caught a REAL defect.** 761's first verify used
   `IF nfacts <> 21`. The destructive case makes `nfacts` NULL, `NULL <> 21` is NULL not TRUE, so it
   printed **`761 OK … register intact at <NULL> facts`** having wiped everything. `IS DISTINCT FROM`
   now, carried into 763, and a LANDMINE.
4. **I stamped two migration notes "UTC" off the BST clock** (20:20/20:35 for 19:16/19:19) — the trap
   the handoff opens by naming. `applied_at` was right; the notes are corrected in place.

### (s7) 759's verdict: APPROVED round 1 — and one advisory objection was real

**APPROVED**, `approved with 2 advisory objection(s) — none high-severity`. editquality's medium was
correct and is fixed forward by **763**: six ids matched `CMA-DRAFT-%` while only five carried
`draft_status`, and 759's verify passed only because it required **both** conditions, so it counted 5
and never saw the sixth. Substance was right (the consultation fact records a settled event and
correctly needs no tag); the **name** misled. Renamed so the prefix means exactly "provisional", and
**the new verify asserts the prefix/tag EQUIVALENCE rather than a count** — a count is precisely what
cannot notice an extra member of the set it is counting a subset of.

`prior_art_librarian`'s two were answerable with measurements already taken (the absence was queried
before writing, not only guarded at apply; 743's objection was verified resolved at the file, the
live config and the Go, per the table in §(s1)).

### (s8) State, and what is open

- **D2: DONE.** 3 pages restored, verified at the served bytes, 0 orphaned sentences, disclaimer back
  on 15 pages, 739's corrections intact.
- **D1: DONE.** Register live, `posture.rung=relied_upon` recorded with declarer/date/basis, the
  `missing_evidence_register` item **closed** with its acceptance test run verbatim through a JOIN on
  `sites` and its evidence in `result`.
- **OPEN:** 761's verdict (`5d54f835-152a-4c6d-a4d1-b3ce289adbd1`) — **read it**; it proposes the
  `posture` key as a shape for RFC_060's Q4 record, which has no built home (0 Go consumers, 0
  existing users, both measured), and that is the claims-verification lane's call, not this lane's.
- **OPEN:** 742 still owes its RESUBMIT (`RESUBMIT_CORR=0d730d51-…`), unchanged from `-03c` §0b(ii).
- **OPEN, and larger than this lane:** nothing enforces `spec.mode` on `content_rewrite` at the
  producer, so the next item minted fleet-wide without it hits the identical destructive
  regeneration. `bugs_open/178`. Named in 743's header as the top follow-up; still true.
- **OPEN:** the three vetcomparison copy errors are RECORDED, not REPAIRED. Owner's call.
- **The 10 remaining registers** in D3's queue.
