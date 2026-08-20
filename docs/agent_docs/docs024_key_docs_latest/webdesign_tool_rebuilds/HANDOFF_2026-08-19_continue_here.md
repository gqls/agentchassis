# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-19 09:05Z. Supersedes `HANDOFF_2026-08-17_continue_here.md`.

**UPDATED 2026-08-20 17:35Z. 23 of 63 replaced AND live-confirmed at the served bytes; 40 ported
slots remain. Read NOTES 17:05Z first: the #22 retire race was LOST (a background watcher's delivery
ran SIX HOURS late and the oklch page served BOTH tools all afternoon — repaired, and "attend" is now
defined in the RUNBOOK as in-turn polling, never a watcher). Also: `bugs_open/339` was rehomed to
this lane and its seam fix is COMMITTED (`e3dee9243`, TL-048, `Council-Submitted: 2ff9e215…`) —
READ THAT VERDICT and act on a REVISE/REJECTED. Next tool: #24 `tool-social-card` (defect analysis
in NOTES 17:35Z entry), then regex-tester, jwt-inspector, token-calculator, clip-path.**
The owner has ruled: framework ownership of all 63 (see "THE PATH TO 63/63"), so there is no open
decision — the work is repetitions of a routine recipe. Newest evidence and per-tool controls: NOTES,
entries 2026-08-20 07:15Z onward. ⚠ Also read `bugs_open/336` (filed by this lane 07:25Z): migration
494 armed a config key the running binary declares on the WRONG action's spec and every page-stamping
workflow failed validation for ~13 minutes; service was restored by running the 315 lane's rollback,
and re-arming 494 before the one-line spec fix has ROLLED will reproduce the outage.
**Phase A remainder after mesh-gradient (smallest first): oklch-picker, aria-builder, social-card,
regex-tester, jwt-inspector, token-calculator, clip-path — then the ≥8 KB Phase B line.**

*(Superseded tally from 2026-08-19 21:30Z: 14 of 63, #16 building — kept below for the recipe and
rulings, which are unchanged.)*

Read: this file → `PLAN_2026-08-15_…` (design + three owner rulings + two corrections) →
`RUNBOOK_…` (every command) → `NOTES_…` (evidence, newest at bottom) →
`SUMMARY_2026-08-19_the_framework_owns_the_fix.md` (the read-aloud account) → `architecture_review/RFC_036` + `bugs_open/303` + `bugs_open/315`.

## The recipe — PROVEN, 14 tools, and now routine

1. **Read the LIVE tool's `<script>`** (fetch with `?cb=<epoch>` so you do not warm the edge with the
   page you are about to replace) and write the spec from its behaviour. Describe *intent* where the
   ported version is defective; do not add features.
2. **File with all THREE gates as pre-asserts inside the transaction**, plus the serial throttle:
   - `idx_cc_tool_function_unique` — fleet-wide on `function`, `WHERE component_level='tool' AND forked_from IS NULL AND is_active`
   - `content_components_name_key` — **`UNIQUE(name)`, NO predicate**; the generator derives
     `name = '<function>-<domain-slug>'`, so a REBUILD needs the old row **renamed** (never deleted)
   - the generator's `already_exists` probe — per-site, joins `page_components`, no `build_status` filter
   - **also assert there is no OPEN `page_rerender` on the target page** (added 2026-08-19): an
     already-queued assemble can fire inside the 60–100 s between build and retire, and it dedupes the
     generator's own rerender away
   - ~~always include the `bugs_open/303` build-constraint sentence~~ **DROP IT — 303 is fixed and live
     (see below). It now costs something and buys nothing.**
3. **Grade the RUN** — `current_step='complete'`, `page_adopted='true'`, no `already_exists`, no
   `__step_error`. A failed build reports the ITEM as `complete` with `error` NULL.
4. **Grade the COMPONENT by locating the MECHANISM, not by grepping wording you imagined.** Enumerate
   its element ids and check the machinery each requirement needs. `{{\.` must be 0. Read the JS.
5. **Retire IMMEDIATELY** — record `page_components` id + length + md5 BEFORE filing; guarded UPDATE;
   md5 after must equal md5 before; assert one deployed slot remains **and that it is the new one**.
