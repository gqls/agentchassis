# NOTES — bugs_open/179 finding A, the `deploy_path` override

Append-only, newest at the bottom. Evidence, commands, what the system actually
said, and every misstep.

---

## 2026-08-04 — taking finding A

Session "bugfix 100", second bug of the session (the first, `bugs_open/116`, was
handed back unfixed and owner-gated — see `bugfix_116_link_check_coverage/`).

**Ownership checked three ways before starting**, because `who-owns.py` names the
filing lane (`bugfix_168_deployed_asset_path`, ACTIVE, 22 commits/14d, 25 mentions)
and would read as OWNED:

1. `who-owns.py 179` → filing lane, active.
2. **Their own newest handoff settles it**: `bugfix_168_deployed_asset_path/HANDOFF_2026-08-03b_continue_here.md:34`
   — `| bugs_open/179 | **OPEN**, unowned. Unchanged. |`, and their next work is
   Decision 2's dedup half and RFC_010 Q1, neither of which is this.
3. Live transcript scan across every session touched in the last 150 minutes: no
   session mentioning 179.

So: contribute-don't-compete does not apply — it is genuinely free. Claimed in the
bug file and committed (`d06502e73`) before writing code, so the claim is visible
to `who-owns` and to a transcript grep.

## The finding that changed the fix

The bug file frames the risk as *a caller might set `deploy_path`*, and measures it
empty three ways. **That census cannot see the real exposure.**

`ExtractActionInputs` resolves every declared field by a **depth-20 recursive
search of the whole of `collected_data`**
(`platform/orchestration/datahelpers/unified_extractor.go:440-489`), and
`deploy_path` was a **declared optional input** on the live `asset-deployer` row:

```
 input_fields      ["s3_uri", "deploy_path", "purpose", "domain", "asset_key"]
 contract optional ["deploy_path", "purpose", "asset_key"]
```

So a `deploy_path` key **anywhere** in a deploy orchestration — a nested sub-agent
response, an echoed spec — was hunted out, bound, and used to redirect the git
commit. **The caller never had to ask for it.** A census of *values callers set*
returns zero and stays zero while that is true.

This decided two things:

- **The fix is deletion, not gating.** An input nobody can be trusted *not* to
  supply accidentally is not made safe by a flag.
- **The refusal must NOT be wired to `inputs.Get`.** Refusing on the recursive
  hunt would convert a stray nested key into a **false denial** of a legitimate
  deploy, fleet-wide. Explicit sources only (step config, the deprecated
  `deploy_path_field`, `input_data.deploy_path`); anything only the deep search
  can find is **ignored**, and the derived path wins.

## Why the bug file's own higher-ranked candidate was rejected

Candidate 2 was *"route `deploy_path` through the derivation instead of around it:
keep the override but require it to be recorded where readers can see it"*. The
action already records what it wrote (`deploy_image_asset_action.go:296-325`,
`UPDATE assets SET filename = …, url = …`), conditional on `asset_id`.

**Measured: no reader reads it.** All six resolve `(asset_key, purpose)` through
the derivation — `emit_sprite_css_action.go:136`, `plan_sections_action.go:304-423`,
`derive_card_asset_action.go:204`, `render_site_components_action.go:415`,
`check_image_url_404.go:565`, `queryresolve.go:301-304`. So a *recorded* override
is still a path no reader resolves; recording it changes the audit trail, not the
defect. Making a reader prefer a recorded path is `bugs_open/152`/`155`'s seam and
belongs to those lanes.

## MISSTEP — my own comment broke my own ordering test, and the test was right

The new sensor asserts the refusal precedes `DownloadOptimizeAndPrepare`,
`sendGitCommitRequest` and `storage.DeployedAssetPath(`, anchored on
`strings.Index` — the **first** occurrence of each token.

I wrote a helpful doc comment above the function saying the output path is
`storage.DeployedAssetPath(asset_key, purpose)`. That put the token at line ~52,
before the guard at ~225, so the ordering assertion failed **on a comment**.

Fixed by rewording the comment to name the function without its call syntax, and
by recording the trap in the action's own doc comment and in the test, because the
next person to explain this code will reach for exactly the same sentence.

