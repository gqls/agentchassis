# NOTES — bugs_open/386 counting-fact drift

Append-only, newest at the bottom. Technical record: what was tried, what the system actually
said, and every misstep.

---

## 2026-08-25 — lane opened; what was read first-hand

Read end to end rather than grepped: `numberSupported` (`claims.go:1069-1121`), `EvidenceFact`
(`:74-96`), the sql refresh arm (`refresh_evidence_base_action.go:490-547`), the supersede
(`:1289-1318`), the gate (`validate_page_content.go:420`, `:449`, `:1320-1333`), and the series
machinery (`claims_series.go`).

Mechanism confirmed at the code: on any non-`fresh` outcome the refresh sets
`fact["value"] = live; fact["verified_at"] = today` and the previous reading is not kept anywhere
in the fact. The bug file's account is accurate.

### Misstep avoided by reading rather than assuming — the series machinery is NOT a drop-in

First instinct was that `claims_series.go` already solved this: it holds many dated observations
per fact, each with its own source, matched exactly. It is the right *shape*, but
`numberSupported` consults it only when `f.Value == nil` (`claims.go:1077`), and `IsSeries()` keys
on `len(Observations) > 0` (`claims_series.go:82-84`). So pushing superseded readings into
`observations` would (a) never be consulted while the fact keeps a current value, and (b) flip
every armed fact into a series, changing behaviour in `ValidateSeries` and the chart path. History
has to be a **distinct field**. Recording this because "reuse the existing mechanism" was the
right instinct and the wrong conclusion, and only the second read showed why.

### Correction to the bug file — candidate 1's premise is false on this estate

The bug file says counting facts "increase every day" and proposes a `monotonic` flag. But
`sql_for_agents/218_evidence_facts_for_043_sites.sql:306` records `work-items-completed` falling
1,267 → 1,051 because the ledger reaps. A `monotonic` flag would be false the first time that
happens, and a range `[previous, current]` vouches for every intermediate value nobody published.
Exact matching against *retained former values* is strictly tighter and needs no monotonicity
assumption. Caught by reading the repo's own migration, not by the census.

### Correction to CLM-027 and the bugfix_380 handoff — the discriminator does not exist

Both say a rotation finding whose `nearest_fact_id` has a `verified_at` newer than the page's
render is a stale render rather than an invention. `grep -rn nearest_fact` repo-wide returns
**zero Go hits**: it is a field in the auditor LLM's output JSON
(`claims_verification/SEED_claims_auditor.sql:70`), so it exists only on the LLM audit path — not
on the Go `unregistered_number` path that actually blocks the rebuild. And at the build gate the
comparison is undecidable in principle: the gate runs now against the current register, so the
fact's `verified_at` and "now" are both today. To be corrected in the register entry visibly, not
just here.

### The census `[MEASURED 2026-08-25]`

295 facts fleet-wide in current `evidence_base` specs; 29 sql-sourced number facts on 6 sites; 13
of those `exact` — fundamentallyai ×6, leopardess ×2, robot-hands ×2, vonc ×3. Only 2 facts
fleet-wide carry no `context_terms`, and both are already `exact`, so the degrade-to-exact rule is
a no-op there.

Disconfirmable, and it came out the interesting way: the archive **does** hold the history. 315
superseded `evidence_base` rows across 15 sites back to 2026-07-16, and fundamentallyai's
`F9-feed-items-collected` reconstructs daily — **11513, the exact value the convicted page
renders, is there dated 2026-08-23**. Had the archive been pruned, the whole backfill half of the
plan would have been impossible and today's stale pages would have needed a different remedy.

Second disconfirmable check: the counters moved again overnight. The bug file recorded the
register at 11646 / 10416 / 437 / 503 on 08-24; today it reads 11828 / 10600 / 454 / 542. The
mechanism re-arms daily, as claimed — had they been unchanged, the "self-inflicted and periodic"
argument would have needed re-examining.

## 2026-08-25 later — the owner ruled, and the ruling's own figure is stale

