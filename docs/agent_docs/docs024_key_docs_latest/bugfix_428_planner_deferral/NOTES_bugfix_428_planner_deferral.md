# NOTES — bugfix_428_planner_deferral (append-only, newest at the bottom)

2026-09-02. Picked up bugs_open/428 (already filed and diagnosed by `gap_planner`)
at the user's request, while already mid-way through bugs_open/427. Read the full
bug file first. Ran `who-owns.py 428` — flagged OWNED/recently active, but only on
the filing session's own commits; `ListAgents` showed `gap planner` idle (filed and
moved on) and a new session `site design planner` (busy, name-collision risk).
Messaged both before touching anything.

Re-verified §4's "13 rows" claim rather than trusting it: a plain `ILIKE
'%entity-page%' OR ILIKE '%entity-directory%'` on the deferred-verdict summaries
returns 16, not 13. Hand-checked the three extras: two are genuine keyword false
positives (one about catalogue depth mentioning "entity-page volume" in passing;
one about imagery compliance on an entity-page template that already EXISTS and is
live on 8 pages — a different defect entirely, not a never-planned-role instance).
The third was a real match. So 13 stands; the discrepancy is exactly the kind of
"keyword census over-counts, hand verification narrows it" the estate's own memory
warns about — recorded in the bug file rather than silently correcting past it.

`site design planner` replied: different mechanism entirely (DES-003/006's
composition/layout resolver, not the page-role planner) — pure name collision, no
overlap. `gap_planner` confirmed no one is working a fix, go ahead.

**The load-bearing finding of this whole bug**, made by tracing the ACTUAL
mechanism behind the "13 deferred verdicts" rather than accepting "the detector is
not the missing piece; the consumer is" at face value: queried
`spec->>'filing_mode'` on every one of the 13/16 rows and got `'record'` on all of
them, e.g. boxingonline's own `e3c2b440-c006-40ec-be7a-88d0b689ed1e`
(`created_by='offer-analysis'`, `handler_agent=''`,
`routed_handler='page-build-handler'`). That is
`write_audit_findings_action.go`'s RFC_056 (read the file's own "WHY IT EXISTS"
comment in full) — a deliberate owner-ruled circuit breaker built 2026-08-25,
eight days before this bug, specifically because an earlier auto-dispatch of this
same finding class (an LLM-audit-seat "aspirational improvement") destroyed live
content. Found the motivating incident: `bugs_closed/238`, finetuning.uk, five
`<img src="">` shipped live plus an unrecoverable rewrite of the case-study copy
underneath them. Confirmed the connection is not circumstantial — the action's own
comment names exactly this shape ("keep the seats — they are the site acceptance
council — but stop the rewrites").

This means bug 428's own §6 candidate 2 ("wire the existing detector to a
dispatcher... lowest-risk, no planner-prompt change") was proposing to rebuild the
exact promoter RFC_056 was built to remove, for the exact finding class it targets.
Corrected the bug file in place (struck through, not deleted, per the estate's
correction convention) rather than letting the next reader build it as described.
Flagged to `gap_planner`, who independently re-verified against the same source
(the action's own header comment, plus the incident citation) and agreed — both of
us escalated to our respective users rather than either building around it
unilaterally. `gap_planner` separately caught a second, independent reason
candidate 2 wouldn't have worked as stated even without the RFC_056 problem: the
`(item_type, status)` pair alone is 1,284 rows, not 13 — any real fix has to filter
to the specific shape both sessions hand-checked, not the general pair. Also
caught, while investigating: bug 428's own §5 had cited `bugs_open/206`'s PRE-FIX
state (`directory-listing` never shipped live) as if still current — 206 closed
2026-08-08, `directory-build-handler` is live — but `business_intel` covers only 3
verticals, none matching any of the 13 affected sites, so dispatching would hit a
live builder with nothing to build from regardless.

Put the choice to the user rather than deciding unilaterally, given the stakes:
build the two safe/independent pieces (prompt formatting + a softer license
tightening) and a human-reviewed release surface, explicitly not an automated
dispatcher. Approved.

Found the exact live `plan_site` prompt text via direct SQL pull (27KB — see
RUNBOOK for the query that avoids dumping the whole thing to the terminal) and the
precedent migration (`640_build_site_planner_prompt_subject_rule_HOLD.sql`) for
editing this exact agent's prompt safely — anchor-and-replace with a pre-flight
drift check and a post-write verify, both aborting rather than guessing. Confirmed
`toJSON` is an EXISTING, already-live template function
(`data_helpers.go:RenderPromptTemplate`'s funcMap) before proposing to use it — no
Go change, no image roll needed, so migration 687 is not a `_HOLD`.

Built the release-surface backend (`HandleReleaseRecordVerdict`,
a `filing_mode` filter on `HandleListWorkItems`) with a dedicated per-predicate
mutation-tested guard suite. Also built the frontend button + filter — could not
run an actual frontend build/lint (no `node`/`npm` in this environment) so verified
by careful manual read of brace/paren balance and cross-checking every new field
access (`selectedItem.spec?.filing_mode` etc.) against how the backend actually
serializes `spec` (confirmed: `json.Unmarshal`'d server-side in both
`HandleListWorkItems` and `HandleGetWorkItem`, so the frontend sees a real parsed
object, not a JSON string needing a second parse).

Dry-ran migration 687 against the live database in a transaction that rolled back
instead of committing, before ever running it for real — confirmed `UPDATE 1` and
both guard blocks passed. Applied for real afterward; verified at the artefact
(pulled the live prompt text again, confirmed both new strings present) rather
than trusting the migration's own `COMMIT` as proof.

All three pieces submitted to council review and committed with
`Council-Submitted:` trailers (not yet resolved as of this note — check the bug
file's status section for current verdicts).

---

2026-09-03, session "428". Resumed the lane. `who-owns.py 428` said OWNED/active but only on the
filing sessions' own commits; `ListAgents` showed `gap planner` idle with an away-summary. Messaged
five lanes before touching anything (`gap planner`, `bugs_open/427`, `gamedesign.uk`,
`designblog.co.uk`, `bugs_open/450`) and all five confirmed clean working copies on
`v3_site_actions.go`, `listing_item_sources.go`, `tool_item_sources.go` and the planner prompt row.
That was worth doing: two of the five came back with findings that changed the design.

**Found migration 687 was NOT in `schema_migrations`** — only `agent_definitions_bak_687` and the
live prompt text proved it had been applied. Verified both, then `--record-only`'d it. A ledger gap
on an applied file is one scoped `--apply` away from a replay.

**The finding that reshaped the whole phase came from the `gamedesign.uk` lane, and I verified it
first-hand before building on it rather than accepting the report.** They reported that
`validate_site_plan` silently drops new children of a section index. Re-read at the orchestration
row (`9fe9660e-7272-4f51-b968-2ff769738086`): `plan_site` 9 pages, `validate_plan` 4, `blog-post`
present in the LLM's page_types and absent from the validated ones, `capability_gaps_emitted` 0.
Also confirmed `site_specs` IS in `collected_data` at validate time (4 recommended types) — which
the check needed and which I would otherwise have had to assume.

**Consequence: I had the design wrong before that message.** My first shape compared
`recommended_page_types` against the validated plan — one snapshot. That check is SILENT on
gamedesign's case, because the type WAS planned. Three snapshots, not one, and the pre-pass snapshot
must sit before Pass A's union as well as Pass C (their caution; verified — it is taken immediately
after the pages array is parsed, ~70 lines before `reconcilePlanWithRealised`).

**Two missteps of my own, both logged in `WRONG_CALLS.md`:**

1. A mutation that "failed" only because it stopped compiling. Deleting the
   `omissionDroppedInValidation` case left the constant unreferenced; `go test` printed `FAIL` for
   the package and I was one step from recording a kill. No test ran. Re-ran as `case false:`,
   which compiles and kills exactly one test. **Check for `--- FAIL: Test` lines, not for `FAIL`.**
2. I wrote the mutation protocol from the mutations' INTENT before running them, predicted one
   failure for the `Live()` mutation and measured three. Corrected the prediction rather than
   narrowing the tests — the two extra failures are the honest ones.

**`090` run: UNVERIFIABLE, stopped by iteration-cap** (run `e2e35519`), the same failure mode as
`d6d350ec` on this bug. It refuted nothing and confirmed one citation supporting the claim. It named
three things it lacked; all three are now verified first-hand and the substitution is declared in
`bugs_open/428` §13.7 per the 2026-07-31 ruling. Worth recording that its third gap was a **good
methodological catch** — its bundle showed `page-build-handler`'s step config rather than
`build-site-planner`'s, and it refused to link them by name-match.

**The frontend defect was found by building the thing, not by reasoning about it.** The Review &
Release button renders on `filing_mode==='record'` alone, while the endpoint requires
`routed_handler` and `routed_status`. My rows are the first record verdicts with no route, so the
button would have 404'd. Fixed, and verified by running `make build-dashboard` under a throwaway tag
and grepping the served bundle with a must-be-present and a must-be-absent control, then deleting
the image. No `node`/`npm` in this environment, but docker is there, which the earlier session's
"could only verify by careful manual read" note did not consider.

**Migration 748 held, not applied**, and rehearsed apply→rollback in one doomed transaction: md5
`369ec3eb…` → `581f9766…` → `369ec3eb…`, byte-exact. Both guards INDUCED rather than assumed (double
apply raises "already applied"; a deliberately damaged rule 20 raises "730/731 damaged"). Live row
re-read afterwards: unchanged.

**Owner ruling arrived mid-session from the `site_delivery_and_editor` lane — "guides should be a
type of their own"** — after two lanes had misrouted it. Recorded, not started, and deliberately not
folded into this bug: the additive half is inert but BLD-030 is now a consumer of the page-type
vocabulary, so a new type nothing may recommend would read as never-planned everywhere. Needs an
owner sequencing call.
