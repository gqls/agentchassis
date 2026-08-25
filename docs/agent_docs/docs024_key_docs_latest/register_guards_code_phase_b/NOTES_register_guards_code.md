# NOTES — register↔tool fact drift (Phase B)

Append-only, newest at the bottom. Missteps are the point.

## 2026-08-16 — session 1: from "take the next bug" to a class fix

**Started at `bugs_closed/225`** on the owner's instruction. First finding, before any
design: **225 was already fixed and live** (2026-08-09), and had been kept in
`bugs_open/` by the owner's 08-06 direction — which the owner **superseded on 08-12**
("if it is fixed and live it should be moved"). So the case itself was paperwork, not
work. Re-verified at the wire rather than trusting the file: `grep -c 625000` = 0 on
both live pages, `FTB_RELIEF_CAP = 500000` present as the positive control.

**Trap found while re-verifying, worth its own line:** the component id the bug file
names in its "Fix landed" section (`55682bc8-…`) **no longer exists**. The LMC lane's
B2 work decomposed that page into `prose-0` / `tool-1` / `prose-2` after the bug was
written. A durable file's own evidence pointer rotted in a week. This is not a footnote
— it is the reason the class fix addresses artefacts by page name and not by component
id, and it is recorded in `bugs_open/288` §"the mechanism that COULD have".

**The real work was the class**, which 225's own section "Why no existing check could
ever have caught this" had already identified and nobody had filed.

### Misstep 1 — I nearly built a second mechanism beside an existing plan

My first design sketch was a register-side `retired_values` + `artifact_check`
extension: teach the daily sweep to assert that an expired figure is ABSENT from a
site's artefacts. It is a reasonable mechanism. It was also **the wrong move**, and
what caught it was reading `mortgagecalculator_couk_adoption/` before writing code:
`PLAN_2026-08-09_facts_into_tool_acceptance.md` had already designed this class fix in
four pieces, owner-visible, with Piece 1 live since 08-10 and Pieces 2+3 **designed,
unclaimed, and sitting last on that lane's list**.

Building my own version would have produced two mechanisms for one job, one of them
undocumented in the plan the owner has read. **The cheap check: `grep -rl` the lane
directories for the mechanism before designing, not just `bugs_open/`.** The plan's
own §5 landmines then bound the implementation (doc_plans has no `site_id`; never
round-trip the register through the typed struct; SKIPPED ≠ PASSED).

### Misstep 2 — a stale figure carried from a subagent into a commit message

Full row in `WRONG_CALLS.md`. Short version: I wrote "0 of 130 facts, 0 of 87 tool
PLANs" into commit `989addb1c` from a research subagent's report without grounding it.
Live numbers are **143** and **90**. The load-bearing half (zero) was right; the
denominators were ~10% stale, and forward-only means the commit message is wrong in
the archive for ever. **A subagent's report is another doc.** It has no special
standing for having been produced inside my own session — if anything it is worse,
because there is no visible seam where measuring stopped and quoting began.

### The measurement that changed the design (and would have shipped an inert check)

I was about to key the fan-out on the acceptance ladder's own eligibility predicate,
`toolEligibilityWhere` — the obvious reuse, and it would have looked like good
citizenship. Ran it live against both SDLT sites first:

> it returns **neither** of the two tools this mechanism exists for.

mortgagecalculator's `tool-stamp-duty` has two components; LMC's
`mortgages-stamp-duty` has three since B2. The predicate's sole-component clause
excludes both, deliberately and correctly for *its* purpose. A fan-out keyed on it
would have been a check that could never fire on the tools that motivated it — and it
would have reported green for ever, which is the estate's most expensive shape.

**Encoding a fact and being acceptance-eligible are different questions.** The fan-out
uses the platform's own name rule instead (the one Tier 4 already uses for a tool's
URL). This is in CLM-022 and TL-045 as a landmine, because the next person will reach
for the same reuse.

