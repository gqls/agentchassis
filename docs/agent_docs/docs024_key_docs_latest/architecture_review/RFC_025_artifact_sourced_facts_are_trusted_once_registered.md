# RFC 025 — Artifact- and attestation-sourced evidence facts are trusted once registered, and 72% of the fleet's facts are in that class

**Status: IMPLEMENTED 2026-08-24** (see §11; RATIFIED 2026-08-12, owner; answers in §9) · opened 2026-08-12 by the
`portfolio_positioning` build-out session,
per the explicit self-classification in `bugs_open/161_HANDOFF_2026-07-31_the_evidence_register_ratifies_the_claim_it_was_built_to_catch.md`
("a new shared vocabulary key on a shared mechanism — architecture scope under CLAUDE.md's
seam rules, RFC before code") and per this session's owner ruling that the whole ~150-domain
fleet build-out is blocked until this bug's generalisable fix is live.

## 1. Problem + evidence

### 1.1 The defect, and why it recurs rather than being a one-off

`gamesdesign.co.uk`'s evidence register asserted, from 2026-07-24 to 2026-07-31, that its
drop-rate tools ran "10,000 Monte Carlo trials per query" — a fact with
`source.artifact: "the figure is hard-coded in the shipped drop-rate tool JavaScript"`.
Neither tool contains any randomness; the 10,000 is an input clamp
(`Math.min(val, 10000)`). The false fact was registered, instructed to the writer as
`writer_block`, published on ten live components over five weeks, and never once flagged —
not by the prose-number scan, not by the stat-field audit, not by the persistence-time
claims guard — because **all three consult the same register the false fact lives in**.
Full narrative, timestamps and the specific code paths: `bugs_open/161` (read in full
before acting on this RFC; this document does not restate its evidence, only what follows
from it).

The single false fact is fixed (migration `270_bug161_...sql`, `banned_claims` armed via
`271_bug161_...sql`). What is not fixed is the mechanism gap that let it stand for five
weeks with every gate reporting green: **`platform/orchestration/actions/refresh_evidence_base_action.go:269-271`**

```go
query := factSQLSource(fact)
if query == "" {
    continue // artifact/attested facts are checked for presence, not re-proved
}
```

and the type this reflects, `platform/orchestration/datahelpers/claims.go:58-69`:

```go
// EvidenceSource records where a fact's proof lives. Exactly one field is set:
//   - SQL: live-verifiable — a query whose result is the fact's value (V4
//     freshness re-runs these).
//   - Artifact: code path or URL evidence — checked for presence in the
//     register, not re-proved.
//   - AttestedBy: human word (e.g. "owner, 2026-07-10") — the honest standing
//     of a claim only a person can vouch for.
type EvidenceSource struct {
	SQL        string `json:"sql,omitempty"`
	Artifact   string `json:"artifact,omitempty"`
	AttestedBy string `json:"attested_by,omitempty"`
}
```

Only `SQL`-sourced facts are re-run on the freshness cadence. `Artifact` facts are, by
design and by comment, never re-proved after the day they are registered — the register
records that a human or an audit *once* looked at a code path and typed a sentence about
it, and nothing ever checks that sentence against the code path again.

### 1.2 Measured blast radius (re-run 2026-08-12, unchanged from the bug's 2026-07-31 figure)

```sql
SELECT s.domain, jsonb_array_length(COALESCE(sp.data->'facts','[]'::jsonb)) AS facts,
       (SELECT count(*) FROM jsonb_array_elements(sp.data->'facts') f
          WHERE NOT (f->'source' ? 'query' OR f->'source' ? 'sql')) AS prose_sourced
FROM site_specs sp JOIN sites s ON s.id=sp.site_id
WHERE sp.aspect='evidence_base' AND sp.is_current ORDER BY 2 DESC;
```

**9 registers, 102 facts, 73 (72%) with no SQL source** — i.e. either `artifact`,
`attested_by`, or (a third, undeclared case found while researching this RFC — see §1.3)
`citation`. None of the 73 is ever re-checked by anything. The false `gd-trials` fact was
one specific instance of this general class, not a one-off typo.

This is fleet-wide and structural, not confined to one site: `evidence-chart` (the one
verified-data chart component this platform has, used live on e.g. `oufe.com`'s Thames
Water case study) and the writer's own citation constraints both depend on the register
being trustworthy. This build-out's own plan (`portfolio_positioning/PLAN_2026-08-12_fleet_buildout.md`)
treats graphs and fact-backed copy across ~150 new domains as depending on this register —
which is the direct reason this RFC exists now rather than being deferred further.

### 1.3 A relevant precedent found while researching this RFC, not in the original bug file

`EvidenceSource` (the typed Go struct) has exactly three fields: `SQL`, `Artifact`,
`AttestedBy`. But `refresh_evidence_base_action.go:246-266` special-cases a **fourth**,
untyped source shape:

```go
if src, ok := fact["source"].(map[string]interface{}); ok {
    if _, has := src["citation"]; has {
        entry := refreshCitationFact(ctx, fact, today)
        ...
```

This is the V5 "researched citations" mechanism (`docs/agent_docs/docs024_key_docs_latest/claims_verification/SPEC_V5_researched_citations.md`,
`platform/orchestration/actions/evidence_citations.go`): a fact backed by an external URL
is re-verified by **re-fetching the URL and checking a stored verbatim quote is still
present in it** — the same shape of check `directory_claims.go` uses for the directory
mechanism this build-out's Phase B also depends on (`portfolio_positioning/PLAN_2026-08-12_fleet_buildout.md`
§Phase B). Two things follow from reading it closely:

1. **This platform has already shipped exactly this class of change once** — adding a new,
   independently re-verified source type to the evidence-fact vocabulary — via
   `evidence_citations.go`'s `VerifyAndRegisterCitationsAction`, whose header references
   "the council's debug_historian seat" asking a question about it. That is evidence of a
   **normal council-gate submission**, not a full RFC, for a fact that is structurally the
   same shape of problem this RFC addresses (a source claim that used to be trusted on
   sight and is now independently checked).
2. **It shipped WITHOUT adding a field to the exported `EvidenceSource` struct.** `citation`
   is read directly off the untyped `map[string]interface{}` inside the one function that
   needs to know about it (`refreshOneSiteEvidence`); `ParseEvidenceBase`'s typed path
   (`claims.go:272-292`, used by `claimscan` and the scanning/matching engine) never sees
   it and does not need to — the source type only matters to the *refresh* step, not to
   *matching*. This is a real, working design precedent for shipping a new source-type
   check without touching a shared, cross-package exported symbol.

This materially affects §2's design and §3's alternatives, and is the central open question
in §8 — I have deliberately not resolved it myself; see §8.

## 2. Design — what this RFC proposes

Not a code change; a design plus the governance question in §8. Two independent pieces,
proposed as separately-shippable stages (per the process's bar 3):

### 2.1 Stage 1 (cheap, ships regardless of §8's answer): a staleness NUDGE for prose-only facts

For the facts that structurally **cannot** be machine-verified — `attested_by` facts, which
are a human's word by design ("truth decisions stay human",
`refresh_evidence_base_action.go:20`) — add a `stale_evidence`-style work item on a long
cadence (e.g. 180 days) asking a human to re-look and re-date the attestation. This does not
check anything; it turns silence into a queue, the same shape `stale_evidence` already uses
for SQL facts that have gone stale. No new vocabulary key, no schema change beyond a
scheduled-task row. **This is ordinary council-gate work and does not need this RFC to
proceed** — named here only so it isn't lost as the cheaper companion to stage 2.

### 2.2 Stage 2 (the actual generalisable fix): `artifact_check`, modelled on the `citation` precedent

For facts whose source is `Artifact` (a code path or URL, as opposed to a human's bare
word), add an **optional** verification key, read the same way `citation` is — off the
untyped map inside `refreshOneSiteEvidence`, not as a new field on the exported
`EvidenceSource` struct:

```json
"source": {
  "artifact": "the figure is hard-coded in the shipped drop-rate tool JavaScript",
  "artifact_check": {
    "component_id": "b381f0db-...",
    "pattern": "Math\\.min\\(val,\\s*10000\\)",
    "must_be_present": true
  }
}
```

