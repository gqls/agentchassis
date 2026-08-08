# NOTES — rfc012_await_findings (append-only, newest at the bottom)

## 2026-08-06 — lane opened on the owner's three rulings

Same sitting as the rulings (recorded in the RFC first, commit `3851e90b5`, so they
survive this session): (d) YES-standing-online-if-possible, census commissioned, B
assigned to this lane "now".

Prior state inherited: 098 CLOSED (its debt 5b is the proven pattern B generalises);
RFC_012 first sitting ruled B DB-backed with addendum-2 binding (the in-memory
namespace is REFUTED, not deferred — it dies at persistAwaitingStateWithRetry's fresh
load); addendum-1 binds any (d) implementation.

Three research agents launched in parallel (helper ground truth / online-check
precedents / the census itself, which writes its own artefact). Ownership checked:
098 lane closed itself; the RFC's ruling block said "implementation: unassigned" until
this sitting assigned it here.

## 2026-08-06 (evening) — B core + the (d) detector shipped; both delegated agents died on quota

Committed: `5f49b4cfd` (B core — the `agenterrors` leaf package, the forwarder, the actions
door, the exemplar conversion) and `abf5e8266` (the `--shared-output-fields` detector, its
tests, the hand-run script, the ack ratchet). Full state + what remains → HANDOFF.

**The 18 remaining INSERT sites**, with the quirks that must be preserved when converting
(from the ground-truth pass; column counts are what each writes TODAY):

| file:line | cols | quirk |
|---|---|---|
| complete_work_item_verification.go:206 | 11 | work_item_id RAW (no NULLIF); no site_id/domain |
| component_link_repair.go:204 | 9 | code+severity are SQL LITERALS; no orchestration_id |
| component_write_guard.go:315 | 13 | canonical |
| content_data_envelope_guard.go:363 | 9 | Go-side nil site_id; no orchestration_id |
| diagnose_council_decide_action.go:636 | 10 | no site_id/domain/work_item_id |
| diagnose_persist_fix_plan_action.go:391 | 8 | orchestration_id LAST; action literal |
| discovery_checks.go:392 | 9 | action+severity literals |
| plan_sections_action.go:1153 | 11 | no domain |
| prepare_link_context_action.go:493 | 9 | NO domain; orchestration_id LAST |
| reconcile_superseded_reviews_action.go:164 | 9 | NO step_name |
| render_content_envelope_guard.go:311 | 9 | Go-side nil site_id |
| save_sections_claims_guard.go:360 | 9 | action literal |
| save_sections_content_data_links.go:176 | 9 | severity literal |
| save_sections_dedup.go:541 | 9 | action literal |
| save_sections_metadata_source.go:300 | 10 | orchestration_id LAST |
| store_generated_component_action.go:1353 | 13 | canonical |
| validate_page_content.go:593 + :698 | 9 each | two sites, different provenance |

Nine of them omit `orchestration_id` entirely — a row that cannot be joined to its run,
which is the defect `save_sections_metadata_source.go:284-287` names and fixed only at
its own site. Existing error_code vocabulary to preserve: CONTENT_DATA_ENVELOPE,
FIX_PLAN_VALIDATION_REFUSED, DISCOVERY_CHECK_ERROR, CONTENT_CLAIMS_FLOOR_DETAIL,
CONTENT_DATA_REGRESSION, CONTENT_DUPLICATE_SECTIONS_COLLAPSED, CONTENT_DATA_LINK_AUDIT,
CONTENT_VALIDATION_BLOCKER_DETAIL, CONTENT_LINK_REPAIR_DETAIL.

**MISSTEPS, recorded:**

1. **A nil `Context` map marshals to `null`, not `{}`.** My first defaults test asserted
   `"{}"` on the strength of the old writer's `contextJSON == nil` guard — but
   `json.Marshal` on a nil map returns `[]byte("null")`, never nil, so that guard has never
   fired in production either. The test failed and was right to. Byte-compatibility means
   pinning what the code DOES; I nearly "fixed" the writer to match my expectation, which
   would have changed live behaviour under cover of a refactor.
2. **My first `emitSharedOutputFields` returned nil slices**, so a clean report marshalled
   `findings_new: null` and a consumer's `len()` crashed — caught immediately by my own
   consumer one command later. Initialise, don't rely on omitempty.
3. **The RFC's own "13 config keys" is one short.** Its prose names 11 after
   `then_step`/`else_step`; the live fleet has 13 config keys, the extra being a
   **config-level `error_step`** (158 occurrences) distinct from the top-level field. I
   only found it because I ran the enumeration query instead of transcribing the list —
   the addendum itself says "worth resolving against a live count before pinning a literal
   list", and it was right.
