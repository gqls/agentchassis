# NOTES — bugs_open/223, the code index's blind spots are narrated as absence

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

Lane opened **2026-08-10** by a fresh session, on the owner's instruction to take the
next unowned bug in `bugs_open/`. 223 was named as the candidate.

---

## 2026-08-10 — ownership check before touching anything

`scripts/who-owns.py 223` reports no owning lane: the four commits on the bug file are
the *filing* and three *contributions* (a correction, a third failure mode from the
`bugfix_209` lane, a recurrence from the `bugfix_201` lane). The three lanes it names
(`bugfix_201`, `bugfix_209`, `bugfix_221`) each merely **cite** it — none has code in
flight against the mechanism. The bug file itself says **OPEN, UNOWNED** and names
`architecture_review` as the owning lane *for routing*, not as a thread already on it.

`who-owns.py` reads commits, so it is blind to a session mid-fix. Checked the live
transcripts too (`grep -c` over `~/.claude/projects/*/**.jsonl` for
`landmine.verifier|bugs_open/223`): seven other sessions mention it, and every hit is
incidental — one was reading the `scripts/landmines-*` toolchain while finishing
`bugs_open/232`, one hit is a `bugs_open/` directory listing, one is a pod list from
08-08. **No session is working the mechanism.** Taking it.

## 2026-08-10 — the bug is still valid, re-measured rather than believed

The bug's central claim is a census, so it is cheap to re-run and it is the claim's
single point of failure (the file says so itself). Both halves still hold today:

```sql
SELECT count(*) FILTER (WHERE path LIKE '%.go')     AS go,      -- 5837
       count(*) FILTER (WHERE path NOT LIKE '%.go') AS non_go,  --    0
       count(*)                                     AS total    -- 5837
  FROM code_symbols;

SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 2 DESC;
-- func 3653 | method 1119 | struct 987 | alias 42 | interface 36
```

682 distinct paths, all `.go`. **No `var`, no `const`, no `type`, no `doc` row exists.**
The corpus has grown since the bug was filed (5755 → 5837) and its composition has not
moved at all, which is the point: the blindness is structural, not a lag.

**A fact the bug file does not record, and it changes the fix.** The table's own CHECK
constraint already permits both missing kinds:

```
"code_symbols_kind_check" CHECK (kind = ANY (ARRAY['func','method','struct',
                                 'interface','alias','type','var','const']))
```

So `var`/`const` are **schema-legal and simply never written**. `\d code_symbols` is how
that was found — the schema-first rule paying for itself.

## 2026-08-10 — where the blindness actually becomes a false verdict

Read the whole chain rather than the symptom. The verifier is DB config on
`agent_definitions.type='landmine-verifier'`, five steps:

`load_entry` (query_database → `doc_notes`) → `derive_checks` (an LLM turns the entry's
`footprint` line into `code_checks` of kind `symbol|ls|content`) → `run_checks`
(`diagnose_code_lookup`, `max_checks:10`) → `verify` (an LLM returns
`{status: STILL_VALID|STALE|NEEDS_HUMAN_REVIEW, rationale, body}`) → `persist_verdict`
(`append_doc_note`, categories `['landmine-verification']`).

**Nothing parses the status mechanically.** The verdict is prose in `doc_notes`, read by
humans, by the next session, and — per RFC_005 §1 — fed into council seats'
`schema_hint`. Which is exactly why a false `STALE` is expensive: it argues for deleting
a correct entry, and it argues to a machine.

The damaging sentence is in the shared lookup action, not in the agent. From
`platform/orchestration/actions/diagnose_code_lookup_action.go`, `emptyAnswer("ls")`:

> `answered: 0 rows — no indexed path has that prefix, out of N indexed symbols. The
> query was RUN; this is not an unanswered question.`

For `scripts/landmines-sync.py` that sentence is **true and misleading**: the query did
run, and it could never have matched. The action's zero-row wording was written (for
`bugs_closed/108` defect B) to stop empty answers reading as silence — it succeeded, and
in doing so it now asserts the strongest available reading of a structurally empty
answer. That is the defect: not a missing guard, an **over-confident guard**.

The same file already contains the correctly-shaped precedent, one census along —
`bodyCoverageNote()`: *"source BODIES ARE NOT INDEXED (0 of N) … read those zeros as
UNKNOWN, not absent."* The fix is to generalise that, not to invent anything.

## 2026-08-10 — the seam is shared by four consumers, measured

```sql
SELECT type FROM agent_definitions
 WHERE default_config::text LIKE '%diagnose_code_lookup%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- feature-designer | fix-proposer | landmine-verifier
```

