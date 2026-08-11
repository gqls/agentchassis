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
  **1,173**, so phase 2 is ~+20% of the corpus. > **CORRECTED later the same day: 1,371 (var 795, const 576), by running the analyser. 1,173 was a better proxy, still a proxy.** *A grep over declaration openers cannot
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

## 2026-08-10 14:41Z — ACCEPTANCE on v1.0.1279: it works, and the gate does not fire

Roll landed. Verified at the artefact, both replicas, against the negative control banked
this morning: `NOT ANSWERABLE BY THIS INDEX` **0 → 1**, plus a positive control
(`this answer is CAPPED` → 1) proving the grep, plus a never-added string → 0 proving the
grep is not matching everything.

Four entries re-fired. Results, and each is a criterion from the plan:

| criterion | result |
|---|---|
| no STALE attributable to an unverifiable footprint | **met** — the 08-08 flat-`STALE` entry now returns `NEEDS_HUMAN_REVIEW` with 3 checks `NOT ANSWERABLE` |
| the genuinely-checkable Go facts still confirmed | **met** — the provocation entry confirmed `loadGateCandidates`, `gate_provocation`, `provocations` **by name** while naming its non-Go half unverifiable, in one verdict |
| every persisted row carries the mechanical suffix | **met** — 4 of 4 |
| the verdict states the fact rather than guessing it | **met** — *"falls outside the Go-only code_symbols index"*, where the morning's run had said *"either not present at the current ref or not indexed"* |
| the gate's TRUE branch is reachable | **NOT SHOWN — see below** |

### The finding: the branch has never fired, and one substring match is why

`no_code_evidence` is `checks_with_rows == 0`. Across **4 of 4** runs the count was
**1, 3, 1, 1** — never 0. Every round routed to `verify`; `verify_unverifiable` has not
executed in production.

The `toolgolden.py` entry shows the mechanism exactly. Its footprint is three `.py` files;
three of its four checks classified `NOT ANSWERABLE`. The fourth, `content: VECTORS`, was
aimed at a **Python constant** and matched **8 Go rows** — `vectorSearchCodeSymbols`,
`pgvectorString`, `RAGIndexAction` — because `content` is an ILIKE **substring** match and
`VECTORS` is inside `vectorSearch` and `pgvector`. That single accidental hit set
`checks_with_rows = 1` and suppressed the gate.

**Two things follow, and I would rather write both down than the flattering one.**

1. **The third false-positive mode is COMMON, not a curiosity.** The council's `editquality`
   seat was right to press on it, and it is the reason it is worth pressing: it is the
   default outcome for a bare-identifier check, not an edge case. The new caveat fired and
   the verdict used it verbatim — *"matched only .go symbols related to vector search …
   which are unrelated to the VECTORS constant described in the entry"*. Pre-fix that is a
   confident confirmation of something never checked.
2. **My own gate is weaker than the submission implied.** I described it as making STALE
   unreachable for a round that confirmed nothing; measured, a round that confirmed nothing
   *of value* still counts one spurious substring match as evidence and takes the other
   branch. The gate is correct — resolution proven by the lockstep test, FALSE branch proven
   live — and **unexercised**. The protective work is being done by the evidence layer: the
   per-check `NOT ANSWERABLE` lines, the shares-a-name caveat, and the mechanical suffix.
   That is the order this lane ranked the candidates in (suffix above branch), so the data
   supports the ranking — but the branch must not be credited with work it has not done, and
   the bug file, the register and this file all now say so.

**The follow-up is NOT mine to make here:** `derive_checks` should not emit a bare `content`
query for a footprint item it can already see is a non-Go file. That is RFC_005's mechanism
(`architecture_review`), it would cut wasted checks, and it is what would make
`no_code_evidence` reachable when a footprint really is entirely unverifiable. Recorded in
the bug file as a follow-up round rather than smuggled into this lane.

## 2026-08-10 — PHASE 2 built: the write path emits var and const

New build `v1.0.1283` deployed; phase 1 re-verified in it (marker 1, negative control 0)
before starting anything new — a fresh image is not evidence a fix survived it.