6. **Grade at the served page, cache-busted**, `http=200` first, with a negative and positive control,
   and include a NEGATIVE that only the old version could satisfy (an old element id) — that is what
   rules out a stale render.

## Where this stands `[MEASURED 2026-08-19 21:30Z]`

| | |
|---|---|
| chassis | `v1.0.1316`, digest `2d0d3def…`, pods 17:13Z, stamp **`07eeba4a1`** — probed on BOTH replicas 20:52Z with a positive and a negative literal. It carries RFC_036 §9.3 (`e24bc9c0f`) and the `bugs_open/303` fix (`6d962bcf8`); check either with `git merge-base --is-ancestor <commit> 07eeba4a1` |
| **replaced, live, graded at the served bytes** | **14** — aspect-ratio, markdown-tables, html-minifier, svg-optimizer, sri-generator, smooth-shadow, json-cleaner, seo-injector, noise-generator, rls-architect, css-variables, prompt-permutator, **ab-test-calculator**, **blob-maker** |
| owner-approved on sight | aspect-ratio, markdown-tables, html-minifier, svg-optimizer |
| **remaining** | **49** — **1** parked (`tool-meme-generator`, only because it is a Phase B rich app), 13 external-script, the rest split simple / larger |
| **Track 1 (generator contract)** | **DONE** — migration 481 applied, now proven FIVE times (every build since has produced contract-rule elements no brief mentioned) |
| **Tracks 2 + 3 (checker, handler)** | **RE-SCOPED 2026-08-19 — much smaller than this file used to say. Read the Track 2 section before writing any Go: two of the six classes are already detected by `tool-auditor`, one of them MUST NOT be checked statically, and the queue defect that made it look futile turned out to be a fixed keying bug.** |

**DB check for the tally, so it is never guessed:**
`SELECT pc.build_status, count(*) FROM pages p JOIN page_components pc ON pc.page_id=p.id
 WHERE p.site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND pc.slot_name='ported-page'
   AND p.name LIKE 'tool-%' GROUP BY 1;` → `removed` = replaced, `deployed` = remaining.

**Five of the first nine ported tools examined were measurably broken in production** — two with a
checkbox whose implementation sat inside its own comment, one corrupting `pre`/`script` content, one
silently disabling truncation on a bad input, one destroying the user's output on a parse error.
**Every one of the five examined since has been broken too** (#11 rls-architect emitted a dropdown
label where it promised SQL; #12 css-variables shipped `/* Auto-generated grey */` on a hardcoded hex;
#13 prompt-permutator would expand ten groups of five into ~9.8 M strings and freeze the tab; #14
ab-test-calculator printed a significance verdict from `NaN`; #15 blob-maker bound no listener to any
of its three controls). On this file's own earlier figure plus those five, the rate is **10 of 14**.
**None was visible from the page.** Reading the live script before writing each brief is what found all
of them, and it costs ~2 minutes per tool. That has become the lane's main product.