Plus a fourth, in Go and invisible to that query:
`diagnose_load_runtime_action.go:479` calls `answerCodeCheck` directly (the runtime
lane). **So any change to the answer wording is seen by the council's fix-proposer and
feature-designer seats too** — they must be told, not merely measured (owner ruling
2026-07-29 §3).

## 2026-08-10 — the write side, and why `var` is missing

`internal/analysis/analyse.go:140` walks the AST and takes `*ast.FuncDecl` plus
`*ast.GenDecl` **only when `d.Tok == token.TYPE`**. `token.VAR` and `token.CONST` are
dropped there, `internal/analysis/types.go:Output` has no field for them, and
`code_symbols_actions.go` therefore has nothing to write. Meanwhile the READER
(`codeKindList` in `diagnose_code_lookup_action.go`) already lists `var` and `const` as
code kinds. Reader and writer disagree, and the writer is the half that is behind.

Rough volume, for sizing: `grep -rhE "^(var|const) " --include=*.go platform/ internal/
pkg/ cmd/ | wc -l` → **930** declaration lines (a lower bound on specs — a `const (…)`
block holds many).

[UNMEASURED] whether adding those kinds would trip the per-kind prune floor on its first
run — `prune_floor.go` cohorts by `kind=`, and a brand-new kind arrives with
stored=0. To be measured before any such change ships, not argued.

## 2026-08-10 — baseline on the RUNNING binary, with a negative control

Before writing anything, established what the live chassis already contains, so the
acceptance grep later has a control that could have failed. Both replicas are
`v1.0.1277`; on `agent-chassis-6dc54d77cd-lftkt`:

```
this answer is CAPPED            → 1   (bugs_open/181's row-cap notice, live)
0 rows AT THAT PATH              → 1   (bugs_closed/163's path-arm fallback, live)
source BODIES ARE NOT INDEXED    → 1   (bugs_closed/108 defect B's coverage note, live)
NOT ANSWERABLE BY THIS INDEX     → 0   ← the marker THIS lane will add
```

The last line is the point: it is 0 today, so a later 1 proves the pipeline shipped
*this* change rather than proving the grep works. The three positive lines prove the
grep works.

## 2026-08-10 — the fourth consumer gets the fix for free, and its own comment asks for it