**The defect was one arm of one switch.** `analyse.go` took `*ast.FuncDecl` and
`d.Tok == token.TYPE`; `token.VAR` and `token.CONST` fell through. Three edits: a
`ValueDef` carrier in `types.go`, the missing arm plus `valueDefs()` in `analyse.go`, one
loop in `code_symbols_actions.go`. Kinds needed no schema change — both were already in the
CHECK constraint and in `codeKindList`.

**Two corrections to my own figures, and the method is the lesson.**

| source | figure | why it was wrong |
|---|---|---|
| my grep, 08-10 morning | 930 | `grep -rhE "^(var\|const) "` counts declaration OPENERS — blind to every member of a grouped block |
| the design pass's awk | 1,173 | counted block members, but by text shape |
| **running the new analyser over the tree** | **1,371** (var 795, const 576) | the parser that will actually emit the rows |

A count of declaration *text* is a proxy. I had a proxy twice and wrote the second one into
phase 1's paperwork. **Build the thing and ask it.**

**[MEASURED] the prune interlock, and it is the opposite of what the design pass assumed.**
`prune_floor.go`'s `ratio()` says in as many words: *"An empty cohort has nothing to lose,
so it reads as fully confirmed rather than as 0/0 — a new class appearing for the first time
must never be able to refuse a prune."* So the FIRST run with var rows is safe by
construction. The exposure is the reverse: an **old** binary indexing against a DB that
already holds var rows sees `kind=var` at 0% confirmed, is below the floor, and refuses the
**whole** prune. Safe direction, self-healing, visible. That closes the `[UNMEASURED]` note I
left this morning — and note that I closed it by reading the code, having earlier written
down the risk in the wrong direction on someone else's authority.

**A design decision worth its own line: the LINE SPAN, not the name, is the point.** For a
value the body *is* the evidence (a spec's `Defaults`, a pattern table). Spec-level spans
inside a parenthesised block — the decl span covers every member, so slicing it per member
would repeat the whole block once per name — and decl-level for a lone declaration, so the
keyword and doc comment land inside the slice. Both halves pinned by tests.

**And a data-loss guard with no symptom:** the blank identifier is skipped. Identity is
`(repo, path, symbol)`, so every `var _ Iface = (*T)(nil)` in one file would UPSERT over the
last and all but one would vanish — silently, with every count still balancing.

**Verified in a clean tree.** Another session has a half-finished `gatherSchema` signature
change in `platform/orchestration/actions`, so the working tree does not build. Mine does:
`git archive HEAD` into a temp dir, copy only my four files over, `go build ./...` clean and
`go test ./internal/analysis/` green. Their WIP is theirs; I did not touch it.

Submitted as its own round (`3af67677-601e-4181-ad09-17c7a789f995`) because the blast radius
is the CORPUS — retrieval search space, embeddings, prune cohorts — not the rendering.

## 2026-08-11 — PHASE 2 ACCEPTANCE: live on v1.0.1284, every criterion met

All five §4a criteria pass. Detail below, but the headline is the pair of verdicts on the
same landmine entry, three days apart, with nothing changed but the index:

> **2026-08-08 — NEEDS_HUMAN_REVIEW:** "`metaCommentaryPatterns` and `placeholderPatterns`
> **no longer resolve as standalone symbols (possibly inlined or renamed)**"
>
> **2026-08-11 — STILL_VALID:** "All cited symbols (`placeholderPatterns`,
> `metaCommentaryPatterns`, …) **confirmed present at expected line ranges** in
> `validate_page_content.go` (commit 5a68d6ca)"
> `[code-lookup evidence: 8 check(s) ran; 8 matched indexed code; 0 NOT ANSWERABLE by this
> index; 0 ran and matched nothing in scope. Scope: 7118 symbols … **kinds with NO rows:
> type.**]`

Nothing was ever renamed. The manufactured hypothesis is gone, and the evidence line now
names only `type` as missing — **phase 1's var/const warning retired itself**, which is
criterion (e) proven on a live run rather than by `TestMissingKindNoteDisappears` alone.

