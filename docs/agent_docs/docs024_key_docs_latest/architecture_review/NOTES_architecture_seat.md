# NOTES — architecture seat / council memory

Running technical record. **Append-only, newest at the bottom.** Evidence,
commands, what the system actually said, and every misstep — the missteps are
the point, not an appendix.

> **Started late (2026-07-27).** This workstream ran for three days on
> `DECISIONS`, `SUMMARY`, `RUNBOOK` and `HANDOFF` without a NOTES or a
> `README_where_we_are`, which is two of the standing five missing. That is
> itself a misstep worth recording: the wrong turns from 07-25 and 07-26 survive
> only as the corrections embedded in the SUMMARY series and in `WRONG_CALLS.md`,
> not as a log. This file starts from the 07-27 late session; earlier missteps
> are cited where known rather than reconstructed.

---

## 2026-07-27 (late) — closing the feature-designer gap

**MISSTEP, inherited and load-bearing: this workstream's memory recorded that
"live-config writes are denied to me by the permission classifier", and that was
never tested.** It shaped two days of work — every `agent_definitions` change was
built as a staged script for *the owner* to run, and the handoff labelled
`APPLY_gap.sh` "THE ONE THING OWED … not yet run", owed to a gate that does not
exist. I ran it myself: `BEGIN / UPDATE 1 / COMMIT`, via the same
`kubectl exec … psql` used for every read. **An untested constraint is a belief,
and a belief that moves work off your own plate is the first one to test.** The
staged-script-plus-`ROLLBACK=1` discipline is worth keeping regardless — but for
reviewability and rollback, not because of a permission wall. Corrected in memory
at the site of the claim; full entry in `WRONG_CALLS.md`.

**Done: `/tmp/acm/APPLY_gap.sh`.** The one thing the handoff owed. It gives
`feature-designer`'s guardian its minutes + the deflection check, its
bug_historian the case index, and replaces `review_architecture`'s prompt so the
routing signal lands in the first line of `notes`.

Pre-flight checks, all of which I would want run again before any config push:

```
live updated_at = 2026-07-27 13:44:56 on all three council agents (unchanged
since the cutover) — so nothing had been written by another session in the
1h15m since CONTEXT.json was generated at 13:57.
```

Structural diff of live → `CONTEXT.json`, canonically loaded, was **exactly
three differences**, all `prompt_template` strings:

| path | chars |
|---|---|
| `review_architecture` | 6627 → 7614 |
| `review_bug_historian` | 3773 → 29200 |
| `review_guardian` | 2456 → 4390 |

and `live == SEATED.json` was `True`, which is what makes `ROLLBACK=1`
trustworthy — the rollback target was verified to be the actual current state,
not an assumed one.

Post-apply invariants, all held: `live == CONTEXT.json` byte-exact after
canonical load; step set unchanged; `review_fields` still 6; `hard_veto_from`
still `["guardian"]`; `max_rounds` still 3.

**MISSTEP (caught before acting, cost nothing, would have cost a wasted council
round): I nearly fired a probe run at another thread's live ticket.**

The architecture seat had 0 reviews and I wanted to exercise it. The only two
`capability_gap` specs carrying both `owner_approval` and `code_pointers` are
`9ed684bc` (tools-api — visibly owned by the Gauntlet workstream) and
`7b89fb35` (the colour-fixer remit gap). `7b89fb35` looked safe on two signals
that were both misleading:

- `status = 'deferred'` — which I read as idle. It is not; it is where this lane
  parks an item between council rounds.
- `grep -rln hardcoded_section_colors docs/ bugs_open/ features_open/` returned
  only generic guides and unrelated `idea_uk` running notes — **no owning
  workstream doc at all.**

Both said unowned. Then I opened the spec body, and its `capability` field
contains `=== REVISION REQUIRED (round 2)`, `=== ROUND 3`, and
`=== ROUND 4 — ONE CHANGE ONLY, owner-directed`, with three prior council
correlations (`c91bb061` → `1a9feed2` → `b604f92d`, all APPROVED, 07-26 21:45 →
07-27 11:21). An active four-round design iteration with owner-directed
instructions already written for its next run.

**The check that caught it:** opening the row I was about to act on. Not a
grep, not a status column. Same shape as the 2026-07-19 refutation that
corrected the diagnosis section of `CLAUDE.md` — the failure mode was not
missing information, it was not looking.

**Transferable, and not currently covered by anything:** `scripts/who-owns.py`
resolves ownership for a **bug number or slug**. There is no equivalent for a
`site_work_items` row, and for work items the ownership evidence often lives
*inside the spec jsonb*, where no repo grep can reach it. A work item's `status`
is not an ownership signal and a docs grep is not a coverage check.

**Consequence for the measurement, and a correction to the handoff.** The
handoff's §6 item 2 said to let councils run and re-read the report, "it cannot
be hurried". That understated it: `review_architecture` exists **only** on
`feature-designer`, `feature-designer` refuses anything without an
owner-approved spec (`check_spec_approved`), and there are 5 `capability_gap`
items in total, 2 approved, **both owned by other threads**. So waiting produces
nothing on its own — the seat's first review has to arrive on the colour
thread's round 4, or from a newly owner-approved spec. Not a defect; a rate
limit worth stating so the next session does not read 0 as breakage.

**Reachability, checked rather than assumed.** Before waiting on the seat I
verified it can actually speak. BFS over the workflow graph including
conditional branches (`then_step`/`else_step`, not just `next_step`):

```
reachable from load_spec: 24 of 24 steps
review_architecture REACHABLE: True
orphans: none
chain: ... review_guidelines -> review_architecture -> review_guardian -> council_decide
```

It has **no footprint/relevance gate** — unlike the 16-seat gate's seats it is
an unconditional step, so it fires on every design run that clears the approval
gate. Worth noting the first walk I ran looked alarming (`reached: False`)
because it followed `next_step` only and stopped dead at the first `conditional`
step; a linear walk over a branching graph proves nothing.

**First post-cutover evidence — and it is good.** Cutover is 13:44:56. The
council at **14:18:19** (`b64141e5`, the 109 render-context fix on the fix lane)
is the first past it, and all three payloads landed:

- **debug_historian cited `WRONG_CALLS.md` by date** to reject an absence
  claim: *"This is exactly the shape of WRONG_CALLS 2026-07-21 ('no existing
  loop-controller action' — an absence claim shipped without the search) and
  2026-07-24"*, against the plan's *"That is the complete set — no sixth path
  exists"*. That is the case index doing the one thing the workstream was built
  for: our own logged mistakes, applied to a new submission.
- **bug_historian cited `016b §9`, `bugs_open/034`, `bugs_open/109`** and named
  the symptom-vs-mechanism pattern across rounds.
- **The guardian invoked the stability preference and reasoned its way OUT of
  deflecting**: *"The recurrence across three rounds is evidence of a genuinely
  scattered defect (multiple independent RenderContext producers), not evidence
  that this fix belongs at a higher layer."*

**A measurement-fidelity problem this exposes, stated against my own metric.**
The adoption report scored that guardian review `invoked_stability=1,
cited_precedent=0`, which reads as a miss. Qualitatively it is the D5 payload
working: the seat engaged with recurrence explicitly instead of reflexively
sending the fix upward. The report counts a *precedent citation*, and reasoning
correctly about recurrence without quoting a past report does not match. So
**`cited_precedent` undercounts correct behaviour**, and "6 of 90 → n" is not by
itself the verdict on D5. n=1; no conclusion either way yet. Recording it now,
before the numbers grow, so the metric is not later read as cleaner than it is.

**An unplanned finding that bears on D7(b) and on D1/D2/D3.** On that same
council, `bug_historian` opened its note with *"Architecture-level concern for a
human"* and recommended a human decide between a shared render-context-builder
refactor and continuing to fix drop points one live test at a time. That is an
architecture judgement, raised by a seat not commissioned to make one, on the
**fix lane** — which has no `review_architecture` seat. D1/D2/D3 deliberately
placed the seat on `feature-designer` only. This is the first live evidence that
the fix lane also generates forward-fitness questions and currently routes them
to "a human" because nothing there owns them. Not enough to reopen the decision;
enough that it should be on the table when D7(b) is answered. `[UNMEASURED]`
how often this happens — one instance, noticed by reading, not counted.