### What was built

`989addb1c` — Piece 2 (`criteria_facts.go` + validator rule P11) and Piece 3
(`refresh_evidence_fact_drift.go` + two wiring lines in the refresher). Register
CLM-022 + TL-045 in the same commit (platform-seams condition 2). Council submitted
`cff364b8` before committing, so the commit carries `Council-Submitted:` — the trailer
that asserts nothing and is credited automatically at report time.

### Verification, and what each check could have come out as

- **Six routing guards, six induced reds.** Each guard deleted in a scratch HEAD tree,
  its test run, FAIL observed, code restored, green re-observed: `no_auto_fix`, fork,
  evidence-vs-value, fetch-error, baseline-present, baseline-precedence. Plus a
  seventh on P11. Without this, thirteen passing tests would have proven only that the
  code runs.
- **The no-op case asserted, not assumed:** a site with no declaration costs exactly
  one SELECT (sqlmock `ExpectationsWereMet`) and its result JSON is byte-identical to
  before the change (golden compare in the test).
- **Built against `git archive HEAD` + my files**, because another session's untracked
  `component_name_resolver_menu.go` does not compile at the working tree. A green
  `go build` there would have been meaningless either way.

### What is deliberately NOT done, and why it is not laziness

- **Piece 4 (the oracle)** — the only piece that answers "is the figure RIGHT". Needs
  an RFC and a Go oracle library. This mechanism only answers "did it MOVE": a tool
  and a register wrong in the same direction agree, and it is silent.
