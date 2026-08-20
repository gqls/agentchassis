# HANDOFF 2026-08-20 — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-08-19_continue_here.md`**, which is
one day old and already wrong about the important thing (it says the 327 fix is nobody's; this lane
took it, and it is live).

> ## ▶ START HERE, IN THIS ORDER
> 1. **`bugs_open/327`** (slug `a_partial_spec_write_silently_shrinks_the_brief_the_writer_reads` —
>    **the number is AMBIGUOUS**, another lane filed a different 327 the same day). Read the status
>    banner and the two council sections.
> 2. **`README_where_we_are.md`, the 2026-08-20 entry** — there is **one decision waiting on the
>    owner** and it is stated there in plain terms.
> 3. **This file's "Next work"**.
>
> **One-line state:** stage 2 unchanged and still undispatched; the writer-brief auditor is built and
> registered (CQ-025); the `327` platform defect is **fixed and LIVE on `v1.0.1319`** with the bug
> file still open because the damage is unrepaired; the council said REVISE twice and the second
> objection is escalated to the owner rather than resubmitted.

## What is live, and how that was established

> ### ⚠ UPDATED 2026-08-20 ~17:00Z — a fresh build landed: **`v1.0.1320`**, and the fleet is **MIXED**
> `7 pods on v1.0.1320, 1 on v1.0.1319` — the first time this lane has actually observed the mixed
> state the corrected recipe below warns about, so **any single-pod answer right now is unreliable**
> and the uniformity check is not a formality.
> **1320 carries the fix**, probed with full controls on a `v1.0.1320` pod: `merged_keys` (new
> literal) **1**, `formatted_len` (pre-existing) **1**, an impossible string **0**. One literal per
> call — a compound `grep` over `/proc/1/exe` times out mid-answer and a partial result reads like a
> whole one.
> **1320 changes nothing else this lane depends on**: `git log` since my fix over
> `site_spec_actions.go`, `format_content_direction.go`, `section_editor_actions.go`,
> `ai_actions.go` and `voicestyle/` returns only my own commit.
> **`copy-editor` is intact.** Its row was updated `16:08:07Z` — along with **198 other agents in the
> same minute**, i.e. the release stamping, not a change to this agent. Migration 462's settings are
> where they always were: ⚠ `max_tokens: 32000` is at
> `config.ai_service.max_tokens`, **not** `config.max_tokens`, and the three-edit budget is **prose
> inside the prompt**, not a `max_edits` key. Querying the paths that sound right returned two empty
> strings and I briefly read that as "the budget is gone". **It was the query, not the system** — the
> lane's most-repeated failure, again.
> **`327` still cannot close:** `[MEASURED ~17:00Z]` **zero `content_direction` writes** since the fix
> went live, so the three fragment briefs are untouched and the repair remains unobserved on a real
> write.

**`v1.0.1319` carries the fix.** Binary-probed, never inferred: `merged_keys` (a literal only the fix
introduces) = **1**; `formatted_len` (pre-existing) = **1**; an impossible string = **0**. Compare
`v1.0.1317` eleven hours earlier: `0 / 1` — same probe, different answer, so it discriminates.

⚠ **The obvious selector samples the wrong machines.** `-l app=agent-chassis` matches **2** pods
while **75 run the image** (70 labelled `dynamic-agent`). Filter on the image:
```bash
kubectl -n ai-persona-system get pods \
  -o jsonpath='{range .items[*]}{.spec.containers[0].image}{"\n"}{end}' \
  | grep 'agent-chassis:' | sort | uniq -c      # >1 line == MIXED fleet, stop
```
Then one `grep -ac <literal> /proc/1/exe` at a time — it takes tens of seconds each and a compound
command times out mid-answer, which reads as a complete result. A council reviewer caught this
defect in my published recipe; I had it wrong.

## What this lane did on 2026-08-20

- **Took the `327` platform fix** (nobody else had; `who-owns.py` still says no owner). `formatted`
  is built from `merged`; the renderer sorts keys at every level. `c9a71388f`, plus round-2 controls.
- **Wrote the tests that did not exist** for either function — driving the real action through
  `sqlmock`, because a formatter-only test passes in the broken world too.
- **Council: REVISE, twice.** Round 1's HIGH was right and produced a real control. Round 2's HIGH
  is a judgement about consequence and is **escalated to the owner, not resubmitted** — per
  CLAUDE.md's rule for seats disagreeing (`architecture` approved this as a point fix; `compliance`
  called it HIGH).