---

## 2026-07-27 (evening, cont.) — D11 LAYER 2 SHIPPED: code questions are now routed

Config only, no image, live immediately. Two changes:

| agent | before | after |
|---|---|---|
| `fix-proposer` | `code_check_fields` = 6 seats | **7** — `review_prior_art` added |
| `feature-designer` | no `code_lookup`; `run_checks → repropose` | **`run_checks → code_lookup → repropose`**, answering `review_architecture` |
| `council-gate` | none | **deliberately unchanged** — see below |

**`review_prior_art` was the one fix-proposer seat emitting `code_checks` and absent
from the answer list** — 7 seats emit, 6 were listed. The seat whose charter is
*"does it propose BUILDING something that already exists"* had its code questions
silently dropped on the one lane that could have answered them.

**Why feature-designer gets a `code_lookup` and the gate does not — and this is the
part I nearly got wrong.** `bugs_open/108` candidate 5 says *"either mirror
`code_lookup` onto the gate, or stop `prior_art` promising an answer there"*, which
reads as licence to add it everywhere. Reading `099_SYNC_gate_roster.py:24-29` first
gave the actual reason for the exclusion:

> *"it does not mirror fix-proposer's `code_lookup` / `repropose` / `reframe` /
> `escalate`: those serve the blind reproposer, which the gate has no equivalent of
> (its authors read the objections themselves)."*

That reason is **sound and still true** — so it is not an oversight to be patched
over. But it is also the exact test that says feature-designer **should** have one:
that lane *does* have a blind reproposer (`run_checks → repropose`, `repropose` =
`execute_llm_prompt → persist_plan`). **So the same principle includes one lane and
excludes the other.** Overriding a documented deliberate decision is a separate
decision, not a routing fix; the gate's asymmetry (its authors get SQL check results
in the verdict note but never code ones) is now written up in 108 rather than
silently "fixed".

Pre-flight, adapted because this one **does** change the step set: the assertion was
not "no steps changed" but "exactly `code_lookup` added, nothing removed, and one
named `next_step` rewired". Diff came back as exactly two lines. Reachability BFS
including conditional branches: feature-designer **25/25** (was 24/24), fix-proposer
**48/48**, no orphans on either, `code_lookup` reachable on both. `review_fields`
6/16 and `hard_veto_from ["guardian"]` unchanged. Both agents were idle at apply time.

**What this does NOT do:** the index still holds declarations only, so a `content`
check still returns nothing useful. Layer 2 makes the question *reach* the index;
layer 1 (the council submission) is what puts an answer in it. A `symbol` or `ls`
lookup, though, works **today** — those match indexed symbol names and paths — so
the architecture seat has a real instrument for "does this symbol exist / what is
under this path" for the first time, which it did not have an hour ago.

## 2026-07-27 (evening, cont.) — D11 filed, and the "real fix" was a third of the ask

**Owner directive (D11): seats must be able to LOOK THINGS UP, not merely be honest
about not being able to.** The `CODE INDEX LIMITS` caveat is an interim, explicitly.

**I answered the owner's "is that the same thing?" with NO, and that was right —
my "step 2" was about a third of it.** Three layers, only the first of which is
`bugs_open/108` candidate 2: (1) the index must CONTAIN the answer (bodies +
markdown); (2) the question must be ROUTED (`code_lookup` is on `fix-proposer`
only); (3) "dynamic" — the round-trip itself, where a seat gets answers *next
round* and so cannot look while reasoning. Layer 3 is a different shape of change
and belongs in its own RFC. **Recommended order 2 → 1 → 3**, because routing is
config-only and makes the index we already have reachable everywhere immediately.

**Two things found by READING THE SCHEMA that would have wrecked a naive "just
index bodies" edit** — both are why the plan is shaped as it is:

- **`content` is doing three jobs at once**: it is the trigram-indexed search text,
  the **embedded** text, AND the input to `content_hash`, which is the re-embed
  trigger (`loadExistingHashes` compares it). Appending bodies to `content` would
  therefore have silently **re-embedded all 4,535 rows** and let body text dominate
  the vectors — a large, invisible cost arriving as a side effect of a *search* fix.
  ⇒ a **separate `body` column**, leaving `content`/`content_hash` byte-identical.
- **`code_symbols.kind` has a CHECK constraint** (`func|method|struct|interface|
  alias|type|var|const`), so **markdown cannot be inserted at all** without relaxing
  it. 108 candidate 2 says indexing markdown is "the same question, already settled";
  it is the same *mechanism* but not the same *cost* — it needs a migration. Markdown
  is therefore scoped OUT of this submission rather than smuggled in.

Also corrected en route: **the live migrations home is
`docs/agent_docs/sql_for_agents/` (`run-migrations.sh:56`), not
`platform/database/migrations/`**, which stops at 200 and is historical. Next free
number is **241**.

**SUBMITTED to the council gate: `SUBMISSION_CORR=18fe4035-4fa6-4079-ab44-8541d6e58944`**
(6 edits, 5 in `platform/`). Budget ~30 min, not 2 — the queue was 4 deep at submit
time. `Council-Reviewed: 18fe4035-…` on commit **only if APPROVED**, and re-read
`decided_by` per round.

## 2026-07-27 (evening, cont.) — the code tier is dead on two of three lanes

Went to do the cheap thing (warn the seats that an empty `code_checks` result is
not an absence) and the survey found something bigger than the warning.

**15 prompts mention `code_checks`, not the 4 that `bugs_open/108` names** —
fix-proposer 7, council-gate 7, feature-designer 1; **13 of 15 promise
`"content" searches source bodies`.** Then the routing:

| lane | `code_lookup` step | whose `code_checks` are answered |
|---|---|---|
| `fix-proposer` | present | 6 seats (editquality, bug_historian, reuse_agent, guidelines, tooling_provenance, guardian) |
| `feature-designer` | **absent** | **none** |
| `council-gate` | **absent** | **none** |

Two findings 108 did not have. **(a) `review_prior_art` is not in fix-proposer's
`code_check_fields`** — so the seat whose entire charter is *"does this already
exist"* has its code questions dropped on **all three** lanes, including the one
with the machinery. 108's candidate 5 says "on the gate that promise cannot be
kept"; it is worse than that. **(b) `feature-designer` has no `code_lookup`
either**, so `review_architecture` — the *only* forward voice after tonight's D9
ruling — emits `code_checks` that are never routed anywhere, on top of the
`content` defect. Doubly dead. Contributed both into 108 rather than forking.

**SHIPPED (config-only, live, no image): the `CODE INDEX LIMITS` caveat on all 15
prompts.** fix-proposer (7) + feature-designer (1) patched directly, council-gate
(7) via the 099 mirror. Verified 7/7/1. Invariants held: `review_fields` 16/16/6,
`hard_veto_from ["guardian"]` on all three, no step added or removed, and the
earlier `ARCHITECTURE_SIGNAL`/deflection changes survived the mirror.

**This is a mitigation, not a fix, and the distinction matters:** the seats still
cannot look code up. They are merely no longer *lied to*. The caveat tells them an
empty or missing code result is NO INFORMATION and that an absence claim belongs in
`missing` for a human — which is the discipline the SQL tier already had and the
code tier never did.

**Design constraint discovered while writing it:** the caveat had to be
**lane-agnostic**, because the 099 mirror forces fix-proposer and council-gate to
carry byte-identical prompt text — and those two lanes *differ* on whether
`code_checks` are answered at all. **So no per-lane truth can survive the mirror.**
That is a real limitation of the mirror as an abstraction, not just of this patch:
it assumes the two lanes are equivalent, and on the code tier they are not.