**A named subclass, now 4 of the 14: the tool asserts something untrue ABOUT ITSELF** — prose where SQL
was promised, a comment claiming a colour was generated, a control wired to nothing, and a verdict
computed from a value that is not a number. These defeat casual checking because the output has the
shape of a working tool's output. 481's rules cannot reach it (they govern behaviour, not whether the
tool's account of itself is true), so it is a standing step when writing a brief: **read what the tool
SAYS about itself and check the code does it.**

## Three platform defects this lane filed (all still OPEN, none this lane's to fix)

- ~~**`architecture_review/RFC_036`** — nobody has built it, which is why 2 tools are parked.~~
  **BUILT, COUNCIL-APPROVED AND LIVE, 2026-08-19 — and this lane demand-proved it the same evening.**
  The `bugfix_311` lane shipped §9.3 (`create_tool_component_action.go:249-285`, commit `e24bc9c0f`,
  council `ceae30f2` APPROVED round 1): when a library entry claims the function, the generator sets
  `forked_from` on the new row, so `idx_cc_tool_function_unique` no longer fires. Their notice is
  `webdesign_tool_rebuilds/CONTRIB_2026-08-19_from_311_lane_RFC_036_s9_3_is_BUILT_and_LIVE_phase_D_is_unblocked.md`.
  **First real exercise: `tool-ab-test-calculator`, 20:57Z — PASS.** New row
  `8a315006-2170-4ba7-b517-4abaf9619e45` carries `forked_from='8c9a6e06…'`, `save_tool` completed with
  no 23505, and the library row's html/schema md5s and `updated_at` are unchanged. Outcome reported
  into `docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/NOTES_311_fix.md`.
  Known follow-up they flagged and it did NOT bite: `deploy_tool_to_site`'s fork lookup (RFC_036 §11)
  is not on this route.
- ~~**`bugs_open/303`** — the tool-birth guard counts tag SUBSTRINGS…~~ **FIXED AND IN THE RUNNING
  BINARY (verified 2026-08-19 21:20Z).** Commit `6d962bcf8` (08-18 19:50Z) replaced raw
  `strings.Count` tag counting with one markup-context scanner (`platform/content/markup_balance.go`)
  at five surfaces including `toolTemplateValid`, the tool-birth guard. It is an ancestor of the live
  stamp `07eeba4a1`, and its own calibration recorded 264 `component_versions` rows, 26 flagged by both
  old and new, **0 flips**. **CONSEQUENCE FOR EVERY FUTURE BRIEF: drop the BUILD CONSTRAINT sentence.**
  It is not merely redundant now — it tells the generator to hide tag names behind string
  concatenation, which is exactly what makes a real truncation harder to detect. The bug FILE is still
  in `bugs_open/`; moving it belongs to the lane that fixed it, not to this one.
- **`bugs_open/315`** — `pages.deployed_at` is stamped whether or not bytes are written (measured stale
  on three pages, including two that were serving correctly). One page was skipped by four completed
  rerenders and published itself ~6 h later.

## TRACK 1 IS DONE — the generator now carries the quality rules (migration 481, APPLIED 2026-08-19)

**Owner ruling 2026-08-19: fixes must extend into the framework.** Six defect classes this lane had
been fixing one brief at a time are now **rules 15-20 of the tool-generator's own contract**
(`agent_definitions` → `{workflow,steps,generate_tool_html,config,prompt_template}`, 2,865 → 4,340
chars, config-only, live immediately, snapshot-backed rollback in `481_..._ROLLBACK.sql`):

| rule | class | measured on |
|---|---|---|
| 15 | never report success you have not verified | copy buttons calling `showCopied()` unconditionally |
| 16 | real listeners, never inline `onclick` / a global | 42 of the 55 remaining ported tools |
| 17 | no `alert()` / `confirm()` / `prompt()` | 25 of 55 |
| 18 | validate parsed values; never silently do nothing | `json-cleaner`'s NaN limit |
| 19 | errors never destroy the user's output | `json-cleaner` writing a parse error into the output box |
| 20 | show input/output sizes on any transformer | `html-minifier` legitimately changing 2.9% and looking broken |

**CONSEQUENCE FOR EVERY FUTURE BRIEF: do not repeat these. Delete them.** A requirement stated in both
the contract and the brief is the drift surface — the day they disagree nobody knows which is
authoritative. **A brief should say only what is true of THIS tool.** The first brief written this way
is `tool-noise-generator` (item `49536dc1`), which is also the live test of 481.

**Still owed (Tracks 2 and 3) — RE-SCOPED 2026-08-19 after reading the package and measuring the
queue. Read this before writing Go; the four findings each remove work.**

1. **Detection of these classes already exists and is not static — it is `tool-auditor`.** On this site
   it had already filed, on 2026-08-15: *"Division by zero is certain when visitors is 0. The z-score
   formula uses `p * (1 - p) / n` under a square root…"* — the exact defect this lane found by hand in
   `tool-ab-test-calculator` four days later. Also a missing copy button, an unguarded negative input,
   `input` and `change` both bound, hardcoded hex. A new detector for the behavioural classes is a
   SECOND opinion, not a first.
2. **`check_tool_health.go` is already driven and already covers both populations** (real forks AND the
   63 ported instances, widened 2026-08-15 under `bugs_open/281`): `design-discovery-agent`'s `checks`
   array holds `tool_health`, `tool_acceptance`, `tool_acceptance_due` `[MEASURED at the live row]`.
   Its 10 sub-checks are structural (script/style present, `@media`, bare hex, `fetch`/external `src`,
   tool-doc header) and cover **none** of rules 15-20. **Extending it needs no config change**, which
   is also how a new check avoids being born undriven.
3. **Rule 18 must NOT be checked statically, and #14 is the measured counterexample.** The obvious test
   — calls `parseInt`/`parseFloat` and contains no `isNaN`/`Number.isFinite` — **fails the
   best-validated tool on the site**: the ab-test rebuild scores `isNaN` = 0 because its guard is
   `/^\d+$/.test(raw)` BEFORE the parse, which is strictly stronger. Rules 15, 19 and 20 are semantic
   in the same way. **Only 16 (inline `on*=` attribute) and 17 (`alert(`/`confirm(`/`prompt(`) are
   literal and decidable** — they cannot be satisfied by a better implementation.
   ⚠ And do not quietly overturn `check_dead_controls.go`'s stated boundary: *"`<button>` with no
   handler is NOT judged statically (JS binds at runtime); the post-hydration equivalent lives in the
   Tier-4 browser tier."* That is the blob-maker defect exactly, and its owner declined it on purpose.
4. **Track 3 is already answered by the package, and the queue is NOT dead — I nearly recorded that it
   was.** `remit.go` exists for precisely this question (`bugs_open/077`): `PartitionByRemit` splits the
   population by whether the HANDLER's literal transform would change it, and `CapabilityGapItem` files
   ONE undispatchable `capability_gap` (status `deferred`, empty `handler_agent`, priority 200) for the
   residue instead of items nobody can clear. First measurement said `improve_tool` = **205
   `unresolved`** against 35 complete, which reads as a graveyard. **Row-level look: 20 of them share
   ONE `item_key` (`audit_fix_webdesign.co.uk`, no per-tool discriminator), all born
   "[unresolved after 2 attempts]" although no row with that key ever reached a terminal status — and
   from 17:24 that day the keys change shape to `audit_fix_<domain>_<page_id>` and DO dispatch (3
   complete, 1 failed). `425_tool_auditor_ported_instances.sql` was applied at 17:17:19Z.** The 205 are
   a pre-425 keying scar, not today's behaviour.

   **So the shape to build, if it is built: ONE sub-check pair inside `check_tool_health.go` for rules
   16 and 17 only; per-INSTANCE `item_key`s; forks → `improve_tool` (in-remit, tool-improver rewrites
   `html_template`); ported instances → residue → ONE `capability_gap` per site, never 42
   `ported_tool_fix` rows. State in the header what it does NOT cover, or its silence on rules
   15/18/19/20 will read as a clean bill.**
- ⚠ **Do NOT put the `bugs_open/303` workaround into the contract.** It would bake tag-obfuscation
  into every generated tool. Keep it per-brief, only for tools that emit or manipulate markup, until
  303 is fixed.

## THE PATH TO 63/63 (owner ruling 2026-08-19: framework ownership of ALL 63 tools)

**The audit question is settled and it is NOT extra work.** Reading the live script is step 1 of every
rebuild, so rebuilding all 55 audits all 55 by construction. A separate audit pass would duplicate it.
What IS worth having is the cheap prevalence sweep already run `[MEASURED 2026-08-19]`, as a
prioritiser and a brief-writing aid — across the 55 remaining ported tools:
`onclick=` **42** · `alert(` **25** · `parseInt`/`parseFloat` **14** · `innerHTML =` **18** ·
`localStorage` **3** · a copy button **26** · external `<script src>` **13**.
⚠ **These are PATTERN PREVALENCE, not confirmed defects.** Inline `onclick` is a code-quality smell,
not a bug; `alert()` is a real UX defect; the `parseInt` 14 are candidates for the NaN-guard class that
silently disabled truncation in `json-cleaner`; the 26 with a copy button are candidates for the
lying-copy class the owner reported. **Each still has to be read.** The value of the sweep is that it
says where to look first, and it sets an expectation: the defects found so far are not incidental.

### Phase A — the 18 simple, self-contained tools (<8 KB, no external script)
The proven path, unchanged. ~5–10 min of attention each; wall-clock is queue depth, not work.
Smallest first (the RUNBOOK's "Scope the batch correctly" query orders them). Serial — the item key
enforces it. **Expect roughly half to have a real defect**, on the run rate so far (5 of 9).

### Phase B — the 22 larger self-contained tools (≥8 KB)
Same recipe, longer briefs. **The rich hand-built apps live here** (mind-map, meme studio, logic
architect, micro-CMS, pasteboard) and by the owner's 2026-08-16 ruling they are reimplementations, not
preservations. **Owner's standing instruction: these go LAST and one at a time, each seen at the served
page.** For these the grade is a feature list checked in a browser, not a tag count — a raw-tag count
cannot tell you a mind-map lost its export.

### Phase C — the 13 external-`<script src>` tools
The page is not self-describing: the logic lives in S3 assets the DB-side checks cannot read (TL-032).
So the brief **must** come from the tool's behaviour in a browser, and **the external asset must be
retired with the slot** or the page keeps fetching a file nothing serves. Do these after Phase A has
made the recipe boring, and expect the spec work to dominate.

### Phase D — the 2 blocked tools (`tool-ab-test-calculator`, `tool-meme-generator`)
**Cannot be reached by any amount of lane effort.** They need `RFC_036 §9`: a ~10-line change in
`create_tool_component_action.go` to set `forked_from` when a library entry already claims the
function, then council + a chassis roll. **There is no config-only interim — proved in §9.1**: the
platform's own definition of a library tool (`forked_from IS NULL AND is_active`) is exactly the index
predicate, so "forkable by other sites" and "blocks a rebuild" are the same condition. Do not spend
another cycle looking for a way round it.

### Ordering, and why
A before C before B: Phase A keeps the recipe warm and cheap; Phase C's cost is spec-writing, not
mechanism; Phase B spends the owner's review attention, so it goes last when everything else is known
to work. Phase D runs whenever RFC_036 lands — it is not a sequencing dependency for A/B/C.

### What "done" means, stated now so it cannot drift
**63/63 replaced, each graded at the served bytes with a cache-buster and a negative control.** If
RFC_036 is never built, the honest terminal claim is **"61 of 63, 2 blocked on RFC_036"** — not "done".

## IN FLIGHT RIGHT NOW (2026-08-20 06:55Z) — pick this up first

**#16 `tool-shadow-stacker` is BUILDING.** Item `2121ba49-8b61-402e-95fc-d5ad54062e2a`,
page `9d3333c8-2122-414d-bc98-cb8c73ebfade`, `/tools/shadow-stacker/index.html`.
**Revert handle already recorded:** ported slot `9b3ec013-1d29-4918-90d0-791b62aafae7`,
6,710 chars, md5 `d970feca108b6e2a84e91d23150471ff`.
When it completes: grade the RUN, grade the COMPONENT, then **retire that slot immediately** and watch
the rerender. Its load-bearing requirement is an invariant on the OUTPUT, not on the inputs: **the code
block must only ever contain valid CSS and the preview must never disagree with it** — the ported
version writes `NaNpx` from an empty field, which voids the whole declaration so the preview loses
EVERY layer while the code box still offers the text as CSS. Negative blur and zero layers
(`box-shadow: ;`) are the same class.
**It is also the test of retiring the `bugs_open/303` sentence** (first brief without it). If the build
is refused for unbalanced tags, that inference is wrong and the sentence goes back into the recipe.
⚠ **A watcher was armed but a fresh session will not inherit it** — re-check the item directly:
`SELECT status, now()-created_at AS age, now()-claimed_at AS claimed_for FROM site_work_items WHERE id='2121ba49-8b61-402e-95fc-d5ad54062e2a';`

## Next actions — start here in a fresh session

1. **Finish #16 (above), then Phase A continues with `tool-diff-checker` (6,731 B)** — then
   `tool-touch-target` (6,732), `tool-grid-generator` (6,828), `tool-text-extractor` (6,908),
   `tool-mesh-gradient` (7,052), `tool-oklch-picker` (7,186). Order the rest with the RUNBOOK's "Scope
   the batch correctly" query. Run the six steps of the recipe.
   **Do not file a rebuild you cannot attend** — measured margins between build completion and the
   rerender being claimed are ~45 min, ~2 min, ~26 min, ~96 min, and it was **lost once at 96**; my own
   two retires this round went in at 94 s and 62 s. There is no floor.
2. **Track 2 is now a SMALL, decided piece of work — read the re-scoped section above before writing
   any Go.** Two sub-checks (rules 16 and 17 only) inside `check_tool_health.go`, routed through
   `remit.go`'s `PartitionByRemit` / `CapabilityGapItem`. Do NOT build a rule-18 checker: #14 is the
   measured counterexample. Council gate + roll as usual.
3. **The "tool lies about itself" class is now 4 of 14 and still has no mechanism.** 481's rules cannot
   reach it and a static check would struggle, because verifying a claim against code is closer to the
   Tier-2 acceptance evaluator's job than to a grep. Decide whether it is a checker at all or a line in
   the tool-acceptance criteria. Until then it is a standing step when writing a brief: **read what the
   tool SAYS about itself and check the code does it.**
4. **`tool-meme-generator` is the only tool the platform no longer blocks but this lane has not
   scheduled** — it is a Phase B rich app, so it goes last, one at a time, seen at the served page, by
   the owner's standing instruction. Its library claim is `6ae53f32-be86-4c29-bc52-983c35d23b18`; the
   §9.3 fork path will handle it exactly as it handled the ab-test calculator.
5. **`bugs_open/315` is still open and still the one that can make a grade lie** — `pages.deployed_at`
   is stamped whether or not bytes are written, so compare `last-modified` against the rerender's
   `completed_at` rather than trusting a status.

## Traps that each cost a real cycle (full detail in NOTES / WRONG_CALLS / LANDMINES)

- A written precondition saying "must be empty" is a STOP, not an input to a judgement.
- **Elapsed time comes from the ROW** (`now()-created_at`, `now()-claimed_at`), never the session clock
  or a `min(created_at)` column. I asserted a stall from a misread timestamp **twice**.
- A negative control alone cannot license a negative finding — something must MATCH in the same run.
- **Grade a requirement by its mechanism, not by a phrase you guessed** — two false negatives on
  correct fixes in one session.
- **Cache vs origin produce the identical symptom:** `?cb=` defeats a stale cache, not a stale origin;
  `last-modified` vs your rerender's `completed_at` tells you which. A PASS can be faked by neither.
- `handler_agent` is NOT NULL on `site_work_items`; for rerenders it is `page-rerender`.
- This lane's own rebuilds are its main queue competitor — each tool spawns a guide page and rerenders.
- **There is no lever to make ONE page assemble sooner, and `priority` is not it.** Measured 2026-08-19:
  a `rerender-pages` sweep filed 117 `page_rerender` items for this site at one timestamp, all priority
  80, draining in alphabetical page order at ~0.7/min — and `tool-*` sorts after every `learn-*` page.
  Lowering two items' `priority` to 5 and 6 changed **nothing**: priority-80 items kept being claimed
  after the write. `load_work_item_actions.go:737`'s `ORDER BY wi.priority ASC` belongs to a loader that
  does not dispatch this item type, and the real selector is still unlocated. Both rows were restored.
  **Verify any dispatch change at the DISPATCHER (what got claimed next), never at the row** — the row
  said `priority=5` and `UPDATE 1`, which is exactly what success looks like. (`WRONG_CALLS.md`, 08-19.)
  In practice the queue drains: all 117 were done within ~70 minutes.
- **The chassis retains ~2 MINUTES of logs.** `logs --since=6h` returns lines starting two minutes ago
  and reads as if the service was quiet before that, so grepping for a line your code emits is only
  valid within about a minute of the event and fails SILENTLY. Before believing an absence, put a line
  you know always fires into the same grep. Better: if the branch WRITES something (a column, a row),
  assert that instead — it has no shelf life. (`LANDMINES.md`, entry refined 08-19.)
- **`id="verdict"`-style ids that exist in BOTH versions are not negative controls.** Enumerate the
  retired slot's own ids before the rebuild and pick ones that appear nowhere else; a guessed
  "old-looking" id can test nothing while looking like a control.