On the refresh cadence, for a fact carrying `artifact_check`: load the named
`page_components.rendered_html` (or a repo file path, for facts sourced outside a
component), and assert the pattern's presence/absence per `must_be_present`. On mismatch,
the same outcome shape `refreshCitationFact` already produces (`drifted`/`error`), feeding
the same `res.Drifted`/staleness reporting `refreshOneSiteEvidence` already has — no new
reporting mechanism, reuse of the existing one.

**A fact with no `artifact_check` key behaves exactly as it does today** — checked for
presence in the register, not re-proved. This is opt-in, unsafe-default-OFF, in the shape
CLAUDE.md's 2026-08-02 seam ruling (§2) asks new authority on a shared seam to take. It does
not change what happens to any of the 73 currently-prose-sourced facts unless and until
someone adds the key to a specific fact.

**What this does NOT do**: it does not make `artifact_check` mandatory, does not retype
existing `artifact` facts automatically (that is a per-fact, human decision — matching "the
register is human-owned by design"), and does not touch `AttestedBy` facts at all (those
get stage 1's nudge, not a check — a human's word cannot be greped).

## 3. Alternatives considered

**A. Do nothing beyond the single-site patch already shipped.** Ruled out by the measured
blast radius: 72% of all registered facts fleet-wide are in the unverifiable class, the
`gd-trials` case is not rare, and this build-out is about to write ~150 new registers into
the same unmonitored class.

**B. Require a full RFC for the specific fact-level fix too (not just this design
question).** Ruled out on cost, for the same reason RFC_002 §3.B rejected requiring an RFC
for every new check type: `artifact_check` is opt-in, reachable by nothing until a fact
names it, and does not retroactively change what any existing fact means.

**C. Add `Citation`-equivalent typed fields to `EvidenceSource` properly, instead of reading
off the untyped map.** Considered because it is the more conventional Go shape — a typed
struct field is discoverable, self-documenting, and IDE-navigable, where an untyped map key
is not. Rejected as the default proposal (though flagged as a legitimate alternative for the
owner to prefer) because it is the one design that unambiguously **adds an exported symbol
other packages depend on**, which is PROCESS_architecture_review.md's own bullet-2 RFC
trigger, applied literally rather than by analogy. The untyped-map shape is not a
workaround invented for this RFC — it is the shape the codebase already chose for
`citation`, and following it keeps the two mechanisms consistent with each other rather
than introducing a second convention for the same kind of thing.

**D. Force every artifact-sourced fact to become an `attested_by` fact instead of building a
checker — i.e. make a human bless each one and stop pretending code-derived facts are
special.** Rejected: this does not scale to 73 facts and does not address the actual defect,
which is that a *code-derived* claim can be verified and a *human* claim structurally
cannot — collapsing the distinction throws away the one class of fact this RFC can actually
make safer.

## 4. Blast radius, named — derived mechanically

**Binaries linking `platform/orchestration/datahelpers`** (`go list -deps ./cmd/<x>/ | grep
datahelpers`, run 2026-08-12): `agent-chassis`, `claimscan`, `component-render-check`,
`config-key-audit`, `content-creator-agent`, `core-manager`, `test-spawning`,
`verifier-remit-check`, `voicescan`, `workflow-monitor` — **10 binaries link the package
that defines `EvidenceSource`/`EvidenceFact`**.