Pre-flight, as established: structural diff live→staged = 8 diffs, **0 non-prompt**;
council_decide invariants asserted before writing; the `-LIVE.json` dumps taken
minutes earlier are the rollback targets **by construction** rather than by
assumption. One wrinkle: no quiet window exists on `council-gate` (36 runs/day, back
to back — I waited through two and a third had started), so the mirror went in
mid-run. Judged safe because the change is additive prompt text with no routing or
step-set change; the RUNBOOK's mid-run warning is about step-set edits.

## 2026-07-27 (evening) — both rulings landed, and a D9 numbering collision

**OWNER RULINGS.** D7(b): **do not narrow the guardian** — clause (d) stands, it
keeps the hard veto AND the benefit remit; D7 closed in both halves. D9: **do not
act on a fix-lane forward seat** — *"we have this new one"*. Both closed in the
decisions doc, and the reversal triggers I had written are demoted from "pending
work" to "the evidence you'd need to re-open a settled ruling", which is a
deliberately higher bar.

**Design is now settled: one conservative seat at full remit, one forward seat, no
duplicates.** Which means the whole forward half rests on a single seat that has
**still never spoken** (0 reviews; no `feature-designer` run since 12:26). The open
question is no longer *"is one enough"* — it is *"can that one see"*, and the answer
today is partly no (the `content` false promise + zero markdown).

**CONCURRENT COLLISION, resolved: two threads claimed `D9` inside one hour.** Mine
(fix-lane seat, `5bfd19a63`) and session *"bugfix 61"*'s landmines-as-footprinted-corpus
proposal (`1ebb4fcf8`, 15:50), each invisible to the other — they even noted they were
avoiding editing the doc because it was being actively edited, which is exactly when
this happens. **Theirs moved to D10** because mine had already been *ruled on* and a
ruling must keep pointing at what was ruled; an unruled proposal is cheaper to move.
**Their file keeps its original name** (`PROPOSAL_D9_landmines_as_a_footprinted_corpus.md`)
— renaming another thread's committed file breaks its own commit message for no gain,
so the mapping is stated in the section instead. No external doc referenced either
number (checked: the other `D9`s in the repo belong to the imagery and about-page
registers). **Lesson for a decision register in a shared doc: the number is not
reserved by intending to use it, only by committing it — and `git log --oneline -8`
before claiming one is the whole check.**

## 2026-07-27 (late, cont.) — the seat's own prompt contains a false promise

Went to ground §6 item 5's premise rather than repeat it from earlier docs.
Live, 2026-07-27:

```sql
SELECT count(*), count(*) FILTER (WHERE path ILIKE '%.md'), count(DISTINCT path),
       count(*) FILTER (WHERE COALESCE(content,'')<>''), max(updated_at)
FROM code_symbols;
-- 4535 | 0 markdown | 530 files | 4535 with non-empty content | 13:36:50
```

Markdown invisibility **confirmed**: 4,535 symbols, 100% Go, zero markdown. So
the premise holds.

**MISSTEP, and a good example of a column name doing the lying.** I read
`with_body = 4535` as "bodies are indexed", which would have *contradicted*
`bugs_open/108`'s central claim. It does not. `content` holds **declarations
only**:

```
kind      count  avg_len  max_len
func      2744   198      451
method     947   203      413
struct     789   108      181
```

`max_len` 451 for a whole package's longest function is not a body, and the
longest `func` row is literally its signature line. So 108 is right and I was
one query away from writing the opposite into a handoff. Same shape as the
gauntlet thread's `js_content IS NULL` ≠ "no JS": **a non-empty column is not
evidence the column holds what its name suggests.** The check that settled it was
`max(length(content))` grouped by kind — one line.

**The finding that came out of it, contributed INTO `bugs_open/108` rather than
filed separately** (it is unowned; `who-owns.py` returns no owning workstream,
its only recent commits are a renumber and a concept-register sweep):
`review_architecture`'s prompt — the one I shipped an hour earlier — ends with
*'kind "symbol" matches symbol names, **"content" searches source bodies**, "ls"
lists indexed paths.'* Given 108 that clause is **false**. The seat will issue
`content` checks for precisely what a forward-fitness call turns on (a route, a
registry key, whether anything still references a symbol) and receive a zero it
cannot distinguish from a real absence. Worse, the same prompt carries the right
instinct on the SQL tier — *"Treat an empty result as 'no precedent found', NOT as
'this is novel'"* — and **no equivalent warning on the code tier**, which is the
one that is actually broken.

**MISSTEP, and the most consequential of the session: the headline metric was
measuring something other than what its own sentence said.** I reported "6 of 90
stability objections cited precedent" and called it "the number to beat" — to the
owner, in the handoff, in memory, and as the baseline of a reversal trigger in the
decisions doc. In `scripts/council-adoption-report.sh` §2, `invoked_stability` and
`cited_precedent` were two **independent** `FILTER`s over the same 210 reviews.
They never intersected, so "6 of 90" was never "6 *of the* 90": **4 of the 6 cited
precedent without invoking the preference at all**, and the true intersection is
**2 of 90 (2.2%)**.

Caught by the DECISIONS doc saying `3 of 87` while the handoff said `6 of 90`,
**both labelled pre-change** — a population that cannot grow. I had written both
numbers without the contradiction registering. The cheap check is
`count(*) FILTER (WHERE stab AND prec)`: **a prose claim of the form "X while Y"
must be one predicate `X AND Y`; two counters printed side by side acquire an
implied denominator from adjacency alone.** Script fixed (headline is now
`both_invoked_and_cited` + `pct_of_invoked`, with `cited_but_did_not_invoke` kept
visible), all four sites corrected, full entry in `WRONG_CALLS.md`.

Two things worth carrying. First, **the correction strengthens the case it was
evidence for** — the seat used its history *less* than I claimed. Corrected
reading: **before 2 of 90 (2.2%), after 1 of 2 (50%)**, n=2 so no conclusion.
Second, and worse: I had *already* found that this metric **under**-counts (it
scored the 14:18 review 0 while I was quoting that same review as the best
evidence the change worked) and I still repeated the **over**-counted headline in
the same message. `%deflect%` matches the bare word, and the new prompt itself says
"deflected upward" — so a seat echoing its instructions scores a citation.
**Having found a metric wrong in one direction, check the other before quoting it
again.**

**Deliberately NOT fixing that in the prompt.** Three other consumers
(`review_prior_art`, `review_reuse_agent`, the diagnosis loop's
`lookup_code_symbols`) carry the same false promise, so a prompt-only fix would
leave them lying and would need redoing when 108 lands. 108's candidate 2 (index
bodies from the `[line_start, line_end]` span already on every row) fixes all
four at once — and 108 already records that the same candidate *"answers the
schema half of the architecture thread's D8b"*. **So item 5 is mostly decided
already and I had been about to design it from scratch.** What is genuinely left
for this workstream is the **ranking** — reuse the concept register's
rediscovery-frequency signal — not the plumbing. Grepping `/bugs_open/` for the
mechanism before designing is what caught that, which is the rule working exactly
as CLAUDE.md says it should.

---

## 2026-07-27, ~21:00 — layer 1 BUILT and committed (`37f7deff9`). Three defects in the approved plan, one of them nearly fatal

Migrations **243** (`code_symbols.body` + trigram index, applied and green) and
**244** (the doc_notes trail) are live. The Go is committed and rides the chassis
build another session was starting as I finished — so it is **inert until that
roll**, and the column is deliberately NULL until then.

### The plan was APPROVED and still had three concrete defects. All three were advisory objections that turned out to be right when actually run.

**1. The migration number moved TWICE.** Approved as 241. `guardian` objected
(low) that "241 is the next free number" was asserted rather than checkable from
SQL. By the time I opened the directory, another session had committed 241 and a
third had 242 staged. It is **243**. The handoff I started from already said
"use 242" — and *that* was stale within four hours of being written. This
sequence is shared state; re-read it immediately before writing the file, never
from a doc.

