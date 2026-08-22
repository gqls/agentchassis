# NOTES — bugfix_342_absent_required (append-only, newest at the bottom)

## 2026-08-22 — state verification before touching anything

- who-owns: "OWNED or recently active" — but the activity is the `bugfix_260_render_fallback`
  lane's own filing + fix work, finished 2026-08-21 evening (their NOTES: APPROVED round 6,
  live-verified on v1.0.1322). The bug file says UNOWNED; RFC_041 §5 says "what it needs now
  is a name". This lane is the name. Tree checked: nothing dirty touches the 342 surfaces.
- **The bug file's warning banner is STALE.** It says the editor-route escalation "is NOT in
  v1.0.1322 and is inert until the next roll". The next roll happened: fleet is on v1.0.1323
  (both replicas, pod images read). Ancestry: `cd90e8b27` (wire editor routes), `65f1b0b95`
  (single emitter), `af4743464` (the seam) are ALL ancestors of the v1.0.1323 stamp
  `70e7b4f9c` (found via another lane's commit `49d90b280`, then verified at the artefact:
  `grep -aq "70e7b4f9c" /proc/1/exe` present on BOTH replicas, nonsense control absent).
- `record_absent_required_fields`: **0 agent_definitions rows name it.** Chrome escalation is
  dormant. **7** live steps as of 2026-08-22 use `render_site_components` (nav-link-fixer, nav-updater,
  pageflow-builder, rerender-chrome, rerender-pages, rerender-site, site-work-orchestrator).
- `required_fields_missing` items: 45 complete / 30 needs_human_review, ALL from the
  post-deploy check (latest 08-18, all naming deployed `ported-page` rows). Zero render-time
  items yet — expected: editor escalation live only since the 1323 roll this morning, chrome
  half unarmed, and the write is per-edit.

### Misstep (mine, caught in-flight): a VACUOUS census read as a finding

I ran the chrome census (site_components rows missing required source:llm fields), got zero
rows, and stated "no chrome row is missing a required field" — then the vacuity check showed
**candidate_pairs_tested = 0**: no site_components row references ANY component declaring a
required llm field, so the query could never have returned a hit. The zero was true but
meant something different: "the chrome store uses only 0-required components", not "all
required fields are supplied". The census was re-read on the corrected join and the
conclusion for arming purposes survives (0 rows fire either way), but the first phrasing
would have been a WRONG_CALLS row if it had reached a doc. Cheap check, now in RUNBOOK:
**count the candidate pairs your census tested in the same query; a zero over zero
candidates is not a finding.** Logged in WRONG_CALLS.md per the tally rule.

### Demand-control failure worth recording

Pod-log grep for render activity (`RenderTemplate: fields rendered empty|Re-rendered
component|render_site_components|RenderComponentAction`) returned 0 on both replicas while
the DB showed 38 page_components writes since pod start — the greps don't match what these
paths actually log at INFO. So "0 absent-required reports in the logs" could not be given a
log-side demand control; the DB-side census is the evidence that matters and is what the
plan uses.

### Populations re-measured (the bug file's figures had moved)

- Active components: **283 as of 2026-08-22** (was 253 at filing, 2026-08-20). No-schema: **100 as of 2026-08-22** — **95 are tools** (schema-less
  by design; `isSelfContainedSection` codifies it). Non-tool no-schema: **5 as of 2026-08-22**, one
  page_components usage each; 2 with `{{.field}}` placeholders (`report-request-form`,
  `audience-check-form`). §5's "expect the 75-of-253 to be the hard part" has dissolved into
  small data work.
- Chrome components WITH required llm fields exist in the library (footer-with-disclaimer 17,
  header-with-categories 16, header-with-search 5, header-with-cart-or-nav 4,
  header-minimal-tool 2) but **no site_components row references any of them** — the store
  uses site-header/site-footer/head (0 required). Hence arming the record = 0 items today.
