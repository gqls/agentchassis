# PLAN 2026-08-24 — bugs_open/352, the invented selector

Design, phasing, decisions **and their reasons**. Corrections to the originating brief live here,
marked as corrections. Drafted with `fable` against my own first-hand grounding; every claim in the
brief was re-checked against the tree by the planner and the three load-bearing ones were then
re-verified by me (§7).

---

## 1. Root cause, and the invariant this establishes

**Root cause.** `auditJS` (`internal/adapters/browserrunner/render_audit_action.go:202`) substitutes
`el.tagName` into the `class` field for a class-less element, and the orchestrator's
`contrastSelector` (`write_render_audit_findings_action.go:780`) faithfully composes it into
`TAG.TAG` — a selector the producer invented and no browser matches.

**The invariant.** *Every selector filed at a fixer has been verified, in the same page session that
measured the failure, to select the very element that was measured* —
`Array.prototype.indexOf.call(document.querySelectorAll(sel), el) !== -1` — and carries its match
count. A finding whose selector fails that assertion, or whose selector is a bare tag (site-wide
blast radius), is **counted and reported, never filed and never silently dropped.**

This is deliberately stronger than deleting the fallback. Deleting the fallback fixes the one
defect we found; the in-page assertion makes *any* future composition defect self-disclosing. That
is the difference between "not produced" and "unrepresentable", and it is the estate's own ordering
rule (rank by what makes the bad state unrepresentable).

## 2. The decision that reframed the job

> **CORRECTION to `bugs_open/352`'s own candidate (1), and to the 198 lane's candidate (6) from
> which it came.** Both say the remedy is to omit the class component so the selector reads `h3`
> rather than `H3.H3`, and both name only the dedup interaction as the risk. That is right about the
> producer and **incomplete about the consequence.**
>
> Today `p.P { color:#fff }` matches nothing, so it is **inert and harmless**. Corrected to `p`, it
> matches — and css-patch-agent's own prompt says *"The platform APPENDS your rules to the END of
> the stylesheet"*, one stylesheet per site, so the rule recolours **every paragraph on the site**.
> `P.P` (77) and `A.A` (44) are 121 of the 181 invented selectors: the two commonest cases are the
> two worst possible things to recolour site-wide.
>
> **So the naive fix is a regression, not a fix.** The 198 lane has accepted this and recorded it
> against its own candidate text; it is why the selector must be *scoped* and *verified*, not merely
> lowercased. Found by asking what the corrected state DOES, rather than whether the diagnosis was
> right.

## 3. The five design decisions

### D1 — the `class` field becomes a frozen legacy echo; the truth ships in new fields

The adapter keeps sending `class` **with the tag fallback intact** and adds `selector` (composed and
verified in-page), `matches`, `selector_verified`, plus `summary.selector_scheme = "verified/v1"`
set unconditionally so a clean page still declares its capability.

This is what makes every version-skew cell **inert rather than wrong**. It matters because
browser-runner-adapter and render-audit-adapter share one image (makefile:107–117) but roll
separately from agent-chassis:

| adapter | chassis | behaviour |
|---|---|---|
| old | old | today, unchanged |
| new | old | old chassis reads `class`, ignores unknown fields → **byte-identical to today** |
| old | new | no `selector` → chassis falls back to `contrastSelector(tag,class)` → legacy shape exactly |
| new | new | verified selectors filed; legacy rows protected by D4 |

**Naively deleting `|| el.tagName` is precisely the version that opens the symmetric window**: a new
adapter against the old chassis would make `contrastSelector` emit bare `H3` keys, and the *old*
binary's retraction would then falsely close every legacy `H3.H3` row on each audited page. The echo
removes that cell entirely. It is a compatibility keel with a stated retirement condition, not
timidity.

### D2 — selector composition, in-page, with bounded blast radius

Per failing element, in order:

1. usable class → `TAG.firstToken`, **raw, not CSS-escaped** — byte-identical to today's
   `contrastSelector` output, so the entire classed population keeps its `item_key`. No false
   retraction and no dedup churn for the ~271 non-`X.X` rows. This is the single most important
   property of the whole design.