**§4a's "cannot be pod-grepped" is now OBSOLETE, and the replacement is better.** The
handoff was right that phase 2 adds no new string literal, so the standing pod-grep recipe
cannot date it — but `bugs_open/153`'s build stamp landed in the same window and answers
the question directly:

```
{"caller":"agent-chassis/main.go:53","msg":"build provenance",
 "git_commit":"55fc8fc35f09a72992a1043c2850965792fb8b69"}
```

`git merge-base --is-ancestor 027bf28a0 55fc8fc35` → yes; likewise `c7c9dd87f` (the council
round). So phase 2 is in the running binary, established in two commands and with no
dependence on my change having a greppable spelling. **This generalises past this lane: the
"a change with no string literal cannot be pod-grepped" trap is retired for every service
carrying the stamp.** The spawned indexer pod was checked the same way — it ran
`agent-chassis:v1.0.1284`, because `resolveAgentImage` inherits the running chassis tag and
neither `code-indexer` nor `landmine-verifier` sets `pin_image_tag`.

**[MEASURED] The 1,371 figure in the handoff was a THIRD proxy, and I nearly accepted a
false shortfall because of it.** §4a set the acceptance bar at "var+const near 1,371". The
live census came to **1,204**. That is not a shortfall — 1,371 is simply the wrong number:

| measurement | var+const | why |
|---|---|---|
| grep over declaration openers (08-10 am) | 930 | blind to block members |
| awk over block members | 1,173 | counts declaration TEXT |
| running the analyser, **no excludes** | **1,373** | ← what the handoff recorded as 1,371 |
| running the analyser **with `exclude_patterns: ["docs/"]`**, as the action actually calls it | **1,204** | ← what the indexer can possibly emit |

`analyse_repo_local` calls `analysis.AnalyseWithExclude(dir, excludes)` with
`defaultAnalyseExcludePatterns = []string{"docs/"}` and no live override. The previous
figure was measured with the analyser but **not with the analyser's arguments**. The
lesson from the 08-10 table — *build the thing and ask it* — was followed; what it missed
is that **you must also call it the way production calls it.** A proxy can survive being
built.

Had I taken 1,371 at face value I would have read 1,204 as ~12% of values silently dropped
— which looks exactly like the identity collision the kind census exists to detect. **The
false-alarm direction is the dangerous one here**, because the remedy for a suspected
collision is to stop and dig.

**[MEASURED] The prediction was made independently and every kind reconciles EXACTLY.**
Rather than "near" anything, I built the deployed analyser (`internal/analysis` is
byte-identical from `027bf28a0`→`c7c9dd87f`→`55fc8fc35`→HEAD) and ran it over the exact tree
the indexer fetches, before the run finished:

| kind | predicted | live | reconciliation |
|---|---|---|---|
| var | 694 | **694** | — |
| const | 510 | **510** | — |
| func | 3702 | **3700** | −2: `init` appears twice in `directory_claims.go` and in `check_phantom_internal_links.go` |
| method | 1135 | **1135** | 31 name clashes in-file, **all immune** — methods are stored `(Recv).Name` |
| struct / interface / alias | 1001 / 36 / 42 | **1001 / 36 / 42** | unchanged, as criterion (b) requires |

7118 total, which is the figure the live verdict's own evidence line reports. A prediction
that could have come out otherwise, and did not.

**The identity constraint does NOT include kind, and that is a standing trap for the NEXT
kind, not this one.** `uq_code_symbols_identity` is `(repo, path, symbol)`. A phase-2
`var Foo` in a file that already has `func Foo` would UPSERT over it — the pre-existing row
would change kind and the old symbol would be **destroyed silently, with every total still
balancing**. Measured directly rather than inferred from the totals: **0 such collisions in
the indexed tree.** Phase 2 destroyed nothing. Filed as a landmine, because the measurement
is only true for today's tree and the next session adding a kind inherits the trap.

