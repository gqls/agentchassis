# PLAN — bugs_open/223: the code index's blind spots must stop reading as absence

Design, phasing, decisions **and their reasons**. Corrections to the originating brief
live here, marked as corrections.

Lane opened 2026-08-10. Design pass run through `fable`; this document records the plan
that was actually built, including where it departs from that pass and why.

---

## 1. The defect, stated so the fix can be judged against it

The bad state, precisely: **a persisted `doc_notes` verdict asserting non-existence —
`STALE`, or "possibly inlined or renamed" — derived from a 0-row answer the index could
never have returned rows for, with nothing in the persisted row letting a reader see the
blindness.**

Three properties of that state shape everything below.

1. **The blindness is deterministic and knowable before the answer is read.** `code_symbols`
   holds Go only (5,837 rows, 100% `.go`, measured today) and 5 of the 8 kinds its own
   CHECK constraint permits. Whether a given check *could* match is a fact about the
   corpus, computable by one query.
2. **The inference is stochastic.** Four verdicts within one hour on identical 0-row input
   ranged from "cannot be mechanically verified" to "the entire described workflow has no
   footprint". Three of four hedged correctly. **So improving the average is not a fix** —
   the entry is degraded by the fourth and no reader can tell which run they hold.
3. **The product is prose that nothing parses**, read months later by sessions and by
   council seats' `schema_hint`. A verdict's status is not machine-read anywhere, so a
   wrong one is not caught downstream; it is simply believed.

From (1) and (2): remove the stochastic step's *input* (state the blindness as a fact) and
then bar the inference *structurally* (make the wrong verdict unreachable). From (3): the
qualifier must be attached by machinery, not by a model's cooperation.

## 2. Candidates, ranked by what makes the bad state unrepresentable

Ranking rule is the estate's own: prefer what makes the bad state impossible to represent
over what makes it less likely.

| # | candidate | verdict |
|---|---|---|
| 1 | **Mechanical evidence suffix on the persisted row** (`note_body_suffix_field` + a Go-composed census) | **BUILT.** The only candidate that survives the stochasticity: the damage-carrier is an *unqualified* false verdict read later, and this makes an unqualified verdict unrepresentable. |
| 2 | **Evidence-gated branch in the verifier's topology** (`conditional_branch` on `no_code_evidence`) | **BUILT.** A round that confirmed nothing cannot *reach* the STALE-bearing prompt. Honest residue: a model can still type "STALE" into free-form JSON — which is why candidate 1 is ranked above it, not below. |
| 3 | **Census-derived answerability in the shared lookup** | **BUILT.** Fixes all four consumers from one edit and is the substrate for 1 and 2. It *is* wording — but computed wording with a machine-readable twin, which is the standard 108/163/181 shipped under. |
| 4 | **Additive machine-readable keys in the action's return** | **BUILT.** Not independently rankable; it is what 1 and 2 branch on. |
| 5 | **Index Go `var`/`const`** | **PHASE 2**, own commit and own council round. The only candidate that removes the third failure mode's blind spot rather than reporting it, and the only one that lets the diagnosis loop CONFIRM a hypothesis decided by a `var`'s content (`bugs_open/231`). |
| — | **Per-class external checks** (`git ls-files`, `information_schema`, an `agent_definitions.type` row lookup) | **REJECTED here, recorded as the open question.** The git class *cannot* run in this action: the shared chassis pod deliberately holds no GitHub read token, which is this action's stated design premise. The others are new check-kind vocabulary on a shared seam and want their own round. |
| — | **Ingest non-Go files** | **REJECTED.** Architecture-scope by the 2026-07-29 ruling, as the bug itself says. Evidence it is a separate *intent* rather than an oversight: the reader-side D12 guard (`kindDoc`, `docBlockHeader`) is built and inert, and `code_symbols`' CHECK constraint does not admit `'doc'` — ingestion needs a schema change, so it is its own lane by construction. |
| — | **Prompt wording alone** | **REJECTED**, as the bug's own candidate 4. It rides along only as the rules inside the two verdict prompts, on top of 1–3. |

## 3. Decisions, with reasons

**The fix goes in the shared action, not in the verifier's prompt.** The damaging sentence
is `emptyAnswer`'s, and `emptyAnswer` serves four consumers: `fix-proposer`,
`feature-designer`, `landmine-verifier`, and the runtime lane which calls
`answerCodeCheck` directly and is invisible to any `agent_definitions` query. That lane's
own comment already states the requirement being generalised — *"'0 rows because bodies
are not indexed' and '0 rows because the code does not do that' must not render
identically"* — so this is finishing an argument the file already makes, not importing one.

**Every new sentence is COMPUTED from a live census; none is written down.** A hardcoded
".go only" would keep printing after it stopped being true, which is the stale-status class
this estate has scars from (a concept-register STATUS line outliving its truth is a
landmine in its own right). The census is one `GROUP BY (kind, extension)` per action run.
`TestMissingKindNoteDisappears` asserts the retirement *in advance*, so phase 2 landing
cannot leave stale prose behind.