- **Measured the thing my own argument rested on.** Exemplar transfer: 52 of 60 phrases reached a
  prompt, **3** appeared verbatim in output, 18 appearances total — against **409** for the single
  *mandated* tagline. ⚠ The 60 were the alphabetically first of 194, so it is indicative, not a rate.
- **Found, and verified in SQL rather than trusting my own tool:** `identity.key_differentiators` —
  this lane's own "lead" register carrier — is **absent on 19 of 25 sites**. Sound advice, largely
  unexercised. (`finetuning.uk`, whom we advised to use it, does have it: 949 chars, fleet's largest.)

## Next work, in the order that closes doors

1. **The owner's decision** (`README_where_we_are`, 2026-08-20): ship the repair as-is, or gate it
   default-OFF. ⚠ **Option B has decayed since it was written** — the fix is now live, so a gate is
   a change to live behaviour rather than a pre-ship choice, and default-OFF would re-disable a
   working repair. Say so before implementing anything.
2. **`327` closes on an artefact, not a probe.** One real `content_direction` write on any affected
   site, then `audit_writer_brief.py <domain>` showing zero non-empty dropped keys. **No write has
   happened since the roll**, so nothing is restored yet and nothing is observed. The three sites are
   other lanes'.
3. **`ai-agent-orchestration.com`'s `example_phrases`** — the payload the compliance seat objected
   to. Their call on their own site config; they have been told twice and offered the edit. **Do not
   write replacement copy** (owner ruling 2026-08-06: the framework writes the content).
4. **A fourth run on `ai-agent-orchestration.com/index`** — unchanged and still the cheapest open
   question about stage 2 itself. ⚠ Deliberately deferred today: that site's brief is about to change
   materially, so editorial output measured against today's brief would not be comparable with
   tomorrow's. Reconsider once its spec is settled.
5. **Dispatch** — wiring `content-quality-auditor` findings to `copy-editor`. Unchanged, held behind (4).
6. **Should the detector be scheduled?** Still an owner/architecture call, unchanged from 08-19.

## Standing cautions (fresh first)

- **Mutation testing bit me twice in one session.** A mutation that FAILS can fail for the wrong
  reason (I changed a call's argument but not its position, and got a one-key failure instead of the
  real six-key one); and a mutation that prints NOTHING may never have been applied (a `HEAD~4`
  revert resolved to a commit that already had the fix, and an empty grep is indistinguishable from
  a control that does not fire). **Revert from a pinned sha, and assert the mutation applied before
  reading its result.** Both in `WRONG_CALLS.md`.
- **Read a council verdict BY CORRELATION.** The newest `doc_notes` row returned another lane's
  verdict; that is the documented trap and I walked into it.
  `SELECT body FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report';`
- **A `090` run's `UNVERIFIABLE` with populated `code_requests` is the loop ordering its next
  bundle, not a verdict.** Read `current_step`, not the outcome field.
- **Do not diff two briefs.** They are rendered in a random key order before `v1.0.1319` and the
  three fragment sites still carry pre-fix text, so a text diff reports ~100% changed either way.
  Compare label presence and phrase position.
- Everything under 08-19's cautions still holds: never size a brief with `length(data::text)`;
  `splitlines()` is not `split("\n")`; an apostrophe is not a quote mark.

## The five living docs

- **PLAN** — untouched; nothing in the plan changed.
- **NOTES** — the 08-20 entries: taking the fix, the council rounds, both measurements.
- **README_where_we_are** — 08-20 entry carries the owner decision.
- **SUMMARY series** — 08-12 · 08-14 · 08-15 · 08-17 · **08-19 (newest)**. ⚠ **Still no new summary,
  and that is deliberate on the second day running.** The five headings would now differ — "the
  fault was never in the writer" has become "and part of the brief never reached the writer either,
  and that half is fixed" — so **the next session should probably write one**, once (2) resolves and
  a real write proves the repair. Writing it before that would record a belief, not a milestone.
- **this HANDOFF.**

**Tooling this lane owns:** `gate_stage2_edit.py` · **`audit_writer_brief.py`** (`--self-test`,
`--fleet`, `--transfer`) · `count_negation_tells.py` · `loanandmortgagecalculator_couk/gate_page_links.py`.

**Platform code this lane now owns:** the `content_direction` derivation in
`site_spec_actions.go`, `datahelpers/format_content_direction.go`, and
`platform/orchestration/actions/site_spec_formatted_from_merged_test.go` (6 tests, all
mutation-proven). **Migrations:** `447`, `462`. None written today.