**A deliberate choice not to push, and it made the acceptance stronger.** The tree is **228
commits ahead of `origin/087_towards_multiple_domains`** (remote tip `5a68d6caf`, 08-10
17:27), and the indexer fetches the **REMOTE** tip. So the index already held exactly the
tree I was about to re-index. Re-running against an unchanged tree makes the census a
controlled comparison: every delta is attributable to the binary. **Pushing first would have
confounded criterion (b)** — a pre-existing kind's count would then move for ordinary code
churn, and the detector could not tell that from an identity collision. Both acceptance
symbols were confirmed present at `5a68d6caf` first, so nothing was lost by waiting. The
push is a separate decision and it is the owner's.

> Checked the remote with `git ls-remote origin <branch>`, **not**
> `git rev-parse origin/<branch>` — the latter reads the local remote-tracking cache and is
> only as fresh as the last fetch. They agreed today; that is luck, not method.

**Criterion (d), `bugs_open/231`'s reproduction, at the data layer:**
`DeployImageAssetInputSpec` is now one `var` row, `deploy_image_asset_action.go:32-44`, body
370 chars — and the body contains the `Defaults` map (`"purpose": "hero"`) plus the
`Deprecated` aliases. That is the declaration *content* 231 stopped for, not a list of use
sites. 231 is unblocked; whether its own `UNVERIFIABLE` now resolves is that lane's run to
make, not this one's claim.

### The same failure mode, one level up — found by verifying my OWN new landmine entries

Having run `landmines-sync.py --apply` (which consumes the `NEEDS_VERIFICATION` signal), I
dispatched the two entries I added today. Both returned an honest `NEEDS_HUMAN_REVIEW` — but
**one of them explained an absence with a manufactured reason, which is exactly the disease
this bug was filed about:**

> "several footprint items (… `ValueDef` type, `valueDefs` identifier, `token.VAR`/`token.CONST`
> references) live outside the index's scope (SQL migrations, **or are of kinds not
> indexed**) and cannot be verified mechanically"

**[MEASURED] That reason is false, and the disproof is in the same file.** `ValueDef` is a
`struct`, and structs are the best-represented kind in the corpus. Its three siblings from
the very same file are indexed right now:

```
FileInfo | struct | internal/analysis/types.go
FuncDef  | struct | internal/analysis/types.go
TypeDef  | struct | internal/analysis/types.go       -- ValueDef: 0 rows
```
`internal/analysis/` holds 26 rows, so the path is in scope too.

**The real reason is that the index cannot see a COMMIT, and nothing says so.** `ValueDef`
was added by phase 2 (`027bf28a0`, 08-10 23:13). The index is pinned to `commit_sha
5a68d6caf`, `commit_time 2026-08-10 16:27` — the last **pushed** tip. Measured now:
**246 commits and 88 changed `.go` files** behind the working tree. (It was 228 when I
measured at 09:45 this morning; the drift is other sessions committing during this session,
which is itself the point.)

So phase 1 taught the lookup to state the kinds it cannot represent, and phase 2 removed two
of them — but **"commits this index has not seen" is a third blind spot of the same shape,
and it is still unstated.** The evidence line reports the extension census and the kind
census; it reports nothing about staleness. A model handed "0 rows" and a caveat listing only
kinds and extensions will reach for a kind or extension explanation, because those are the
only two it has been given. That is what happened, verbatim, above.

**Why this is not a re-opening of `bugs_closed/108`.** 108 was "the refresh runs, resets the
freshness clock, and re-indexes the same stale commit", and its fix — pin the index to the
live working branch — **is working**: `ref` is `087_towards_multiple_domains`, the refresh
ran today, and the corpus matches that branch's pushed tip exactly. This is the residual it
leaves behind: **the index can only ever be as fresh as the last PUSH**, and on a tree where
sessions commit far more often than anyone pushes, "current with the branch" and "current
with the code" are different claims. Nothing currently distinguishes them for a reader.

**The self-reinforcing part, which is what makes it worth writing down.** The entries most
likely to be verified are the ones just written; the symbols a session writes a landmine
about are the ones it just added; and those are precisely the symbols least likely to have
been pushed. **The verifier is systematically at its blindest on exactly the entries it is
most often asked about** — and it currently reports that blindness as a property of the
code rather than of its own corpus.