4. **Both delegated agents (the 18 conversions, the reader census) terminated on a Fable 5
   usage limit.** Neither left partial edits (verified: nothing outside my own two files
   references the new helpers; the dirty `actions/` files are other sessions' WIP). The
   census agent got as far as noting `enrich_fingerprint_with_css` is wrapper-adapted and
   was starting a third sweep for mid-string/condition references. Delegating both halves
   at once was the wrong call under an unknown quota — the census alone is a session's work.

## 2026-08-06 (late) — B core PROVEN LIVE on v1.0.1259

Pod-verified on BOTH replicas (`-54xsx`, `-ldx5z`, started 10:50Z; my commit `5f49b4cfd`
was 08:55Z, so the image postdates it — but a roll is not evidence, so:)

- positive `agenterrors` -> **5** on each replica (the new leaf package is in the binary);
- negative `retract_page_deployment: failed to record condition` -> **0** on each (the
  per-site log message the conversion DELETED; the generic one in agenterrors replaced it).

A discriminating pair: the INSERT statement itself is byte-identical before and after, so
grepping the SQL would have proved nothing either way — the removed log line is the only
string that distinguishes the two binaries. Choosing the needle was the whole check.

The (d) detector is a `cmd/` binary, NOT in the chassis image — it needs no roll, and the
online CronJob half will ship it as its own image (component-render-check's pattern).

## 2026-08-07 (early) — B is FINISHED in code: all remaining INSERTs converted, one left on purpose

Committed `f930de86b` (25 files). The 18 sites NOTES enumerated on 08-06 are converted,
minus one deliberate exclusion, plus one that did not exist when the census was taken.

**What each converted row gains.** Canonical 13 columns everywhere. Nine sites gain
`orchestration_id` — the run join `save_sections_metadata_source.go:284` named and fixed
only at its own site. `reconcile_superseded_reviews` additionally gains `step_name`, which
it had never written.

**The `domain` NULL→'' question, MEASURED before accepting it.** Nine sites used
`NULLIF($2,'')`; the canonical writer passes `$2` raw, so they now write `''` where they
wrote NULL. Rather than argue it:

```sql
SELECT CASE WHEN domain IS NULL THEN 'NULL' WHEN domain='' THEN 'EMPTY_STRING' ELSE 'set' END, count(*)
FROM agent_error_log GROUP BY 1;   -- EMPTY_STRING 9,964 | set 4,696 | NULL 128
```

Both forms already coexist live and `''` is the overwhelming majority, so the converted
rows join the shape every consumer already has to handle. The only reader that filters on
domain — `diagnose_load_runtime_action.go:267`, `($2::text IS NULL OR domain = $2::text)` —
excludes `''` and NULL identically when a filter is supplied and admits both when it is
not. `site_id`/`orchestration_id` carry both forms live too. **This is the disconfirmable
version of the check: had the census come back 9,964 NULL to 128 empty, the conversion
would have needed a NULLIF in the shared writer.**

**A NEW hand-copy landed DURING the work.** `plan_sections_action.go` gained an
11-column `FACT_SCOPING_EMPTY_COMPOSITION` INSERT in `ff515351e`, committed by another
session after this lane's 08-06 census and before it landed. Found only because I re-ran
the grep before committing instead of trusting my own table. **A census of copies is stale
the day it is written** — which is the argument for the shared door existing at all,
rather than for a one-off tidy-up. Converted here too, so the total is 19, not 18.

**The one site left unconverted, and why it is not laziness.**
`store_generated_component_action.go:1353`. Three reasons, in order of weight:
1. It is the site an earlier council round's **edit-quality and guardian seats named
   directly** — the objection is quoted at `:1343-1351` and ends "Left standing on purpose
   — the duplication is cheaper than the blast radius."
2. It already writes the **canonical 13 columns**, so there is no drift for the conversion
   to remove. The benefit is cosmetic; the objection is not.
3. It is **dirty with another session's in-flight change** (a `PageWantedLivePredicateFor`
   predicate swap at ~:869). A pathspec commit still takes a same-file passenger.
   **When the file is clean**: convert it with `LogActionEntry` and set
   `AgentType:"component-creator"`, `StepName:"store_component"`,
   `Action:"store_generated_component"` **explicitly** — those literals are exactly what
   the seats warned would be misfiled, and the merge must never be allowed to supply them.

**MISSTEPS, recorded:**

5. **I nearly used `LogActionError` at the provenance sites.** It resolves agent/step from
   `ActionParams`, which for `component_link_repair`, `validate_page_content`'s link
   recorder and both envelope guards would have *silently* replaced the ORIGIN's provenance
   with the running step's — the precise misfiling the guardian seat objected to, arriving
   under cover of a refactor with every test still green (they pin codes and messages, not
   agent_type). Caught by reading the objection at `store_generated_component:1343` before
   touching that file, then re-reading every other site for the same shape. **The fix was
   structural, not vigilance:** `LogActionEntry` takes caller fields as AUTHORITATIVE and
   fills only zero fields, so a named provenance cannot be overwritten by anything.
6. **Two tests asserted a code/severity against the SQL TEXT, because they were literals.**
   `mock.ExpectExec("'warning'")` and `mock.ExpectExec("CONTENT_LINK_REPAIR_SKIPPED")`.
   Under the shared writer those are bind parameters, so both tests failed — and the
   tempting fix (match `INSERT INTO agent_error_log` and let the values go to `AnyArg`)
   would have left a test that passes for a row carrying **any** code at all, which is what
   its own comment says it exists to prevent. Both assertions were MOVED to the argument
   and pinned by value. **A test that changes shape is the moment its assertion is most
   likely to be quietly weakened.**
7. **`agent_type` is NOT NULL and three sites carry their own `"unknown"` fallback.**
   Delegating that to the merge would yield `''` when both params sources are empty — a
   constraint violation, and a best-effort writer would swallow it as a warn. Preserved
   explicitly at each of those sites.

**Proven NOT live, with a discriminating count.** Live `v1.0.1261` (both replicas,
started 2026-08-06T19:54Z; the commit is 2026-08-07T01:24Z, so the image predates it):

```
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "INSERT INTO agent_error_log"'   # 14 on BOTH
grep -rn "INSERT INTO agent_error_log" platform/ --include=*.go | grep -v _test.go | wc -l   # 2 at HEAD
```

14 distinct statements in the shipped binary against 2 in the tree. **That count is the
acceptance test for the next roll: after a build from a commit ≥ `f930de86b` it must be
`2` on every replica.** It is a good needle precisely because it cannot be satisfied by a
roll that happened to include someone else's work — the number only falls when these
conversions are in.

## 2026-08-07 — council verdict: REVISE. The objection is correct, and it caught a real miss

Corr `5c2bc265-84ac-452b-bd8b-22fd7b875427` → `complete_revise`, **gating objection from
`editquality`**, 7 seats abstained.

**The gating objection is about the SUBMISSION, not the code, and it is right.** My prose
asserted explicit provenance-setting at `component_link_repair`, `validate_page_content`'s
link recorder and both envelope guards, and mentioned "four other updated test files" —
**none of which appear in the `edits` array**. The array holds 7 entries (the cap is 8) and
the change spans 25 files, so I described work the council was not shown and could not
judge. The seat's read is exact: *"Either real edits are missing from the plan the council
is asked to approve…"*. **Fix on resubmit: make the edits array the REPRESENTATIVE SET and
say so explicitly in the summary — name the file count, state that the array is a sample
chosen for the distinct shapes (merge door / provenance site / loop site / corrected
comment / new-copy-found-mid-work / test arity), and stop describing files it does not
contain.** Resubmit with `RESUBMIT_CORR=5c2bc265-84ac-452b-bd8b-22fd7b875427`.

**The genuinely valuable finding — a seat's read-only check pulled a landmine I had never
seen.** `` `WHERE domain IS NULL` on `agent_error_log` sees 1.3% of the rows that have no
domain — three of the twenty writers store `''` ``, filed 2026-08-06 by another lane,
**with `agenterrors.go` in its own footprint**. My NULL→`''` risk argument re-derived half
of it from scratch and missed the other half: the *filtering* reader really is indifferent,
but this conversion makes the `domain IS NULL` census undercount **worse**, and I shipped
that without saying so. Actions taken: dated correction on the landmine itself (title left
alone, it is the sync key); `WRONG_CALLS.md` entry for the check I skipped; RSH-008's open
review question now cites it and states plainly that **adding the `NULLIF` is the real fix
and is not done**.

**Why I missed it, which is the transferable part.** The `SessionStart` hook only surfaces
landmines for files **already dirty**. I was *calling* `agenterrors.go`, not editing it, so
the entry naming it was never shown and I read that silence as coverage. My own memory
index says to grep by SYMBOL for exactly this reason. **The trigger needs to be "I am about
to make a durable claim about a shared table's semantics", not "I am about to edit a file".**

**One unresolved discrepancy, left unresolved rather than smoothed.** The seat's re-run of
my census returned **13,765** empty-domain rows against my **9,964**, hours apart — while
the NULL count reproduced **exactly at 128**. The claim's direction is unaffected and the
discriminating figure matched, but I have not explained the gap (accretion? a different
bucket definition?), so it is marked [UNRESOLVED]. Note also what the landmine says and my
figures did not: this table is reaped at 14 days resolved / 30 unresolved, so **no census
of it is ever "all history"**.

## 2026-08-07 (morning) — the conversions are LIVE on v1.0.1262 — and the acceptance test I wrote for this moment was WRONG

Both replicas (`-5ghft` started 05:47:39Z, `-dfk4b` 05:47:14Z) carry `f930de86b` (01:24Z).
Proven with a discriminating pair on each replica independently, not with the roll:

```
POS  "failed to write some discovery check error records"   -> 1   (the PLURAL line f930de86b ADDED)
NEG  "failed to write discovery check error record"         -> 0   (the SINGULAR line it DELETED)
NEG  "content_data envelope: failed to write record"        -> 0   (per-site line, deleted)
NEG  "claims floor: failed to write finding record"         -> 0   (per-site line, deleted)
POS  "orchestration/agenterrors"                            -> 8
```
The singular/plural pair is the good needle: one string must be absent and its near-twin
present **in the same binary**, so neither a stale image nor a lucky substring can satisfy
both. Tree at HEAD agrees (plural 1, singular 0).

**MISSTEP 8 — the acceptance test I published for exactly this moment gave the wrong
answer, and it would have read as FAILURE.** NOTES 08-07 (early), the HANDOFF and RSH-008
all said: *"after a build from a commit ≥ `f930de86b` it must be `2` on every replica"*
(14 distinct `INSERT INTO agent_error_log` statements before, 2 INSERT sites in the tree
after). The measured value on a **correct** binary is **1**.

Cause: the two surviving sites — `agenterrors.go:89` and the deliberately-unconverted
`store_generated_component_action.go:1353` — hold **byte-identical** SQL literals, and the
Go linker **deduplicates identical string constants**. Two sites, one string in `.rodata`.
The pre-conversion 14 was 14 because the hand-copies had genuinely drifted (8/9/10/11/13
columns) — the very drift the conversion removed. **So converging the copies is what
collapsed the count, and my needle counted the symptom of success as if it were a count of
sites.** `1` is in fact a *stronger* pass than `2` would have been; but a session running
the documented check would have seen 1 ≠ 2 and concluded the conversion had NOT shipped.

The transferable part: **a needle that counts occurrences of a string literal cannot count
SITES.** It counts *distinct* strings after linker dedupe, and the number moves when two
sites become textually identical — which is precisely what a de-duplication refactor is
for. The class is filed in `LANDMINES.md` (footprint: `strings <binary> | grep -c`).
Corrected in place at all three publication points (this file above, HANDOFF, RSH-008).

**What this does NOT change:** the conversion is live and correct. The count fell 14 → 1,
and the singular/plural pair proves the specific commit rather than merely "some newer
build". Both are recorded because the count alone is now known to be a poor instrument.

## 2026-08-07 (morning, 2) — the `NULLIF` follow-up is REVERSED: measured, and adding it would be a regression

The handoff carried one piece of follow-up code: RSH-008's *"adding the `NULLIF` is the real
fix and is NOT done."* I wrote that yesterday after the council's read-only check pulled the
`agent_error_log.domain` landmine I had missed. **I did not measure it. Measured now, the
answer is no.**

```sql
SELECT CASE WHEN occurred_at >= '2026-08-07T05:47:39Z' THEN 'post-roll' ELSE 'pre-roll' END,
       count(*) FILTER (WHERE domain IS NULL), count(*) FILTER (WHERE domain = ''),
       count(*) FILTER (WHERE domain <> ''), count(*)
FROM agent_error_log GROUP BY 1;
-- post-roll:   0 NULL |     29 '' |    16 real |     45
-- pre-roll:  128 NULL | 13,885 '' | 4,762 real | 18,775
```

**Zero NULL rows since the conversion went live**, and grouping the 128 NULL rows by
`agent_type, action` returns nine groups, **every one of them a site this conversion
changed**, newest `2026-08-05 20:14Z` — all pre-roll. The NULL bucket is closed; the 14/30-day
reaper empties it. Adding a `NULLIF` would send 100% of new rows into the shape **0.9%** of
rows use, re-split a table that has just converged, and strand 13,885 `''` rows behind a
`domain IS NULL` query that would newly *appear* to work. The remedy the landmine already
prescribes — `COALESCE(domain,'') = ''` — is unaffected and correct for both eras.

The "be consistent with `site_id`" argument fails too: `site_id`/`work_item_id` take `NULLIF`
because they are **uuid** columns and `''::uuid` raises. It is a type necessity, not a
null-discipline precedent for a `text` column. I had been reading it as one.

**Two real findings the `NULLIF` framing was hiding.** Both fell out of re-locating the
writers instead of trusting my own file:line table (which the landmine on this very table
tells you to do):

1. **The census behind "nineteen copies" grepped `platform/` only.** There is a third live
   INSERT site — `internal/agents/contentcreator/claims_guard.go:184` — which **omits the
   `domain` column** and the column has **no DEFAULT**, making it the last latent NULL
   producer. It has written **0 rows in the entire retained window** (oldest row in the table:
   2026-07-08), so it is dormant, not benign.
2. **It cannot use the shared door, structurally.** `contentcreator` holds a `*pgxpool.Pool`;
   `agenterrors.Write` takes a `*sql.DB`. **"The ONE writer" is true of the `database/sql`
   half of the estate only.** Converting that last site is a driver problem, not a
   copy-paste job — worth knowing before anyone files it as a tidy-up ticket.

**Left unchanged deliberately:** `claims_guard.go`. Dormant, another lane's package
(`fa3b5207a`, bug 123, whose whole point is that this producer has no site), and a
null-discipline patch without the driver fix is a half-measure that would make the file look
converted when it is not.

**MISSTEP 9 — I accepted a reviewer's implied remedy as "the real fix" without measuring it,
one paragraph after being caught not measuring.** Yesterday's failure was under-checking (I
argued NULL→`''` inert without grepping `LANDMINES.md`). The correction I then wrote
over-accepted: the landmine says `domain IS NULL` undercounts, and I converted that true
observation into a prescription — *therefore write NULL* — that I never tested and that the
data contradicts. **Both errors have the same root: a claim about a shared table's semantics
made without a query.** Being caught is a prompt to measure, not a licence to adopt whatever
the catcher's framing implies. Filed in `WRONG_CALLS.md`.