`diagnose_load_runtime_action.go:455-500` (the diagnosis loop's runtime lane) does not
call the action — it calls `answerCodeCheck` directly, and it already renders
`bodyCoverageNote()` and `mixedCommitNote()` through the same shared helpers. Its comment
states the requirement this lane is generalising, in as many words:

> "the verdict prompt's cite-or-abstain acts on absence, so '0 rows because bodies are
> not indexed' and '0 rows because the code does not do that' must not render
> identically."

So placing the answerability statement inside `answerCodeCheck`/`emptyAnswer` reaches the
council's `fix-proposer` and `feature-designer` seats, the landmine verifier and the
diagnosis loop with one edit and no config change for three of the four. That is the
argument for the placement, and it is the file's own precedent rather than a new idea.

## 2026-08-10 — an adjacent latent trap, measured while reading the sync library

`landmines_lib.slugify` caps the slug at `s[:80]`. `doc_notes.subject_key` is `text` with
no length limit, so the cap is the script's own choice, and the slug is the identity every
row and every verdict is keyed by.

Measured rather than assumed (`L.parse` on the live file):

```
entries: 356          duplicate slugs: 0          slugs AT the 80-char cap: 333
```

**Zero collisions today, and 333 of 356 entries (94%) sit exactly at the cap.** A
disconfirming result was available — a non-zero duplicate count — and did not occur, so
this is not a bug. It is a trap: two entries whose titles agree for the first 80 slug
characters would share one `source`, the sync would overwrite one with the other, and the
verifier's `load_entry` (`WHERE source = $1 … ORDER BY created_at DESC LIMIT 1`) would
verify whichever body was written last while reporting the other's slug. Landmine, filed
separately from this bug.

**Misstep, for the record:** `landmines_lib.parse` takes a **path**, not file content — I
passed the file's text and got `OSError: File name too long` with the whole document echoed
back as the filename. Cost: one command. Not a WRONG_CALLS entry (no claim was published),
but the correct call is in the RUNBOOK now.

## 2026-08-10 — the blind spot also produces FALSE POSITIVES, which the bug file does not say

Measured the three motivating lookup shapes directly against the index, expecting three
zeros. Two came back zero. **The third did not, and it is the more dangerous case.**

```sql
SELECT count(*) FROM code_symbols WHERE path LIKE 'scripts/%';   -- 110  (!!)
SELECT count(*) FROM code_symbols
 WHERE body ILIKE '%landmine-verification%' OR content ILIKE '%landmine-verification%'; -- 0
SELECT count(*) AS rows_at_path,
       count(*) FILTER (WHERE symbol IN ('metaCommentaryPatterns','placeholderPatterns'))
  FROM code_symbols WHERE path LIKE '%validate_page_content.go';   -- 30 rows, 0 of the two vars
```

**`scripts/` resolves to 110 indexed paths** — `scripts/documentation_project/01/analyser.go`,
`scripts/goscripts/…` and friends. There are Go programs in subdirectories of `scripts/`,
so the prefix is *represented* while every `.py` and `.sh` file directly under it is
invisible. An `ls` check written at directory altitude therefore comes back with a
generous listing that reads as **confirmation that the footprint resolves**, and the
files the entry actually named are not in it. The `ls` kind presents itself as a
directory listing and is in fact a *Go-file* listing; nothing in the answer says so.

That is worse than the 0-row case the bug was filed about, because a false STALE at least
looks like an accusation a reader might check. A flattering partial confirmation looks
like diligence — which is the shape 223's own scope section warns about for mixed
footprints, arriving here through the `ls` kind rather than through a mixed entry.

**Design consequence:** the coverage statement cannot be attached only to empty answers.
An `ls` answer must carry it whether it returned 0 rows or 110.

**And the `var` case is now exactly diagnosable, mechanically:** the path is indexed (30
symbols), the named symbol is absent, and the corpus contains no `var`/`const` row at all
— so "absent from the index because this kind is unrepresentable" is distinguishable from
"absent from the code", without guessing. That is the discriminator the verifier needed
when it invented *"possibly inlined or renamed"*.

## 2026-08-10 — built, and the one mutation that survived

Design pass run through `fable`; its ranking and my departures are in
`PLAN_2026-08-10_223_index_answerability.md` §2–3. Two things it corrected in my own
notes, both worth having in the log:

- my `grep -rhE "^(var|const) "` figure of **930** is a lower bound — a grouped
  `var ( … )` block counts once however many specs it holds. Counting specs gives
  **1,173**, so phase 2 is ~+20% of the corpus. *A grep over declaration openers cannot
  see a block's members.*
- `conditional_branch` exists with string conditions and dotted paths, and
  `compareValues` handles a bool robustly — so `no_code_evidence` is a **bool**, not a
  count.

**Seven mutations, six died, one passed.** Replacing the `ls` arm's
`b.WriteString(scope.lsReachNote())` with `b.WriteString("")` — deleting the fix at its
call site — kept twelve green tests green, because every test exercised the rendering
*helpers* and nothing asserted that anything calls them. The half left unguarded was the
FALSE-POSITIVE half, the one this lane discovered itself. Full account in `WRONG_CALLS.md`,
2026-08-10; the transferable check is **"which test fails if I delete the CALL, not the
function?"** Three `sqlmock` wiring tests through `answerCodeCheck` now kill it.

**The `\d` habit paid twice in one file.** The first draft of seed 365 hand-wrote its
snapshot INSERT and was wrong in two ways at once — `agent_definitions` has no `name`
column (it is `display_name`), and `agent_definitions_type_version_key` is UNIQUE on
`(type, version)` so a same-version copy cannot be inserted at all. Both were caught by
reading the schema before running anything, and the fix was to use the estate's existing
`snapshot_agent(type, reason)` function, which also sets `previous_version_id` — something
the hand-rolled version did not.

## 2026-08-10 — the pre-roll safety claim, PROVEN behaviourally rather than read

The seed changes live config immediately; the Go half is inert until a roll. I claimed
that window is safe from reading two functions. Then I tested it, because a claim about
behaviour is not the behaviour.

Applied seed 365 (snapshot captured, in-transaction `DO`/`RAISE` verification passed), then
fired the verifier at the same entry on the **unchanged** binary `v1.0.1277`:

```
evidence_gate → {"condition_met": false, "next_step_override": "verify"}
status        → COMPLETED, verdict NEEDS_HUMAN_REVIEW (unchanged from the pre-seed run)
persisted note → has_suffix = f
```

All three as predicted: the absent field resolved to nil, compared false, and the gate took
`else_step`; `verify_unverifiable` did not run although it exists in config; and
`note_body_suffix_field` was ignored rather than breaking the note. **It could have failed
in three ways and did not** — the gate could have errored on an unresolvable field, it could
have routed to the new branch, or the unknown config key could have failed the persist. A
run that COMPLETED with `condition_met: false` recorded in its own state is the artefact,
not my reading of `compareValues`.

Council submitted first: correlation `495df717-4010-491f-aec0-92c13aaf3809`, committed as
`1058b5366` with a `Council-Submitted:` trailer, because HEAD is shared and any session's
roll ships the code regardless.

## 2026-08-10 — council APPROVED, and five of six objections were real enough to act on

Round 1, correlation `495df717-4010-491f-aec0-92c13aaf3809`: **APPROVED**, 6 advisory
objections, none high-severity, 14 reviewers, 3 abstained. Verbatim seat notes are in
`doc_notes` (`categories ? 'council-gate'`) and the full report in `diagnosis_artifacts`.
Recording what each cost, because an approved round whose objections are ignored is the
dishonesty surface the coverage report exists to catch.

**`editquality` MEDIUM — a real gap, fixed in code.** My own diagnosis named THREE
false-positive modes and I had guarded two. The third — a `content` check aimed at a non-Go
file answered by a **same-named Go symbol** (the `slugify` case: six confident hits on
`slugifyPathSegments`/`slugifyForCompositionName`) — was named in the submission and left
unaddressed. Now `contentMatchReachNote()` rides on **non-empty** content answers: *"every
match above comes from a .go file … a .go symbol that merely SHARES A NAME with what you
asked about does not confirm it."* Mutation-proven at the call site.

**`bug_historian` MEDIUM — the strongest objection in the round, and it was right.** The gate
"depends on runtime resolution of `lookup.no_code_evidence` across a step boundary … the fix
could ship, look wired, and **never actually gate a single verdict**, with no error surfaced
anywhere." My live proof covered the FALSE branch only; a resolution miss would look
identical. Closed at build time: `codeEvidenceGateField` is now a Go constant, and
`TestSeedConditionResolvesAgainstTheActionsReturnShape` asserts seed 365's condition string
equals `"lookup." + codeEvidenceGateField + " == true"`, so a rename fails a test instead of
silently unwiring production. **Its first version did NOT close it** — it built its own literal
map and so tested only that the *evaluator* resolves a dotted path; renaming the key in the
action left it green. That is the *same* helper-versus-wiring hole that let a mutation survive
earlier today, one level up, in the test I wrote to close an objection about wiring. Caught by
re-mutating. The half that cannot be bought at build time — that the TRUE branch is reachable
live — stays a named acceptance step.

**`debug_historian` MEDIUM — needle gate missing, added and PROVEN.** The seed mutates a live
jsonb workflow blob, and I shipped only its verify/rollback half: no pre-write occurrence
count, no idempotency guard keyed on pre-state. Added, and **induced against the
already-applied state** rather than asserted: `needle gate: run_checks.next_step is
gate_evidence, expected the pre-365 value 'verify'`. Also recorded 365 in
`schema_migrations` via `--record-only` (the 270/273 precedent) — without that the runner
would have counted it pending.

**Cheap enumerations, all four done:** `answerCodeCheck` callers swept repo-wide (2, both
updated); `append_doc_note`'s 8 live consumers listed, 0 naming the new key; the "no compose
action exists" claim swept rather than resting on `transform_data` alone (only
`format_research_content`/`format_crawl_for_analysis`, both web-domain formatters); and this
council's own precedent on the seam checked — 4 prior reports, newest 2026-08-06, all
approved, none touching answerability, so no verdict is being repeated or contradicted.