**The census failure path does NOT set `s.err`.** `s.err` means "we know nothing about
scope" and suppresses the older guards that depend on the row counts. A failed or partial
census degrades to "composition unknown", which the classifier treats as *claim nothing* —
the pre-223 wording, never a false NOT ANSWERABLE. A partial read discards both maps
rather than keeping half, because a half-read census would report an extension as absent
because iteration stopped: the exact false "does not exist" being fixed.

**The classifier is deliberately narrow.** It speaks only about a file extension the census
says is absent. A directory prefix, a bare identifier, free text, or an unknown census all
keep today's wording. **A false NOT ANSWERABLE would suppress a real absence**, which is
the mirror of the bug — so the conservative direction is chosen explicitly, matching
`looksLikePath`'s stated posture in the same file.

**`checks_run` keeps its old meaning.** It counts checks that EXECUTED, unanswerable ones
included. Redefining it would have been tidier and would silently change a field other
consumers might read. Measured first: 0 live definitions reference it, or any of the four
new keys. The landmine is registered instead — `checks_run > 0` does not mean anything was
verified, and it never did.

**`codeRows` counts CODE-TIER rows only.** The D12 guard exists because a document *says*
where only code *shows*; a field named "did this check find evidence" must honour the same
line or it launders prose into proof one layer up.

**The evidence suffix lands on EVERY verdict, not only bad ones.** A qualifier appearing
only in the failure condition is a signal a model could learn to avoid triggering, and a
reader cannot calibrate a caveat they never see in the healthy case.

**An empty-resolving suffix field writes a loud line rather than skipping.** A missing
qualifier is precisely the condition the key exists to make visible.

### Correction to the brief

> **CORRECTED 2026-08-10 — my own figure was a lower bound and the design pass caught it.**
> This lane's first NOTES entry sized phase 2 from `grep -rhE "^(var|const) "` → **930**
> declaration lines. That undercounts: a grouped `var ( … )` block counts once however many
> specs it holds. Counting specs inside blocks gives **1,173**, so phase 2 is ~+20% of the
> corpus rather than the "15–25% of doubt" the brief carried. The lesson is the one already
> in `WRONG_CALLS.md`'s neighbourhood: a grep bounds what it can *see*, and a grep over
> declaration *openers* cannot see a block's members.

### Departures from the design pass, and why

- **The register entry went to `diagnosis-loop.md` (DIAG-036 + new DIAG-042)**, with a
  cross-reference note added to `documentation-system.md` DOC-067 rather than a new DOC
  entry. DIAG-036 is where this mechanism's siblings live and where 163's capability change
  already sits; DOC-067 is where a *reader of the corpus* arrives, so that is where the
  "how to read a verdict" caveat belongs.
- **No existing test needed amending.** The pass predicted `TestCodeIndexScopeEmptyAnswer`
  would need its "no UNKNOWN" assertion updated. It did not: the existing fixtures carry no
  census, so they take the fail-open path and pass byte-identically. That is the fail-open
  design confirming itself, and it is also a warning — see below.
- **`lsReachNote` on NON-EMPTY answers** is this lane's addition, from a measurement the
  bug file does not contain (§4).
- **The seed's snapshot uses `snapshot_agent()`**, not a hand-rolled INSERT. The first draft
  hand-wrote it and was wrong twice — no `name` column, and `(type, version)` is UNIQUE —
  both caught by `\d agent_definitions` before running anything.

## 4. What this lane found that the bug file does not contain

1. **The blind spot also produces FALSE POSITIVES.** `ls` is a path-*prefix* listing over an
   index of Go symbols and presents as a directory listing. `scripts/` returns **110**
   indexed paths (Go programs in its subdirectories) while every `.py`/`.sh` directly under
   it is invisible — so a generous listing reads as *confirmation* that a footprint
   resolves. **Worse than a false STALE:** a wrong accusation invites checking; a flattering
   partial confirmation reads as diligence.
2. **A `content` check aimed at a non-Go file can be answered by a same-named Go symbol.**
   `slugify` returned six confident hits on `slugifyPathSegments`/`slugifyForCompositionName`
   — a false positive *with citations*.
3. **The careful branch is also a total loss.** The run banked in this lane's EVIDENCE file
   reached *"either not present at the current ref or not indexed"* — a disjunction one
   census collapses — then guessed the census correctly and hedged everything on the guess.
   Two LLM calls and eight queries to arrive at "cannot confirm or deny".
4. **The kinds are schema-legal already.** `code_symbols`' CHECK constraint permits `var`,
   `const` and `type`, and the reader's `codeKindList` already treats them as code. Phase 2
   is an unfinished write path, not a design gap — which is why it is not RFC-scope.

## 5. Phasing