- page_components writer-side census: **131 rows as of 2026-08-22** missing a required llm field top-level;
  breakdown: ported-page/body 77 deployed + 23 removed (the post-deploy check's existing,
  already-itemised population), hero/headline 14 deployed, scattered singles after that.
  NOTE this is the WRITER question (pre-gate); the seam's render-time answer is a strict
  subset after `contextToInterfaceMap` defaults.

### Code reads that shaped the plan

- Editor routes: emit item, then **persist the blank anyway** — refusal is the missing half
  (`section_editor_actions.go` applyContentEdit ~:1042, applyComponentSwap ~:1169).
- `ApplySectionEditAction` already has the ONE-persist-switch idiom (link repair :~445,
  envelope refusal :~455) — the refusal gate belongs there, not in the two branches.
- `section-editor.apply_edit` has **no error_step** (live row) → a refusal meets
  `bugs_open/344`'s completion-trample on the DRIVING item. Live page still protected; the
  item status is 344's bug, not this lane's. Interaction stated in the code comment.
- Six unwired sites re-verified: GateConvertedTemplate + tool_birth_instance_scope (raw
  candidate templates, no component row), legacy head render (template load only),
  RenderTemplateWithMap (contact-info block, callers have no component schema),
  offline audit ×2 (probes with fields REMOVED by design). No change owed.
- `assemble_from_library` renders a stitched TEMPLATE (content arrives later) — report-only
  is correct there; no escalation owed.
- RFC_022 budget: `apply_section_edit` counts 7 optional keys; ConfigKeys are not counted by
  the audit; the new key changes nothing. `render_site_components` has NO registered input
  spec at all (pre-existing; not widened by this lane).

## 2026-08-22 late morning — built, tested, submitted, committed

- Refusal half implemented: key + deciding arm in `mistyped_llm_fields_gate.go`
  (`refuse_absent_required_fields`, `refusePersistForAbsentRequired`), outcome field +
  branch copies + ONE gate at the persist switch in `section_editor_actions.go`, ConfigKeys
  declared. Tests: two table tests + a seam→outcome→gate chain test; **mutation-proven**
  (inverting the deciding arm fails TestEditorRefusalNeedsBothArmingAndAFinding AND
  TestSeamFindingSurvivesOntoTheEditOutcome; reverted, green). Honest limit written into the
  chain test's comment: the in-branch copy needs a DB — the post-roll canary covers it.
- Migrations: 550 (chrome record arm, appliable) + 551_HOLD (editor refusal arm, after the
  roll), both with rollback sidecars, DO/RAISE verifies, double-apply refusals; 550 also
  refuses if a render_site_components step has moved into a sub_workflow (the write only
  reaches top level).
- Council: corr `3626629a-f2bc-4089-9118-c1d6dd007807`, submitted 09:32Z, dispatched almost
  immediately (no queue wait this time). Two client-side schema rejects first: `create` is
  not a valid operation (use `add`), and `risks` must be a STRING not an array.
- Committed `0ee442cfb` with `Council-Submitted:` trailer, 13 files, scope report all mine.
  Clean `git archive HEAD` build verified after commit (platform/internal/cmd all compile).
- **My WRONG_CALLS entry (vacuous census) was swept into the 337 lane's commit `9e23fb852`
  as a same-file passenger** between my append and my commit — the exact CLAUDE.md case;
  stated in my commit message, nothing lost.
- MEMORY_workstreams: lane registered with the owed follow-ups (verdict read, 550 apply,
  post-roll 551 + canary, the 5 no-schema components decision).
- 550 deliberately NOT applied pre-verdict: migrations are council scope (314 — live the
  moment they apply), and the arm fires on 0 rows, so waiting costs nothing.

### The 9-of-15 arithmetic re-derived first-hand (not inherited from the bug file)