2. own id → `TAG#` + `CSS.escape(id)` — matches exactly one element.
3. else walk ancestors to the nearest carrying an id or class token → `#anchor TAG` / `.anchor TAG`.
   The ai-agent-orchestration case composes `.differentiators-section H3`, scoped to the section
   rather than every `h3` on the site.
4. else bare `TAG`, verified, with `matches` = its full count.

Then verify in-page and carry `matches` + `selector_verified`. An exotic class token that breaks the
raw arm (1) simply fails verification and is counted — which is why raw-not-escaped is safe there.

**The filer refuses two categories, each with its own result counter** (the `skipped_locked`
doctrine — "not filed" must be visible, never silent): `selector_verified == false` →
`skipped_unverified_selector`; verified but containing no `.` and no `#` →
`skipped_unanchored_selector`.

**Categorical refusal, not a numeric `matches` threshold** — chosen deliberately. The two real
hazards are "selects nothing / the wrong thing" and "selects the whole site", both categorical; a
tunable N is a number nobody owns, and every anchored selector is already bounded by its section.
`matches` still rides the spec and the description so the fixer and any promoter can see blast
radius.

> **The unanchored refusal fixes a hazard that is LIVE TODAY, not only a post-fix one.** A
> whitespace-only `className` is truthy in JS, so it survives the `||` fallback as `" "`, and
> `strings.Fields(" ")` is empty — so `contrastSelector` returns a **bare tag** and the current
> binary files a site-wide selector already. Found by the planner; I had missed it.

### D3 — the 73 open `X.X` rows are CANCELLED by migration, not rekeyed

> **CORRECTION to my own brief.** I argued for rekeying over cancelling because rekeying preserves
> the attempt count the two-strike rule depends on. **That argument is dead.** The two-strike
> counter reads `status IN ('complete','failed') AND created_at > NOW() - INTERVAL '7 days'`
> (`load_work_item_actions.go:1519-1523`, verified by me). The park dates from 2026-08-11, so there
> is no attempt history left for a rekey to preserve — and the attempts that were spent were aimed
> at a selector that could never match, so a fresh cycle for the re-filed successor is *correct*,
> not a loss. This removes the main cost of cancel-over-rekey.

And no truthful rekey exists anyway: the correct new key is **DOM-derived** (the anchor), which SQL
cannot compute, and a bare-tag rekey would hand a parked fixer exactly the site-wide selector D2
exists to refuse.

Cancellation is honest: `cancelled` asserts **withdrawal, not resolution**; every one of these rows
is unactionable *by construction* and that is measurable; `cancelled` is in
`workItemClosedStatuses`, so no retraction path can ever touch them again; and by migration 157 it
frees the dedup slot, so the next render audit re-files the still-failing ones under verified
selectors (**up to a fortnight** — see the correction in §11; this paragraph originally said
"the next weekly audit"). **No index conflict is possible** — a plain `UPDATE` to a terminal status only *leaves*
`idx_swi_dedup`'s partial predicate, and the 42P10 trap belongs to `ON CONFLICT` inference, which
this design never goes near.

### D4 — retraction made shape-tolerant in BOTH directions

- **Alias keys.** `stillFailing` inserts, per finding, both the key from the filing selector **and**
  `workItemKey("contrast_failure", page+"#"+contrastSelector(c.Tag, c.Class))` — the legacy
  composition. For classed findings the two coincide (a harmless duplicate insert); for class-less
  findings the second is the `TAG.TAG` alias, so **a legacy row on a still-failing page can never
  read as resolved.** A genuinely repaired pairing produces no finding, hence neither key, and
  retracts honestly.
- **Scheme guard.** New filings stamp `spec.selector_scheme`; `decide()` returns
  `retractionOutOfScope` for a stamped candidate when `payload.Summary.SelectorScheme == ""`, so an
  old adapter's reply cannot retract a new-shape row. The candidate's raw `Spec` is already carried
  on `auditRetractionCandidate` for exactly this kind of judgement
  (`work_item_retraction.go:74-80`, whose own comment cites the design-audit path reading
  `spec.audit_source`) — so this reuses an existing seam rather than adding a query.