**Commit 1 (done, `1058b5366`)** — the census, the classifier, the four additive keys, the
runtime lane's adoption, the opt-in suffix on `append_doc_note`, seed 365, tests, register.
Council correlation `495df717-4010-491f-aec0-92c13aaf3809`, committed with
`Council-Submitted:` because the code is on a shared HEAD and any session's roll ships it.

**Commit 2 (next, same lane, own submission)** — index Go `var`/`const`:
`internal/analysis/types.go` (+`ValueDef`, `Output.Vars`), `internal/analysis/analyse.go`
(walk `*ast.GenDecl` for `token.VAR`/`token.CONST`, one entry per name per `ValueSpec`,
spec-level line spans so the body slice captures the literal — the deciding evidence
`bugs_open/231` needs), `code_symbols_actions.go` (+one loop, kinds already in the CHECK
constraint and in `codeKindList`), plus `internal/analysis` tests. **Ships after commit 1
has rolled**, because its blast radius is the corpus (embeddings, prune cohorts, every
diagnosis run's search space) and the guardian seat is right to want that reviewed apart
from a rendering change.

**Deferred to their own lanes** — non-Go ingestion (architecture-scope), and any new
external check kind.

## 6. Ordering, stated honestly

**No ordering constraint is claimed** (owner ruling 2026-07-29 retired that condition). The
seed is live immediately and the Go half is inert until a roll, so there is a window in
which the gate is wired and never fires. That window is *today's behaviour*, and it is safe
by two properties verified in the code rather than assumed:

- `conditional_branch` → `resolveFieldValue` returns nil for an absent field with no error,
  and `compareValues(nil,"true")` is false, so the gate takes `else_step` = `verify`;
- `append_doc_note` declares neither `ConfigKeys` nor `CheckConfig` nor `StrictConfig`, so
  per `action_inputs.go` it is "not checked at all" for unknown config keys — the new key
  is inert, not even a warning.

Registration in the same commit satisfies the surviving condition (2).

## 7. How a reviewer tells this is real rather than cosmetic

Every guard dies under a named mutation, and one of them **did not, first time**:

| mutation | test that must fail |
|---|---|
| invert `!s.representsExt(ext)` | `TestUnanswerableReason{ClassifiesNonGoPaths,IsConservative}` — both directions |
| drop `missingKindNote` from `emptyAnswer` | `TestEmptyAnswerNamesMissingKinds` |
| silence `lsReachNote` **at the call site** | `TestLsArmWiresTheReachNoteOnANonEmptyListing` |
| unwire `notAnswerableAnswer` in an arm | `TestLsArmRendersNotAnswerableAndFlagsIt` (prose **and** the outcome flag) |
| under-report `unanswerable` in the census line | `TestCodeEvidenceLine` |
| count `[doc]` rows as evidence | `TestContentArmDoesNotCountDocRowsAsEvidence` |
| no-op `applyBodySuffix` | `TestApplyBodySuffix` |
| delete `note_body_suffix_field` from seed 365 | live: no verdict row ends `[code-lookup evidence:` |
| flip the seed's condition to `== false` | live: `llm_call_log` shows the STALE-bearing prompt rendered |

> **The third row is the lesson of this lane, and it is in `WRONG_CALLS.md`.** The first
> version of the test suite exercised the rendering *helpers*. Silencing
> `b.WriteString(scope.lsReachNote())` to `b.WriteString("")` killed **no test** — the fix
> could be unwired and the suite stayed green, on exactly the half of the bug that this
> lane discovered itself. A helper with no caller looks precisely like a finished fix.
> Wiring tests through `answerCodeCheck` with a populated census now kill it.

## 8. Acceptance

Paired before/after on one entry, plus the bug's own criteria.

- **BEFORE** is banked verbatim: `EVIDENCE_2026-08-10_prefix_run_verbatim.md`, a run
  dispatched on purpose on v1.0.1277, five of eight checks unanswerable by construction.
- **AFTER the roll**, re-fire the same entry and require: (a) the two `.py` `ls` checks
  render `NOT ANSWERABLE BY THIS INDEX`, not "the query was RUN"; (b) the `doc_notes` and
  `subject_key` content checks still return their ~24 rows each — **the fix must not buy
  abstention by checking less**; (c) the verdict is an explicit unverifiable-by-index
  abstention rather than a hedge built on a guess; (d) the persisted row ends with
  `[code-lookup evidence: …]`.
- **Then a MIXED-footprint entry**, which is the case a partial confirmation flatters: it
  must route to `verify` (not `verify_unverifiable`), confirm what it confirmed, and name
  what it could not check in the same verdict.
- **Pod-grep with the banked negative control**: `NOT ANSWERABLE BY THIS INDEX` is **0** on
  v1.0.1277 and must be ≥1 on the roll, in every replica, alongside a positive control
  (`this answer is CAPPED` → 1) that proves the grep itself works.