**What would have falsified the reversal:** post-roll NULL rows still arriving, or any
NULL-domain group attributable to an unconverted writer. Both checked; both empty.

## 2026-08-08 — the census is delivered; council round 2 came back REJECTED on SCOPE VISIBILITY

**Still live.** v1.0.1263 (both replicas, started 08:54Z) re-verified with the same
discriminating pair: POS plural 1 / NEG singular 0 / NEG envelope 0 on each.

**The census landed** — `architecture_review/CENSUS_2026-08-07_rfc012_await_step_readers.md`,
commit `40992cbce`. Headline: **the config side breaks NOWHERE and the Go side breaks in
exactly 3 places, silently.** 138 of 221 awaited steps already merge under `.response`, so
(a) is a far smaller change than RFC_012 §4 feared. Full reasoning in the artefact; the two
findings worth carrying are that `ExtractNestedField` retries every unfound segment through
`["response"]` (so a wrapper is transparent to every dotted-path reader), and that
`hero_deployed`/`logo_deployed` are read by two-level direct map access with an `ok` guard,
so under a wrapper the page renders with no image and **nothing records it**.

**MISSTEP 10 — my first pass at "which actions await" returned 24, and the true answer is
40.** I grepped for the map-literal form `"await_response": true` and *excluded* lines
carrying `json:"await_response"` on the reasoning that a struct tag is an outbound request
field, not a result. But adapter actions signal await through a **typed result struct** —
`web_search_action.go:221` returns `&WebSearchResult{… AwaitResponse: true …}`. My exclusion
silently dropped **every adapter dispatch**: `web_search`, all five `firecrawl_*`,
`scrape_web`, `git_commit`, `generate_image`, `batch_webscrape`, browser-run, render-audit,
repo-analysis — i.e. the steps most likely to be awaited at all. Caught only because
`git_commit` steps were obviously missing from a list that had `deploy_page` in it.
**A census of a BEHAVIOUR must enumerate every way the behaviour is EXPRESSED**; I enumerated
one syntax and called it the behaviour.

## 2026-08-08 — council round 2: REJECTED, hard veto from `guardian`. The catch-22 is the point

Corr `5c2bc265-…`, verdict at 2026-08-07 08:39Z. Seats: `guardian` **veto**,
`editquality` object (medium), `reuse_agent` **approve**, `tooling_provenance` **approve**,
5 abstained. `decided_by: "hard veto from guardian"`.

**The veto is explicitly NOT about craftsmanship.** Its own notes say so: *"the merge-fill
design, the withdrawn-and-corrected NULLIF fix, the discriminating pod-verification, and the
disclosed dormant third writer are all genuinely careful work."* The objection is
**scope-visibility**: *"Guardian cannot assess blast radius on 26 files it cannot see — this
is the textbook 'MANY packages at once' signal the charter says to veto rather than approve
piecemeal."*

**And that is a genuine catch-22 between the two rounds, which is the transferable finding.**
Round 1's gating objection (`editquality`) was that my prose described files **not** in the
edits array. The fix was to declare the array a representative sample and say so plainly.
Round 2's `editquality` seat confirms that worked — *"The sample-not-exhaustive framing is an
honest and adequate resolution of the prior gating objection — the array now matches the
prose."* — and the **same honesty is what the guardian vetoed**: by stating that 26 files
were out of view, I gave the guardian the exact fact its mandate requires it to veto on.
**The 8-edit cap and a 34-file change are structurally irreconcilable inside one round.**

**The way out is named in the verdict itself, and it is cheap.** Guardian's `missing` field:
*"Full 34-file diff (or at minimum **the list of all 27 non-test files**) was not
supplied."* A list is not an edit and does not consume the cap. This is NOT the `bugs_open/124`
situation where a scope veto must not be answered by resubmitting — that veto was about *how
a capability reached production*. This one names a supplyable artefact.

**Three more actionable objections, all worth doing regardless of the verdict:**
1. **`editquality` (medium) and `guardian` (low) both say edit 8 — the (d) detector — is
   scope creep.** It is: it shares an RFC number and an owner sitting with B, and nothing
   else. *"Bundling an unrelated detector into the same reviewed plan violates (c) and makes
   the round judge two different bugs at once."* **Split it into its own round.** I bundled
   them because they arrived in the same sitting, which is a reason about my calendar, not
   about the code.
2. **`reuse_agent` (approve, but):** `sharedoutputs.go`'s routing-graph walk was never checked
   against **`relaygaps.go` in the same package**, which also walks `agent_definitions`
   routing config. Extend rather than reimplement, or state why not. I did not look.
3. **`tooling_provenance` (approve, but):** `cmd/config-key-audit` is an existing tool with
   history, and I added a detector to it without leaving a `doc_notes` PLAN/NOTES row for
   that subject — so the next lane must re-derive the ack-ratchet design and the
   13-vs-11-routing-key correction from source.

**No trailer written.** `f930de86b` carries `Council-Submitted:`, which asserts nothing and
is now simply uncredited. **Do not write `Council-Reviewed:` anywhere** — there is no
approved verdict.

**A small thing I got wrong and am recording rather than amending (forward-only):** I put a
`Council-Submitted:` trailer on `40992cbce`, a **docs-only** commit. The gate refuses docs
client-side, so the trailer is meaningless there. It asserts nothing so it cannot be a false
claim, but it is noise in 098's join and I should not have copied it across.

## 2026-08-08 (later) — round 3a APPROVED, and the reuse objection I had not looked at was a real defect

**Round 3a: APPROVED**, corr `5c2bc265-84ac-452b-bd8b-22fd7b875427`, verdict 15:25:32Z,
`decided_by: "approved with 5 advisory objection(s) — none high-severity"`. Thirteen seats
fired, 2 abstained. The split worked and the file list worked: `editquality` says the
plan "correctly answers the prior veto's request" and notes "Scope corrected: unrelated
detector split out". Nothing was rewritten to earn this — the code is identical to the
rejected round. What changed was showing the whole scope and removing the second bug.