- **Reason rewording.** The current string over-claims: *"this pairing is no longer below its
  contrast threshold"* asserts something about the **element** when only the **selector pairing** was
  observed. It becomes *"…the pairing keyed by this selector no longer reproduces"*, which stays
  true under anchor churn.

**With D1 + D4, every ordering of the three deployables is inert.** That is the only real answer to
"ordering alone cannot fix the symmetric window" — the tolerance lives in the code, not in a
sequence a human has to get right.

### D5 — locked components: narrowed, and stated without overclaiming

Lock-check tokens are derived from the **filing selector** — its `.`/`#` tokens — via a
`selectorLockTokens()` helper, falling back to `strings.Fields(c.Class)` for old-shape findings.
Substring containment **stays**: the existing comment's conservatism ("a false skip costs one
unfiled finding; a false file edits a locked component's look") is correct, and the defect was the
*input* (a bare tag), not the containment.

Anchoring **narrows** the class-less-inside-locked hole rather than provably closing it. The
residual case is a locked component whose entire subtree carries no class at all — and that case
defeats **today's** check too (uppercase `"H3"` does not substring-match lowercase `<h3` markup), so
the change cannot regress it. The false-*skip* arm (single-letter tag substrings like `"P"` matching
any capital P in locked markup) ends, because a bare tag is never again the token.

### Rejected: converging on `contrast_check.go`'s `describe()`

My brief asked whether the two probes should share one helper. **No** — and the reason is the same
property that makes D2 safe. `describe` lowercases the tag and joins *all* class tokens, so adopting
it would change the `item_key` of **every classed finding** (`H2.card-title` →
`h2.card-title.muted`), converting a 73-row transition into a whole-population one for zero
measurement benefit. The two probes also serve different contracts: `describe` labels a check detail
for a human; `auditJS`'s selector is a dedup key **and** a patch target with a live keyed
population. What they now share is the *principle* — never invent a token, verify in-page.
`contrastMathsJS` stays byte-shared as before.

## 4. My one correction to the planner: the golden test must not become self-proving

`TestAuditJSComposition` (`contrast_check_test.go:236-281`) byte-pins `auditJS` against
`testdata/audit_js_golden_2026-08-22.txt`, **and** pins that golden's sha256 to
`preRefactorSHA = 4ec6cb73…` — a digest taken from `b32aa9cd9~1`, i.e. from a commit *other than the
one under review*. Its own comment says why: *"a golden silently regenerated from the post-refactor
code would otherwise prove the refactor against itself."*

The plan as drafted says to repoint the test at a new golden and its new sha256. **That would
discard the anti-self-proving property**, because both artefacts would then come from my own new
code, and the planner's stated answer (the two-artefact rule) does not cover it — two artefacts from
one source are one artefact.

**Instead the guarantee changes shape, from IDENTITY to CONFINEMENT.** The 2026-08-22 golden and its
historical digest **stay**, and the assertions become:

1. `sha256(golden) == preRefactorSHA` — unchanged, still anchored outside this commit.
2. `contrastMathsJS` is a byte-substring of **both** the golden and the new `auditJS` — the shared
   maths fragment, which I am not touching, keeps the original join-point guarantee in full.
3. **Confinement:** compute the common prefix and common suffix of `auditJS` and the golden; assert
   the differing middle on the *golden* side is exactly the old push block, and on the *new* side
   contains `indexOf.call(nodes,el)`. A single contiguous divergence region, pinned by content on
   both sides.
4. `strings.Count(auditJS, "function effBG") == 1` — kept.

Why this is mutation-resistant: re-adding the tag fallback into the selector path, or dropping the
verification, either changes the confinement region's content (assertion 3 fails) or splits the
divergence into two regions (the prefix/suffix arithmetic fails). And the digest that could not be
forged from this commit is still doing its job.

