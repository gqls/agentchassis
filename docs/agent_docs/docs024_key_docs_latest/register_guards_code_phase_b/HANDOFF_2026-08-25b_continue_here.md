# HANDOFF 2026-08-25b — continue here

**Lane:** `register_guards_code_phase_b` (`bugs_open/288`, the class behind `bugs_closed/225`).
**Supersedes `HANDOFF_2026-08-25_continue_here.md`.**

**State: the mechanism is LIVE, has run for real, and has now been ADOPTED for the first
time. §5d's "single highest-value action" is DONE — `mortgages-stamp-duty` declares 13
facts and Phase 3a has its first sample carrying information. Two morning fixes are still
committed-and-unrolled; the tag is bumped and the release is the owner's to run.**

## 1. ✅ DONE — the roll landed 19:07Z 2026-08-25 on `v1.0.1339`

**Both fixes are LIVE and artefact-verified on both replicas** (three literals present,
`stale_attestation`=5 positive control, nonsense negative 0). Note the tag: the release went out
as **`v1.0.1339`**, not the `v1.0.1338` this lane bumped — another lane's tag. *Read the artefact,
never the tag.* Full table and the post-roll sweep in `bugs_open/288` §5f.

⚠ **Neither fix is PROVEN, and the two reasons differ — do not let §5f's green table imply
otherwise.** `6ad4a8046` has **no live case**: 0 misplaced fleet-wide against a control of 6
correctly placed / 304 facts, because agritec repaired the only instances before the finder
shipped. `bba8a892d` has **no observable surface at all**. Presence + unit test is the whole
claim for both. An induced positive on a live register would close the first and needs an owner
decision; the second needs the `Ambiguous` result field (§3b).

### ~~The one thing owed to the owner~~ (discharged)

**`IMAGE_TAG` is bumped to `v1.0.1338` and committed (`6902665ff`). The release has not
run.** `bba8a892d` (duplicate-value ambiguity guard) and `6ad4a8046` (naming a misplaced
top-level `artifact_check`) are in no running binary — probed with controls, and the pods
started 09:27Z *before* either commit was authored. `v1.0.1337` is what the cluster serves,
and a same-tag re-release re-serves the cached digest, hence the new tag.

```
date; make release redeploy-agents ENVIRONMENT=production REGION=uk001; date
```

Post-roll probe (every pod, controls in the same exec) is in the RUNBOOK's new
"Post-roll probe" section. **Read what it does NOT establish before quoting it** — §3 below.

⚠ Until it rolls, `misplaced_artifact_checks` is absent from every sweep payload for a
**binary** reason, not a data one. Do not read that zero.

## 2. What was done 2026-08-25 (do not redo)

- **LMC `mortgages-stamp-duty` declares 13 facts, applied and live.** `bugs_open/288` §5e.
- **Thirteen, not the seven the sweep proposed.** The six extra are rates: the register
  stores percentages (`2`,`5`,`10`,`12`), the code stores fractions (`0.02`,`0.05`,`0.10`,
  `0.12`). No value probe can match them, and two digits are below the floor of 1000
  anyway. **A binding suggestion is a LOWER BOUND on the work, never its size** — landmine
  filed.
- **`install_fences.py` (LMC) now carries `facts` from the criteria file.** Necessary, not
  cosmetic: its `--apply` rebuilds the whole body, so the fragment installed exactly as our
  own CONTRIB instructed **would have been deleted on their next run**. Landmine filed.
- **The CONTRIB to LMC was corrected in place** — its warning 1 described *mcalc's*
  installer, not theirs (`WRONG_CALLS.md` 15).
- **Three stale register STATUS lines corrected** (CLM-022's "HAS STILL NEVER RUN FOR REAL",
  and the CLM-022 / TL-045 index cells' "Go inert until roll" — TL-045's was 8 days stale).

**Proof of the adoption is 0 → 13, not the row readback**: the same scoped dry run returned
an empty `fact_drift` array before (corr `2bebb885`) and 13 `unreconciled_declaration`
entries after (corr `d4dd59e2`).