**Round 3b submitted** as its own correlation `7b6497d7-8147-4b15-aee4-fa6e361827f2`
(FORCE=1 — every file is under `cmd/`/`scripts/`, and the 2026-07-17 scope rule exists to
stop *docs* spending credits, so I overrode it and said why in the rationale rather than
letting a path filter written for prose decide). Its edits array covers **all five files
of the change exactly** — `diff` of the commits' file set against the array is empty — so
the round-2 veto is structurally unavailable to it.

### The correction that matters: my file-count arithmetic, and the split's effect on it

The handoff told me to put "the full 27-non-test-file list" in 3a. **That figure was for the
THREE-commit set, and splitting the round splits it.** Measured, not transcribed: 3a is
`5f49b4cfd` + `f930de86b` = 29 files / 23 non-test; 3b is `abf5e8266` + `867037f5a` = 5 / 4.
The sets are disjoint (`comm -12` empty), so 23 + 4 = 27 and 29 + 5 = 34. I stated that
arithmetic in the submission explicitly, because handing the guardian a *shorter* list than
its `missing` field asked for, with no explanation, would have read as a dodge.

Both headline figures were also re-measured at the pre-B base `3e92c6a7a` rather than
carried forward from my own prose, and both reproduce: **19 INSERT sites across 18 files,
9 missing `orchestration_id`** — counted per SITE, because `validate_page_content.go` holds
two and a per-file count gives 18/8.

### MISSTEP 11 — the reuse objection was right, and "I did not look" was the whole defect

The round-2 handoff recorded `reuse_agent`'s point as an open question: had
`sharedoutputs.go`'s routing-graph walk been checked against `relaygaps.go` in the same
package? It had not. I looked, and it was not a style point.

`relaygaps.go:207` walks through `validation.WalkSteps` **on purpose**, and says why in its
own comment: "bugs_open/144's rule that a second hand-written descent goes blind in its own
direction". `sharedoutputs.go` had written exactly such a descent, with a stated
justification — "the walker does not expose containment" — **which is false**. `WalkSteps`
hands over a qualified path; the container is its third-from-last segment; `containerOf` is
four lines.

And it *had* gone blind, in the direction 144 predicts. A loop's body is declared as either
`substeps` or `sub_workflow`, and **`substeps` wins at execution** (`loop_actions.go:91`,
precedence mirrored by `validation.subWorkflowsOf`). The private descent read `sub_workflow`
only. Two consequences, and the second is worse than the first:

1. a `substeps` body is invisible → **0 findings**, indistinguishable from clean;
2. on a step carrying **both**, it walked the `sub_workflow` half — **reporting a hazard in
   config the executor ignores.** A reader chasing that finding would not find the behaviour.

**Fixed in `867037f5a`.** Three measurements, because "0 findings" is what this bug class
produces for free:

- **the gap was inert:** `$.** ? (@.substeps != null)` over live definitions → **0** at any
  depth. Widened to every row regardless of state → exactly 2, both *soft-deleted*
  `multipage-website-builder` rows. So the shape has been used here and retired, and no live
  run could ever have shown the gap. **This is why the proof had to be tests.**
- **the tests fail without the fix:** both new cases were written first and run red —
  `SubstepsShapeIsSeenToo` returned 0 findings, `BothShapesResolveToTheExecutedOne` named
  `inert_inner` by name. Green after.
- **the fix is a proven no-op on today's fleet, not an assumed one** — which mattered
  because **17 live agents do use `sub_workflow`**, so the container derivation and the
  executor's nested decode had to be shown not to move them. One 1,097,081-byte export of
  177 agents, both binaries (old from `git archive HEAD`, new from the tree),
  **byte-identical reports**.

Landmine filed with ten parseable footprints, and two things went wrong writing it that are
worth knowing: `landmines-sync.py` splits footprints on **commas**, not the `·` the file's
prose style uses, so my first append became one unsearchable blob (the hook could never have
matched it); and the entry format is `###`, not `##` — a `##` heading is treated as a section
divider and the sync warns. **Check the sync output for your own slug rather than assuming
`applied:` covered it**, and verify the rows by querying `body`, not the heading text.

### Three challenges from the 3a verdict, settled with measurements rather than left standing

- **`debug_historian`: "`grep -c` prints NOTHING on a zero count", so the NEG half of the
  pod-proof might be capturing an empty result. REFUTED, measured:** `grep -cF` prints `0`
  and exits **1** — `stdout=[0] exit=1`. The printed number is trustworthy. **But the seat
  was pointing at something real one step over:** the *exit code* is not, and with a NEG
  grep last, `kubectl exec` returns "command terminated with exit code 1" on a **correct**
  result. Any wrapper branching on exec status reads a pass as a failure. RUNBOOK now ends
  the probe with `true`.
- **`debug_historian`: "`-l app=agent-chassis` is 2 pods; 41 run that binary". CONFIRMED and
  sharper than stated:** 42 pods run an agent-chassis image under **four** app labels
  (`dynamic-agent` 38, `agent-chassis` 2, `business-intel` 1, `vet-intel` 1), of which **19
  are Running** — the rest are completed job pods, and exec'ing one fails with "cannot exec
  into a completed pod" rather than answering wrongly. So this lane's "both replicas" proof
  covered 2 of 19. What licenses the generalisation is **tag uniformity**, a separate query,
  now in the RUNBOOK. Re-proved 1/0/0 on **v1.0.1264** across two labels — and note the
  fleet had rolled twice since the last proof (1262 → 1264) with nobody in this lane looking.
- **`prior_art_librarian`: the import-cycle justification is asserted, never code-checked.
  Fair, and it holds:** `platform/orchestration/coordinator.go:23` imports
  `platform/orchestration/actions`, and `agenterrors` imports only stdlib + zap. The leaf
  package is necessary, and now that is a check rather than a claim.

### The objection FOUR seats raised independently, which I am recording rather than fixing

`bug_historian` (medium), `guardian` (low), `architecture` (low) and my own risks block all
name the same gap: `LogActionEntry`'s merge fills only zero fields, so a provenance you
**deliberately omit** is safe but one you **forget** is filled silently from the running
step — no error, no warning, and no test in the package would catch it, because they pin
codes and messages, never `agent_type`. `bug_historian` classes it with `missingkey=zero`:
the 18 known sites got the rigorous treatment while the shared mechanism stays exploitable
for caller 19. The named remedy is a required-provenance variant. **This is a real open
item, not a disagreement** — it is exactly the "make the bad state unrepresentable" move
this estate prefers, and it is now written into the `agenterrors` provenance row so caller
19 meets it before the trap does.

`tooling_provenance` also asked for a travelling `doc_notes` row for `agenterrors` itself,
not just for the audit tool. Both are filed (`subject_type='tool'`, keys
`cmd/config-key-audit` and `platform/orchestration/agenterrors`, `created_by='rfc012-lane'`).

`constitution` (low) objected to the rationale's tone — all-caps, combative — and noted the
plain-tone rule covers plan rationale, not just generated content. That is fair; 3b is
written plainly, which cost nothing.

## 2026-08-08 (evening) — the (d) CronJob is live, and two items closed as decisions rather than tasks

**All three owner rulings are now delivered.** `22ed9aa04` ships
`shared-output-fields-check`: daily 07:10 UTC, own image `v1.0.1265`, own tag sequence (it
does not ride the fleet release). First manual run 15:45Z — **CLEAN, 177 live agents, 0 new /
2 acked / 0 stale, container exit 0**, and it wrote its `doc_notes` row.

Three things the binary needed that it did not have, and each one is a trap avoided rather
than a feature:

1. **Direct Postgres.** Every mode read its fleet from stdin via `kubectl exec`; a CronJob
   cannot, because `ai-persona-app` has no `pods/exec` RBAC — and a kubectl-only tool fails
   there **looking like a clean run**. `PG_CLIENTS_HOST` is how the job declares itself. The
   DB and stdin routes **share one decoder**: `loadLiveAgentsFromDB` runs the wrapper's own
   query and hands the bytes to `decodeLiveAgents`, so the two can differ in how they FETCH a
   fleet and never in how they PARSE one. A third hand-written decoder is how 144 happened.
2. **Reporting on every run, clean ones included.** A check that only speaks when it fails is
   indistinguishable from one that has stopped. The body states the **scope** (agents scanned,
   routing keys read), because a finding count cannot distinguish "looked at 177 and found
   nothing" from "looked at 3 and found nothing" — and the second is a broken export reporting
   success. An undecodable agent row is disclosed in the row itself.
3. **Arguments scanned, not positional** — see the RUNBOOK correction. `--ack` at a fixed
   argv index was ignored silently anywhere else, and the image's CMD puts `--report` first,
   which is exactly the shape that broke. Fixed, and pinned by `TestParseSharedOutputArgs`.