Note the legacy echo (D1) keeps `||el.tagName` **in the file**, so a naive
`!strings.Contains(auditJS, "||el.tagName")` tripwire would be wrong. The confinement assertion is
what distinguishes "the echo still exists" from "the selector still derives from it".

## 5. Edit list — 8 edits (council-submittable)

| # | file | operation |
|---|---|---|
| 1 | `internal/adapters/browserrunner/render_audit_action.go` | JS selector composition + verification; 4 new fields on 3 structs; unverified counter |
| 2 | `internal/adapters/browserrunner/contrast_check_test.go` | golden test → confinement (§4), not repointing |
| 3 | `internal/adapters/browserrunner/render_audit_action_test.go` | copy-through, counter, and an env-gated real-Chromium composition test |
| 4 | `platform/orchestration/actions/write_render_audit_findings_action.go` | selector preference + 2 categorical refusals + alias keys + scheme guard + reason rewording + `selectorLockTokens` |
| 5 | `platform/orchestration/actions/write_render_audit_findings_test.go` | 6 new tests, all pinned on effects; every existing test passes unchanged |
| 6 | `docs/agent_docs/sql_for_agents/586_retire_invented_contrast_selectors_HOLD.sql` (+`_ROLLBACK`, +`_VERIFY`) | withdraw the 73 |
| 7 | `scripts/render_audit.py` | mirror the composition so the local probe stops misleading investigations (it misled 211) |
| 8 | prose: `bugs_open/352` arm banner + decision record; `bugs_open/211` §4 dated line; register `VIZ-013`/`WII-016` key-shape statement |

**Not an RFC and not an optional-key-budget matter:** no action config key is added
(`WriteRenderAuditFindingsInputSpec` untouched); the new reply fields are opt-in *by presence* with
today's behaviour as the absent default; and the consumer set is enumerated — exactly one action
reads them. That is RFC_022's three conditions met, with the consumers named rather than asserted.

## 6. Migration 586 — and why `_HOLD` is for tidiness, not safety

`_HOLD` so it is applied by hand after `git merge-base --is-ancestor` against **both** services'
build-provenance stamps. But every misordering is already inert by D1/D4 — applying early merely
lets the old producer re-fill freed slots with `X.X` rows, which a re-run sweeps. Full SQL, the
`DO`/`RAISE` premise guard (a verify block of bare `SELECT`s cannot stop a `COMMIT`), and the
guarded `_ROLLBACK` are in the migration file; the predicate is
`item_key ~ '#([A-Z][A-Z0-9]*)\.\1$'`, whose uppercase backreference cannot match a real kebab-case
class. The one theoretical false positive — a site genuinely using `class="H3"` on an `<h3>`, 352's
own candidate-4 caution — is why `_VERIFY` lists the distinct matched keys for an eyeball pass
before applying.

## 7. What is measured to prove it worked — each could come out otherwise

Producer-side, after both rolls and one rotation:

- fresh rows matching the `X.X` regex → **must be 0**; any row is the defect alive.
- for a sampled fresh row, `document.querySelectorAll(spec->>'selector')` on the live page returns
  ≥1 **and includes** an element whose computed pairing is the filed one → could return 0, which
  would refute the in-page verification claim outright.
- **False-retraction control:** rows closed since the roll whose `item_key` matches the `X.X` regex,
  on sites with a post-roll audit, **before 586 is applied** → must be 0. Any row is a false
  completion and refutes D4. This is the control that matters most, because it is the one that would
  catch me being wrong about the alias keys.

Live-artefact-side:

- after css-patch-agent completes a new anchored item, the appended rule's selector must match >0
  elements on the **served** page — the exact measurement that read 0 for dartsonline's `H3.H3` —
  and a single-selector re-measure shows the ratio moved, **or does not, which is arm 2's territory
  and is why arm 2 stays open.**
- after 586: the previously-open 73 all `cancelled` with `result.cancelled_by='migration_586'` and
  **none** newly `complete`. A `complete` appearing would mean the retraction path closed them
  first, refuting the ordering claim.