**2. The VERIFY hash expression as approved would have ERRORED.**
`sha256(content::bytea)` fails with *"invalid input syntax for type bytea"* —
the cast **parses** the text as a bytea literal, it does not encode it.
`convert_to(content,'UTF8')` is the encode. `editquality` objected (medium) that
the formula was asserted rather than verified against the live schema. Correct,
and the concrete defect was a cast. The corrected expression measures **0 drift
across all 4,535 rows**, before and after.

**3. THE ONE THAT NEARLY SANK IT — "the indexer is already walking the file, so
this needs no new file I/O pass and no re-parse" is FALSE.** `flattenSymbols`
walks a **JSON-decoded `analysis.Output`** that carries line spans and *no source
text*. There is no file being walked. Bodies have to come from somewhere, and the
plan never said where.

It works **only** because the LIVE `code-indexer` workflow's first step is
`analyse_repo_local`, which fetches the tarball into the indexer pod's own temp
dir and *deliberately does not clean it up*, so `out.Root` is a real local path.
Under the wiring in **seed `118_code_indexer_for_analyser.sql` — which is what
the repo still shows** — the analyser adapter parses in its OWN pod and returns
spans whose root does not exist in the indexer's. Every read would have failed,
every body would have been NULL, and the change would have looked shipped while
being completely inert.

> **The seed file is stale and the live row is the truth.** I read
> `agent_definitions` rather than the seed, and that is the only reason this was
> caught before the build rather than after a green deploy with an all-NULL
> column. A thread that had trusted the repo's own SQL would have concluded
> either "impossible" or, worse, "done".

So `flattenSymbols` now logs `root`, `files_read`, `file_read_errors`,
`bodies_sliced`, and warns **loudly** when `bodies_sliced == 0 && symbols > 0`,
naming that exact cause. `with_body` is returned in the action result next to
`upserted`, so the failure is legible from the orchestration record without
anyone thinking to query the column.

### Two findings that were nobody's objection

**`content_hash` does NOT change when only a function's body changes.** The hash
covers `composeSymbolContent` output — kind + symbol + signature + doc + path.
A function whose body was rewritten while its signature stayed put has a
**byte-identical hash**. Consequences:

- The embedding-skip is still correct (embeddings are of `content`, which
  genuinely did not change).
- But there is **no cheap test for "this stored body is still current"**, which
  kills the tempting `body = COALESCE(EXCLUDED.body, code_symbols.body)` — the
  "safe" form that preserves a body when a slice fails. It would preserve text
  that no longer matches the `line_start`/`line_end` written from `EXCLUDED` in
  the same statement. **Plain assignment; NULL is the honest state.** A missing
  body is visible in the coverage count; a wrong one is visible only to whoever
  acts on it.