**Proven on the IMAGE, not just with `go test`** — which is the check worth copying: `docker
run --rm -i <image> < live_agents.json` exercises the CMD, the ack path inside the image and
the argument order in one command, and it is how I confirmed
`/app/shared_output_fields_ack.txt` was actually readable. All four exit codes verified there
before deploying: 0 clean, 1 NEW pair, 2 unusable input (empty fleet, bad JSON).

⚠ **A shell artefact nearly gave me a false pass, and re-testing directly is what caught it.**
My first exit-code check printed `EXIT=0` for the empty-fleet case, which would have meant a
broken export PASSES. It was `${PIPESTATUS[0]}` in my own one-liner, not the binary: measured
directly, empty fleet → 2, bad JSON → 2, clean → 0, no ack list → 1. **A wrong exit code
would have been invisible in production** — the Job would simply have gone green.

### Item 3 CLOSED as won't-do, and the handoff instruction retired

`store_generated_component_action.go:1353`. The old handoff said "when clean, convert". I have
retired that, on two decisive grounds plus one practical:
- the **round-3a plan the council APPROVED says this site is unconverted.** Converting it now
  contradicts an approved plan and reopens an objection two seats raised;
- **"no drift to remove" is now a CHECK, not a claim:** extract both surviving statements,
  normalise whitespace, `diff` — **byte-identical**. Which is precisely why the Go linker
  dedupes them and why `strings | grep -c` reads 1 rather than 2. A conversion changes nothing
  about the row and buys only the risk of a provenance-literal slip;
- it is *still* dirty with another session's one-line `PageWantedLivePredicateFor` change, two
  days on. (`git diff -U0 <file> | grep -c agent_error_log` → 0, so it does not touch the
  INSERT.)

### Item 2 — the census understated it, and my attempt to confirm it FAILED. Filed, not asserted.

Reading `deploy_image_asset_action.go` made the finding stronger than the census had it.
`DeployImageAssetAction` has **one** path: it calls `sendGitCommitRequest` — whose return map
carries `await_response: true` and **no** `image_url` — and *then* assigns
`result["image_url"]` from `processed.Paths.RelativeURL` onto that same map. So the action's
own computed workings and the await signal travel in ONE map under
`hero_deployed`/`logo_deployed`. **That is RFC_012's own mechanism biting the very keys the
census flagged**, not a future risk conditional on (a) shipping. The readers
(`v3_site_actions.go:1010`/`:1021`, `assemble_from_library.go:452`) are two-level direct map
access with an `ok` guard and **no else branch**, so a mismatch sets no `hero_url`/`logo_url`
and logs nothing.

**And I could not establish the empirical half — so I did not claim it.** `hero_deployed` and
`logo_deployed` appear in **0 of the 1,667 retained `orchestration_states` rows** (window
opens 2026-07-13), so I observed no live shape in either direction. The config side IS durable
and measured: 4 steps declare these keys, on `pageflow-builder` and `site-work-orchestrator`,
`await_response` unset in the definition because it comes from the action's return.

Per CLAUDE.md's default for a structural, cross-cutting claim, filed to the loop rather than
written up: **`RUN_CORRELATION_ID=dce40cf4-5a8a-4316-93c0-0f3c37d2f3a7`** (intake
`59780c12-…`; the printed intake corr is NOT the artifact key). Queue and both bug dirs
grepped first — the adjacent `deploy_image_asset` bugs (152, 155, 179, 209) are all about
**source resolution**, not this overwrite. **A REFUTED verdict is a good outcome here** and I
would rather have that than a confident paragraph in a handoff.

Gotcha, cost me one intake: **the 090 trigger does not escape double quotes in the symptom** —
`"image_url"` dies server-side with `invalid input syntax for type json`, *after* printing a
correlation, so it looks like it worked. Now in the RUNBOOK.

### Two small self-corrections, recorded rather than amended (forward-only)

- **I put `Council-Reviewed:` on a docs-only commit** (`d42dcdba5`). The verdict is genuinely
  approved and I had read it, so it is not a false claim and 098 confirms **MISMATCH: 0** — but
  the gate refuses docs client-side, so it is noise in the join. This is the SECOND time this
  lane has done this (the earlier one put `Council-Submitted:` on `40992cbce`). Trailers belong
  on the commit whose CODE was reviewed.
- **I briefly told the user a landmine append had been lost.** It had not: another session had
  committed it minutes earlier as a *declared* same-file passenger (`e492c2abc`, whose message
  names my entry and its line count — good citizenship). My grep said 0 because I searched
  `substeps is the half that RUNS` while the heading reads ``substeps` is the half that RUNS``
  — **a grep proves absence only for the spelling it searches**, and I walked straight into it
  seconds after quoting the same rule about someone else's needle.

**098 coverage, for the record:** `f930de86b` is now credited `[5c2bc265, by correlation, via
submitted]` — the `Council-Submitted:` trailer resolved itself at report time, with no amend,
exactly as designed. `5f49b4cfd` remains in UNREVIEWED because it carries no trailer at all
(it predates the submission); forward-only, so it stays that way. `867037f5a` and `22ed9aa04`
are outside the report's scope, which is `platform/`-shaped commits only.

## 2026-08-08 (late) — v1.0.1266, the mixed-fleet trap, and the 090 came back UNVERIFIABLE

**A fresh chassis build went out mid-session (v1.0.1266) and re-proving it caught a trap my
own RUNBOOK had only half-covered.** The fleet was **mid-roll: 20 pods on v1.0.1264 and 5 on
v1.0.1266 at the same moment.** My documented loop picks one pod per *label*, which is only
sound when every pod runs the same tag — and here a label-picked pod could have answered for
either build, so "it is live" would have been a coin toss dressed as a measurement. Fixed the
RUNBOOK to enumerate one Running pod **per distinct image tag**. Both tags read `1 / 0 / 0`,
so the conversions are live on the new build — **but that was a finding, not an assumption**.

Generalises beyond this lane: **tag uniformity is a precondition of the pod-proof, and on this
cluster it is false for minutes at a time after every release.** The council's
`debug_historian` seat got me half-way here by pointing out the label selector covers 2 pods
of 42; the other half is that the 42 are not necessarily running the same thing.

### The 090 verdict: UNVERIFIABLE — which is NOT a refutation, and the distinction matters

`dce40cf4-…` ran **5 iterations** and stopped: `status: UNVERIFIABLE`, `conclusion: NOT
CONFIRMED (stopped: evidence-not-growing)`, `is_fix: false`, *"Hand to a human with the full
trail; do NOT auto-conclude."* It hit **exactly the wall I hit** — there is no retained
runtime evidence to grow into, because those keys appear in 0 of 1,667 orchestration rows.

**Read against `an-unverifiable-verdict-does-not-say-your-premise-was-false`: this says the
QUESTION could not be settled by the evidence available, not that the premise is wrong.** And
what it did settle cuts my way: its final round tested a RIVAL hypothesis — that the
downstream reader uses a plain `hero_url`/`logo_url` — and returned **`last verdict:
REFUTED`**, on the ground that *"BuildRenderContextAction's actual, PRIMARY read for both
images is the exact two-level raw access the original symptom describes … executed BEFORE any
check of a plain hero_url/logo_url key … the raw two-level read … is in fact the primary,
unmitigated path"*. It also found no evidence that `DeployImageAssetAction` writes
`hero_url`/`logo_url` unconditionally.

**One genuinely new fact from the trail: a `hero_url`/`logo_url` FALLBACK exists**, running
after the raw read and labelled `(fallback)` in the code. Whether anything populates it is now
the crux — it decides live defect vs latent one — and neither I nor the loop could answer it
statically. **So the next move is a CANARY, not more reading**: trigger a page build and catch
`collected_data->'hero_deployed'` in flight, then look at the rendered page. One run settles
what five iterations could not. Recorded in the handoff.

**Cost of the exercise: one 090 run, and I would spend it again.** It refuted a rival
explanation I had not considered and surfaced the fallback path, which is more than the
confident paragraph I would otherwise have written into a handoff.

### The measurement that kills the obvious fix for the merge gap

Before handing on §1 I measured which call sites actually rely on the merge, because "make
provenance required" is the remedy four seats gestured at and it **does not work**: **13 sites
name their provenance EXPLICITLY, 7 are MERGE-FILLED and are RIGHT to be** (the running step
IS the correct provenance for `complete_work_item_verification`, `diagnose_council_decide`,
`multipage_actions`, `plan_sections` ×2, `reconcile_superseded_reviews`,
`retract_page_deployment`). A hard requirement breaks those 7.

The design that survives is the estate's own: RFC_010's owner ruling of 2026-08-02 — *"make X
a field with the unsafe default OFF"*. Here the unsafe branch is silent inheritance, so it
becomes `Entry.InheritProvenance bool`, default `false`; the 7 declare it, the 13 are
untouched, and an Entry with neither an `AgentType` nor the opt-in is a state the writer can
refuse instead of papering over. Go's zero value makes the unsafe default OFF for free. Full
working in the handoff so the next session does not re-derive it.

### §1 BUILT — and I did NOT build the design the last entry above specified

> **CORRECTION to the entry immediately above, recorded rather than edited away.** That entry
> concluded the design should be `agenterrors.Entry.InheritProvenance bool`, default `false`.
> **I did not do that, on a structural ground the entry had not considered:** `Entry` is the
> ROW, and `agenterrors.Write` — the package that OWNS the type — would ignore the field
> entirely. Worse, `orchestration.AgentErrorEntry` is a type ALIAS of `Entry`, so agentbase,
> messaging and the coordinator would each acquire a knob that does nothing for them. An inert
> field on a shared leaf type is its own trap, and this estate has been bitten by that class.
> **Shipped instead: a named door** — `LogActionEntryInheritingProvenance` /
> `LogActionEntryFindingsInheritingProvenance`. It satisfies the RFC_010 ruling's *purpose*
> (unsafe branch declared at the call site, unsafe default OFF) while keeping the leaf type
> row-shaped, and it makes the census a grep rather than a struct-body read. The departure is
> disclosed in the council submission's risks block as item 2, for a seat to overrule.

What shipped: the merge is split into a **JOIN half** (`orchestration_id`, `agent_id`,
`pod_name`, `work_item_id`) still inherited when zero, and a **PROVENANCE half**
(`agent_type`, `step_name`) never inherited unless declared. `actionErrorEntry` is gone,
replaced by `actionJoinIdentity` + `runningStepProvenance`, so the dangerous half is **not
reachable** from the function that performs the fill — the control is structural, not a comment.

**A forgotten provenance LANDS as `agent_type='unattributed'`, and I want to record why, because
I nearly refused the write instead.** `component_write_guard.go:276-279` says "a row in
`agent_error_log` that misattributes the writer is worse than no row", which reads like a
licence to refuse. It is not: `agenterrors.go`'s own header says **this table is the only sink
that survives an awaited step**, so a refusal silently destroys a finding — and a sentinel does
not misattribute, it declares ignorance. Landing the row with the running step **demoted into
`context`** (where it asserts nothing) satisfies both. It also closes the second half of the
old landmine for free: `''` can no longer reach a NOT NULL column, so the vanishing-row case is
now unrepresentable from this door.

### The measurement I did not expect, and it makes the 7 "merge-filled" sites look worse than the last entry said

The entry above says the 7 merge-filled sites are "RIGHT to be" — the running step IS their
correct provenance. That is true as a statement about intent. **What the running step actually
resolves to is another matter, and I only found out because I queried the live table:**

```
 error_code                        | agent_type | step_name | rows | last
 REVIEW_SUPERSEDED_BY_PASSING_SAVE | generic    |           |   25 | 2026-07-23