`grep -rn "RenderTemplate(" --include=*.go platform/ internal/ cmd/ | grep -v _test | grep -v
"func RenderTemplate"` → **14**, plus `RenderTemplateWithMap` = **15**. Schema-wiring sites,
counted with a grep that NAMES THE RECEIVER (the bug file's own landmine: a bare
`grep -c 'InputSchema = '` also matches `ci.InputSchema`, a different struct) → **9**:
assemble_from_library:302, v3_site_actions:2464, section_editor:1069 + :1206,
render_site_components:1051, component_library:1730/1805/2092, rerender_page_sections:655.
**9 + 6 = 15 closes.**

The six unwired, each read at its own call site rather than taken on trust:
- `GateConvertedTemplate(function, converted string, …)` — signature takes a raw string; no
  component row is in scope, so there is no schema to pass.
- `ScopeToolBirthTemplate(html, function string, …)` — same shape, raw candidate template.
- the legacy head render (`rerenderSinglePage` :538) — its loader
  `rerenderLoadHeadTemplate` selects `defaults->>'head'` then `html_template` only, never
  `input_schema`; and **`RerenderSitePagesAction` is in no `GlobalActionRegistry` entry**
  (grep: 0 hits), matching RFC_041 §4 — a dead path, so wiring it would be inert anyway.
- `RenderTemplateWithMap` — a different executor (contact-info block), callers hold no schema.
- the two `cmd/component-render-check` probes — they render with fields REMOVED on purpose, so
  a report there fires on every probe by design.

## 2026-08-22 midday — council round 1: REVISE, and the objection that changed the design

**Gating (`prior_art_librarian`, HIGH):** 551 UPDATEs by `type='section-editor'` while a landmine
documents four agent types carrying TWO active rows, of which only the higher version loads — so
an UPDATE-by-type can silently arm the row nobody reads. **The seat was right that I asserted it
rather than checking it.** Checked: section-editor has exactly one live row (`2ed3b581`, v1); the
four are `chief-strategist`, `content-creator`, `content-creator-contact`,
`site-component-architect`. My precondition was already the correct guard — it now NAMES the four
and says why aborting (not writing both) is the right response.

**The one that changed the code (`bug_historian`, medium):** arming DETECTION on chrome while only
the section editor got PROTECTION reproduces **this bug's own shape** on the sibling call site —
016b §9's *"one call site of a shared judgement gets the rigorous fix, the sibling stays
heuristic"* — and my "0 rows fire today" is a population snapshot, not a structural guarantee.
Acted on rather than argued: the chrome store now has the SAME refusal, through the SAME decision
function, **default OFF with no migration arming it**. That is the honest middle: arming today
would arm an unexercisable refusal, while leaving the capability out would put a code change, a
review and a roll between the first adopting site and its protection. The flip trigger is named —
the first `required_fields_missing` item with `surface='site_component'`. A new test asserts the
two paths agree on all four arming×finding combinations, so a forked chrome implementation fails.

**`guidelines`, medium:** 550's post-verify walked only the top level — i.e. it certified exactly
the coverage the write can reach and called it full. It now walks `config.sub_workflow.steps` too
and RAISEs on any unarmed nested occurrence. **This is the sharpest kind of objection: the check
that would have passed was the one I wrote to prove the write was complete.**

**`guardian`, medium:** name the owning pipelines, don't say "chrome". Enumerated in the header.
Its two low objections taken: both migrations relabelled `config_change`.

Everything else was answered with a query and attached rather than re-asserted: one consumer of
`apply_section_edit` (top-level 1, nested 0), both keys armed 0 **re-measured this round**,
`render_site_components` **7** top-level / **0** nested as of 2026-08-22, optional-key budget **7** of 10 as of 2026-08-22 (ConfigKeys are not
counted).

**What I got right and want to keep doing:** I re-read the submission against the actual diff
before re-firing (`git diff` + a grep per claimed symbol), which is the 260 lane's own hard-won
rule — three of their six rounds were the submission describing code that had moved.

## 2026-08-22 afternoon — APPROVED at round 2, and the advisories acted on rather than banked