## 3. ⚠ TWO THINGS THE NEXT SESSION WILL OTHERWISE OVERCLAIM

**(a) The new distribution cannot estimate the number Phase 3b needs.**
7 `present_in_script` / 6 `not_probed` / **0 `absent`** / 0 `present_in_markup_only`. It is
the first sample with any information in it (agritec's was 24 × `not_probed`). But **this
declaration was authored FROM THE CODE, so `absent` is structurally near-impossible in
it** — and `absent` is exactly what 3b must be armed against. An `absent` can only arise
from a declaration authored from the register alone, or one that has aged past a rebuild.
**7/6/0/0 establishes that the probe fires and discriminates. It is not a distribution.**

**(b) The ambiguity guard has NO behavioural surface, even after the roll.**
`Ambiguous` reaches only the doc_note body (`refresh_evidence_fact_suggest.go:284`), never
a per-site result field; `planFactBindingSuggestions` skips any declaring subject (`:167`),
and agritec — the one register with duplicates — now declares; and the note cooldown is 30
days (`:246`), so all five noted subjects are suppressed to ~2026-09-24. **This corrects
`HANDOFF_2026-08-25`, which implied agritec's nine value-sharing pairs would exercise it.
They cannot.** Post-roll the honest claim is *presence at the binary plus the unit test*.
The cheapest thing that changes it is exposing `Ambiguous` on the per-site result (~10
lines + a test, beside `FactBindingSuggestions` at `refresh_evidence_base_action.go:190`),
which is council-gated and would then be dry-run-provable by inducing a duplicate.

## 4. WHAT IS OWED, in order

1. **The roll** (§1), then the probe, then update §5e and the two register cells' residual
   clauses.
2. **gamesdesign.co.uk — an owner decision, not a task.** Three `fact_binding_suggested`
   notes, all proposing `gd-trials`=10000 (their only fact above the floor). **Deliberately
   not adopted**: their PLANs read `created_by='tool-generator'`, a platform agent that
   rewrites the whole body — measured, 28 of 41 lines on `tool-spawn-rate-balancer` between
   07-29 and 08-21 — so a hand declaration expires at the next rebuild. **No lane directory
   owns gamesdesign.** `drop-rate-simulator` has no `doc_plans` row at all and needs a PLAN
   before it can have a fence. The durable fix is teaching `tool-generator`/`tool-improver`
   to carry `facts` (CLM-021's undelivered third).
3. **Phase 3b still needs an `absent`-bearing sample** (§3a). Candidates that clear the
   floor: mcalc (already declares 13, so its batch is spent), and any site adopting from the
   register rather than from the code. This is a real gap, not a waiting game.
4. **`bugs_open/288` §5.4** — the `improve_tool` arm has still never run in production;
   reachable on 91 of 178 exposed tool pages.
5. **§5.5, the prose half** — untouched; `bugs_closed/093`'s resolution (a second SURFACE
   for the existing scanner, never a second scanner) is the template.
6. **Piece 4, the oracle** — *is the figure RIGHT* — out of scope, behind its own RFC.

## 5. Open with another session

**`bugs_open/387` proposed an opt-in `writer_block_guidance` key in `composeWriterBlock`,
in `refresh_evidence_base_action.go` — this lane's file.** They offered to build it
themselves in a window this lane names, and will not touch the file without a word back.
**Answer given: yes, their option (b).** The functions are disjoint (`composeWriterBlock`
versus the fact-drift / `artifact_check` / suggest paths); the only real hazard is two
sessions holding uncommitted edits to one file at once, since a pathspec commit takes a
same-file passenger. **This lane holds no uncommitted edits to that file as of this
handoff, so the window is open now.**

**UPDATE, same day — DONE, and reviewed.** They committed `c17a18620` (`writer_block_guidance`
in `composeWriterBlock` + `plan_sections_action.go` + a test, register CLM-029, council corr
`0de22385`). Re-read by this lane rather than taken on trust, because a whole-fleet release is
pending on this HEAD:

- **one additive hunk** at `composeWriterBlock` (+19, no deletions), disjoint from every path
  this lane touches. No same-file passenger either way. Their register edits are additive and
  this lane's CLM-022 correction is intact beside their CLM-029.
- **`scripts/verify-head-builds.sh ./platform/orchestration/...` → OK.** That check was owed by
  *this* lane, not theirs: `pinned_sweep` resolves `REF=HEAD` once for the whole release, so the
  release this lane asked for ships their commit too.
- `bba8a892d`, `6ad4a8046` and `c17a18620` are all ancestors of HEAD.

**Closed 2026-08-25: APPROVED at round 2** (trail `0de22385`, 8 approve / 3 advisory / 6 abstain;
round 1's gating objection was the typed-`EvidenceBase` round-trip question, answered by
enumerating all 9 `ParseEvidenceBase` callers and pinning both write paths with round-trip tests).
Re-verified after their follow-up: `14ec48b89` is the **only** commit to
`refresh_evidence_base_action.go` since `c17a18620` and is **comment/blank-only** — checked, not
taken on trust — and `verify-head-builds.sh ./platform/orchestration/...` is still OK at the
current HEAD.

⚠ **FIELD CONTRACT, for anyone editing `composeWriterBlock` in this file:
`writer_block_guidance` is NEGATIVE / PROHIBITIVE guidance ONLY** — stated at the carry site and
in CLM-029, where its relationship to `banned_claims` is reconciled (detective vs preventive;
sites hold both). Do not put positive instructions through it.

⚠ **So v1.0.1338 will carry several lanes' commits under one tag — `bugs_open/249`'s straddling
case. Probe per SERVICE, not per fleet.** Their discriminator (`numeric stand-in placeholder`)
and this lane's three are independent and will settle on the same roll.


## 6. Landmines this lane earned today (both filed, both dispatched to the verifier)

- **A lane's fence installer REBUILDS the `doc_plans` body from its source files**, so a
  key added to the live row by hand is deleted on the next `--apply` — clean exit, no
  error. **Check `created_by` FIRST**: it is usually the installer's own hardcoded literal
  and therefore names the script that will overwrite you. Two files named
  `install_fences.py` behave *oppositely*. Extended the same day with the worse form: the
  writer may be a **platform agent**, which no lane can patch.
- **A `fact_binding_suggested` note proposes only what a value probe can MATCH**, so
  adopting it verbatim declares a fence that looks complete and watches half the tool.
  Reconcile **both directions** against the tool's script before declaring. Do not lower
  the floor — it is measured.

Earlier ones still standing: probe SCRIPT TEXT never the whole page · never tokenize a
`string_agg` of `rendered_html` · the floor is MEASURED · a trailing comma is a list
separator · delete the CALL not just the body · a fixture built to DEMONSTRATE a rule will
not TEST it · grep the file for the word your comment uses.

## 7. Housekeeping

- `825445ae7`'s commit message is **mangled**: backticks in `git commit -m` were executed
  by the shell, blanking two `` `absent` `` tokens so one sentence reads "so is
  structurally near-impossible … and is the number Phase 3b needs". Forward-only, so it
  stands; the intact text is in `bugs_open/288` §5e and the lane NOTES. **Use single
  quotes or a heredoc for a message containing backticks.**
- `b891a67dd`/`0c304c9a6` carry a disclosed same-file passenger in `LANDMINES.md` — the
  `web_admin_console` lane's `POST /c/<token>` entry. Named in the message; not this lane's.

## Register / bugs / docs

`bugs_open/288` §5e (the adoption, the distribution's limit, the writer-dependent trap,
and the §5c correction) · §6 (what is owed) · CLM-022 and TL-045 in the concept register,
both status-corrected today · `WRONG_CALLS.md` 15 · `LANDMINES.md`, two new entries ·
lane NOTES 2026-08-25 session 3 · RUNBOOK "Adopting a `fact_binding_suggested` note".