```

`generic`, and an **empty** `step_name`. Then the fleet-wide shape: `generic` holds **559 rows
across 25 distinct `step_name`s** — the widest step spread of any `agent_type` on the table,
against e.g. `vet-practice-verifier`'s 9,696 rows across 5. That is the signature of a filler
being inherited across unrelated sites, on a table whose main investigation index is
`(agent_type, occurred_at DESC)`.

**And the estate already knows.** `types/context.go:62` documents `RunAgentType` as *"the
RESOLVED real agent type whose workflow is executing … as opposed to the dispatch-path sender
which is often 'generic'"*, citing `bugs_open/060`; `coordinator.determineOwnerAgentType` reads
it precisely so `owner_agent_type` stops recording `generic`. Corroborating, from a different
direction: `agent_error_log_test.go:187` lists `"generic"` alongside `""` in the set of values
that must NOT pass a review gate — this estate already treats it as a non-identity.

⚠ **I deliberately did not fix it, and the reason is this lane's own scar.** `actions` cannot
call `determineOwnerAgentType` — it is a `*SagaCoordinator` method in the package that imports
`actions`, i.e. the exact import edge that forced `agenterrors` into existence. The structural
fix is to hoist the ladder onto `*types.ExecutionContext` so both sides share ONE copy, and
that is a **second shared seam** in a commit that already ships one. Round 2 of this lane was
REJECTED on precisely that (two seats called an unrelated detector scope creep). So it is
recorded in three places a future session will actually hit — the register entry, the LANDMINES
entry, and a comment at the `reconcile_superseded_reviews` call site — and bundled nowhere.

> ⚠ **Scope honesty, marked because it is an inference:** [INFERRED] that the other 3
> merge-filled `LogActionEntry` sites also resolve to `generic`. Only `reconcile` has live rows
> (25); the other three have **written none in the retained window**, so I could not measure
> them and have not claimed them. What IS measured is the fleet-wide 559/25 shape and the
> `RunAgentType` doc comment saying the sender is "often" generic.

### MISSTEP 12 — my own channel corrupted a Go source file, and `gofmt` was clean either way

Writing the test file, I typed a pair of ASCII apostrophes (`''`) inside a comment. What
reached disk was **U+201D, a right double quotation mark**:

```
$ grep -n "does not produce a bad row" log_action_error_test.go | cat -A
151:// so M-bM-^@M-^] does not produce a bad row, it produces NO row.$
```

`gofmt -l` was silent (UTF-8 in a comment is legal Go), the package compiled, and all 12 tests
passed. I only caught it because the harness echoed the changed line back and the glyph looked
wrong. **In a comment this cost nothing. In a string literal — an error code, a SQL fragment, a
`banned_claims` needle — it would be a silent behaviour change that no formatter, compiler or
test in this estate would flag.** The cheap check, and it is now how I finish any file I write:

```bash
grep -o -P '[^\x00-\x7F]' <file> | sort | uniq -c        # inventory; — and § are intentional
grep -n -P '[\x{2018}\x{2019}\x{201C}\x{201D}]' <file>   # smart quotes are NEVER intentional here
```

Both files came back with exactly one offender, in the test file, now fixed.

### The mutation runs, because a green suite was the defect and could not also be the proof

Five mutations, each caught by exactly the expected tests and no others — the "no others" half
matters, because a mutation caught by *everything* usually means the harness is wired wrong:

| mutation | tests that fail |
|---|---|
| restore the pre-split merge (inherit unconditionally) | **3** — incl. the reported defect's own test |
| strict door stops inheriting the JOIN half | **10** — the RFC_012-B run-join win is guarded |
| the opt-in door silently becomes strict | **3** |
| annotate the caller's context map in place, not a copy | **1** |
| sentinel becomes the empty string | **4** |

Restored: **12/12 pass**. The map-copy mutation is worth its own line: `RecordFindings` shares
one base `Entry` across a batch and **replaces `Context` per finding**, so annotating in place
would both leak into the caller's own map and be silently dropped for every findings row. That
is why the diagnostics are applied per finding in `recordFindings`, not once on the base.

### Two things about the TREE, not the code, that cost me time and would cost the next session more

1. **`complete_work_item_verification.go` was being edited by another session while I read it.**
   My census grep put the call site at line 198; a `sed` of 190-215 two minutes later showed
   comments, because the file had gained 89 lines in between. **A line number from a grep is
   stale the moment you print it on this tree.** I re-derived by symbol
   (`grep -n "LogActionEntry(" <file>` → 265), verified the concurrent diff does not touch the
   INSERT (`git diff -U0 <file> | grep -c agent_error_log` → **0**), and confirmed their WIP
   compiles and its own tests pass before taking it as an unavoidable same-file passenger.
   > **CORRECTED, same session, ~40 minutes later — there was NO passenger in the end.** They
   > committed their own work first (`1c5d9ceb5`, RFC_017), so `git diff` on that file came back
   > as my one call site and nothing else. **The council submission still discloses the passenger
   > as a risk, because it was dispatched before this resolved** — an over-disclosure of a risk
   > that evaporated, not an over-claim, and not worth a duplicate round to correct. Worth
   > recording as the *good* outcome of the concurrency rule: the answer to "their WIP is sitting
   > in my file" is often just **wait until you are ready to commit, then re-check**, because on
   > this tree everyone else is committing narrowly too. Their commit message is also a neat
   > independent corroboration of this change's design — *"fails CLOSED, with fail-open an opt-in
   > field whose unsafe default is OFF"* — the same RFC_010 §2 shape, reached the same day by a
   > different lane on a different seam.
2. **`LANDMINES.md` changed under me mid-edit** — the Edit tool reported the file had been
   modified on disk since I read it. The edit still applied because I anchored on an exact
   quoted span rather than a line number. `landmines-sync.py --apply` then reported
   `content changed: 1` and `orphaned: 0`, which is the check that my edit disturbed nobody
   else's entry.

Full-tree baseline, so the pre-existing failures are not read as mine: `git archive HEAD | tar
-x` into a scratch dir at `6f2f2b1ce` reproduces the **identical** failure set (thunder adapter,
`test/integration/*`, `test/performance/*`, `test/e2e/*`). `./platform/orchestration/...` is
**9 packages green** with the change in.

### The verdict: APPROVED round 1 — and two of its objections were worth acting on, not noting

`5d200313-f6c3-4fec-8457-503ac620d5ef`, 2026-08-08 17:32Z. **10 seats, verdict APPROVED, 2
advisory objections counted, none high, no veto.** Dispatch→run start was ~1 second, not the 29
minutes the runbook warns about — so the lane was idle; do not read my timing as the norm.

**The scope question got answered by the seat I asked, which is the durable part.** I had written
into the submission's risks that I judged this council-gate scope rather than RFC scope under the
2026-07-29 §1 test, and invited the `architecture` seat to disagree on record. It approved, with
`ARCHITECTURE_SIGNAL: point_fix` and an explicit trigger test: *"stays inside
platform/orchestration/actions, adds no schema column, no new reserved key on a shared action
config, no wire-shape change — agenterrors.Entry and agent_error_log are untouched. New exported
symbols are additive wrappers; every existing call site's behavior is provably unchanged (backed
by mutation te…"* (truncated in the artifact). **That is now a worked precedent for the next
seam of this shape**, and it is better than my own reasoning because it names the test rather
than the conclusion.

`bug_historian`'s note is the one to hand the next author: *"unlike that case it audits ALL 20
call sites rather than patching one and leaving the mechanism generic — that is the correct fix
shape per the historian's own charter (c)."*

#### Objection 1 (bug_historian medium + editquality): the step_name asymmetry. THEY WERE RIGHT.

I had scoped it out: a caller naming `agent_type` but leaving `step_name` zero still had
`step_name` filled from the running step, and I justified that as preserving two
council-approved sites and avoiding scope creep. `bug_historian` called it *"a narrower
recurrence of the exact defect being fixed … leaving the same failure shape live for any future
caller of that shape"*, and `editquality` flagged the same thing independently.

**The scope argument was a false trade-off and I should have seen it.** The two sites did not
need the asymmetry — they needed *inheritance*, which there is now a door for. So: the strict
door is strict on **both** columns (naming only `agent_type` yields an empty `step_name` plus a
`logger.Warn`), and `prepare_link_context` + `diagnose_persist_fix_plan` call
`LogActionEntryInheritingProvenance`, keeping their own explicit `agent_type` while declaring the
step. **No live row's contents change**, and that is pinned by a table test over both sites
rather than asserted. Mutation 6 — restore the asymmetry — fails exactly
`TestLogActionEntry_NamingOnlyAgentTypeDoesNotBorrowTheRunningStep` and nothing else, while both
migrated-site subtests keep passing, which is the proof the migration is behaviour-preserving.

The reasoning that matters for next time: **an empty value asserts nothing; a borrowed one
asserts something no caller claimed.** The defect class here is false attribution, not absence,
so "leave it blank and warn" is the honest floor and "fill it from whatever is nearby" never is.

#### Objection 2 (guardian medium ×2): facts a council cannot check. Both now checked.

- *"whether any caller outside `platform/orchestration/actions` … also calls LogActionEntry
  directly and would be silently switched from lenient to strict."* Repo-wide grep across all
  six doors: **0 callers outside that package.** The `agenterrors.Write` callers (agentbase,
  messaging, coordinator) build a full `Entry` and never touch this merge.
- *"the 20-call-site census … is the single fact the whole containment argument rests on.
  Flagging so a human with full repo access confirms the count before merge."* Recounted, and
  **the count had MOVED: it is 21.** See the next section — this is the finding of the hour.
- The same-file passenger objection was **discharged by fact, not by argument**: the concurrent
  session committed its own work (`1c5d9ceb5`) before I committed, so there was no passenger.

#### Objection 3 (debug_historian + prior_art_librarian): re-run the figures, do not carry them

Both seats flagged that I cited live counts from earlier in the session without re-running them
at commit time, `debug_historian` naming the WRONG_CALLS shapes it matches (*"a two-day-old
figure carried into a fix's own header"*). **Correct, and it cost one query.** Re-run at commit
time, all six in one statement: `unattributed` **0**, `agent_type IN ('','unknown')` **0**,
`generic` **559** rows over **25** distinct `step_name`s, `REVIEW_SUPERSEDED_BY_PASSING_SAVE`
**25** rows of which **0** are non-`generic`. Every figure held. **That they held is not the
point — that they were checkable and I had not checked them is.**

### THE 21st CALL SITE ARRIVED WHILE THE CHANGE WAS IN COUNCIL, AND IT LANDED RIGHT WITHOUT KNOWING

Recounting for `guardian` turned up `page_build_failure_guard.go:110`, which was not in my
census and therefore not in the approved plan. It is not a miss — the file did not exist when I
counted. Another lane committed it as `2c3efc9f5` (PBP-038, deploy-stamp refusal) while my round
was running.

**How the tree hid it, which is the transferable bit.** `git log --oneline -3 -- <path>` returned
**empty** and `git status` showed nothing, so my first read was "untracked file, another session
is mid-write". Both were wrong in the same way: the commit had landed *between* my two commands,
and a `git log` pathspec query is answered from **HEAD at the moment it runs**. `git ls-tree -r
HEAD -- <path>` plus `git log --all` settled it in one command. **An empty `git log` for a file
that visibly exists means your HEAD is older than the file, not that the file is untracked** —
this is the `a-quiet-git-log-is-not-silence` trap wearing a different hat.

**And the design news is good.** The new site calls `LogActionError` — the door whose contract
already *is* running-step identity — so it required no migration and got the safe behaviour
**without its author knowing this change existed.** That is precisely the "caller 19" scenario
the four seats were worried about, arriving in the wild and landing correctly. It is also the
*second* time this seam gained a caller mid-flight; `f930de86b`'s own commit message says *"the
19th arrived DURING the work"*. **So: never write a call-site count here as a constant. It is a
measurement with a timestamp, and on this tree the timestamp expires in minutes.** Corrected
everywhere it appears: 21, not 20 — 13 strict-door (all naming provenance), 4 declared
inheritance, 4 through the simple doors.

### CORRECTION to `f993554f6`'s own commit message — the landmine did NOT ship in that commit

> **`f993554f6` asserts "Register + landmine in this same commit (owner ruling 2026-07-28
> condition 2)". The REGISTER half is true and verified (`git show f993554f6 --
> register/resilience-self-heal.md | grep -c "THE MERGE IS NOW TWO HALVES"` → 1, so condition
> (2) IS satisfied — the register is the artefact the ruling names). The LANDMINE half is
> FALSE.** Another session's `111e5b817` (18:26:22, an RFC_017 verifier correction) swept my
> `LANDMINES.md` edits into itself four minutes before I committed at 18:30:34. My commit's
> `LANDMINES.md` diff is `11 0` — **eleven lines of THEIR uncommitted `page_build_failed` park
> entries**, and none of my text. So my commit took their passenger while my own content left
> in their commit: the collision ran in both directions at once.

**Nothing is lost and forward-only holds** — the landmine text is at HEAD, correct and synced to
`doc_notes`. What is damaged is only the record: `git log -S"FIXED AT SOURCE 2026-08-08" --
LANDMINES.md` now credits an RFC_017 commit with authoring a provenance landmine.

**What caught it, and it was luck dressed as diligence.** The `pre-commit` pattern check on my
*second* commit flagged "1 line(s) removed from LANDMINES.md, a fleet-wide append-only ledger". I
went to prove the removed line was my own from 40 minutes earlier — and
`git show f993554f6 -- LANDMINES.md | grep -c "<my own text>"` came back **0**. Every other
signal said fine: the file on disk held my text the whole time, both
`landmines-sync.py --apply` runs reported `content changed: 1, orphaned: 0`, the commit succeeded
and listed the file, and `git status` was clean afterwards.

**The check I am adding to my own routine** (and to `WRONG_CALLS.md`, because the tally is the
point): after committing, verify the change is in **your commit**, not merely in the file —

```bash
git show <sha> --numstat -- <path>                        # did the file move at all?
git show <sha> -- <path> | grep -c "<text YOU added>"     # is it YOUR change that moved?
```

A non-zero numstat is **not** evidence: mine was `11 0`, all of it somebody else's. This is the
`git mv` landmine's "verify at HEAD, not at the tree" rule generalised past renames, and it is the
missing half of `a-pathspec-commit-still-takes-a-same-file-passenger` — that rule warns you about
inbound passengers, and would never have told me my own content had *left*.

**Why `LANDMINES.md` specifically is the high-risk file for this:** it is fleet-wide and
append-only, so it is the file most likely to be dirty in several sessions simultaneously. Same
goes for `WRONG_CALLS.md` and `MEMORY*.md`. For those three, commit narrowly and then grep your
own needle out of your own sha.

### Round 5 also APPROVED — and its objections were all ONE objection, which is the useful signal

`5d200313` now carries two approved reports: 17:32:13Z (the split) and 17:49:21Z (the symmetry).
Round 5: **APPROVED, 2 advisory objections counted, none high.** Ten seats; `guardian`,
`prior_art_librarian` and `editquality` all objected, and — read together — **they objected to the
same thing three times**: the claim "exactly two call sites have `agent_type` set and `step_name`
zero" is a Go-source fact the council structurally cannot check, and a miscount fails *quietly*
(an empty `step_name`, not a marked row).

**That convergence is worth more than any single objection, and it is not answerable by asserting
harder.** What answers it is the shape of the failure, and it is worth stating plainly because I
did not say it well enough in the submission: **the `agent_type` half of this seam does not depend
on the census at all** — a missed site writes `agent_type='unattributed'`, which is one SQL query
away, and no count is needed to find it. The `step_name` half genuinely is weaker: a missed site
writes an empty `step_name` visible only in a pod log, and pod logs do not survive a rollout. So
the seats put their finger on the one asymmetry that survives, which is a good showing for the
gate. **A source-scanning test is the obvious next lever and I am deliberately not reaching for
it** — the landmine `a-source-scanning-test-makes-comments-load-bearing` says first-occurrence-wins
scanning would be satisfied by a commented-out `AgentType:`, so it would create a control that
reports clean for the wrong reason. Left as a named question rather than a half-answer.

**`prior_art_librarian` caught a real unverified assertion, and it was mine.** It objected that
*"types.ExecutionContext.RunAgentType … unreachable from actions across the import edge"* was
*"an unfixability claim resting on an import-cycle assertion. Not verified against the actual
package graph in this round."* **Correct — I had written that into the register, the LANDMINES
entry AND a code comment on the strength of reading the type's receiver, never once asking the
compiler.** Three documents resting on one unchecked inference. It took one command:

```bash
go list -f '{{range .Imports}}{{println .}}{{end}}' ./platform/orchestration \
  | grep -x github.com/gqls/agentchassis/platform/orchestration/actions      # found
```

The claim **holds** — `platform/orchestration` does import `actions`, so the reverse edge is a
genuine cycle. But holding is not the point; **it could have come out the other way and I would
have shipped an unfixability claim into three durable documents.** [MEASURED 2026-08-08]

**And the same command answered a question I had not thought to ask, which is the better outcome.**
Running it over *both* packages shows `platform/orchestration/types` is imported by each of them —
so the proposed hoist target for §1a is **confirmed reachable from both consumers** rather than
merely plausible. The next session no longer has to discover whether the shape compiles. That is
the objection paying for itself: I ran the check to defend a claim and came away with a fact that
de-risks the hand-off.

Third objection, cheapest and also fine: *"assumes `LogActionEntryInheritingProvenance` already
exists rather than being newly invented here"* — `git show f993554f6 -- log_action_error.go | grep
-c "^+func LogActionEntryInheritingProvenance"` → **1**, so round 4's commit did ship it and round
5 really was a pure call-site swap.

### Owed, not done: the landmine verifier, and why the bare `--apply` cost me the trigger

Another session filed a landmine at 18:44 (`c7d4af7cc`) that lands squarely on what I had already
done twice: **`landmines-sync.py --apply` consumes the `new`/`changed` status, so a later
`landmines-verify-dispatch.sh` fires for nothing.** Its remedy is to run the wrapper *instead of*
the bare apply, and, if you have already applied, to fire the entry by hand with the slug from the
first sync's `NEEDS_VERIFICATION:` line. I ran the bare `--apply` twice before reading it, so I did
exactly the thing the entry warns about — then used its escape hatch:
`./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#logactionentry-s-merge-fills-a-provenance-you-meant-to-set-and-every-test-in-the'`
→ `CORRELATION_ID=1ffae4bf-521d-462b-bd48-d26a4988a6bd`. **Verdict not read — genuinely
outstanding, and a session picking this lane up should read it before trusting the entry's
numbers.** Worth noticing that the useful thing here was another lane writing its mistake down
within the hour; I would not have known the status was consumable otherwise, because both my syncs
reported success.

## 2026-08-08 (night) — the hardening is LIVE on v1.0.1268, and verifying it broke two of my own instruments

A fresh chassis build rolled. **Both commits are in the binary, pod-proven on both chassis
replicas independently** (`agent-chassis-67ddcc695f-jvfmc`, `-dwsdl`, v1.0.1268, started 18:57Z):

| needle | expect | got |
|---|---|---|
| `Failed to write to agent_error_log` (unchanged — proves the pipeline) | 1 | **1** |
| `recorded as unattributed rather than credited to the running step` (the split) | 1 | **1** |
| `names an agent_type but no step_name` (the symmetry — separates commit 2 from commit 1) | 1 | **1** |
| `provenance_running_agent_type` (the context diagnostic) | 1 | **1** |
| `credited to the walking step` (exists in no version) | 0 | **0** |

The third needle is the load-bearing one: it exists only after `0dc2d71a2`, so it distinguishes
"round 4 shipped" from "both rounds shipped". Without it a build made between my two commits
would have passed.

**PICKING THE PODS was the first trap and the RUNBOOK already warned me.** `-l app=agent-chassis`
is the wrong selector, so I picked by image TAG — and the first two pods I picked were *gone by the
time I exec'd*, one `Succeeded` and one `NotFound`. The fleet spawns ephemeral per-job pods that run
the same image, so a tag census returns dozens of pods that will not exist in thirty seconds.
**Add `--field-selector=status.phase=Running` and prefer the long-lived `agent-chassis-*`
deployment replicas.** Tag histogram over Running pods: 44 on v1.0.1268, 0 on anything else — the
v1.0.1266 straggler I had lined up as a free negative control evaporated mid-check.

### MISSTEP 13 — my anchored needle returned 0 on a binary that contains the string

I checked the sentinel constant with `strings /app/agent-chassis | grep -c "^unattributed$"` and got
**0**, on the same exec where four other needles returned 1. Not a shipping failure — **a broken
instrument.** The Go linker packs string constants into contiguous blobs, so `strings` emits them
concatenated dozens-to-a-line; there is nothing for `^…$` to anchor to. Proven in place: unanchored,
the same binary returns **4**, and one of the lines reads
`…conditions_recordededitorial_referrerscomponents_examinedassembled_page.html…`.

**Why this one is worth a landmine rather than a shrug: it fails in the direction that reads as
"your fix did not ship".** Had I run only that needle I would have concluded the roll missed my
commit, gone looking for a build problem that does not exist, and possibly rebuilt. Filed as a
second failure mode on the existing `strings | grep -c` entry, whose first mode (counts SPELLINGS,
not SITES) bit this same lane on 08-07. **Needle on a distinctive full PHRASE, never a short bare
constant, never anchored.**

### The landmine verifier said NEEDS_HUMAN_REVIEW, and it is structurally guaranteed to for any fresh entry

`1ffae4bf-521d-462b-bd48-d26a4988a6bd` came back **NEEDS_HUMAN_REVIEW**: *"`LogActionEntryInheritingProvenance`
… returns 0 results from the code index (commit 93c57696), meaning either it was never merged to
this branch, was removed, or the entry's new-API section is stale."*

**The verifier is honest — it names all three possibilities and declines to pick.** I can pick,
because I have tools it does not: `git merge-base --is-ancestor 93c576963 f993554f6` → **true**, so
the index is pinned to an **ancestor of the commit that created the symbol** and could not contain
it under any circumstances. `git grep` finds it at HEAD; both pods carry the enclosing code. The
index's newest row is 08-08 08:28Z, about ten hours before my commit.

**The generalisation, which is the part worth keeping: a landmine describing a NEW symbol will
ALWAYS come back `NEEDS_HUMAN_REVIEW` if filed the same day, so that verdict on a fresh entry
carries no information whatsoever.** It is not a signal to weaken the entry. Settle it with
`git grep` at HEAD plus a pod-grep and write the disposition INTO the entry — which I have done,
because the next reader will otherwise meet a scary-looking verdict with no resolution beside it.
This also means my earlier note ("verdict not read — genuinely outstanding") is now discharged.

**And the wrapper lesson from `c7d4af7cc` applied properly this time:** ran
`./scripts/landmines-verify-dispatch.sh` INSTEAD OF the bare `--apply`, and it reported
`content changed: 1 / orphaned: 0` and dispatched in one go.