Owner ruling arrived as bug file §4b (commit `2a9091c7d`), relayed by the `bugs_open/364` session:
a counting fact is expressed as "at least" N, or the claim is cancelled or minimised. That
promotes candidate 3 (`gte`) to the default and puts *don't mint the claim* above it. §4b
explicitly does **not** supersede candidate 1, so the lane runs the ruling first and the durable
fix second.

§4b's caveat is the load-bearing part and it checks out: `numberSupported` gates on
`context_terms` via `strings.Contains` against a ±70-byte window (`claims.go:1086-1096`), a
substring test, so a `gte` fact vouches for every smaller number near a term it names.

**But the figure §4b cites is stale, by the very mechanism this bug describes.** It quotes
`ai-agent-orchestration.com` carrying `4068 gte / context_terms ["orchestration"]`. Read live
2026-08-25: the fact is `aao-orchestrations`, the terms are indeed the single broad
`["orchestration"]` — and the value is **7281**. The ceiling of what that one fact silently
supports has risen by 3,213 since the figure was taken. So the caveat is *understated*, not wrong.
Reported back to the 364 lane. This is the third document in two days to quote a counting fact
that had moved by the time it was read, which is itself the argument for the ruling.

## 2026-08-25 later still — the 13 facts do NOT take one remedy, and the motivating case takes the ruling's *other* half

Prompted by an FYI from the `bugs_open/387` session (its lane: the writer shipped the literal
`NNN` placeholder 14 times because the unscoped prompt carries only `writer_block`, never the fact
values). That made me check how each exposed fact's number actually reaches the page before
designing anything, and the answer splits the population.

`[MEASURED 2026-08-25]` `writer_line` + `writer_block_managed` for all 29 sql-sourced number facts:

**(a) The ruling is already implemented on 5 facts, by hand, and it works.** `F1-live-sites` is
`gte` 26 with `writer_line` = *"more than 10 live production sites … (live count {value}; state a
FLOOR, never the exact number)"*. Same shape on `F2-council-seats` ("more than a dozen"),
`C1-records-verified` ("more than 2,000"), `C4-agent-definitions-catalogue` ("more than 150
… ({value} at the last live count)"). A rounded-down floor, with the live value available to the
substituter and an explicit instruction not to state it. **That is the owner's ruling, in
production, predating the ruling.** Phase A copies this template rather than inventing one — and it
is also evidence the ruling is implementable without the accidental-support hole, because all four
carry narrow multi-word terms.

**(b) The five facts that actually convict the page have NO `writer_line` at all.** F9, F10, F11,
F12, F13 — precisely the bug's §1 evidence. `composeWriterBlock` composes from `writer_line`, so
these five contribute **nothing** to the writer's instructions while still being used by
`numberSupported` to convict. The numbers on the convicted page were therefore never written under
instruction from these facts.

Where they came from instead:
```
capabilities | evidence-chart | evidence-chart | comp_updated 2026-08-23 | deployed 2026-08-24
old_in_content_data = t | old_in_rendered = t | new_in_content_data = f
```
The stale value is **frozen into `content_data`**, written on 08-23 when the register said 11513;
today's 11828 is absent from it. So this is a stored snapshot produced by the component that exists
to render the register.

**Three consequences, and they reorder the lane:**

1. **The ruling's prose remedy cannot reach the motivating case.** There is no `writer_line` to
   rewrite, and "express it as at least N" is a *prose* instruction; the convicted content is a
   chart. §4b anticipated this by explicitly preserving candidate 1 — for this component class
   candidate 1 is not a tolerance widening, it is the *semantically correct* answer, because the
   chart already renders its own `verified 2026-08-23` stamp. "11513 verified 2026-08-23" is a true
   statement for ever, and needs no re-render at all. The register simply cannot currently agree.
2. **An assemble-mode rerender would republish the stale bytes.** Only a regenerating rerender
   (`rerender_sections`, i.e. reason ∈ `image_landed` / `section_data_resolved` / `cta_links_stale`)
   recomputes `content_data`. Any Phase C design must pick the reason deliberately; the default
   route is the one that cannot fix this.
3. **The real exposure is far smaller than 13.** Of the 13 `exact` sql facts, the fast movers are
   fundamentallyai F9/F10 (+~180 a day) and F11/F12 (+17 / +39 a day), plus leopardess
   `C1-ch-vet-mirror` and `C1-records-enriched`. The rest are small counts of *enumerable* things —
   `vonc-archetypes` 8 (and its writer_line names all eight), `vonc-guides` 4, `vonc-tools` 6,
   `rh-manufacturers` 6 (names all six), `rh-grippers` 10, and `F14-interactive-tools` 5 whose
   writer_line says *"an EXACT count — do not round it or state a floor"*. For those, exact is the
   honest form and the ruling's stronger option does not apply; converting them to `gte` would be
   the accidental-support mistake for no benefit.

So: **the ruling applies cleanly to about two prose facts on leopardess; the bug's own motivating
damage takes Phase B.** I had committed the order as "ruling first, durable fix second" an hour
ago. That is right as a default and wrong for the case that filed the bug — recorded here rather
than quietly re-ordered.

`writer_block_managed` is `true` on fundamentallyai and leopardess, unset on robot-hands and vonc —
though both unmanaged sites already use `{value}` in every writer_line, which is worth passing back
to 387: whatever blocks unmanaged sites from machine substitution, it is not the absence of
`{value}` in their lines.

## 2026-08-25 — M3 run properly, and THREE wrong turns of my own first

The premise that gates the whole durable fix — *every stale rendered value was once the register's
current value* — is **CONFIRMED, 5 of 5**. But I reached it after measuring the wrong surface twice
and drawing a false conclusion once, and the missteps are more instructive than the result.

### Misstep 1 — I scanned `content_data`, which is not a published surface

First premise test regexed `page_components.content_data`. It returned three numbers matching no
register value, one of which (`125`) I started to chase as a possible real finding. It is part of
`"…(122/125 on the first full day; most traffic is search-engine crawlers — always say so)"` — a
register fact's **claim text including its writer guidance**, staged in `content_data`.

Checked whether that guidance reaches the page: `guidance_in_content_data = t` on all three
components, `guidance_in_rendered_html = f` on all three. So it is not published and there is no
second defect — but `content_data` is a staging field holding the whole register snapshot, and
scanning it over-reports by construction.

### Misstep 2 — I then regexed raw `rendered_html`, which is chart markup

Second test extracted every 3+ digit number from `rendered_html` and reported 13 values per page as
"NEVER a register value → real finding": 127, 320, 555, 600, 620, 625, 666, 700, 766, 875, 1200,
9375, and 8125. **Every one is SVG geometry** — `evidence-chart` draws a chart, so viewBox bounds
and coordinates are in the markup. The identical list on all three pages was the tell and I should
have read it as one.

Worse, `8125` and `9375` *did* match former F10 values, so a coordinate coincided with a real
former reading. Had I trusted that, I would have "confirmed" the premise partly on chart geometry.
That is the accidental-support failure mode running in reverse, inside my own measurement.

**The rule: the scanners do not read markup, they read extracted text blocks. Do not hand-roll a
third text-extraction formulation** — which is exactly what the bugfix_380 handoff §3 warns about
when it notes a shared `page_visible_text(uuid)` would stop 601's query being a third formulation.
Use `cmd/claimscan`, which runs the same engine as the gate.

### Misstep 3 — "three pages are convicted" was false

From `content_data LIKE '%11646%'` I concluded three pages were convicted, and said so. The real
engine finds **one**. `digital-asset-recovery` and `index` store 11646 in `content_data` and never
render it — the chart draws only some of the snapshot it is given. **Conviction requires the value
to be RENDERED**, and `content_data` presence does not imply it. Both pages are `page_type=landing`,
which is *not* on `editorialPageTypes` (guide, blog-post, news-index, tool, game), so the skip list
is not the explanation and I should not have reached for it.

### The real measurement

Export honesty asserted first: 91 TSV rows against 91 in the DB, stderr 0 bytes — the documented
4-column recipe with `page_type`, without which every page reads UNKNOWN and the tool disagrees with
the gate it exists to predict.

`go run ./cmd/claimscan -evidence <live> -components <tsv>` → **5 findings across 91 components**,
all on `capabilities` / `evidence-chart`, all five carrying their own `verified 2026-08-23` stamp —
precisely the bug file's §1 list, no more and no less. Zero practice claims, zero suppressed.

Premise test against the archive, all five values:

| fact | value | first held | last held | current? |
|---|---|---|---|---|
| F9-feed-items-collected | 11513 | 2026-08-23 | 2026-08-23 | no |
| F10-feed-items-scored | 10194 | 2026-08-23 | 2026-08-23 | no |
| F11-council-rounds-revise | 428 | 2026-08-23 | 2026-08-23 | no |
| F12-council-rounds-approved | 483 | 2026-08-23 | 2026-08-23 | no |
| F13-council-rounds-rejected | 23 | 2026-08-21 | 2026-08-23 | no |

5 of 5 were the register's own value on the day the component was written (08-23 17:26), and none
is current. **Disconfirmable and it could have failed**: a rounded or paraphrased figure would have
returned no row. So history-exact covers 100% of the motivating case, and Phase B is provably
sufficient for it rather than merely plausible.

### Engine parity FAILED, and it matters for how these numbers are read

`git diff 4c996e1b5..HEAD -- platform/orchestration/datahelpers/` is **not** empty: `claims.go` is
**122 lines ahead** of the rolled chassis, from `52958897f` — the 364 lane's Phase 2 / RFC_053,
moving the number scan to COMPONENT grain, committed 11:50 today and not yet rolled.

So the claimscan run above predicts the **post-roll** gate, not the live one. That is the right
engine to design Phase B against (it ships after a roll too), but it must not be quoted as "what
the fleet does today".

Checked for collision: `52958897f` adds `normaliseSurfaceKey` and surface/grain handling and touches
**none** of `numberSupported`, `ContextTerms`, `Tolerance`, `seriesSupports`. Phase B lives entirely
in the fact-matching path, so this is same-file proximity, not a design conflict — but two sessions
editing `claims.go` means whoever commits takes both edits, and no hook prevents a same-file
passenger. Flagged to the 364 session.

## 2026-08-25 — my claims.go half was SWEPT into another lane's commit, and the cost is a reporting one

`63d95be1f` is my Go-slice commit. It carries five files. It does **not** carry
`platform/orchestration/datahelpers/claims.go`, which is the half the whole design lives in — and I
had named that path explicitly on `git commit`.

Not a loss, and not a mystery: the file was already committed by the 364 lane, which swept my
uncommitted working-tree changes into its own commit.

> **CORRECTED 2026-08-25, within the hour — I named the wrong commit, and I had named it in the one
> place it would be used for hand-tracing.** I wrote `001211abf`. It is **`6548e8d79`** ("a gte fact
> that is a ROLLING WINDOW falls…"), which added **98 lines** to `claims.go` and introduced
> `FactHistoryEntry`, `RetainHistory`, `History`, `FactHistoryMaxEntries` and the `historySupports`
> arm. `001211abf` added **25 lines of which ZERO are non-comment** — it appeared in my search only
> because its prose mentions `historySupports` while describing the function.
> **What caught it:** the 364 lane, checking my claim against its own commit rather than accepting
> it. **My error was the command.** I ran `git log -3 -- claims.go`, which answers "what last
> touched this file" — a question I did not ask. The question was "what introduced these symbols",
> and its command is the pickaxe: `git log -S 'FactHistoryEntry' -- <file>`. A path-filtered log
> will always name the most recent commit, and it will always look like an answer. This is the exact failure CLAUDE.md describes — committing
per task stops *you* sweeping up *others'* WIP; it cannot stop a session still running `git add -A`
from sweeping up *yours*. By the time my pathspec commit ran there was nothing left to commit for
that path, so it silently carried five files instead of six.

Verified rather than assumed, because a sweep can also mangle: HEAD's `claims.go` carries exactly
one each of `RetainHistory bool`, `History []FactHistoryEntry`, `func (f *EvidenceFact)
historySupports` and the `if f.historySupports(val)` arm, and both full packages pass fresh with
`-count=1` at the HEAD combination. The code is intact.

**The cost is not the code, it is the audit trail, and it is worth naming because it is invisible
otherwise:**

1. **My commit message describes code that is not in my commit.** Anyone bisecting or reviewing
   `63d95be1f` will find the tests and the register entry but not the mechanism they test.
2. **The council trailer is now on the wrong commit.** `098` joins commits to verdicts by trailer.
   My `Council-Submitted: 18dba069` sits on the commit that carries the *tests*; the in-scope
   platform file landed in `001211abf`, which carries no trailer at all. So the coverage report will
   show an in-scope commit as unreviewed while the review that actually covers it is credited
   elsewhere. Nothing dishonest has been written — the trailer asserts a submission, and that
   submission is real and does cover this code — but the join is wrong and no amend can fix it
   (forward-only).
3. **A same-file passenger is the one thing a pathspec commit cannot exclude**, and this is the
   mirror image: not a passenger riding *in*, but my own work riding *out* under someone else's
   message.

Nothing to undo. Recorded here, reported to the 364 lane so their commit's contents are not a
surprise to them either, and the correlation is written down in both this file and CLM-028 so the
verdict can still be traced by hand when `098` shows the gap.

**The checks — and it takes TWO, because there are two failures here and neither check sees the
other's.** Recorded as a pair after the 364 lane pointed out that mine only catches half of it:

- **A path you NAMED that is ABSENT from the yellow scope block** = someone committed it between
  your edit and your commit. That is *my* failure above, and the scope block catches it.
- **A path that is PRESENT but arrived FATTER than you thought** = you have swept up someone else's
  work. The scope block cannot see this: it lists *files*, not lines, and the file was present and
  expected — only its size was wrong. The check is `git diff --numstat <file>` before committing,
  which prints `added deleted path`. `98 2` against an expectation of about twenty is not a number
  anyone scrolls past. That is the 364 lane's failure, and it is the same "gate on the COUNT, which
  no content can fool" rule that already exists as a landmine for deleted markdown bullets.

Neither is a substitute for the other, and the second is the one with teeth: the sweeper is the
party who can still prevent it, and the swept party finds out afterwards or not at all.

## 2026-08-25 — the wrong-sha error is TWO instances, not one, and my "I checked" was not a check

Follow-up from the 364 lane. When I corrected my sha I grepped for `001211abf`, found it in two
`site_ai_agent_orchestration` docs, and told the owner they were "a legitimate separate use — that
lane's own work was swept into 001211abf. Not my error propagating."

**I did not check that. It is false.** Verified now, properly:

- `git show 001211abf -- WRONG_CALLS.md --numstat` → `34 0`, and `grep '^+#\{2,3\} '` on that diff is
  **empty**. It added no new entry; it appended to the 364 lane's own existing entries 12 and 13.
- The site lane's entry — *"I wrote an EXEMPLAR into a prompt and it shipped to the public as copy"* —
  was introduced by **`3d31b86a9`** (the `bugs_open/381` lane), which added 100 lines to that file.

So **two different lanes, on the same day, independently attributed a sweep to the same innocent
commit**, and neither was swept by it. The cause is the one I had already diagnosed an hour earlier:
`git log -N -- <file>` answers "what last touched this file", and at the moment you go looking, the
most recent commit to touch a hot shared file is almost always someone else's. It is a plausible
wrong answer delivered with no tell.

**My own failure here is the second-order one and is worth more than the first.** I had just been
corrected for attributing a sweep with the wrong command. In the very next action I passed judgement
on someone else's attribution — "true, leave it alone" — *using no command at all*, and reported
that to the owner as though I had verified it. The cost of checking was one pickaxe run, which I had
already typed once that hour.

**And a trap on top of the fix, which matters because I have just told everyone to use the pickaxe:**
`git log -S '<symbol>' -- <file>` matches any change in the symbol's OCCURRENCE COUNT, **including a
commit that merely mentions it in prose.** That is exactly why `001211abf` was so magnetic: it
surfaced in my `-S historySupports` search while containing none of the code, because the 364 lane's
comment *names* the function while describing it. So the pickaxe alone can also hand you a plausible
wrong answer.

**The discriminator, two seconds:**
```bash
git show <sha> -- <path> | grep '^+' | grep -v '^+\s*//'   # empty ⇒ comment-only, not your carrier
```

**The full attribution recipe, since three commands are each individually insufficient:**
1. `git log -S '<symbol>' -- <file>` — candidates that changed the symbol count. NOT `git log -N`.
2. `git show <sha> --numstat -- <file>` — a carrier is fat; a comment-only commit is small.
3. the non-comment filter above — the one that separates "contains the code" from "mentions it".

## 2026-08-25 — council APPROVED round 1 (corr 18dba069), and the three advisories answered with checks

11 reviewers, 6 abstained, **approved with 3 advisory objections, none high-severity**. Approval is
not the interesting part; the objections are, and two of them were answerable only by running
something. Answered here rather than left in the artifact.

### 1. `bug_historian`, MEDIUM — "one guarded call site, the mechanism still generic elsewhere"

The sharpest objection, and it used my own submission against me: my `grounded_in` names **two**
writers of the raw `evidence_base` facts map, and I wired `recordFactHistory` into only one. If
`evidence_citations.go` can overwrite an armed fact's `value`, the outgoing reading is lost through
the sibling door — the exact loss this exists to prevent (016b §9, case 7).

**ANSWERED — the sibling door does not exist, and it is structural rather than incidental.**
`evidence_citations.go` is **append-only**: it builds a set of existing fact ids (`:232-237`, reading
only `id`), `continue`s on any candidate whose id is already present (`:277-280`), constructs a
brand-new map for the rest, and does `facts = append(facts, fact)` (`:304`). `grep '\["value"\]'`
over that file returns **nothing**; the only `"value"` reference is the copy-from-candidate loop
populating a NEW fact. There is no assignment into an existing facts element anywhere in the file —
the sole `fr.(map[string]interface{})` cast is the id-collection read. A fact that already exists is
never mutated, so its value can never be overwritten without history.

### 2. `reuse_agent` MEDIUM + `prior_art_librarian` LOW — why not read the superseded rows directly?

Fair, and unargued in the submission: my own evidence says the history is already recoverable from
315 superseded `site_specs` rows, so why a second store? **The rationale I owed:**

- **`numberSupported` has no database handle and must not acquire one.** It is a pure method on an
  already-parsed `EvidenceBase`. Its consumers include `cmd/claimscan`, which imports **no database
  package at all** — only stdlib plus `datahelpers` — and scans from an evidence JSON file exported
  by hand. Making history a live read would mean threading a DB handle into a pure scanner that an
  offline CLI depends on, and would make the operator tool unable to predict the gate it exists to
  predict. That is a worse coupling than a duplicated array.
- **The superseded rows are an audit trail, not an index.** Answering "did this fact ever hold N"
  from them means unmarshalling every superseded revision of the *whole register* for that site —
  per number, per component, per page, inside a scan that already runs per-component.
- **They are unbounded.** `site_specs` has no retention job (that is why the archive reaches back to
  2026-07-16). Reading them directly would accept every value the fact ever held, for ever,
  reintroducing precisely the unbounded-acceptance problem the 90-entry cap exists to prevent. The
  cap is not a storage convenience; it is the bound on what the scan will accept.

So the duplication is deliberate and the superseded rows keep their job: they are the **backfill
source** for Phase 2 and the audit record, not the read path.

### 3. `guardian`, MEDIUM — "an eyeballed claim about someone else's work, not a verified one"

Correct to insist, and my submission did assert it from a `git show` rather than from the file.
**ANSWERED against the actual current file:** extracting `numberSupported` from the rolled chassis
(`4c996e1b5`) and from HEAD and diffing them yields **exactly my 11 added lines** and nothing else.
The other lane's two commits touch nothing inside that function, nor `ContextTerms`, `Tolerance`, or
the tolerance switch my arm sits after.

### Advisories carried forward rather than answered now

- **`guardian` LOW — arming is not a bulk toggle.** The gate is per-fact *data state*, not a
  code-level kill switch, so a careless seed could arm many facts at once and turn a contained
  feature into a fleet-wide behaviour change with no code review in the loop. Recorded in CLM-028:
  arming is a one-fact-at-a-time reviewed operator act.
- **`compliance` — the Phase 2 control needs a sharper acceptance test than "read what disappears".**
  Check that every disappearing finding sits on the register-rendering component (the
  `evidence-chart` class). **If a disappearing finding is free-text PROSE, that is the
  accidental-support signal, not a stale render, and arming that fact would quietly authorise
  invented copy to pass.** This is the best thing in the verdict and it goes into Phase 2's gate.
- **`debug_historian` — verify at the binary.** When this rolls, confirm by pod-grep for the new
  symbols against the running binary, not by git or image tag, with a must-be-present and a
  must-be-absent control in the same breath.
- **`architecture` — the thing to watch, and it is not volume.** `EvidenceFact.History` is a
  general-purpose "formerly-true value" primitive, not counting-fact-specific. If a second consumer
  class starts arming it for a *different reason*, that accumulation crosses into `needs_rfc`
  territory — and the cap and the opt-in default "won't self-police against semantic scope creep the
  way they police against volume". Recorded in CLM-028.

### One honest limitation of the verdict itself

Three seats (`editquality`, `prior_art_librarian`, and by implication `guardian`) flagged that
`site_specs` is **not in their available schema**, so the numeric claims this design rests on — 5/5
findings, 8-of-29 falling facts, 315 superseded rows, the zero-consumer enumeration — **stand
unverified by the council**. They are verified by me, with the queries in RUNBOOK §§1-4 and §7, and
`prior_art_librarian` is right that a human should re-run them before arming any fact. An approval is
not independent confirmation of a number a reviewer could not see.

### Addendum, same day — the limit is bigger than the allowlist, and I had it too small

I wrote above that three seats could not see `site_specs`, so my numbers "stand unverified by the
council". The 364 lane generalised it: a seat reads rows but **runs nothing**, so any figure produced
by executing code is unverifiable regardless of which tables are visible. I checked the mechanism
rather than accepting either version, and it is stronger than both.

`[MEASURED 2026-08-25]` A `council-gate` `review_*` step's config carries exactly seven keys:
`ai_service`, `error_step`, `input_fields`, `output_format`, `prompt_template`, `temperature`,
`tolerate_truncation`. **There is no SQL key, no tool key, no query key.** No review step has any
query capability at all — probed directly: `s.step->'config' ?| array['sql','query','tools','tool',
'db','database','queries']` over every `review_*` step returns **0**.

> **CORRECTED 2026-08-25, within the hour.** I wrote that "the only `query_database` steps in the
> entire workflow are `compose_verdict` and `compose_verdict_checked`". **There are THREE** — I
> missed `load_schema_hint`. Caught by the 364 lane enumerating them itself instead of taking my
> word. **The conclusion is unchanged and is now better supported**: `load_schema_hint` injects a
> schema *description as text* into the prompt, which is precisely why `editquality` wrote "SQL
> checks against the given schema can't reach…" — reasoning about what a query would show, never
> reporting one it ran. But the evidence I gave for it was false.
> **The cause: I truncated my own output and then made a completeness claim on it.** I listed the
> non-review steps with `| head -20`, and `load_schema_hint` sorts after the twenty `gate_*` rows.
> A `head`/`tail` silently converts an ENUMERATION into a SAMPLE, and "the only X" built on a sample
> is false with no tell — the row you need is exactly the one past the cut. If a claim contains the
> word *only*, *every*, *no* or *none*, the command behind it must not contain a `head` or `tail`;
> use `count(*) OVER ()` (or `wc -l`) so the total is in the output and disagrees with you.

**So a review seat executes nothing whatsoever** — no SQL, not even read-only; no Go; no tests; no
build. The "given schema" the seats referred to is text in a prompt template, not a query capability.

**What that means for this verdict, stated plainly because it is easy to let an approval imply more
than it does:** every figure in my submission was unverifiable by every seat. The register numbers
(5/5 findings, 8-of-29 falling, 315 superseded rows, zero consumers) because no seat can query
anything; the engineering claims (14 tests green, 7 mutations killed, both packages fresh with
`-count=1`) because no seat can run anything. **The council reviewed my reasoning and my plan. It did
not, and structurally cannot, verify a single number in either.**

**And it refines the 364 lane's own distinction, which I would otherwise have adopted whole.** They
split an *un-shown* check (a process failure, fixed by putting the query in the submission) from an
*unseeable* table (a limit of the instrument, unfixable by submission quality). The first half is too
optimistic: if a seat executes nothing, then a shown query is also just text, results included.
Including it makes the **inference** auditable — a reader can see whether the query would answer the
question — but it never makes the **number** verified. So:

> A submission can make its reasoning checkable. Nothing a submission can do makes its evidence
> checked. The only thing that verifies a figure is someone re-running it.

Which is exactly why `prior_art_librarian`'s line — *"a human should re-run the plan's own cited
queries against live `site_specs` before arming any fact"* — is the single most load-bearing sentence
in the verdict, and why it is written into CLM-028's verify-later rather than left in an artifact
nobody opens once the trailer exists.

## 2026-08-25 — the slice is LIVE at the binary, and the pre-arming control passes with a negative control

The fleet rolled to `v1.0.1339` while this lane was working (the 364 lane's Phase 2 shipped on it).

**Live, verified at the artefact, not at the table.** `[MEASURED 2026-08-25]` Both `agent-chassis`
replicas carry the new capability. Probed as a CAPABILITY (the JSON tag), not a commit, with both
controls in the same breath as the landmine requires:

| probe | r5bj7 | vx8b6 | meaning |
|---|---|---|---|
| `context_terms` (present-control) | FOUND | FOUND | the probe can find things |
| **`retain_history` (mine)** | **FOUND** | **FOUND** | the slice is live |
| `zzz_not_a_real_symbol_386` (absent-control) | absent | absent | the probe is not matching everything |

**⚠ And do NOT date this roll from `service_binary_capabilities`** — the 364 lane found that table
stale for this very roll: its newest `agent-chassis` row named a commit containing none of the Phase 2
code while both binaries plainly carried it, and a third lane consequently dated `v1.0.1339` from it
and got a commit that predates the fix. **The binary was right and the table was wrong.** Both of us
had been treating that table as the no-shelf-life authority. It is a hint.

So the first half of CLM-028's "inert twice over" is discharged: the code is live. The second half
stands — no fact is armed, so nothing has changed for any page.

### The pre-arming control, run in full (this is Phase 2's gate)

Fresh export, asserted 91 rows against 91 in the DB, stderr empty.

- **Baseline** (live register, as the fleet sees it today): **5 findings**, all
  `capabilities/evidence-chart` — 11513, 10194, 428, 483, 23. Unchanged from this morning despite the
  roll to component-grain.
- **Armed** (F9-F13 given `retain_history` plus history seeded from the 479-row archive, deduped by
  the same rule `recordFactHistory` uses, capped at 90, current value excluded): **0 findings**.
- **Disappeared: exactly those 5, and every one on `evidence-chart`.** Zero in free-text prose —
  which is the `compliance` seat's sharper acceptance test passing, and the result that would have
  stopped the arming had it come out otherwise.
- **Appeared: none.**

### The negative control, which is what makes the above worth anything

Zero findings after arming makes "nothing else disappeared" trivially true — the site had nothing
else to lose. So the diff alone cannot distinguish "spares stale renders" from "spares everything".
Following the 364 lane's suggestion of a behavioural probe over a symbol probe, I injected a value
the register has **never** held into the same component and rescanned against the ARMED register:

```
11513 -> 11514   (archive runs … 11373, 11513, 11646, 11828 — 11514 never existed)
armed register + synthetic copy  ->  1 finding: NUMBER capabilities evidence-chart "11514"
```

**Still flagged.** One away from a genuinely-held value, in the identical context window, and caught.
So arming vouches for exactly the values the register held and nothing adjacent to them — exact-match
doing the job it was chosen for over a range or a `gte`.

**What this does and does not license.** It licenses arming F9-F13 on fundamentallyai as a reviewed,
one-fact-at-a-time operator act. It is not itself the fix being live: no register has been written,
and `bugs_open/386` stays OPEN until a fact is armed and a real stale render is spared in production.