- **A retired-figure scan over prose/scripts** — would close the other half of the
  class (blind spots 2 and 3), but it is a second scanner over the same components
  (`bugs_open/093`'s shape) and needs a measured false-positive rate first.
- **RFC_025 stage 2b** (`page_name` addressing, reachable for every source kind) —
  small and adjacent, but it changes an existing mechanism's semantics and deserves
  its own round rather than riding in on this one.

All three are named as residual in `bugs_open/288`, not left as folklore.

## 2026-08-16 (later) — council round 1: REVISE, and it found the defect I could not see

**The verdict was REVISE, gated by `editquality`, and it was worth every minute.** The
memory line "a REVISE round is cheaper than the defect it finds" held again.

### The gating objection, which I had reasoned my way INTO

> classifyFactDrift's baseline is defined purely from PRIOR REGISTER state … So a tool
> that is wrong from the day it opts in, against a register fact whose value has been
> stable since (exactly bug 225's shape: correct register, stale code, no subsequent
> legislative change), produces baseline == current and … emits nothing.

I had a test asserting that silence — `TestClassifyFactDrift_NoBaselineIsSilent` — with
a comment explaining that "is the current number right" is Piece 4's question. **The
reasoning was sound and the behaviour was still wrong**, which is the interesting part.
The distinction I missed: I do not need an oracle to notice that *nobody has ever
checked this pair*. That is not a claim about arithmetic; it is a claim about our own
records, and we hold those.

So a first declaration now files one `unreconciled_declaration` review item per (fact,
tool). It is self-quieting — the item carries the value, which becomes the next pass's
baseline — and the mutation test proves the point: **M8 reverts to the old behaviour and
the new test fails**, so the objection is closed by construction, not by assertion.

**The transferable lesson, which is not "listen to reviewers".** My silence had a
*justification*, and the justification was true. A defensible reason for a behaviour is
not evidence the behaviour is right — it only proves the behaviour is not accidental.
The test I wrote asserted my reasoning back to me, which is the most comfortable kind of
green there is. What the seat did that I did not: it took the mechanism's own motivating
bug and ran it through the new code. **The check I should have written first: does this
fire on 225?** It did not, and nothing in my thirteen tests asked.

### Two more real findings, and three answered by measurement

- **`compliance` (medium)** — a risk inversion I had built in without noticing: every
  human-routed finding sat at medium/35, so a *confirmed* value move on a `no_auto_fix`
  tax calculator ranked BELOW an auto-fixable drift at high/30. The harder something is
  to fix automatically, the more urgent it is, not less. Now three bands.
- **`debug_historian` (high)** — `p.status='active'` is the estate's LIVENESS predicate
  and is documented as wrong for anything choosing which pages to JUDGE. The sharp part
  of the objection was not the predicate itself but the asymmetry: *I had measured the
  eligibility predicate and not this one, in the same query*. Now uses the shared
  `NeverDeployedPagePredicateFor`. **Measured before changing it: both motivating pages
  are `active`/`deployed`, so it would NOT have shipped inert — 198 active vs 6
  archived tool pages fleet-wide. The fix closes a door rather than repairing a miss**,
  and I have said so rather than letting it read as a catch.
- Answered without code changes, each by a query or a citation: the fact id **is** in
  the `item_key` (the sketch hid it); a persistently-403 source is **not** permanent
  silence (the citation arm ages a fact past `staleness_days` regardless of fetch
  success, and that arrives here as evidence_drift); **one** agent type calls the
  action; the optional-key budget audit reports **0 shared actions over N=10**, so P11
  does not flip this to architecture scope.

### A design call the round forced into the open

Filing per (fact, tool) means seeding mortgagecalculator's 13-fact fence produces 13
reconciliation items at once. Collapsing to one item per tool would be tidier and is
**wrong**: a key coarser than the finding drops every finding after the first
(`bugs_open/091`), which the `bug_historian` seat raised in the same round. Kept
per-fact, banded low, and the burst is named in the CONTRIB so the receiving lane is
not surprised by it.

## 2026-08-16 (later still) — council round 2: REVISE again, and it killed my round-1 fix

**Round 2's gating objection is the sharpest thing that happened to this lane**, and it
was aimed at the fix I had just written for round 1:

> `unreconciled_declaration` only fires when baseline is nil, but nil requires BOTH no
> lastItem AND no previousRow(factID). previousRow is … prior REGISTER state, the same
> signal round 2's own rationale names as the broken baseline definition … Since the
> register is re-verified daily and … the register's cited VALUE was correct and
> unchanged the whole time, previousRow will almost always resolve.

Correct, and I should have seen it: **I fixed the symptom and left the mechanism that
caused it in the fallback chain.** On mortgagecalculator the register is superseded every
morning, so a previous row always exists carrying 500000 — baseline never nil — the
first-declaration case never fires — the mechanism is still blind to 225. My round-1 fix
was inert for the exact case it was written for, and my tests passed because I tested the
fix in isolation rather than against the bug.

**The resolution made the code smaller, which is usually the sign of a right answer.** I
had conflated two questions:

- *has the register's VALUE moved?* — answered by comparing register rows
- *has this TOOL ever been reconciled?* — answered only by whether we ever filed for it

`previousRow` answers the first and I was using it for the second. Deleted. The baseline
is now the newest `fact_drift` finding for (fact, tool) and nothing else; its absence
means "never reconciled", which is a real state rather than a missing measurement. The
value-move case still works, because a moved register value stops matching what the tool
was told.

**Two mutations passed before one bit — see `WRONG_CALLS.md`.** M11 mutated a path the
test bypassed; M11b read the row without using it. Only M11c (declared, populated,
consulted) failed, with the right message. I nearly recorded a false proof.

**Also done this round:** reused `discovery_checks.ToolSubjectKeyExpr` instead of the
hand-written equivalent (two seats objected; they were right — a second spelling of the
subject-key rule can drift from what the acceptance tiers resolve), and verified live
that it resolves both tools: LMC `mortgages-stamp-duty` → `mortgages-stamp-duty`, mcalc
`tool-stamp-duty` → `stamp-duty`, both non-fork. Measured for the guardian seat: the
validator has exactly **one** production caller (`write_experience_pattern_action.go`),
and **0 of 161** current PLANs and 0 of 11 experience patterns use a top-level `facts`
key — no collision.


## 2026-08-17 — live on the fresh build, and the zero that proves nothing

**The roll landed** (pods 22:07:55Z / 22:08:17Z on 08-16) and the code is in it. Verified
at the binary rather than at the tag, both replicas, with a positive control in the same
exec: `fact_drift_review` **2**, `unreconciled_declaration` **1**, control
`stale_attestation` **5**.

**Which revision** matters here and I checked it rather than assuming: two strings unique
to the final, post-council, advisory-corrected commit (`6b3b0510e`) — the
`"may have stopped matching"` warning and the `DISTINCT ON (dp.subject_key` join — are
both present. So the binary is not an earlier build of the same feature. This is worth
doing every time: "my symbol is present" only tells you *some* version of your change
shipped, and I have shipped four.

**The first production exercise happened without me.** The `evidence-freshness` sweep ran
at **09:04:14Z** with this code live: 8 register revisions written, **0 errors**. That is
the no-op proof — the new query path executed against all 13 register-bearing sites in
production and broke nothing, which is the thing that could most easily have gone wrong
in a shared daily sweep.

**And `fact_drift` items: 0 — which is NOT evidence of anything.** There is no demand
behind it. 0 of 90 current tool PLANs declare a `facts` list, so the fan-out has nothing
to act on; a clean sweep here is a check with nothing to check
(`a-post-fix-zero-needs-a-demand-control`). I have written that into `bugs_open/288` in
those words, because "the sweep ran clean" is exactly the sentence a later reader would
quote as "the mechanism works".

**The one remaining step is a coordination question, and I did not take it myself.**
Firing the mechanism needs a declaration, which means writing to mortgagecalculator's
fence. That lane was active on 08-16 (guides, imagery, dead links) though it has not
touched the fence since 08-10. Seeding files low-severity items into their review queue —
a change to their backlog, not just a config row, and `bugs_open/033` says that queue has
no working surface. **Asked the owner; the owner said ask the lane.** No live session
matched that lane (its transcript stopped at 11:41, as the current sessions started), so
the ask now sits at the top of their newest handoff — the file a fresh session there reads
first — with three options and no preference. The CONTRIB had been sitting unread since
08-16, which is why the handoff was the right channel and the CONTRIB alone was not.

---

## 2026-08-24 — session 2: the lane resumed, and four phases landed

Picked up cold after seven days idle (last lane commit 2026-08-17 19:30). Ownership
checked three ways before assuming it: `who-owns.py`, a grep of every lane dir for
`288`/`CLM-022`/`fact_drift`/`artifact_check` dated ≥08-18, and yesterday's live
process-table roster (`RESTART_2026-08-23_open_threads.md`) — no session on it in any
of them. The mcalc lane discharged its arm on 08-17 and had moved to guides and
`bugs_open/348`.

### What the live system said before I touched anything

| | 2026-08-16 (as filed) | 2026-08-24 (re-measured, same queries, same controls) |
|---|---|---|
| register facts / sites | 143 / 12 | **294 / 15** |
| facts carrying `artifact_check` | 0 | **0** (control: 185 carry `citation`) |
| tool PLANs declaring `facts` | 0 of 90 | **1 of 132** — and later the same day, **1 of 134** |
| `fact_drift_review` items | 13, filed 08-17 | **13, still `needs_human_review`, untouched** |

**The register more than doubled in eight days while adoption of the one mechanism on
the right surface stayed at zero.** That is the finding that shaped the whole session:
the gap is widening, not closing, and the binding constraint is adoption rather than
the checker. It is also a clean instance of the count-carries-its-date rule — every
figure in the bug file had gone stale by ADDITION and still read as current.

### The measurement that decided the design, and the control that saved it

The obvious census — "do the 13 declared SDLT values appear in the tool's stored
HTML?" — returned **13 of 13 present**, which reads like the design is validated.

It is worthless. Four of the thirteen values are `5`, `2`, `10`, `12`. The probes that
made it mean something were the ones that had to come out FALSE: `625000` (bug 225's
own expired cap) **absent**, `777000` and `314159` **absent** — and then the noise
controls `99` and `7`, **both present and neither registered**. Recorded in
`WRONG_CALLS` 2026-08-24 §1.

Then the surface. Two design agents independently said "script text, not the whole
page", so I tested it on the live page rather than taking it:

```
15,111 bytes = 6,132 script + 8,962 non-script
  "500,000"  script YES  prose YES     <- the register's own writer_line put it there
  "500000"   script YES  prose no
  "625000"   script no   prose no
```

**A whole-page check matching the comma form would have passed bug 225's page every
day for sixteen months.** That is the single most important line in this lane.

### Four phases, in the order they close doors

1. **The declaration stops failing silently.** Two defects the bug file recorded as
   already handled: P11 had never validated a tool fence (one production caller, and
   it is the experience-pattern register), and `parseCriteriaFacts` failed open on a
   fence that DID declare — which also disarmed the round-3 zero-rows warning, since
   that is gated on `issues` being non-empty. Council **APPROVED at round 1**.
2. **RFC_025 stage 2b.** `artifact_check` reachable for every source kind (the citation
   arm `continue`d before it, and every legislated figure is a citation fact), and
   addressable by `subject_key` instead of a component id that dies on decomposition.
3. **The byte probe**, annotation only, script text only, floor measured at 1000.
4. **Adoption**: propose the bindings already visible — 15 across 3 tools, one of them
   the estate's second SDLT calculator.

### The missteps, which are the useful part

**Nine mutations across the change; three passed and were worthless.** All three are
the same family — a check that cannot tell the world it is testing from the world it is
not — and all three are in `WRONG_CALLS` 2026-08-24.

- **Phase 1**: four tests of the note writer stayed green when the CALL to it was
  deleted, because every one invoked the writer directly.
- **Phase 2**: eight tests stayed green when the pre-pass itself was deleted, for the
  same reason. Writing the end-to-end test then **found a real defect no unit test
  could see**: an `attested_by` fact carrying an `artifact_check` was evaluated TWICE,
  once by the pre-pass and once by the original branch, appending two entries under one
  fact id and bumping `verified_at` from the second.
- **Phase 3a, the worst one**: the headline test for the headline rule. Mutating the
  probe to read the whole page left it GREEN — the prose writes the comma form and the
  code surface is searched for raw literals, so both readings failed the raw search and
  agreed. **The test asserted the right answer for the wrong reason.** Fixed with a
  fixture carrying `data-relief-cap="500000"` in markup against `625000` in code, plus
  three premise assertions so it cannot silently stop discriminating.

**I knew this lesson. It is this lane's own, from 2026-08-16, and its register entry
sits three bullets above the code I was editing.** Knowing it did not stop me repeating
it twice more in one sitting. The mechanical version that would have: *for every guard,
delete the CALL and not just the body.*

### Two more traps worth the ink

- **The guard's own first version was wrong.** Excluding every trailing comma to avoid
  matching `1,500,000` made `{ upTo: 1500000, rate: 0.10 }` invisible — the real band
  table, on the very page the rule was written for. A trailing comma is a list
  separator; it is only a thousands separator when a digit follows. And **Go's RE2 has
  no lookaround**, so the rule is hand-rolled byte checks, not a regexp.
- **An objection can be sound with the wrong reason attached.** Two seats said an
  unscoped note cooldown would let two same-named tools on different sites silence each
  other. That cannot happen — 134 current tool PLANs, 134 distinct subject keys. But
  **one fleet-global PLAN resolves on many sites (6 do)**, and half the finding is
  per-site, so site A really would have silenced site B. Checked the premise, kept the
  fix.

---

## 2026-08-25, session 3 — the first adoption, and two misses of mine that produced the good bit

### State at open

Two commits from the morning (`bba8a892d` ambiguity guard, `6ad4a8046` misplaced
`artifact_check`) committed and in no running binary. Probed rather than assumed:
all three new strings `0`, `stale_attestation` = 5 (positive control),
`ZZZ_must_be_absent` = 0 (negative). Both replicas started **09:27:24Z / 09:27:48Z**,
which is *before* either commit was authored (09:35Z, 09:53Z) — three independent axes
agreeing. Bumped `IMAGE_TAG` to `v1.0.1338` (`v1.0.1337` is what the cluster is already
serving, and a same-tag re-release re-serves the cached digest) and asked the owner for
the release. **Still unrolled at the close of this session.**

### MISS 1 — I warned another lane about their neighbour's file

`WRONG_CALLS.md` 15, in full there. Short version: the CONTRIB this lane filed into
`loanandmortgagecalculator_couk/` told them their `install_fences.py` would "refuse,
silently — its rule 2 skips a tool that is not ladder-eligible" and sent them to the
mcalc lane for `--allow-ineligible`. **I had read mcalc's `install_fences.py`.** LMC's
is a 233-line fork with no rule 2, no eligibility predicate and no such flag. One
`grep -c allow-ineligible` on a path I had already typed into my own CONTRIB returns 0.

The check is embarrassingly cheap and the reason I skipped it is the interesting part: I
had genuinely read *an* `install_fences.py` that day, so nothing felt unverified. **A
shared filename is not a shared file, and a fork is the shape that punishes this hardest**
— everything you remember about the original stays approximately true, so the wrong
paragraph reads fine to you and to them.

### MISS 2, which is the one that mattered — the true trap was in the slot the wrong warning occupied

LMC's installer `--apply` does an unconditional supersede + INSERT of a body **rebuilt
from scratch** out of `acceptance/criteria/<slug>.criteria.json`, and the fence it builds
carries only `profiles` / `no_auto_fix` / `no_auto_fix_reason` / `checks`. No `facts`
handling at all. The live `mortgages-stamp-duty` row's `created_by='operator:bugfix224-session'`
is that script's own hardcoded literal — **so it is the writer of the fence standing there.**

Which means the paste-ready fragment my own sweep had just filed, installed the way my own
CONTRIB told them to install it, **would have been deleted the first time anyone re-ran
their installer.** Clean exit, no error, fence still parses, tool silently undeclared again.

My advice ("install through the lane's own fence installer; never hand-edit the doc_plans
row") was *correct* and had exactly one failure mode, which I had not looked for. That is
worse than being wrong: a wrong warning is recoverable, a wrong warning that fills the
space where the right one goes reads as diligence.

`[MEASURED 2026-08-25]` over all **7** `doc_plans`-writing lane scripts: **1** injects into
the live body (agritec — safe, and why *their* declaration survives), **1** rebuilds but
carries `facts` (mcalc — safe), **1** rebuilds and drops it (LMC), **4** write unrelated
PLAN kinds. **The exposed population was exactly one, and it was the one I had warned
about the wrong thing.** Filed as a landmine, because its victim has no symptom to search on.

### MISS 3 — I nearly sized the adoption by the machine's list

The note proposed **7** bindings. I read the tool's script before declaring, on agritec's
"verify both directions" rule, and **the tool encodes 13**. The six the suggester could not
see are the rates: the register stores `2`, `5`, `5`, `5`, `10`, `12` and the code stores
`0.02`, `0.05`, `0.10`, `0.12` in `SDLT_BANDS`, plus `SURCHARGE_ADDITIONAL = 0.05`, plus a
bare inline literal `(price - FTB_NIL_BAND) * 0.05`. **No value probe can match `5` to
`0.05`, and at two digits the measured floor of 1000 forbids the attempt anyway** — two
independent reasons, and the note says nothing about either.

Declaring the seven would have left every rate in a stamp-duty calculator drifting behind
a fence that reads complete — `bugs_closed/225`'s class arriving by omission, inside the
document written to prevent it. **And this lane told the agritec lane exactly this rule
the day before** ("declare what the tool ENCODES, not the subset that happens to be
fenced") and then nearly failed it on its own first adoption. Second landmine.

### What was done, and what proves it

- **LMC `mortgages-stamp-duty` declares 13 facts, applied and live in `doc_plans`.**
- `install_fences.py` now carries a criteria file's top-level `facts` into the fence, so
  the declaration is reproduced on every future `--apply` instead of surviving until the
  next one. Additive: `--only simple` (no facts) prints no `declares` clause.
- **`--apply` was gated on a diff, not on confidence.** Regenerated the body to a file
  (`--body-out`) and diffed it against the live `doc_plans.body`: the only difference was
  the `facts` block, byte-identical once stripped. Safe because these bodies are
  deterministic — hardcoded date, no `now()` — and both files were unchanged since
  2026-08-09, the day the live row was written.
- **The proof is 0 → 13, not the row readback.** Scoped `refresh_evidence_base` dry run on
  the LMC site **before** the write returned an empty `fact_drift` array
  (corr `2bebb885`); the same dry run **after** returned 13 `unreconciled_declaration`
  entries (corr `d4dd59e2`).

### Phase 3a's first discriminating distribution

| probe verdict | n | which |
|---|---|---|
| `present_in_script` | **7** | the thresholds — 40000, 300000, 500000, 1500000, 250000, 925000, 125000 |
| `not_probed` | **6** | every rate — below the 1000 floor |
| `absent` | **0** | |
| `present_in_markup_only` | **0** | |

Against agritec's 24 × `not_probed`, this is the first sample carrying any information at
all. ⚠ **But read what it is**: on a tool whose declaration was authored *from the code*,
`absent` is structurally near-impossible, so this sample cannot estimate the `absent` rate
— which is the number Phase 3b actually needs. A declaration authored from the register
alone (or one that ages past a rebuild) is where an `absent` can come from. **Do not let
7/6/0/0 stand in for a distribution.**

### gamesdesign: NOT adopted, deliberately, and the reason generalises

Three notes, all proposing `gd-trials` = 10000 (their only fact above the floor; the other
three are 10, 11, 4). **No lane directory owns gamesdesign.** Applied my own new landmine's
check first — `created_by` on all three PLANs is **`tool-generator`**, a platform agent, and
`tool-spawn-rate-balancer`'s body was **fully rewritten between 07-29 and 08-21 (28 of 41
lines)**, so regeneration is observed behaviour.

So the LMC fix does not transfer. There the writer was a lane script I could patch in three
lines; here the writer is a platform agent that (per CLM-021) names neither `writer_block`
nor `evidence_base`. **Declaring by hand would knowingly write something the next rebuild
deletes.** Escalated rather than adopted; the durable fix is council-gated Go in the
generator. Also: `drop-rate-simulator` has no `doc_plans` row at all, so it needs a PLAN
before it can have a fence.

### Still true at close

- Both fixes unrolled. `misplaced_artifact_checks` is absent from every payload for a
  **binary** reason, not a data one — do not read that zero.
- **The ambiguity guard has no observable surface even after the roll.** `Ambiguous` reaches
  only the doc_note body (`refresh_evidence_fact_suggest.go:284`), never a result field; the
  suggester skips any declaring subject (`:167`), and agritec — the one register with the
  duplicates — now declares; and all five noted subjects are cooldown-suppressed for 30 days
  (`:246`) to ~09-24. I briefed a design pass that agritec's nine value-sharing pairs would
  exercise it. **They cannot.** Post-roll the honest claim is presence-at-the-binary plus the
  unit test, and it should be written that way rather than as a behavioural proof.