**Of those, exactly one actually executes `refresh_evidence_base`**: it is registered in
the generic action registry (`platform/orchestration/actions/registry.go:640`, category
`"site"`) that only `agent-chassis`'s orchestration worker dispatches. The others
(`claimscan`, `voicescan`, `verifier-remit-check`, `config-key-audit`) link
`datahelpers` to **read** facts (via `ParseEvidenceBase`'s typed path) for scanning/auditing
purposes — they do not call the refresh action and are unaffected by either stage's change,
since both stages only touch behaviour inside `refreshOneSiteEvidence`, not the typed parse
path `claimscan` etc. use.

**Under design 2.2 (untyped map key)**: the exported `EvidenceSource`/`EvidenceFact` Go
types are unchanged, so every one of the 10 binaries above compiles identically and none of
their read paths sees a new field to handle or ignore. **Under alternative C (typed
field)**: all 10 would still compile (an added `omitempty` field on a JSON struct is
backward-compatible), but the exported-symbol trigger applies regardless of whether
anything downstream actually breaks — which is exactly the "adds" clause the process
document's 2026-08-09 addendum was written to catch.

## 5. Staged rollout plan

1. **This RFC, ratified** (or the owner rules stage 2 doesn't need one — see §8).
2. **Stage 1** (staleness nudge) ships first, independently, through the normal council
   gate — it has no dependency on §8's answer.
3. **Stage 2** (`artifact_check`), once designed per whichever of §2.2/alternative-C the
   owner prefers: implement, unit-test `numberSupported`/refresh behaviour with a
   deliberately mismatched `artifact_check` fact (must fail closed — see §7), mutation-check
   it (revert and confirm the new test fails, per bug 161's own verification section), then
   retype the `gd-trials` fact itself (already textually corrected; this would additionally
   attach a real `artifact_check` pointing at the input-clamp line) as the induced-fault
   canary — a fact that would have caught the ORIGINAL false claim, proving the mechanism on
   the motivating case rather than a synthetic one.
4. **Watch**: after stage 2 ships, the 73-fact prose-sourced count should only fall as
   individual site owners choose to add checks — it is not expected to move on its own,
   and a sudden fleet-wide drop would indicate something auto-migrating facts, which nothing
   in this design does or should.

## 6. Rollback plan

Stage 1 is a scheduled-task row; deleting it undoes it completely, no schema question.

Stage 2 (design 2.2): the new key lives inside the existing `jsonb` `source` object, which
every current reader already treats as "whatever the human wrote, preserved verbatim"
(`refreshOneSiteEvidence`'s own comment: "Work on the generic map so unknown keys ... survive
the rewrite untouched"). Removing the `artifact_check`-reading branch from
`refreshOneSiteEvidence` returns every such fact to today's behaviour (checked for presence,
not re-proved) with **no data migration** — the key simply stops being read, exactly as an
unknown key is ignored today. The previous binary already tolerates the new schema, because
it already tolerates arbitrary unknown keys in this column by design.

## 7. Acceptance evidence

- **Owed, stage 1**: first cadence firing confirmed against a real `attested_by`-only site,
  and confirmation the item routes to a human, not an auto-action (attestation cannot be
  machine-resolved).
- **Owed, stage 2**: the `gd-trials`-shaped induced-fault test (§5.3) — a fact whose
  `artifact_check` pattern is made to NOT match the artifact must fail/flag, not silently
  pass; this must be demonstrated failing BEFORE the check exists (mutation-check) as well
  as passing after.
- **Owed, stage 2**: confirm the verifier fails closed on a fetch/read error (per
  `RFC_017_verifier_registry_fails_open_on_error.md`'s established rule for this codebase —
  an `artifact_check` that cannot read its target must not silently pass it as fine).
- **Owed**: a fleet sweep, post-stage-2, of how many of the 73 prose-sourced facts get a
  real `artifact_check` attached in the following month — this is expected to be slow and
  owner-paced (facts are retyped per-site, not migrated in bulk), and a report of "0 this
  month" is not itself evidence of failure.

## 8. The question for the owner, stated plainly

Following `RFC_002`'s own pattern (§8 there): this is a governance question, not a technical
one, and I have deliberately not resolved it myself.

1. **Does `source.artifact_check` need this RFC's ratification before it can ship, or does
   it qualify for the normal council gate under the 2026-07-29 ruling ("an addition to a
   shared vocabulary needs an RFC only when it changes what the shared mechanism
   GUARANTEES")?** My own reading, stated so it can be argued with rather than assumed: it
   is structurally closer to `citation` (§1.3) — an opt-in, per-fact addition that changes
   nothing for any fact that doesn't name it — than to `attribute_absent`/`attribute_matches`
   (RFC_002's case), which changed an *existing, stated* guarantee ("confirm, never refute")
   for a mechanism that had no such branch before. The evidence-base mechanism's existing
   guarantee is narrower than Tier 2's was: it already says plainly, in its own doc comment,
   that artifact facts are "checked for presence... not re-proved" — i.e. it never promised
   verification for this class, so adding verification does not revoke a promise, it adds
   one where none existed. If that reading is right, this RFC's real job is stage 1 +
   design 2.2, both ordinary council-gate work, and this document's purpose is establishing
   that reading on the record — not blocking the build-out for a multi-round review.
2. **If it does need full ratification: is design 2.2 (untyped map key, matching the
   `citation` precedent) preferred over alternative C (a typed `EvidenceSource` field)?**
   2.2 avoids the exported-symbol trigger and matches existing convention; C is more
   conventional Go and more discoverable to a future reader who doesn't know to grep for
   untyped map keys. Both are safe; this is a maintainability trade-off, not a safety one.
3. **Should stage 1 (the attestation staleness nudge) be decoupled entirely and shipped now,
   independent of this RFC's outcome?** It has no dependency on §8.1's answer and closes
   part of the 72% gap (the `AttestedBy` share) on its own.

**Related**: `bugs_open/161` (the motivating case, full evidence); `bugs_closed/043` (the
audit whose remediation seeded the false fact); `RFC_002` (the precedent this RFC's §8.1
argument is built on, and whose format this document follows); `RFC_017` (fail-closed
verifiers, which stage 2's acceptance evidence must satisfy); concept register
`claims-verification.md` (`CLM-003`/`CLM-014`/`CLM-018`); this session's build-out plan,
`portfolio_positioning/PLAN_2026-08-12_fleet_buildout.md`, which is blocked on this RFC's
resolution per this session's owner ruling.

---

## 9. RATIFIED — the owner's two answers, 2026-08-12

Asked as the two plain questions in §8, directly, the same way `RFC_002`'s owner questions
were put and answered.

**Q1 (§8.1) — does `artifact_check` need full ratification, or does the `citation` precedent
let it take the cheap, normal-council-gate path? → FULL ARCHITECTURE-REVIEW RATIFICATION.**
The owner chose the more cautious path over this RFC's own argued reading (that it's
structurally closer to `citation` than to RFC_002's `attribute_absent`/`attribute_matches`
case). **This RFC's ratification, right now, IS that full review** — the process
(`PROCESS_architecture_review.md`) does not require a separate second round once the owner
has read and ruled on the design in front of them; §§1-7 above are the review. Stage 2
implementation may proceed under this ratified design. Unlike `RFC_002`, this is not
retrospective — no code exists yet, so there is nothing to reconcile after the fact.

**Q2 (§8.2) — untyped map key (matches `citation`) or a typed `EvidenceSource` field? →
UNTYPED, MATCHING `citation`.** Design 2.2 as proposed: `artifact_check` is read off the raw
`map[string]interface{}` inside `refreshOneSiteEvidence`, the same way `citation` already
is; the exported `EvidenceSource`/`EvidenceFact` Go structs gain no new field. Alternative C
(a typed field) is not built.

**Q3 (§8.3, stage 1) was not separately asked** — it has no dependency on §8.1's answer and
proceeds as ordinary council-gate work regardless.

**What proceeds now**: stage 1 (attestation staleness nudge) and stage 2 (`artifact_check`,
untyped-map design) both implement against this ratified RFC. Each stage still goes through
the normal council gate before shipping, per the process's own flow (§"The flow", step 4) —
ratification is not a substitute for that, it is what makes stage 2's submission answerable
without re-litigating whether it needed an RFC at all.

## 10. Implementation status, 2026-08-12

Both stages implemented and committed (`3129cceea`,
`platform/orchestration/actions/refresh_evidence_base_action.go` +
`refresh_evidence_base_rfc025_test.go`). Confirmed by direct inspection, not assumed:
`datahelpers/claims.go` (`EvidenceSource`/`EvidenceFact`) is byte-for-byte unchanged —
`git diff --stat` on that file is empty — satisfying §9 Q2. The §5.3/§7 induced-fault
canary is in: `TestArtifactCheck_MismatchedPatternFlagsDrift` is the `gd-trials` shape (a
fact whose cited pattern is no longer found in its named artefact flags `drifted`, not a
silent pass), plus three explicit fail-closed cases (unresolved component, invalid id,
invalid regex — RFC_017). Full existing evidence/citation/refresh suite re-run clean, no
regressions. `go build`/`go vet` clean. Council-submitted, corr
`9fd94852-ff79-496b-96b5-78a8d3619162` (submitted after the commit — no
`Council-Submitted:` trailer on it; resolve the verdict by correlation, not by the commit
message).

**Council round 1: REVISE** (2026-08-12 18:08). Real findings, not process nitpicks —
addressed rather than argued around: [HIGH] the `artifact_check` regex had no anchoring,
so a bare `10000` pattern would substring-match `100000` — the platform's own documented
landmine, reproduced INSIDE the fix meant to close it; refused at parse time now.
[MEDIUM, 4 reviewers independently] `component_id` wasn't scoped to the fact's own site —
now joined through `pages` and site-scoped, fails closed on cross-site references.
[MEDIUM, architecture seat] the `changed`/stale_evidence-raise decoupling touched the
pre-existing citation branch, outside this RFC's ratified scope — reverted for
citation/sql (exact prior behaviour restored), the new capability scoped to a dedicated,
unit-tested `shouldRaiseStaleEvidence` predicate. One HIGH objection (does write-back
silently delete untyped keys via the typed struct?) was a **false alarm**, verified
directly by reading `writeRefreshedEvidenceBase` — cited as evidence rather than argued
from memory. Commit `9652f4d52`.

**Council round 2: APPROVED** (2026-08-12 20:42, corr `9fd94852-ff79-496b-96b5-78a8d3619162`).
**Note on trailers**: both implementation commits (`3129cceea`, `9652f4d52`) were made
*before* their respective council submissions, so neither carries a `Council-Submitted:`
trailer — the automated `098` coverage report will not join them to this verdict
automatically. This note is the durable, human-readable record of the actual approval;
forward-only rules forbid retroactively amending either commit to add a trailer.

**Still short of `IMPLEMENTED`**: the mechanism is built, tested, and council-approved, but
no real fact anywhere has actually been retyped with a working `artifact_check` yet — the
`gd-trials` fact itself is still a plain `artifact` fact in the live register (its false
claim was already independently corrected; RFC_025 never required migrating it, only
building the mechanism that could). This is deliberately left as future, site-by-site,
human-paced work — not a blocker on the code being considered live. Mark `IMPLEMENTED`
once the next chassis roll carries this code and at least one real fact uses either new
mechanism.

**LIVE 2026-08-13**: chassis `v1.0.1295` (rolled by another lane ~13:53Z) carries this
code — verified per the standing recipe at the artefact, not the tag: build-provenance
stamp `69612d692` probed present in `/proc/1/exe` on BOTH replicas with an absent-sha
control, and both implementation commits (`3129cceea`, `9652f4d52`) confirmed ancestors
of the stamp via `git merge-base --is-ancestor`. The one remaining condition for
`IMPLEMENTED` is a real fact using `artifact_check` or the attestation nudge firing on a
real register (the daily evidence sweep will now exercise stage 1 automatically —
attested facts older than 180 days will begin raising `stale_attestation` items without
further action).

## 11. IMPLEMENTED — the §5.3 canary armed and proven live, 2026-08-24

The bugs_open/161 close-out session executed the §5.3 step eleven days after go-live
(zero facts had used either mechanism in between — measured, with `citation` as the
positive control returning 61+). Migration
`docs/agent_docs/sql_for_agents/585_bug161_arm_artifact_check_canary_on_gd_trials.sql`
(council APPROVED round 1, corr `a9e1a0de-ff04-4193-83dc-ad67f2d4d83d`) attached the
first real `artifact_check` fleet-wide to `gd-trials` itself: pattern
`Math\.min\(val,\s*10000\)` against tool-drop-rate-simulator's hero component
`15f1f798-51fb-41d0-8a07-18148b39a293`, `verified_at` deliberately left at `2026-07-31`
as the demand control. A single-site `evidence-freshness` dispatch the same day
(orch `ac49d67e-3f86-4034-a666-64737ed1b001`, `sites_checked=1`) proved the loop end to
end: per-fact outcome **`fresh`**, tolerance `artifact_check`, `verified_at` bumped
`2026-07-31 → 2026-08-24`, register rewritten by `evidence-refresher` **with the
`artifact_check` key surviving the rewrite** (the §10 round-1 write-back question,
now verified at the artefact rather than by code-reading). The fact that motivated this
RFC is the first fact its mechanism guards, checked daily from now on.

Still open elsewhere, not here: the drift/error branch is unit-proven only (the 585
council round's bug_historian suggests an owner-sanctioned induced-drift test);
stage 2b (`page_name` addressing) and the encoded-figure prose half are tracked in
`bugs_open/288`; stage 1's earliest possible firing is ~2027-01 (every attested fact
is younger than 180 days as of 2026-08-24). Adoption beyond the canary stays per-site
and human-paced by this RFC's own design — **27** artifact-sourced facts (as of
2026-08-24) carry no check yet, and §7's expectation stands: a slow count is not
failure.

## 12. Addendum 2026-09-03 — §11 has three stale claims, and stage 1's population widens

Written by the `bugs_open/161` residual lane while re-verifying that bug. **The RFC's design
and its ratification stand; three factual statements in §11 do not.**

**12.1 Corrections to §11, in place.**
- ~~"stage 2b (`page_name` addressing)"~~ → **stage 2b shipped 2026-08-24 as `subject_key`
  addressing** (`eecd99b0a`; `bugs_open/288` §5.6). `page_name` was never proposed here —
  §2.2's worked example is `component_id` only. The label was coined in `bugs_closed/161`'s
  close-out and copied here; two documents carried it for ten days.
- ~~"the drift/error branch is unit-proven only"~~ → **an induced live drift was proven
  2026-08-24**, the same day this section was written, on mortgagecalculator.co.uk's
  `sdlt-ftb-relief-cap` (`bugs_open/288` §5b): pattern pointed at the expired `625000`,
  `outcome: drifted`, dry run, restored byte-identical.
- ~~"stage 1's earliest possible firing is ~2027-01 (every attested fact is younger than 180
  days)"~~ → **it fired on 2026-09-01**, for boxingonline.com. The claim reasoned from the
  facts that existed on 08-24 and treated the 180-day threshold as the only route to the
  queue; `checkAttestationStaleness` also treats an **undated** fact as due immediately, by
  design and by its own comment. A register gaining an undated attested fact fires the next
  day. **The general form is worth keeping: a projection over a population that is still being
  written to is a statement about today's rows, not about the mechanism.**

**12.2 The `artifact_check` count is unchanged, and now MEASURED rather than sampled.**
`[MEASURED 2026-09-03]` **27** artifact-sourced facts across 5 sites still carry no check, on
405 facts in 27 registers (was 294 in 19 on 08-24). §7's expectation holds: adoption is slow
and that is not failure. What §7 also owed — *"a fleet sweep of how many"* — had only ever
been answered by ad-hoc SQL; `siteRefreshResult.FactsUnverifiable` now reports it on every
daily pass.

**12.3 Stage 1's producer predicate widens from `attested_by` to every unverifiable fact.**
The residue arm in `refreshOneSiteEvidence` nudged `attested_by` facts and dropped every other
unverifiable fact through `continue` **uncounted** — no counter, no entry, nothing. So the 27
facts in 12.2 were invisible to the one mechanism that could have said "nothing has ever
checked this". They are now counted and nudged on the same cadence, with a `Detail` that names
the shape: for a bare `artifact` fact, attach an `artifact_check`; **for the 12 whose
`artifact` is an external URL (all relojistas), retype to `citation`**, which already
re-fetches and re-checks a verbatim quote every sweep. Item type, key and threshold unchanged.
This is inside stage 1's ratified remit — *"it does not check anything; it turns silence into
a queue"* — applied to the whole silent population rather than one subset of it.

**12.4 A precondition this RFC assumed and did not state: the register has to PARSE.**
`[MEASURED 2026-09-03]` two of 27 registers did not, so no mechanism in this RFC — nor any
claims gate — ran on them at all: `finetuning.uk` (since 08-24) and `noted.co.uk` (since
08-25). One text-valued fact was enough, because `ParseEvidenceBase` decoded `facts` as one
array and every caller reads a parse error as "site not opted in". Fixed at source
(`3f221f99f`): facts decode one at a time and a bad fact costs that fact. Full case and the
before/after control: `bugs_open/456`.

**12.5 The open option this deliberately does NOT take.** Giving `EvidenceFact` a real
text-valued shape is the honest end state — `Value` is a `*float64`, and licences, retention
windows and opening hours are facts a register should be able to hold. It is a
shared-vocabulary addition and belongs in its own round. It is not urgent now: with 12.4 in
place a text fact is skipped and reported rather than catastrophic. **Named here so the next
author finds it as a decision, not as a gap.**