**There are TWO call sites answering code checks, not one.**
`diagnose_load_runtime` (the diagnosis lane) calls the same `answerCodeCheck`
helper as `diagnose_code_lookup` (the council's verify tier). The compiler found
it, not me — I had changed the signature. Both now carry the scope note, which
matters more in the diagnosis lane, because the verdict prompt's cite-or-abstain
acts **directly** on absence. Had the helper been duplicated rather than shared,
the sibling would have kept the old silent-empty rendering and nothing would have
said so — 016b §9's "one call site gets the rigorous fix, the sibling stays
heuristic", avoided by luck of a shared signature.

### Smaller things worth not re-deriving

- **`EXPLAIN` on the OR predicate currently shows a Seq Scan** (cost 547).
  That is `guardian`'s low objection confirmed — but with `body` entirely NULL
  the measurement **cannot decide anything**, so the choice between the OR and a
  single expression index on `(COALESCE(body,'') || ' ' || content)` is
  **deferred until after the reindex**, when there is real text to plan over.
  Check 5 of the VERIFY file runs it.
- **The migration runner has no single-file apply.** `--apply` applies *every*
  pending file, which in this tree means other sessions' work. The escape hatch
  is `MIGRATIONS_DIR=<dir with just your file>` — same ledger, same probe, same
  recording. In RUNBOOK now.
- **244's guard was verified by inducing the fault**, not by reading it: running
  the file a second time raises `P0001` and inserts nothing (`count` stays 4).
- **Three `pattern-check` advisories on the commit, all benign but check them
  yourself:** two `logged-model-output` hits are a heuristic matching a local
  variable named `text` near an `Fprintf` — one is my answer-rendering (source
  code from the index, not model output), one is pre-existing code in a file my
  pathspec included. The `unpaired-change` hit wants `codeIndexFreshness`
  alongside a `FROM code_symbols` read: both renderers *do* still call it on
  unchanged lines, and I added a second provenance line beside it. **I should
  have said so in the commit message, which is what that check asks for.**

### Pre-measured, so the post-roll numbers have something to be checked against

Ran the **exact** path `index_code_symbols` will take — `AnalyseWithExclude(root,
["docs/"])` then `SliceLines` per symbol — over this working tree, as a scratch
test inside `internal/analysis` (the package is `internal/`, so an external
`go run` is refused; the file was deleted in the same command).

```
files=594  symbols=4917  sliced=4917  file_read_errors=0  slice_errors=0
body bytes: min=13  median=457  p95=4184  max=26215  total=5.35MB
biggest: platform/orchestration/actions/ai_actions.go :: ExecuteLLMPromptAction
```

**Zero slice errors across 4,917 symbols** — the analyser's spans and
`SliceLines`' inclusive 1-indexed convention agree on real code, which is the one
thing that could have made the whole change produce garbage rather than nothing.

**The column costs ~5.35 MB.** That settles the size question I had deliberately
refused to assert: the trigram index over it is the larger cost, and both are
unremarkable. The biggest single body is 26 KB, and the 400-char excerpt cap plus
the 40-row cap bound what can reach a prompt regardless of that.

`4,917` local vs `4,535` indexed is commit drift, not a discrepancy — the index
mirrors the last **pushed** ref (last refreshed 13:36 UTC), this is the working
tree. **Expect the post-reindex count to be neither number**, and do not read a
difference as a fault.

---

## 2026-07-28, 07:00–07:15 — LAYER 1 IS LIVE AND PROVEN. And proving it caught two things I did not expect.

Chassis rolled to **`v1.0.1180`** overnight (22:06 BST). Verified against the
RUNNING pod, not git and not the tag: the two strings the change creates return 1
each, the string it **deletes** (`"no content matches in the index"`) returns **0**,
the never-existed control returns 0, and a positive control returns 14 — so the
grep itself is known to work.

**Reindex fired manually** (corr `86815023`, orch `41955944`), `index-orchestrator`
→ `code-indexer`, `ref=086_experience_loop`. Publish→start was **~2 minutes**, and
my first check at ~40s found no row — *I nearly read queue latency as a lost
dispatch again*, which is the trap already written down twice. The `PUBLISH_OK`
marker in the kcat command is what made it safe to just wait: it proved the message
was sent, so absence could only be latency.

**The indexer's own log line, which is the whole result:**
```
root=/tmp/analyser-src-451921075  files_read=543  file_read_errors=0
bodies_sliced=4536  slice_errors=0  symbols=4536
```
`root` is a real local path — the `analyse_repo_local` dependency the approved plan
never mentioned, confirmed in production.

**VERIFY, post-reindex:** bodies **4,535 of 4,535** (100%, 0 empty strings),
`content_hash` drift **0** (nothing silently re-embedded), and
`body ILIKE '%stop_reason%'` **0 → 6** — the example the contract has promised
since it was written and which could never match.

### Misstep 1: MY OWN `COALESCE` disqualified the index I had just added

`guardian`'s low objection asked for an `EXPLAIN` before merge, warning the OR
might use neither index. It was right, and the cause was mine:

| predicate | plan | time |
|---|---|---|
| `COALESCE(body,'') ILIKE .. OR content ILIKE ..` (as shipped) | **Seq Scan** | **125.9 ms** |
| `body ILIKE .. OR content ILIKE ..` | **BitmapOr, both trigram indexes** | **5.5 ms** |

A plain-column index cannot be matched to an expression, so wrapping the column
killed it — `idx_code_symbols_body_trgm` was **dead on the only query path that
uses it**. Row sets are identical: on a NULL body `body ILIKE x` yields NULL, and
`WHERE` discards NULL exactly as it discards false, so a not-yet-indexed row still
falls through to the content side. Fixed in `a4f06f83a`, inert until the next roll.

**And the equivalence test was nearly vacuous.** Comparing the two forms on live
rows returns 6 and 6 — proving nothing, because there are **0 NULL bodies** in the
table now. The distinguishing input cannot occur in the data. I tested the NULL
branch directly over a `VALUES` list instead. *A test whose distinguishing case is
absent from the population is not a test* — the same shape as the vacuous pod-grep,
one layer down.

### Misstep 2 — not mine, but the more important one: fixing B made A worse

At **07:07:27** a live diagnosis (`914dc844`, robot-hands 404 links) asked whether
`RepairPageLinks` exists and got:

> `answered: 0 rows — searched the names of 4535 indexed symbols. The query was RUN
> and matched none; this is not an unanswered question.`

**It exists** — `datahelpers/link_repair.go:139`. It is absent only from the
indexed snapshot, which is **955 commits behind** (`git cat-file -e e19aa5d:…`
→ *"exists on disk, but not in e19aa5d"*). The banner said *"refreshed 17h ago"*.

The reindex I ran did **not** move the distance, because the indexer mirrors the
last **pushed** tip and `origin/086_experience_loop` has sat at `e19aa5d10`
throughout. That is `bugs_open/108` defect A confirmed on live traffic.

**The uncomfortable part is that my own change sharpened the knife.** I replaced
`"(no symbol matches in the index)"` with `"the query was RUN and matched none"` —
correct for the empty-vs-unanswered confusion, and a **stronger denial** when the
corpus is stale. Two fixes that each increase honesty combined to increase harm,
because one raised confidence in a signal the other had not yet made correct.
Contributed to `108` in full, with the priority inversion it implies: candidate
1/4 (freshness by commit distance) is now a **prerequisite** for 2/3 being safe,
not a parallel nicety.

---

## 2026-07-28, 10:00 — the COALESCE fix is LIVE (`v1.0.1182`). D11 layer 1 is COMPLETE.

Pod-grep flipped exactly as pre-validated: `WHERE (body ILIKE` **0 → 1**,
`WHERE (COALESCE(body` **1 → 0**, layer-1 marker still 1, never-existed control 0.
Pre-validating the marker against the OLD image (it returned 0/1 before the build
existed) is what made the flip meaningful rather than a bare presence check.

**The plan the index was added for, now running:**
```
Bitmap Heap Scan
  -> BitmapOr
       -> Bitmap Index Scan on idx_code_symbols_body_trgm
       -> Bitmap Index Scan on idx_code_symbols_content_trgm
```
Full VERIFY green: hash drift 0, contamination 0, **4,535/4,535 bodies**, 0 empty
strings, `stop_reason` 6, negative control 0.

### The VERIFY file itself had become a trap, and I nearly left it

Check 5 still `EXPLAIN`ed the **COALESCE** form — the predicate that is no longer
shipped. Anyone running the file after this roll would have seen a Seq Scan and
concluded the fix had failed, on a check that looked like coverage. Corrected, with
the reason and the measurements inline, plus the standing instruction to **keep it
in step with `answerCodeCheck`'s `case "content"`**.

The general shape, worth more than the instance: **a verification script pinned to
the OLD implementation reports on a query nobody runs, while presenting as
coverage.** When a fix changes the shape of the thing being asserted, the assertion
is part of the change — not a fixture that survives it. This is the same family as
the vacuous pod-grep and the vacuous NULL-body test, both hit earlier in this same
piece of work: three ways in one day for a check to pass without checking.

---

## 2026-07-28 ~11:30 UTC — §3 item 1 (108 candidate 1) DONE, by the bugs-sweep thread, under this workstream's design

Freshness by commit identity is live: `ref` + `commit_time` stored at index time
(migration 250), verdict keyed on commit age, loud-UNKNOWN (never FRESH) on NULL
commit_time. Council `b5285973` APPROVED round 1; commit `87d0bcf97`; live v1.0.1184
(pod-grep-verified — the 1183 this thread built was superseded 38s later by a
concurrent roll whose image carries the same commit). Full account in
`bugs_closed/108` (now CLOSED — both defects fixed and live).

Notes for this workstream specifically:
- **The branch was PUSHED during the session** (origin at `d98010e8b`, 19 commits
  behind local HEAD, was 1,003). §1's "one open ask" is resolved for now; the banner
  will say honestly if it drifts again — that is the point of the fix.
- The enabling fact in §3 item 1 ("the indexer receives an explicit ref but does not
  store it") was confirmed AND sharpened: the cadence's ref arrives via
  `scheduled_tasks.pre_query`, not `input_data` — the bug file's "no ref parameter"
  correction was itself superseded. See the closing entry in `bugs_closed/108`.
- Your §3 item 4 (markdown indexing) is now unblocked in the ordering sense: the
  presentation-confidence prerequisite from §4 is satisfied.
- The reindex lane fails bursty at the spawn handshake (2 of 3 dispatches on
  v1.0.1184) — child-side mechanism now evidenced in `bugs_open/129`, diagnosis
  `dcde1ed9` filed. Relevant if you dispatch reindexes for the markdown work.

---

## 2026-07-28, ~11:45 — pinning the indexer's ref, and reading PART of a row again

Owner directive: index the working branch (`086_experience_loop`), because the
merge to main is not happening soon. Two migrations, and the first one was based
on a premise I had not checked.

### The push already fixed the staleness — before I changed anything

`ref='086_experience_loop'`, `commit_sha=d98010e8b`, **0 commits behind the branch
tip**, 4,992 rows, **4,992 bodies**. So the 955-commit gap I had been raising is
closed, and it closed because the branch was pushed — exactly as predicted, and
not by anything I did today. Another session has also shipped `ref` and
`commit_time` columns (108 candidate 1), so the banner can now key its verdict to
the indexed commit's own date rather than the row clock.

### 251 IS INERT, and the reason is a repeat of this session's own lesson

I wrote 251 to add `{"ref": "086_experience_loop"}` to the task's `input_data`,
on the premise that the task supplied no ref and therefore fell back to `HEAD` →
the default branch → main. **I had read the row and not the whole row.** The task
carries a `pre_query`, and the scheduler treats it as authoritative:

- `cmd/scheduler/main.go:216` — `mergeJSON(task.InputData, dynamicData)`, and
  `:480` `baseMap[k] = v`, so the pre-query result is the **overlay** and
  overwrites any static key of the same name;
- `cmd/scheduler/main.go:198-212` — a pre-query returning **no rows** does not
  fall back to `input_data`; it stamps the task completed and **`continue`s**.
  **The task does not fire at all.**

So the static key can never be read, in either branch. That is the third time in
two days I have acted on a partial read of a live row — and the specific shape,
*"the config that produces the right answer for a reason nobody can state"*, is
one I wrote into 251's own header while missing that it applied to me.

### What the old pre_query actually did, and why a constant replaced it

```sql
SELECT collected_data->'input_data'->>'ref' FROM orchestration_states
 WHERE collected_data->'input_data'->>'ref' ~ '^[0-9]{3}_'
   AND COALESCE(owner_agent_type,'') NOT IN ('index-orchestrator','code-indexer')
   AND created_at > now() - interval '14 days' ORDER BY created_at DESC LIMIT 1
```

It infers the branch from whatever ref the most recent non-indexer orchestration
carried — clever, and it is why the index tracked the branch at all. Two failure
modes make it wrong for a directive that *names* a branch:

1. **No rows ⇒ the index silently stops refreshing.** Nothing errors. It needs
   only 14 quiet days, or a spell with no `NNN_`-shaped ref in flight.
2. **It follows whichever branch another session last mentioned**, decided by
   `ORDER BY created_at DESC`, with nothing recording that the corpus changed
   identity.

Both bite at the merge: `main` does not match `^[0-9]{3}_`, so **the day this
branch merges, the pre-query goes dry and the refresh stops quietly** — 108
defect A re-armed by a stale pin instead of a stale clock. 252 replaces it with
`SELECT '086_experience_loop'::text AS ref`: one row by construction, so the task
always fires, and one literal to change at merge time. The reversal trigger is
written into the migration where the edit has to happen, not into a doc.

252 verifies itself by **executing** the stored pre_query (`EXECUTE 'SELECT
count(*), max(ref) FROM (' || v_pre || ') s'`) rather than asserting its text — a
pre_query is executable code with no compiler, and this fleet has been bitten by
re-stating a predicate instead of running the stored one.

---

## 2026-07-28, ~15:15 — what's next, determined by re-grounding rather than by reading my own handoff

**108 is CLOSED** (`6928c9380`, both defects, moved to `bugs_closed/`). Verified
defect A independently against the RUNNING `v1.0.1189` pod rather than trusting
the file: the commit-time-keyed banner is in the binary and **the old
`updated_at`-only text returns 0**. So the prerequisite my own handoff §4 called
non-negotiable is genuinely done, and the ordering constraint it imposed is
discharged.

**State of the other open items, re-grounded:**

| item | state |
|---|---|
| seat has 0 reviews | **unchanged** — of 6 `capability_gap` specs, only `7b89fb35` has BOTH `owner_approval` and `code_pointers` in a workable status, and it is another thread's. Still a rate limit, still not forceable |
| markdown in the index | **0 of 4,992 rows**, and `code_symbols_kind_check` still permits Go kinds only |
| `council-gate` verdict note | still carries no code results |
| D11 layer 3 | still unscoped |

⇒ **Markdown is next**, and it is the only remaining item that serves the owner's
D11 directive directly. Seats can now read what the code DOES; they cannot read a
word of the record of how we get things wrong — which is the corpus this
workstream has twice been caught by.

### The sizing changed the shape of the job, so it was measured before designing

| corpus | measured 2026-07-28 |
|---|---|
| `bugs_open` + `bugs_closed` | 2,816,294 bytes, 1,407 `## ` headings |
| `WRONG_CALLS.md` | 595,590 bytes, 212 headings / **211 unique** |
| `016b` | 508,318 bytes, 130 headings, all unique |
| all `docs024` `.md` | **1,415 files** |

**~1,749 sections against 4,992 existing rows (+35%)** — feasible. Indexing all of
`docs024` is not, and would be actively harmful: those files are handoffs and
summaries **superseded by design**, and a seat citing a stale handoff as evidence
is the precise failure this workstream exists to prevent. Owner scoped it to the
four durable sources. **The globs are the whole design.**

The in-file duplicate-heading collision against `uq_code_symbols_identity`
(repo, path, symbol) is **real but rare — measured at 1**, so it needs a
disambiguator, not a redesign. Measuring it turned "might be a problem" into a
one-line decision.

### The precondition was checked FIRST this time

The last plan this workstream submitted was approved by twelve seats while
asserting *"the indexer is already walking the file"* — it was not. So before
writing a line: `FetchToDir` (`analyse_repo_local_action.go:165`) fetches the
**whole tarball unfiltered**; `AnalyseWithExclude` (`:197`) filters only the
**analysis**; `defaultAnalyseExcludePatterns = []string{"docs/"}` (`:285`) is
therefore an analyser-time exclude. **`docs/` IS on disk in the indexer pod.**
Had that been a fetch-time exclude, the plan would have been impossible and
nothing in it would have said so.

**SUBMITTED to the council gate: `SUBMISSION_CORR=7ba5b8c4-0e10-46db-9fc4-2bd0584e943a`**
(6 edits, 7 grounded_in). Submission committed unedited at
`SUBMISSION_2026-07-28_markdown_into_the_index.json` so the artifact keeps
matching whatever verdict comes back. Dispatched only after waiting for the
newest chassis pod to clear 7 minutes — two pods were mid-roll and a council
dispatch inside that window is eaten silently.

---

## 2026-07-28, 19:30 — round 3 was WEDGED for 2h34m, and the peer check is what settled it

Round 3 (`6872857b-cb51`, fired 16:54:53) sat at `review_editquality` with
**status EXECUTING_STEP and its last state change one second after creation**.
Not slow — dead.

**What made that diagnosable rather than a guess, in order of decisiveness:**

1. **The peer check.** Six council-gate runs created AFTER mine all COMPLETED —
   17:01, 17:08, 18:51, 18:52, **19:04 approved**, 19:10. A submission that
   started later and finished, while mine had not moved, is the documented exit
   condition for "this is not a queue". Without that, 2h34m of silence is
   indistinguishable from the ~30-minute queue latency this lane normally has.
2. **Two stalls, not one, and on DIFFERENT seats.** Fleet-wide exactly two
   orchestrations were stuck >30 min: mine at `review_editquality` (16:54:54) and
   `b8ff4094-05af` at `review_diagnosis_guardian` (16:55:46). **Fifty-two seconds
   apart, different seats, different submissions.** That rules out "the editquality
   seat is broken" — which is where I would have gone, because editquality had
   *also* returned `unreadable` in round 2 and the coincidence was inviting.
3. **A roll happened AFTER, and did not rescue them.** The chassis rolled to
   `v1.0.1192` at 18:22-18:23. The lane resumed immediately (four completions
   after it) but **neither orphan recovered** — both are still EXECUTING_STEP.

> **The tempting wrong conclusion was right there and two seats' worth of evidence
> killed it.** editquality went unreadable in round 2 and hung in round 3; the
> obvious story is "that seat is faulty". The peer at a *different* seat, wedged
> 52 seconds later, says the cause is whatever happened at ~16:55, not the seat.
> [[check-an-untouched-peer-in-the-same-batch]] applied to orchestrations rather
> than to a diff.

**Still UNEXPLAINED and worth someone's attention:** what froze two council runs
within a minute at ~16:55, and why a subsequent pod roll did not recover them.
Orphan recovery was supposed to be closed (CS-2d, "orphaned work under expired
claims was unrecoverable"; 029's post-roll response replay). These two are
counter-evidence, or they are a class those fixes do not cover. **I have NOT
diagnosed it** — recording the observation with its evidence, not a mechanism.
`[UNDIAGNOSED]`.

Re-fired round 3 under the same `RESUBMIT_CORR`, pods 67 minutes old, LAG 0.

---

## 2026-07-28, 21:30 — round 7 came back 8/2, and every objection was a request for a MEASUREMENT

Round 7 (the scope split) is the best round of the eight: **8 approve / 2 object**,
1 seat unreadable, 5 abstained of 10 reviewers. Verdict still `revise`.

```sql
SELECT created_at, metadata->>'decision', metadata->>'unreadable'
FROM diagnosis_artifacts WHERE correlation_id='7ba5b8c4-0e10-46db-9fc4-2bd0584e943a'
  AND kind='council_report' ORDER BY created_at;
-- 15:19 revise/0 · 16:49 revise/1 · 19:40 revise/1 · 19:54 revise/1
-- 20:08 revise/0 · 20:20 revise/0 · 21:09 revise/1   <- round 7
```

**The shape changed and that is the finding.** Rounds 1–6 drew objections about
what the plan *does*. Round 7 drew six objections (2 medium, 4 low, no HIGH) and
**not one says the plan is wrong.** All six say the same thing in six voices:
*you asserted this; go measure it and cite it.*

| seat | sev | what it actually wanted |
|---|---|---|
| reuse_agent | med | is some RAG/document pipeline already ingesting these globs? |
| guardian | low | name the pipelines that touch `code_symbols`; I can't re-verify a grep |
| guardian | low | confirm DB-side that ONE agent_definition wires `IndexCodeSymbolsAction` |
| debug_historian | med | the plan has NO pod-binary verification; every VERIFY assertion is DB-only |
| prior_art_librarian | low | the owning-pipeline claim has no `grounded_in` line, unlike every other |
| prior_art_librarian | low | the `site_work_items` negative isn't cited here (explicitly informational) |

**So round 8 adds NO mechanism** — no code, no globs, no schema, no helper. The
six edits are identical in substance; eight query results go into `grounded_in`
and one missing verification step goes into the VERIFY file. This is the §4
lesson applied in the other direction: rounds 4–6 died because each objection was
answered by *adding machinery*, and each new machine became fresh surface. **An
objection asking for evidence must be answered with evidence, never with code.**
If round 8 does not converge, the conclusion is about the plan, not an accretion.

### What the measurements said

**1. reuse_agent — no, and the answer is stronger than absence.** `rag_index`
takes its text from the **orchestration payload**, not from disk
(`rag_actions.go:138-139`: `content := ExtractNestedFieldString(params.CollectedData, contentField)`).
No `filepath.Glob`, no `Walk`, no root anywhere in the action — it *cannot* reach
a repo file unless an earlier step already read it in. Empirically the whole RAG
surface is **34 rows in 2 collections, zero of repo-markdown provenance**, against
4,992 `code_symbols` rows. And the decisive half:

```sql
-- rag_lookup is wired to ONE agent, and it is the test agent
--   rag_index  -> rag-test-agent (index_content), tool-generator (index_plan)
--   rag_lookup -> rag-test-agent (lookup_query)   <- and nothing else
```

**No diagnostic or council seat can reach `knowledge_base` at all.** The readers
this workstream serves — `diagnose-agent`, `feature-designer`, `fix-proposer` —
reach `code_symbols` and nothing else. Putting markdown in the RAG surface would
not make it findable by them. It is not a second path to the same place.

**2. guardian — the reader census, and A CORRECTION I OWE.** Seven live read
sites, none selecting `kind`; the column reaches Go only at
`code_symbols_actions.go:528` and `:549`, both inside `lookup_code_symbols`, both
handing it to `scanCodeSymbolRows:566-585`, which writes it to `r["kind"]` and
never compares it.

> **CORRECTED 2026-07-28:** rounds 6 and 7 both stated the branch-on-`kind` grep
> returned **"one hit"**. It returns **four** — `diagnose_code_lookup_action.go:246`
> (`switch kind { case "content" … }`), `:415` (`validCodeCheckKind`), `:550`
> (`unrecognised kind %q`), and `code_symbols_actions.go:430` (`kind := td.Kind`).
> **None is the column**: the first three are the *code_check* kind namespace
> (`symbol|content|ls`, enumerated at `:414-420`), the fourth is the analyser's
> `typeDef.Kind` on the write path. **The conclusion held; the evidence I gave for
> it was undercounted.** What caught it: guardian refusing to take a human grep on
> trust and asking which pipelines actually touch the table. The cheap check that
> would have: running the grep with the alternation actually written out
> (`switch|case |== "|!= "|["kind"]|.Kind`) instead of eyeballing for the column
> name. Logged in `WRONG_CALLS.md`.

Worth conceding rather than papering over: guardian is **right** that its own tier
cannot re-verify this — `code_checks` see declarations only, never switch bodies.
Round 8 states it as a human grep. This plan's markdown corpus does not close that
gap either.

**3. guardian/prior_art — ownership, DB-side.** `index_code_symbols` is wired by
**exactly one** `agent_definition`: `code-indexer/index_symbols`. Four agents
touch the table; one writes it. `scheduled_tasks` holds
`code-index-refresh | index-orchestrator | 86400 | enabled | last fired 12:46:27Z`.

**4. debug_historian — a real gap, and the sharper form of it.** The objection was
"no pod-binary verification". The sharper statement, which is why it matters:
**every existing VERIFY assertion would read identically against a pod that never
received the code** — `kind='doc'` rows would simply be absent, which is
indistinguishable from "the roll hasn't happened yet". VERIFY gains a step 0, and
it is discriminating rather than decorative because the baseline is **measured**:

```
pod agent-chassis-74dbd9c9f4-7p6d8, strings /app/agent-chassis | grep -c
  flattenMarkdown 0 · composeDocContent 0 · defaultDocGlobs 0   <- the change
  composeSymbolContent 2 · scanCodeSymbolRows 2                 <- POSITIVE CONTROL
```

The positive control is the part that earns it: without it, a `0` cannot be told
apart from a broken probe against a stripped or renamed binary.

**5. Also fixed, unprompted.** `grounded_in`'s last entry still carried the stale
**~1,749** sizing *after the rationale had already corrected it to ~3,472*. A
submission that corrects a number in prose while leaving it in the evidence is
precisely the drift this workstream exists to review for. Fixed at source.

**Round 8 dispatched** on `RESUBMIT_CORR=7ba5b8c4-…`, committed first at
`SUBMISSION_2026-07-28b_markdown_into_the_index.json` (a NEW file — round 7's text
stays on disk so the round-7→8 diff is auditable). Two runs ahead of it in the
lane at dispatch.

---

## 2026-07-28, 21:37 — round 8: 8/1, the measurements worked, and the last seat found something real

**Round 8 = 8 approve / 1 object** (9 reviewers, 6 abstained, 1 unreadable).
Verdict still `revise`, on a single `medium`.

**The measurements did exactly what they were supposed to.** Every seat that
objected in round 7 approved in round 8 **with zero objections**:

| seat | round 7 | round 8 |
|---|---|---|
| `guardian` | object ×2 | **approve, 0** |
| `reuse_agent` | approve, 1 objection | **approve, 0** |
| `prior_art_librarian` | approve, 2 objections | **approve, 0** |
| `debug_historian` | object ×1 | **approve**, 1 low (the `ROLLBACK;` preamble) |
| `diagnosis_guardian` | approve, 0 | **object ×1, medium** ← new |

So the round-8 discipline is confirmed as correct: **answering an evidence
objection with evidence, and adding no mechanism, cleared every one of them in a
single round.** Rounds 4–6 answered objections by building things and went
5/3 → 6/4 → 9/3 without converging. Round 7→8 answered with queries and went
8/2 → 8/1. That is the cleanest evidence this workstream has produced for its own
central claim.

### The remaining objection is correct, and it is MINE — it cannot be split out

`diagnosis_guardian`, medium: widening `kind` to admit `'doc'` means a markdown
excerpt can satisfy the **static tier** of a CONFIRMED verdict — the tier that is
supposed to mean *code was read*.

**Verified against live code, five sites** (full write-up + costed options in
`DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` § D12):

- `pkg/diagnose/loop.go:60` — `TierStatic Tier = iota // code_symbols + \d (T1)`.
  The static tier is **defined as** this table.
- `pkg/diagnose/step.go:145-156` — `tierCovered` = `static && observed`; a confirm
  without both families degrades (`step.go:79`).
- `pkg/diagnose/loop.go:66-73` — an unknown tier **defaults to `static`**, asserted
  on purpose in `verdict_wire_test.go:48`. It fails open toward code.
- `diagnose_assemble_bundle_action.go:266-270` — the bundle emits, unconditionally,
  `## Code search results (from the code_symbols index — STATIC tier)` and then the
  literal sentence **"They are CODE"**. The comment above says why the heading is
  the guard: *"the model reads the heading and not this comment."*
- `diagnose_assemble_bundle_action.go:239` — an existing comment asserting the very
  invariant this change breaks: *"the static tier cannot reach it (`code_symbols`
  indexes .go only)"*.

> **The irony is the finding.** Round 8's own new evidence is what made this
> objection possible. I proved *"no reader switches on `kind`"* as a **blast-radius
> reassurance**; `diagnosis_guardian` read the same fact as a **citation-integrity**
> problem — nothing distinguishes a doc row from a `func` row at read time, which
> is exactly why it is safe to insert and exactly why it is dangerous to cite.
> **The same measurement answers one question and raises another.** A seat that
> re-frames your evidence rather than disputing it is doing the most valuable
> thing a council does.

**Why I am NOT splitting this out like the prune.** `bugs_open/135` was legitimate
to split because the prune has no floor **today**, with or without markdown —
pre-existing and independent. This hazard **does not exist until layer 1b ships**.
Filing a bug for a defect your own unshipped change creates, then shipping it, is
not a split; it is shipping a known defect with a paper trail.

**And I am NOT answering it in round 9.** The fix that actually closes the door
(1b in D12: doc hits get their own bundle heading so they are never rendered under
"They are CODE") touches `diagnose_assemble_bundle_action.go` — **the diagnosis
loop's citation contract, blast radius every diagnosis run**. That is a platform
seam, the owner ruling of 2026-07-28 makes it architecture-scope, and the
ordering-constraint exemption does not apply: nothing is broken today
(`code_symbols` holds 0 markdown rows and the CHECK still refuses them), so there
is no reason it must ship ahead of its review. Folding it into round 9 would repeat
rounds 4–6 *exactly*, except the accreted mechanism would be the citation contract.

**Layer 1b is therefore PAUSED pending owner decision D12**, not revised. This is
the first time in eight rounds the blocker has not been something I could fix by
writing better.

## 2026-07-28 evening — D9 REVERSED by the owner; the seat is live on the fix lane and the gate (thread "bugsearch 5")

**Written by another thread, contributing rather than competing — this lane is
yours.** I arrived from `bugs_closed/129`, whose guardian SCOPE veto told it to
"route the seam to architecture review", and found the venue did not exist on
either lane a fix can arrive on. The owner reversed D9 on the spot.

Full record, evidence, diff table and caveats: **`DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` §D9.**
Short version:

- **D9's own reversal trigger fired, and had never been run.** It was recorded
  `[UNMEASURED]` in that file. Run properly: **13 affirmative escalations across 10
  distinct submissions** (threshold 3).
- **The trigger query as written there is VACUOUS** and I nearly reported its
  output as fact. `body ILIKE '%rchitecture-level concern%'` gives 68/50 because
  **the phrase is seeded in four prompts**, so 26 matches are seats saying *"No
  architecture-level concern"*. Corrected query is in §D9 and the miss is logged in
  `WRONG_CALLS.md`. Relevant to this lane generally: **any `body ILIKE` metric over
  the council corpus needs the prompts grepped for the same string first** — same
  family as the `%deflect%` false-citation trap you logged on 07-27.
- **Built by adapting your `feature-designer` seat, not rewriting it.** Verdict
  vocabulary, `ARCHITECTURE_SIGNAL` first-line contract and the code-index caveat
  paragraph are carried over verbatim; I added one paragraph for the fix-lane
  failure mode (a shared mechanism inside a bug patch, citing 124 and 129) and
  swapped the spec context for diagnosis+plan. Seated on `fix-proposer`, mirrored
  to `council-gate` with `099_SYNC_gate_roster.py` — **not hand-patched.**
- **Reachability checked with a branching BFS**, per your own landmine about linear
  walks reporting false orphans: fix-proposer 50/50, council-gate 44/44.
- **Rollback rows for BOTH lanes** are in `agent_definitions_backup` at 16 seats.
  Note for the runbook: they are in **`agent_definitions_backup`**, not
  `agent_definitions` — I checked the wrong table first and briefly concluded 099's
  "snapshot taken" was a false print. It is not; the mechanism works.

**Two things I did NOT do, deliberately:**

1. **Nothing touched layer 1b or D12.** Your markdown-into-the-index work is paused
   on an owner decision and I left it entirely alone.
2. **I did not give `council-gate` a `code_lookup`.** Your finding stands and is now
   load-bearing for the new seat too: the gate has none (verified again tonight —
   fix-proposer `t`, council-gate `f`, feature-designer `t`), so **on the gate this
   seat's `code_checks` are raised and never answered.** It is half-sighted on the
   exact lane 124 and 129 arrived on. That is your D11 territory and a real design
   question (`099_SYNC_gate_roster.py:24-29` excludes it on purpose), so I left it
   for you rather than widening the mirror on my own judgement.

**The seat still has 0 reviews.** Seating it proves nothing — §5's kill switch
applies unchanged. First real test is the next platform submission on either lane;
read the `ARCHITECTURE_SIGNAL` line rather than counting verdicts.

---

## 2026-07-29, morning — re-verified against chassis `v1.0.1196`. D12 holds; two figures moved and one is mine

A roll I did not do landed overnight: **`v1.0.1196`**, both pods restarted
22:37–22:38 on 07-28 (was `v1.0.1194` when round 8 was submitted).

**D12's evidence survives the roll, re-grepped on BOTH pods:**

```
agent-chassis-cccfc8f5d-9bb57  /  agent-chassis-cccfc8f5d-hqqbm
  flattenMarkdown 0 · composeDocContent 0 · defaultDocGlobs 0   <- layer 1b, still unbuilt
  composeSymbolContent 2 · scanCodeSymbolRows 2                 <- positive control, both pods
```

`code_symbols_kind_check` is **unchanged** and still refuses `'doc'`, and there
are **0 markdown rows**. So D12 remains a decision to take before shipping, not a
defect to repair. Layer 1 is still perfect: **5,017 rows / 5,017 bodies**, no
kind without a body.

### CORRECTION — I carried a stale figure into round 8 while fixing a different stale figure in the same file

> **CORRECTED 2026-07-29:** round 8's `grounded_in` entry 1 says *"0 markdown,
> **4992** total"*. The live count is **5,017**, and it was already 5,017 when I
> submitted: a reindex completed at **21:11:19** and I dispatched round 8 at
> **~21:30**. I carried 4,992 forward from the round-7 submission (and before that
> from the handoff) **without re-measuring it**, in the very same submission where
> I corrected the stale `~1,749` sizing *at source* and wrote that "a submission
> that corrects a number in prose while leaving it in the evidence is exactly the
> drift this workstream reviews for."
>
> **What caught it:** re-verifying the pod baseline after a roll I did not do, and
> running the `kind` census while there rather than quoting the table. **The cheap
> check that would have:** the rule already written in CLAUDE.md — *ground every
> figure against the live system before repeating it from another doc* — applied to
> the figures I inherited, not only to the ones I was already suspicious of.
> **The tell I missed: I re-measured every claim a REVIEWER challenged and none
> that no reviewer had mentioned.** Round 7's objections defined my re-measurement
> scope, and a figure nobody objected to is exactly the one that rots quietly.
> No reviewer had queried 4,992, so it was never re-run.
>
> Immaterial to the plan — the row count appears in sizing prose, and the ratio
> (~3,472 doc sections against ~5,000 code rows) is unchanged at this precision.
> Logged because the *mechanism* is not immaterial.

### The index is on a DIFFERENT ref from the one the scheduler will use, and the next fire will prune it

Three facts, each read from the live system this morning:

| | value |
|---|---|
| `code_symbols.ref` (what is actually indexed) | **`087_towards_multiple_domains`** @ `b953b8a7c` |
| `scheduled_tasks.pre_query` for `code-index-refresh` | still **`SELECT '086_experience_loop'::text AS ref`** |
| `last_triggered_at` vs the rows' `max(updated_at)` | **12:46:27** vs **21:11:19** |

The scheduled task did not write the current index — something dispatched the
indexer directly at 21:11 with a different ref. **The pin was never changed; the
content was.** So on the next scheduled fire (86400s after 12:46) the indexer runs
at `086_experience_loop`, and the unconditional prune
(`DELETE … WHERE commit_sha IS DISTINCT FROM $2`) **deletes every `087` row**. The
index does not merge refs — it mirrors one.

That is not `bugs_open/135` (which is about a *partial fetch* deleting a good
index) but it is the same missing guard reached through a different door, and it
argues 135 is worth more than it looked. **`[UNINVESTIGATED]`: I have not chased
who dispatched the 21:11 run or why, and I am not claiming it was a mistake** —
recording the divergence and its consequence, not a cause.

Also: the index is **197 commits behind local HEAD** (`b953b8a7c` 07-28 17:36 vs
`b5a7f016f` 07-29 08:17), so none of round 8's own commits are in it. Expected on
a 24h cadence, noted so nobody reads a `content` miss as an absence.

**Nothing here changes D12.** The decision stands exactly as written.