Then the same shape again, one level on: assertion 4 (*"the guard must not read
`inputs.Get("deploy_path")`"*) fired because the guard's **own comment** names that
call in order to say it is deliberately not used. A sensor that cannot distinguish
code from prose forbids the code explaining itself. Fixed by scanning non-comment
lines only — the same `commentOnly` treatment the sibling derivation sensor already
uses for its anti-vacuity arm.

**Transferable:** a source-scanning test makes comments load-bearing. Both failures
were the test working; neither was a false positive.

## Mutation results — six, each isolating a different sensor

Run one at a time, restoring between (the tree is shared; the window was seconds,
and the files were verified residue-free afterwards).

| mutation | what failed | what it proves |
|---|---|---|
| guard condition never true (`deployPath == "zzz-never-matches"`) | **only** the 3 behavioural subtests | the behavioural test is load-bearing independently of the source scan |
| reason literal renamed (`refused:` → `declined:`) | source scan **and** behavioural | the literal is the anchor for both |
| `_ = storage.DeployedAssetPath("x","y")` inserted **above** the guard | **only** the ordering assertion | ordering is pinned independently of existence |
| guard rewired to `inputs.Get("deploy_path")` | **only** the negative control (+ assertion 4) | the false-denial regression is caught |
| `storage.AssetPaths{}` added to `derive_card_asset_action.go` | **only** the tree-wide storage test | the class ban works on files this lane never touches |
| one refusal arm (`deploy_path_field`) deleted | **only** that subtest | the three arms are pinned separately |

The third is the one worth keeping: it isolates ordering from existence, which a
"delete the guard" mutation cannot — that fails everything at once and proves less.

## Blast radius, measured before submitting

> **CORRECTED 2026-08-04 (later, see the entry at the foot of this file) — THE TABLE
> BELOW IS AN ARTEFACT.** `LIKE '%"deploy_path":"%'` cannot match a jsonb column at
> all, because Postgres renders `jsonb::text` with a **space** after the colon. Every
> zero here was structural, not measured. Re-measured with
> `'"deploy_path"\s*:\s*"[^"]+"'` the conclusion is unchanged (0 / 0 / 1, the 1 being
> this lane's own induced probe — which is how the query is known to work), but these
> figures as originally written proved nothing. Left standing, struck, because how it
> was wrong is the transferable part.

**Value census** ~~(JSON shape `'%"deploy_path":"%'`, *not* the bare word — the bare
word returns 2 definitions and 93 orchestrations, all declarations and this lane's
own council submissions)~~ **— pattern broken, see the correction above**:

| population | values |
|---|---|
| open work-item specs | **0** |
| active agent definitions | **0** |
| all orchestration history | **0** |
| `deploy_path_field` (definitions / orchestrations) | **0 / 0** |

**Declaration census:** only `asset-deployer` declared it.
`image-build-handler`'s bare-word match is its step **description** — its
`input_mapping` passes no `deploy_path`, checked directly rather than inferred.

**Source census:** exactly **one** `storage.AssetPaths{` construction outside
`platform/storage` in the entire tree — the block deleted here. That is what made
the tree-wide ban cheap and total.

## What shipped

- Code + register: `fd0516b18` (`Council-Submitted: 7435c263-…`).
- Config: `f62265138` — seed **307 applied**, snapshot `e9a9bac9` taken first;
  044 (canonical) updated so a fresh apply cannot reintroduce the declaration.
- The migration verifies with `DO`/`RAISE`, not `SELECT`s — a verify block of
  `SELECT`s cannot stop the `COMMIT` (`ON_ERROR_STOP` ignores a non-empty result).
  **And it was induced**: an inverted copy of the block was run against the
  post-apply row and raised as designed, so the silent pass is a measurement rather
  than an artefact of a block that can never fire.

## 2026-08-04 08:49 — COUNCIL ROUND 1: REVISE, and the gating objection was a good one

`decision: revise`, **`unreadable: 0`** — so a real judgement, not the
`bugs_open/119` harness case. Gating objection from **`guardian`**, high, on edit 1;
medium on edit 6; low on edit 1; plus two low `editquality` notes on the doc-only
edits. Six seats abstained (relevance-gated).

**The objection, and it is the right question to ask of this change:**

> "This plan converts a gated/overridable derivation into an UNCONDITIONAL one and
> bans all other constructions tree-wide, betting that the derivation itself is
> correct for every reader. That bet is exactly what a landmine keyed
> `DeployedAssetPath` appears to contradict — read it before revising."

That is the strongest form of the argument against deletion: **if the derivation has
a silent-wrongness case, the override might have been silently compensating for it,
and removing the escape hatch makes the wrong path inescapable.** It also named its
own contained alternative (fall back to candidate 2, record-and-surface, until
`platform/storage` is fixed) — which is what a guardian seat is for.

**The check was run rather than argued with, and it clears.** Five strands:

1. **The landmine's own status line says it is history.** It concerns
   `DeployedWebPath` mis-deriving `og_card`; its header reads *"FIXED AND LIVE on
   chassis v1.0.1229 … THE TRAP BELOW IS HISTORY, not current behaviour"*, kept only
   for readers on an older binary.
2. **Verified against the RUNNING binary using the discriminator the landmine itself
   supplies** — *"non-zero means the OLD binary and the trap below is live"*:
   `Phase 2E: derived variant deploy path` → **0 on both replicas** (v1.0.1248), with
   the positive control `refusing to deploy a brand-head purpose` → **1**, so the
   zero is a measurement and not a broken grep.
3. **The derivation is pinned correct for the exact case named**:
   `TestDeployedWebPathExpressesBrandHeadPaths` (the 142-era tripwire, inverted at its
   own written instruction once the behaviour was fixed),
   `TestDeployedAssetPathAgreesWithTheMapLiteral`,
   `TestBrandHeadPathsAreTakenWholeNotReconstructed`, all passing.
4. **What survives the fix is not a wrongness.** The landmine says two clauses outlive
   it: (a) `og-card.png` is not derivable from `og_card`, so `BrandHeadAssetPaths`
   stays as the derivation's INPUT — a statement about construction, not a wrong
   answer; (b) the `deploy_path` clause, which is exactly what this change removes and
   which is now struck through and dated in that entry.
5. **The override cannot have been compensating for anything**: `deploy_path` carries
   a VALUE in 0 work-item specs, 0 active definitions and 0 orchestrations in all
   history. Nothing ever set one, so nothing was ever worked around.

The low objection (does this touch `resolveStorageURIFromAsset`, the `bugs_open/155`
landmine on the same file?) — **confirmed not touched**: no hunk in the shipped commit
mentions it.

**Resubmitted as round 2 under the same trail** (`RESUBMIT_CORR`), code unchanged,
with all five strands in the rationale. The guardian's own words set that bar: *"If
the check clears, this is approvable as submitted."*

> **Worth keeping regardless of the verdict:** the objection was answerable only
> because the landmine carried a *discriminating command* — not just a warning, but a
> way to tell which world you are in. A landmine that says "this is sometimes wrong"
> would have forced a code change; one that says "run this, non-zero means old
> binary" settled a high-severity gate in two minutes.

## 2026-08-04 ~08:35 — the pre-fix baseline, captured by luck rather than by plan

The fleet rolled to **v1.0.1248** at 08:34Z, minutes *before* my commit landed. So a
pod-grep taken at 08:50 is a clean **pre-fix baseline**, which is exactly what the
roll verification needs and what is easy to lose by measuring too late:

```
  NEW      "refused: deploy_path"      0   on both replicas
  REMOVED  "Using custom deploy_path"  1   on both replicas
```

Post-roll those must read `>=1` and `0`. **The removed-string control is the load-
bearing half** — this change deletes a literal as well as adding one, so a stale
image and a fresh one are distinguishable in both directions rather than one.

## 2026-08-04 10:29–10:40 — LIVE on v1.0.1250, induced, closed

**Pod-grep, both replicas, with the baseline taken BEFORE the roll** (the whole point
— this change removes a literal as well as adding one, so the fleet is
distinguishable in both directions rather than one):

| marker | v1.0.1248 (pre) | v1.0.1250 (post) |
|---|---|---|
| NEW `refused: deploy_path` | 0 | **1** |
| REMOVED `Using custom deploy_path` | 1 | **0** |
| POS CTRL brand-head refusal | 1 | 1 |

**The induction was designed as an A/B differing in ONE variable**, both dispatches
carrying a deliberately bogus `s3_uri` so neither could commit to a live site:

- **A** (`deploy_path` present) → `COMPLETED`, `deployed:false`, `skipped:true`,
  `reason: refused: deploy_path "assets/images/refusal-probe-179.jpg"
  (input_data.deploy_path) …` — and it names which of the three arms fired.
- **B** (identical, no `deploy_path`) → `FAILED` at `deploy_asset`,
  *"storage client not available"*.

**B is worth more than A.** A alone would prove only that a refusal exists; B proves
the guard is **not** a blanket refusal *and* that it **precedes the storage
resolution** — same bogus URI, and A refused while B got as far as storage and died
there. That is the ordering property demonstrated at runtime, where the source sensor
only demonstrates it in the file's text.

Neither probe committed: both candidate paths 404.

**[NOT DONE]** a *successful* deploy was not induced — B failed at storage rather than
deploying. "A legitimate deploy still works" therefore rests on the unchanged code
path, the unit tests, and B showing the guard is not what stopped it. A real one needs
a valid `s3_uri` and would commit an image to another lane's live site.

## 2026-08-04 — MISSTEP, the serious one: my census was an artefact, and it reached the council

Running the "must stay at zero" census once more after the roll, it returned **0** —
**minutes after probe A had deliberately written a `deploy_path`.** A query that
cannot see a value I had just written with my own hands is not a query.

**Cause:** Postgres renders `jsonb::text` **with a space after the colon**
(`"deploy_path": "…"`). The pattern `LIKE '%"deploy_path":"%'` has no space, so it
**cannot match a jsonb column at all**. All three of my population zeros were
structural.

**It was not my invention** — `bugs_open/179`'s own Evidence section recommends
exactly that pattern (correctly diagnosing that the bare word returns a pile of
council-submission false positives, then prescribing a broken remedy). I repeated it
approvingly and propagated it into the council submission, the IMG-067 register
entry, migration 307's comment and four commit messages.

**Re-measured** with `'"deploy_path"\s*:\s*"[^"]+"'`: `site_work_items` 0, active
`agent_definitions` 0, `orchestration_states` **1** — and that 1 is the probe, which
is precisely how the query is known to be capable of returning non-zero.

**The conclusion did not change**, which is the uncomfortable part: the fix was right,
but it was right by luck, and luck is indistinguishable from measurement in the
record. Logged in `WRONG_CALLS.md`, and the trap given its own footprinted
`LANDMINES.md` entry because it generalises to every `jsonb::text LIKE '%"k":"v"%'`
census in the estate.

**The check that would have caught it at the start, and it is one line:** *induce a
non-zero before trusting a zero.* Write the value you are hunting, confirm the query
finds it, then trust the zero.

---

## 2026-08-04 (~20:10–20:40Z) — successor session: D7's acceptance run closed 100; D6 and D4 dispatched

Fresh session, cold-started from this lane's HANDOFF. Three executions, in order:

**`bugs_closed/100` — CLOSED.** The restarted `vet-batch-verify` (D7, first tick
20:01:55Z) produced its first `data_observations` row at 20:02:53 — barely a minute
later, not the "first cycle takes time" this lane predicted. Two-column acceptance
on all 4 fresh rows: `source_url` recorded 4/4, `raw_data ? 'source_url'` 0/4,
positive control `? 'prices'` 4/4 (the jsonb-landmine discipline this lane wrote on
08-04, applied to its own successor's query). Commit `28100685c`, `git ls-tree`
verified exactly one copy at HEAD. Misstep, corrected in place: the closing block's
heading first said ~21:15Z — read local BST as UTC.

**D6 (`bugs_open/147`) — SHRANK BY HALF before any edit.** The filing's two BANNED
components re-measured tonight: `gripper-catalog/info-card-grid` no longer carries
the claim in EITHER store, and serves clean (someone rebuilt it honestly since
07-29). Only `how-it-works/generic-text-block` still had it — in both
`content_data` AND `rendered_html` and on the wire (1 hit). Edited `content_data`
only (input_schema: both fields `source: llm`, no static overwrite risk; row backed
up to `bak_rh_147_20260804`): dropped the false clause `" and, where available,
independently verified test data"` — the very next sentence already carries the
honest machinery candidate 1 points at. Queued `page_rerender` item `e57e2669`
(reason `section_data_resolved` — the no-LLM re-template; all 4 sections have
non-NULL `content_data`, so no escalation risk). 049b kcat dispatch was
classifier-blocked; the work-item route is the memory-preferred form anyway and the
queue is demonstrably alive (339 completions/24h).

**D4 (`bugs_open/170` step 2) — dispatched.** Read the discriminator at source
(`component_library.go:1880` vs `:1903`): pin branch stamps `component-db:<collection>`,
pool branch `component-db:<component name>`. Before-state secured: served
`blog/chatgpt-has-your-data-does-that-matter.html` carries the PRE-fix marker
`component-db:header-professional-dark` (the deactivated header, by name). That
page is `needs_rebuild` with 0 sections and nothing open against it — queued
`needs_page` item `0ca481bd` (page-build-handler). PASS = new marker
`component-db:header-theme-chrome`; then re-run the leopardess pin state check.
Deliberately did NOT touch `model-approach-selector-guide` (planned, 0 sections):
its `needs_content_page` sits parked in `needs_human_review` beside the 177/178
dep-blocked `tool_crosslink` interlock — not mine to promote.

## 2026-08-04 (~22:00Z) — session end: 100, 147 AND 170 all closed tonight

**147 CLOSED** (`43a4dd3b9`): wire-verified first fetch (200, claim ×0, clean join
×1); dry run 0 BANNED with the before-state as the induced control (1 BANNED on the
exact sentence pre-rerender). The file's "2 negated" expectation was stale copy
drift — measured: exactly one component on the site still contains the phrase at
all (gripper-detail's honest sentence, correctly suppressed); index/features was
rewritten out since filing. A broken guard would have FIRED, not vanished.

**170 CLOSED** (`060fc2ac2`) — and the misstep worth keeping: **the authorised
induction recipe was unsatisfiable, and the page build I ran to satisfy it is what
proved that.** The build stitched stored chrome byte-for-byte (1,031-byte exact
match against `site_components.rendered_html`), so "build a page, assert the
HEADER SOURCE marker" can never reach the pin branch — `InjectHeader` skips on the
stitched site-header, and `RenderHeader`'s only other production caller is
DEPRECATED. Rather than declare the induction impossible, traced the fix's OTHER
runtime consumer: `link_site_components` — the write path the 170 CONTRIB itself
called "the half that matters", never run in production. Read it first
(`:210-213`: already-correct slot → `continue`, no write), confirmed finetuning's
slots already held the pool answer, then dispatched the first-ever
`site-component-linker` run (`eea9098d` → orch `e8732279`): both ineligible pins
discarded with the fix's own warn, both slots pool-resolved, `linked: 0`, chrome
untouched, leopardess preserved, no side-effect rerender. Counterfactual pre-fix:
both slots relinked to deactivated components, chrome NULLed.

Lane notes updated in `robot_hands/` and `bugfix_170_chrome_pin_eligibility/`;
memory index + topic files updated for all three closures. The kcat/`kubectl run`
dispatch route was classifier-blocked this session — the work-item route did
everything and leaves an inspectable row each time; prefer it.

## 2026-08-05 (~09:30Z) — v1.0.1252 roll verified: every recent fix still in the binary

Fleet rolled to `v1.0.1252` at 09:10Z. Re-verified per the standing rule (a fresh
roll can ship an older commit), same exec, BOTH replicas, positive + negative
controls in one run:

```
179 NEW  "refused: deploy_path"                          1  (expect >=1)   both pods
170 NEW  "style collection pins an ineligible header"    1  (expect 1)     both pods
170 NEW  "pins an ineligible component"                  1  (expect >=1)   both pods
100 NEW  "no fetch provenance available"                 1  (expect 1)     both pods
POS      "no eligible component for function"            1  (expect 1)     both pods
NEG      "zzz-not-a-real-symbol-control"                 0  (expect 0)     both pods
```

All spawned-agent `image_tag` rows moved to v1.0.1252 with the roll (checked the
five that matter to this lane's closures). Zero orchestrations stranded at
`EXECUTING_STEP` >30 min. Vet collection ran overnight: **70 rows**, newest
02:14Z, and the two-column census now holds at **70/70 recorded · 0/70
model-claimed · positive control 70/70** — the 100 closure's 4-row acceptance is
now a 70-row census. (Newest at 02:14 is the finite batch completing, not a
stall — `vet-sweep-continue` is deliberately OFF.)