**`guardian` process objection ACCEPTED, unremedied and recorded:** the workflow-JSON edit
should have been filed `operation: "config_change"` naming the owning pipeline, not `"add"`
on a new `.sql` file. A submitted plan cannot be amended and forward-only forbids rewriting
the round, so: **next submission in this lane uses `config_change` for a workflow edit.**

**`architecture` MEDIUM, `ARCHITECTURE_SIGNAL: needs_rfc` — routed, not argued.** The seat
holds that a new reserved key on a widely-shared action is architecture-scope by its trigger
test "regardless of the author's declaration", *while acknowledging in the same note* that
opt-in/default-OFF is "the sanctioned pattern from the 2026-08-02 owner ruling". So the
remedy the owner mandated for shipping new authority is itself the trigger. That is a
governance question, not a measurement one — and the 2026-07-28 ruling says a scope objection
is not answered by resubmitting with better numbers. Filed as
`architecture_review/RFC_022_an_opt_in_default_off_field_is_the_owners_own_remedy_and_the_seats_own_trigger.md`
with three costed options, and this lane's recommendation (trigger on the accumulated
optional-key COUNT, not on any single addition).

**Misstep worth its own line:** I nearly wrote `Council-Reviewed:` retrospectively by
amending the commit. Forward-only forbids the amend, and CLAUDE.md is explicit that `098`
resolves a `Council-Submitted:` correlation at REPORT time and credits the commit
automatically once the verdict turns approved. The trailer already on `1058b5366` is correct
and needs nothing; the follow-up commit carries `Council-Reviewed:` because by then the
verdict had been read.