Verdict: **approved with 3 advisory objections, none high** (trail `3626629a`, round 2). An
approval is not a reason to stop reading the objections — one of them named a real property of
my own change that I had not stated.

**`guardian`, medium — where does the new gate sit relative to the `!force` idempotence exit?**
The honest answer is *after it, so it never runs when the exit fires*, and I had not said so.
Checked at the code (`:850`): the exit returns when the slot already holds non-empty,
non-`pending` HTML. **For the REFUSAL that is correct and not a gap** — the exit fires precisely
when nothing is about to be written, and a gate that prevents writes has nothing to prevent.
**For DETECTION it is a real blind spot and a pre-existing one**: an already-populated slot whose
`content_data` lacks a required field is never inspected by a non-forced refresh, because no
render happens to inspect it. Now written into the code at the gate, so the next reader does not
have to rediscover it: *what is checked is every slot this function is about to write*, not every
chrome slot continuously.

**`bug_historian`, medium — is `chromeSlotHasStoredHTML` a fresh ad-hoc "does this hold real
content?" discriminator, the class this estate has had to re-harden twice (016b §9 items 3, 5)?**
No, on both counts, and the objection was worth answering rather than waving off. It is
**pre-existing** (`:1706`), and it is already the discriminator the execution-failure branch
~90 lines above uses for exactly this fatal-vs-degraded split (and the caller at `:333`) — so
this is reuse, not a second lineage. And it is not that class of judgement: it asks whether any
non-empty `rendered_html` is STORED, never whether the HTML looks meaningful, so there is no
sparse-but-real content for it to misread.

**`prior_art_librarian`, medium — the chrome zero was quoted, not re-run, while being the
load-bearing argument for shipping the refusal unarmed.** Fair. Re-run at approval, one statement
with its own vacuity guard: `candidate_pairs 0 | rows_missing 0 |
components_with_required_fields 813 | chrome_rows_total 72`. **Read in that order it says
something the bare zero does not**: neither side is empty — 813 component×field pairs in the
library do declare a required field, and 72 chrome rows exist — what is empty is the JOIN. The
figure is now in 550's header in that form.

**`editquality` + `guardian`, low — enumerate every caller of the signature you changed.** Done:
one production caller (`:312`) and three test call sites (`site_component_lock_guard_test.go`
:292/:327/:353, all updated). "It compiles" was already evidence, but the enumeration is the
thing that could have come out otherwise.

**`debug_historian`, low — bake the row-count assertion into 551's UPDATE.** Done, with
`GET DIAGNOSTICS`. The point is sharper than it first reads: a precondition that ran *beforehand*
cannot see a second active row inserted between it and the write, and on this tree that is not
hypothetical.

**`bug_historian`, low — this is still per-call-site patching of a generic root cause
(`missingkey=zero`), and the historian's index says narrow per-site patches on this class have
not historically been the fix that stuck.** True, and deliberately so: refusing at the SEAM is
new authority over content that renders successfully today at sites that never asked (owner
2026-08-02 §2). Recorded here as the standing argument rather than settled — if a future owner
ruling licenses a seam-level default, this is the note that says why we did not take it.

**Applied `550` after the verdict**: 7/7 steps armed, verified by reading the live rows rather
than trusting the migration's own post-check, with a negative control (the REFUSE key must still
be armed nowhere — it is). Ledgered via `--record-only`.

## 2026-08-22 evening — v1.0.1326 rolled, 551 APPLIED, and the canary PASSED on both arms

### The build check, and one probe arm that was invalid

Fleet on **v1.0.1326** (both replicas, started 15:10Z). Probed at the binary, both replicas:

| literal | novel? | want | got |
|---|---|---|---|
| `refuse_absent_required_fields` (editor key) | **yes**, 0 at `0ee442cfb^` | present | **present** |
| `REQUIRED content field(s) absent — refusing to store` (chrome) | **yes**, 0 at `a39a16555^` | present | **present** |
| ~~`refusing to persist`~~ | **NO — 3 pre-existing files** | — | **INVALID ARM**, see WRONG_CALLS |
| nonsense literal | — | absent | absent |
| `Go template execution failed, using regex fallback` (deleted) | — | absent | absent |