> `[UNMEASURED]` how often this actually produces a wrong verdict across the 392 entries —
> I have one instance, found by looking at my own two. Sizing it is a query over
> `doc_notes` verdicts joined against symbol ages, not an argument, and I have not run it.
> `[NOT DIAGNOSED]` I am deliberately **not** asserting a root cause or a fix here: the
> mechanism above is first-hand verified, but "what the caveat should say and where it
> should be computed" is a change to a shared seam and belongs in a `090` round, not in a
> NOTES paragraph at the end of an acceptance. Recorded as a follow-up, not a diagnosis.

#### [MEASURED] The owner pushed the branch, and the staleness mechanism is now proven both ways

The push landed mid-session (remote tip `286884b65`, phase 2's source on it, 2 commits
behind local). Re-indexed against it and re-fired **the same entry**, so this is a
controlled before/after on the *staleness* claim specifically — nothing changed but which
commit the corpus had seen:

| | 10:00Z, index at `5a68d6caf` | 10:19Z, index at `286884b65` |
|---|---|---|
| verdict | `NEEDS_HUMAN_REVIEW` | **`STILL_VALID`** |
| the reason | "`ValueDef` type, `valueDefs` identifier … **are of kinds not indexed**" | "the analyser files and types (`ValueDef`, `valueDefs`, `token.VAR`/`token.CONST` arm) **are present**" |
| evidence line | 5 of 10 matched, **2 NOT ANSWERABLE**, 3 matched nothing | **9 of 10 matched, 0 NOT ANSWERABLE**, 1 matched nothing |

And directly in the corpus: `ValueDef | struct | internal/analysis/types.go`,
`valueDefs | func | internal/analysis/analyse.go`. **A `struct` and a `func` — the two
best-represented kinds in the index.** The earlier "of kinds not indexed" was a manufactured
reason, exactly as claimed, and the one thing the verifier still cannot check
(`uq_code_symbols_identity`, an SQL constraint name) it now names honestly.

This **strengthens** the finding rather than closing it. The gap was real, it produced a
false explanation, and the only thing that fixed it was a human happening to push. Nothing in
the lookup reported the staleness at the time, and nothing reports it now — the caveat still
says only `kinds with NO rows: type`. **The remedy is still unbuilt and still belongs in a
`090` round.** The landmine stands.

## 2026-08-11 ~12:45Z — v1.0.1286 verified through, RFC_022 implemented, staleness 090 filed

**The fresh roll (v1.0.1286) changes nothing for this lane, verified rather than assumed:**
pods' imageID digest matches the local image; the image revision label is `c3b424c8e`, and
phase 2 (`027bf28a0`) is an ancestor of it; the phase 1 marker greps 1 with a never-added
negative control at 0; the census still reads var 700 / const 525 with every old kind at its
post-push value. Migrations 381/383 are DB config and roll-independent. One practical note:
`kubectl logs | grep 'build provenance'` came back EMPTY here — the stamp line had rotated
out of a busy pod's log within ~25 minutes. The image label
(`docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'`)
answered instead, with the digest match closing the local-image-vs-pod gap. The stamp is the
front door; the label + digest is the fallback the runbook should mention when logs rotate.

**RFC_022 ruled and shipped** (see the RFC's STATUS block, CLAUDE.md's new section, and
commit `bacfa2e12`): option 3 with option 1 interim, live in both rosters via 381+383. The
099-reverts-377 tripwire found on the way is filed in LANDMINES, CLAUDE.md, and the
council_gate_cost lane's own NOTES.

**The staleness finding is now a 090 round, not a paragraph.** Filed ~12:40Z after the
dedup checks (queue: two unrelated items, both status `failed`; bugs dirs: no match).
`RUN_CORRELATION_ID=520b2f7e-5473-4655-8f41-9a04b7b9eab1` — the run key, stamped by the
dispatch loop; the intake correlation joins to nothing. The symptom named the mechanism and
pointed at the two doc_notes verdicts whose flip is the evidence; it asserted no counts and
no consequences. Whatever the loop concludes gets read against the both-ways proof above —
a REFUTED verdict would be a success too, and gets recorded visibly here and in LANDMINES.