⚠ If I quote any `render_audit.py` figure as evidence, it must say whether `--sitemap` was passed
(homepage-only understates ~100×) and must discount `overImage` rows (~8.5%, the probe's own guess).
Both are LANDMINES entries on this footprint.

## 8. Scoped out, with reasons

- **Arm 2 — the appended rule losing on source order.** Different defect, different remedy; stays
  open on 352 with an arm banner. Sketch for the bug file: css-patch-agent's workflow gains a
  measurable precondition — grep the editable stylesheet for a declaration governing the filed
  selector's property; if the offending declaration lives in page-level component CSS emitted after
  it, refuse and park with a `parked_by` marker (198's `mark_base_unsafe` shape) rather than append a
  rule that cannot win; and completion consults the spec's own `acceptance_test` at the
  `checks.GetVerifier`/`verifyBeforeComplete` choke point, which
  `write_audit_findings_verifier_join_test.go:85` confirms **nothing reads today**. Not designed
  further here.
- `run_checks_action.go:1140` — its fallback names a **component** for attribution, never a filed
  CSS selector. Out of scope; the 352 file's own read agrees.
- `contrast_check.go`'s `describe` — already truthful; extending verification to it is a separate,
  evidence-free improvement.
- Unifying the two probes on one `describeJS` — rejected in §3 on key-shape grounds.
- A numeric `matches` threshold — rejected in favour of categorical refusal plus a carried count.
- A skew-window guard that fingerprints `class == tag` — rejected: that is 352's candidate (4),
  guessing at intent from a lossy string, in another coat. The re-runnable 586 sweeps the same
  stragglers without it.

## 9. Risks, and which a reviewer raises first

1. **Guaranteed first objection (guardian): "a migration closes 73 findings inside a bug fix."**
   Answer: `cancelled` asserts withdrawal, not resolution; each row's selector provably matches
   nothing — that *is* the bug; re-detection is automatic and weekly; the prior status is preserved
   in `result` with a guarded rollback; and the alternative is a standing population of parked rows
   whose specs aim fixers at impossible selectors.
2. **Anchor churn** — a renamed section class re-keys a class-less finding: the old row retracts
   (with the reworded, still-true reason) while the successor files in the same transaction. State
   converges to one correct open row; a reviewer may still call the interim reason weak.
3. **The JS is untestable ungated.** The golden confinement pin is the review-time guard; the
   behavioural test needs `BROWSER_RUNNER_IT=1`. Named plainly rather than papered over.
4. **Keeping the lying `class` field** — it is the skew keel (D1), recorded with a retirement
   condition rather than left as a puzzle.
5. **`skipped_unanchored_selector` silently drops real failures.** It is counted, not silent — but a
   reviewer should push on whether an unanchored finding deserves a different item type rather than a
   skip. My answer: not in this change; it needs its own remedy and inventing one here would be the
   same over-reach as arm 2.

---

## 10. COUNCIL VERDICT — APPROVED, round 1 (`acadbe8b-f131-4d4b-b4de-5b61f0898f93`)

*"approved with 4 advisory objection(s) — none high-severity"*, 2026-08-24 13:37Z. Twelve seats
reported; three objected on record without blocking (`bug_historian`, `guidelines`, `guardian`,
`prior_art_librarian`). **Every medium objection was answered with a measurement, and one of them
corrected me.** Recorded here rather than summarised away, because the objections are the useful
part of an approval.

| seat | objection | severity | how it was closed |
|---|---|---|---|
| `bug_historian` | "the plan ASSUMES an old chassis drops unknown JSON keys; nothing proves the decode path is not strict. The whole 4-cell skew argument depends on it." | medium | **VERIFIED, not assumed.** `write_render_audit_findings_action.go:785` uses plain `json.Unmarshal`, lenient by default. The only three `DisallowUnknownFields` calls in the tree are `provocation_gate_action.go:549`, `provocation_generator_action.go:238` and `internal/tools-api/gripper/prompt.go:146` — none on this path. |
| `prior_art_librarian` | "no evidence the render audit runs on a live schedule, so 'the next weekly audit re-files them' is unshown." | medium | **RIGHT, AND IT CORRECTED ME — see §11.** |
| `guidelines` | "nested-field addition to an already-declared object, READ by a workflow step, so the response-body exemption does not apply: it MUST be named in the seam's concept-register entry in the commit that ships it. No edit does." | medium | **Owed and now paid.** Register **VIZ-016** (the selector contract + the shared `item_key` shape, per owner ruling 2026-08-02 §1) and **WII-016** (why alias keys were needed), plus the index row. |
| `guardian` | "586 frees the dedup slot so the next audit re-files under verified selectors — but only if BOTH images have rolled. State the gate explicitly rather than bundling it as one more edit." | medium | The migration now carries the gate in its header with the per-service provenance commands and says plainly that early application produces churn, not corruption — a tidiness gate, not a safety one. |
| `bug_historian` | "the new counters are the fail-loud mechanism, but the plan never says who READS them — same shape as 'a cap on a read path reports a backlog as nothing to do'." | medium | **Accepted as a real residual and NOT closed.** The counters ride the action's result map and its log line; no automated consumer reads them today. That is honest but weak, and it is recorded in §12 as owed rather than claimed. |
| `reuse_agent` | "two selector-composition schemes now live in one package; a future reader could converge them and reproduce the hazard this plan avoided." | low | **LANDMINES entry added** — converging them re-keys ~271 unrelated rows into false completions, and the divergence is stated as deliberate and permanent in VIZ-016. |
| `debug_historian` | "586's pre-check regex is a hand-written mirror of the producer's composition, not derived from it." | low | Flagged in the migration header as needing re-derivation if the composition changes again. |
| `architecture` | "this is the SECOND retrofit of alias-key handling into a retraction path because `item_key` embeds a derived value; a third is the trigger to generalise." | low | Recorded in **WII-016** as the trigger condition. Deliberately not built — inventing a "finding identity versioning" abstraction inside a bug patch is the same over-reach as folding in arm 2, and the seat said so itself. |

## 11. CORRECTION — "weekly" was wrong; the window is a FORTNIGHT

> **CORRECTED 2026-08-24, prompted by the council's `prior_art_librarian` seat.** §3 D3, the
> submission and the first draft of migration 586 all said the withdrawn rows would be re-filed by
> *"the next weekly audit"*. **Measured, and it is not weekly.** Of the 13 affected sites, **all 13**
> had a `contrast_failure` filed within 14 days, but only **3 within 7 days** and **2 within 3**;
> the oldest last-audit is 2026-08-10. The honest statement is **"up to a fortnight"**, and the
> migration, the bug-file banner and the withdrawal reason string all now say that.
>
> The safety argument survives intact — every affected site *is* re-audited, and the withdrawn row
> records its own prior status — but the interval I quoted was an assumption wearing a measurement's
> clothes. I had verified that audits *run* (8 distinct days in 21) and let that stand in for *how
> often each site is covered*, which is a different question. **The seat asked for the check I had
> skipped, not for reassurance.**

## 12. What actually shipped, versus the plan

- **Migration is `587`, not `586`.** Another lane took 586 while this was in the council round —
  exactly the staleness the RUNBOOK §8 warns about, arriving inside one session.
- **Edit 2 landed as designed** (confinement, not repointing) and is **mutation-proven**: removing
  the in-page verification, the legacy echo, or `selector:sel` each fails a named assertion.
  ⚠ One of those three guards was **vacuous on first writing** — it matched my own explanatory
  comment rather than the code. See `WRONG_CALLS.md` and NOTES.
- **Still owed, and not claimed as done:** a consumer for `skipped_unverified_selector` /
  `skipped_unanchored_selector` (the `bug_historian` objection above); applying `587` after both
  rolls; and the post-roll measurements in §7, none of which can be taken until the images ship.
- **Arm 2 remains open** and `bugs_open/352` stays in `/bugs_open/` with a banner saying so.