The startup provenance line had scrolled (`--tail=3000`, absent), and probing the binary for my
own commit shas returns OUT for all three — **a binary carries the ONE sha it was built from, not
its ancestors**, so that is uninformative by construction, not evidence of absence. The two novel
literals are the evidence, and they are capability-level, which is the stronger form.

### 551 applied, and verified independently of its own post-check

Preconditions re-checked live first (8 hours had passed): 1 live section-editor row, `apply_edit`
still `apply_section_edit`, key absent, and this morning's chrome arm still 7/7. Applied clean.
Read back from the live row: key `true`, **both sibling keys survived** (`strip_literal_markdown`,
`allow_rendered_html_transform` — a wrong `jsonb_set` path replaces the whole config branch, which
is what that check exists to catch), version 1→2. Two negative controls: no other agent gained the
key (0), and the **chrome refusal is still armed nowhere** (0) — its deliberate state.
`_HOLD` suffix dropped in the same commit; ledgered via `--record-only`.

⚠ Another session held `.git/index.lock` during the rename. **Waited, did not remove it** — that
file is another session's commit in progress.

### The canary — BOTH ARMS, at the artefact

Targets deliberately NOT `deployed`, so a failing canary could harm no live page.

**Refusal arm** (`0a1498b3`, tool-cta, 3 required fields absent), corr `63f2eab4`:
- step **FAILED** with the exact text: *"refusing to persist — 2 schema-required field(s)
  rendered empty (headline, trust_note); the live section is left unchanged and a
  required_fields_missing item has been filed"*;
- **artefact byte-identical**: md5 still `69a2f28c…`, and `updated_at` still **2026-07-17** —
  the row was not touched at all;
- **item filed**: `required_fields_missing`, status `detected`, naming both fields,
  `surface=page_component`, `route=content_edit`. Refusing did not cost the record.

**Positive control** (`9737d0d9`, use-cases-list, required field present), corr `3eb5318a`:
all orchestrations **COMPLETED** through `deploy_page`. The artefact's bytes are identical
(expected — I set the field to its own value), but **`updated_at` moved to 18:05:04**, which is
the discriminator that separates *"persisted"* from *"skipped"*. **An arm that only stops edits
would have failed here.**

### A finding the canary produced that no test could: the subset behaviour, observed

The census predicted three absent fields (`headline`, `description`, `trust_note`); the refusal
named **two**. `description` was filled by a default in the merged map the template actually sees.
That is the documented seam-vs-gate divergence — *"did the WRITER supply it?"* vs *"will it RENDER
EMPTY?"* — **confirmed in production for the first time**, having been pinned only by
`TestSeamReportsASubsetOfThePreRenderGateAndSaysWhy` until now. The seam's report really is a
strict subset. Do not "fix" the count.

### HONEST LIMIT: check (c) of the canary did NOT fire

551's header says the canary reads the DRIVING work item's terminal status, expecting `complete`
until `bugs_open/344` lands. **It could not: a CLI-dispatched canary has no driving work item.**
Measured — 0 trampled rows in the window, and the only `required_fields_missing` item is the one
the refusal filed. So the 344 interaction remains **[UNEXERCISED]**, not "verified benign", and
closing that needs a work-item-driven edit (the `section-editor` route reached from a queue item),
not another CLI run.

### Two measurement missteps of mine, both in WRONG_CALLS

1. The invalid probe arm above — a positive control on a literal that already existed.
2. My canary monitor waited for `count(*) = 2` across BOTH correlations, and the refusal arm
   alone writes two rows (child + parent), so it fired **BOTH ARMS TERMINAL** while the control
   was still running. A count aggregated over two populations can be satisfied by one; the fix is
   `count(DISTINCT correlation_id)` or asserting each separately.
