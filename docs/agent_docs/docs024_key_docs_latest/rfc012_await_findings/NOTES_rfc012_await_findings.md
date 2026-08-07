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
